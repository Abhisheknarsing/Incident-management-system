package handlers

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"time"

	"incident-management-system/internal/errors"
	"incident-management-system/internal/logging"
	"incident-management-system/internal/models"
	"incident-management-system/internal/monitoring"
	"incident-management-system/internal/services"
	"incident-management-system/internal/storage"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// UploadHandler handles file upload operations
type UploadHandler struct {
	db                *sql.DB
	fileStore         *storage.FileStore
	logger            *logging.Logger
	processingService interface {
		ProcessUpload(ctx context.Context, uploadID string) (*services.ProcessingProgress, error)
		GetProcessingStatus(ctx context.Context, uploadID string) (*services.ProcessingProgress, error)
	}
}

// NewUploadHandler creates a new UploadHandler instance
func NewUploadHandler(db *sql.DB, fileStore *storage.FileStore, processingService interface{}) *UploadHandler {
	return &UploadHandler{
		db:        db,
		fileStore: fileStore,
		logger:    logging.GetGlobalLogger().WithComponent("upload_handler"),
		processingService: processingService.(interface {
			ProcessUpload(ctx context.Context, uploadID string) (*services.ProcessingProgress, error)
			GetProcessingStatus(ctx context.Context, uploadID string) (*services.ProcessingProgress, error)
		}),
	}
}

// UploadFile handles Excel file uploads
func (h *UploadHandler) UploadFile(c *gin.Context) {
	start := time.Now()
	logger := h.logger.WithContext(c.Request.Context()).WithOperation("upload_file")

	logger.Info("Starting file upload")

	// Parse multipart form
	file, err := c.FormFile("file")
	if err != nil {
		apiErr := errors.NewAPIError(errors.ErrMissingFile, "No file provided").
			WithUserMessage("Please select a file to upload")
		monitoring.TrackError(c.Request.Context(), apiErr, "upload_handler", "upload_file")
		errors.SendError(c, apiErr)
		return
	}

	logger.Info("File received",
		logging.GetGlobalLogger().WithMetadata(map[string]interface{}{
			"filename": file.Filename,
			"size":     file.Size,
		}))

	// Validate file size (max 50MB)
	const maxFileSize = 50 << 20 // 50MB
	if file.Size > maxFileSize {
		apiErr := errors.FileUploadError("file_too_large")
		monitoring.TrackError(c.Request.Context(), apiErr, "upload_handler", "upload_file")
		errors.SendError(c, apiErr)
		return
	}

	// Save file to storage
	filename, _, err := h.fileStore.SaveUploadedFile(file)
	if err != nil {
		apiErr := errors.FileUploadError("invalid_format").WithDetails(err.Error())
		monitoring.TrackError(c.Request.Context(), apiErr, "upload_handler", "upload_file")
		errors.SendError(c, apiErr)
		return
	}

	// Create upload record
	upload := &models.Upload{
		ID:               uuid.New().String(),
		Filename:         filename,
		OriginalFilename: file.Filename,
		Status:           models.UploadStatusUploaded,
		RecordCount:      0,
		ProcessedCount:   0,
		ErrorCount:       0,
		Errors:           []string{},
		CreatedAt:        time.Now(),
	}

	logger.Info("Creating upload record",
		logging.GetGlobalLogger().WithMetadata(map[string]interface{}{
			"upload_id": upload.ID,
			"filename":  filename,
		}))

	// Save upload record to database
	if err := h.createUploadRecord(upload); err != nil {
		// Clean up file on database error
		h.fileStore.DeleteFile(filename)
		apiErr := errors.DatabaseError("create upload record", err)
		monitoring.TrackError(c.Request.Context(), apiErr, "upload_handler", "upload_file")
		errors.SendError(c, apiErr)
		return
	}

	logger.LogDuration("upload_file", start,
		logging.GetGlobalLogger().WithMetadata(map[string]interface{}{
			"upload_id": upload.ID,
			"success":   true,
		}))

	monitoring.UpdatePerformance(time.Since(start))

	c.JSON(http.StatusCreated, gin.H{
		"message": "File uploaded successfully",
		"upload":  upload,
	})
}

// GetUploads returns a list of all uploads
func (h *UploadHandler) GetUploads(c *gin.Context) {
	start := time.Now()
	logger := h.logger.WithContext(c.Request.Context()).WithOperation("get_uploads")

	logger.Info("Retrieving uploads list")

	uploads, err := h.getUploadRecords()
	if err != nil {
		apiErr := errors.DatabaseError("retrieve uploads", err)
		monitoring.TrackError(c.Request.Context(), apiErr, "upload_handler", "get_uploads")
		errors.SendError(c, apiErr)
		return
	}

	logger.LogDuration("get_uploads", start,
		logging.GetGlobalLogger().WithMetadata(map[string]interface{}{
			"count": len(uploads),
		}))

	monitoring.UpdatePerformance(time.Since(start))

	c.JSON(http.StatusOK, gin.H{
		"uploads": uploads,
	})
}

// GetUpload returns a specific upload by ID
func (h *UploadHandler) GetUpload(c *gin.Context) {
	start := time.Now()
	logger := h.logger.WithContext(c.Request.Context()).WithOperation("get_upload")

	uploadID := c.Param("id")
	if uploadID == "" {
		apiErr := errors.NewAPIError(errors.ErrMissingUploadID, "Upload ID is required")
		errors.SendError(c, apiErr)
		return
	}

	logger.Info("Retrieving upload",
		logging.GetGlobalLogger().WithMetadata(map[string]interface{}{
			"upload_id": uploadID,
		}))

	upload, err := h.getUploadRecord(uploadID)
	if err != nil {
		if err == sql.ErrNoRows {
			apiErr := errors.NotFound("Upload")
			errors.SendError(c, apiErr)
			return
		}
		apiErr := errors.DatabaseError("retrieve upload", err)
		monitoring.TrackError(c.Request.Context(), apiErr, "upload_handler", "get_upload")
		errors.SendError(c, apiErr)
		return
	}

	logger.LogDuration("get_upload", start,
		logging.GetGlobalLogger().WithMetadata(map[string]interface{}{
			"upload_id": uploadID,
			"found":     true,
		}))

	monitoring.UpdatePerformance(time.Since(start))

	c.JSON(http.StatusOK, gin.H{
		"upload": upload,
	})
}

// createUploadRecord inserts a new upload record into the database
func (h *UploadHandler) createUploadRecord(upload *models.Upload) error {
	query := `
		INSERT INTO uploads (
			id, filename, original_filename, status, record_count, 
			processed_count, error_count, errors, created_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	// Convert errors slice to JSON string for storage
	errorsJSON := "[]"
	if len(upload.Errors) > 0 {
		// For now, store as simple JSON array - in production, use proper JSON marshaling
		errorsJSON = fmt.Sprintf(`["%s"]`, upload.Errors[0])
	}

	_, err := h.db.Exec(query,
		upload.ID,
		upload.Filename,
		upload.OriginalFilename,
		upload.Status,
		upload.RecordCount,
		upload.ProcessedCount,
		upload.ErrorCount,
		errorsJSON,
		upload.CreatedAt,
	)

	return err
}

// getUploadRecords retrieves all upload records from the database
func (h *UploadHandler) getUploadRecords() ([]models.Upload, error) {
	query := `
		SELECT id, filename, original_filename, status, record_count, 
			   processed_count, error_count, errors, created_at, processed_at
		FROM uploads 
		ORDER BY created_at DESC
	`

	rows, err := h.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var uploads []models.Upload
	for rows.Next() {
		var upload models.Upload
		var errorsJSON string

		err := rows.Scan(
			&upload.ID,
			&upload.Filename,
			&upload.OriginalFilename,
			&upload.Status,
			&upload.RecordCount,
			&upload.ProcessedCount,
			&upload.ErrorCount,
			&errorsJSON,
			&upload.CreatedAt,
			&upload.ProcessedAt,
		)
		if err != nil {
			return nil, err
		}

		// For now, initialize empty errors slice - in production, parse JSON
		upload.Errors = []string{}
		uploads = append(uploads, upload)
	}

	return uploads, rows.Err()
}

// getUploadRecord retrieves a specific upload record by ID
func (h *UploadHandler) getUploadRecord(uploadID string) (*models.Upload, error) {
	query := `
		SELECT id, filename, original_filename, status, record_count, 
			   processed_count, error_count, errors, created_at, processed_at
		FROM uploads 
		WHERE id = ?
	`

	var upload models.Upload
	var errorsJSON string

	err := h.db.QueryRow(query, uploadID).Scan(
		&upload.ID,
		&upload.Filename,
		&upload.OriginalFilename,
		&upload.Status,
		&upload.RecordCount,
		&upload.ProcessedCount,
		&upload.ErrorCount,
		&errorsJSON,
		&upload.CreatedAt,
		&upload.ProcessedAt,
	)

	if err != nil {
		return nil, err
	}

	// For now, initialize empty errors slice - in production, parse JSON
	upload.Errors = []string{}

	return &upload, nil
}

// ProcessUpload triggers processing of an uploaded file
func (h *UploadHandler) ProcessUpload(c *gin.Context) {
	start := time.Now()
	logger := h.logger.WithContext(c.Request.Context()).WithOperation("process_upload")

	uploadID := c.Param("id")
	if uploadID == "" {
		apiErr := errors.NewAPIError(errors.ErrMissingUploadID, "Upload ID is required")
		errors.SendError(c, apiErr)
		return
	}

	logger.Info("Starting upload processing",
		logging.GetGlobalLogger().WithMetadata(map[string]interface{}{
			"upload_id": uploadID,
		}))

	// Check if upload exists and is in correct status
	upload, err := h.getUploadRecord(uploadID)
	if err != nil {
		if err == sql.ErrNoRows {
			apiErr := errors.NotFound("Upload")
			errors.SendError(c, apiErr)
			return
		}
		apiErr := errors.DatabaseError("retrieve upload", err)
		monitoring.TrackError(c.Request.Context(), apiErr, "upload_handler", "process_upload")
		errors.SendError(c, apiErr)
		return
	}

	// Check if upload is in correct status for processing
	if upload.Status != models.UploadStatusUploaded {
		apiErr := errors.NewAPIError(errors.ErrInvalidStatus,
			fmt.Sprintf("Upload cannot be processed in current status: %s", upload.Status)).
			WithUserMessage("This upload has already been processed or is currently being processed").
			WithSuggestions([]string{
				"Check the upload status",
				"Wait for current processing to complete",
				"Upload a new file if needed",
			})
		errors.SendError(c, apiErr)
		return
	}

	// Start processing in background
	go func() {
		ctx := context.Background()
		_, err := h.processingService.ProcessUpload(ctx, uploadID)
		if err != nil {
			logger.Error("Processing failed for upload", err,
				logging.GetGlobalLogger().WithMetadata(map[string]interface{}{
					"upload_id": uploadID,
				}))

			// Track processing error
			apiErr := errors.ProcessingFailed(err.Error())
			monitoring.TrackError(ctx, apiErr, "processing_service", "process_upload")
		} else {
			logger.Info("Processing completed successfully",
				logging.GetGlobalLogger().WithMetadata(map[string]interface{}{
					"upload_id": uploadID,
				}))
		}
	}()

	logger.LogDuration("process_upload", start,
		logging.GetGlobalLogger().WithMetadata(map[string]interface{}{
			"upload_id": uploadID,
			"started":   true,
		}))

	monitoring.UpdatePerformance(time.Since(start))

	c.JSON(http.StatusAccepted, gin.H{
		"message":   "Processing started",
		"upload_id": uploadID,
	})
}

// GetProcessingStatus returns the processing status of an upload
func (h *UploadHandler) GetProcessingStatus(c *gin.Context) {
	start := time.Now()
	logger := h.logger.WithContext(c.Request.Context()).WithOperation("get_processing_status")

	uploadID := c.Param("id")
	if uploadID == "" {
		apiErr := errors.NewAPIError(errors.ErrMissingUploadID, "Upload ID is required")
		errors.SendError(c, apiErr)
		return
	}

	logger.Info("Getting processing status",
		logging.GetGlobalLogger().WithMetadata(map[string]interface{}{
			"upload_id": uploadID,
		}))

	ctx := context.Background()
	status, err := h.processingService.GetProcessingStatus(ctx, uploadID)
	if err != nil {
		apiErr := errors.DatabaseError("get processing status", err)
		monitoring.TrackError(c.Request.Context(), apiErr, "upload_handler", "get_processing_status")
		errors.SendError(c, apiErr)
		return
	}

	logger.LogDuration("get_processing_status", start,
		logging.GetGlobalLogger().WithMetadata(map[string]interface{}{
			"upload_id": uploadID,
		}))

	monitoring.UpdatePerformance(time.Since(start))

	c.JSON(http.StatusOK, gin.H{
		"status": status,
	})
}
