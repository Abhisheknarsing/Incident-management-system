package handlers

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"time"

	"incident-management-system/internal/models"
	"incident-management-system/internal/storage"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// UploadHandler handles file upload operations
type UploadHandler struct {
	db               *sql.DB
	fileStore        *storage.FileStore
	processingService interface {
		ProcessUpload(ctx context.Context, uploadID string) (interface{}, error)
		GetProcessingStatus(ctx context.Context, uploadID string) (interface{}, error)
	}
}

// NewUploadHandler creates a new UploadHandler instance
func NewUploadHandler(db *sql.DB, fileStore *storage.FileStore, processingService interface{}) *UploadHandler {
	return &UploadHandler{
		db:               db,
		fileStore:        fileStore,
		processingService: processingService.(interface {
			ProcessUpload(ctx context.Context, uploadID string) (interface{}, error)
			GetProcessingStatus(ctx context.Context, uploadID string) (interface{}, error)
		}),
	}
}

// UploadFile handles Excel file uploads
func (h *UploadHandler) UploadFile(c *gin.Context) {
	// Parse multipart form
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "No file provided",
			"code":  "MISSING_FILE",
		})
		return
	}

	// Validate file size (max 50MB)
	const maxFileSize = 50 << 20 // 50MB
	if file.Size > maxFileSize {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "File size exceeds maximum limit of 50MB",
			"code":  "FILE_TOO_LARGE",
		})
		return
	}

	// Save file to storage
	filename, _, err := h.fileStore.SaveUploadedFile(file)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
			"code":  "INVALID_FILE_FORMAT",
		})
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

	// Save upload record to database
	if err := h.createUploadRecord(upload); err != nil {
		// Clean up file on database error
		h.fileStore.DeleteFile(filename)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to create upload record",
			"code":  "DATABASE_ERROR",
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "File uploaded successfully",
		"upload":  upload,
	})
}

// GetUploads returns a list of all uploads
func (h *UploadHandler) GetUploads(c *gin.Context) {
	uploads, err := h.getUploadRecords()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to retrieve uploads",
			"code":  "DATABASE_ERROR",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"uploads": uploads,
	})
}

// GetUpload returns a specific upload by ID
func (h *UploadHandler) GetUpload(c *gin.Context) {
	uploadID := c.Param("id")
	if uploadID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Upload ID is required",
			"code":  "MISSING_UPLOAD_ID",
		})
		return
	}

	upload, err := h.getUploadRecord(uploadID)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Upload not found",
				"code":  "UPLOAD_NOT_FOUND",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to retrieve upload",
			"code":  "DATABASE_ERROR",
		})
		return
	}

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
	uploadID := c.Param("id")
	if uploadID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Upload ID is required",
			"code":  "MISSING_UPLOAD_ID",
		})
		return
	}

	// Check if upload exists and is in correct status
	upload, err := h.getUploadRecord(uploadID)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Upload not found",
				"code":  "UPLOAD_NOT_FOUND",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to retrieve upload",
			"code":  "DATABASE_ERROR",
		})
		return
	}

	// Check if upload is in correct status for processing
	if upload.Status != models.UploadStatusUploaded {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": fmt.Sprintf("Upload cannot be processed in current status: %s", upload.Status),
			"code":  "INVALID_STATUS",
		})
		return
	}

	// Start processing in background
	go func() {
		ctx := context.Background()
		_, err := h.processingService.ProcessUpload(ctx, uploadID)
		if err != nil {
			log.Printf("Processing failed for upload %s: %v", uploadID, err)
		}
	}()

	c.JSON(http.StatusAccepted, gin.H{
		"message": "Processing started",
		"upload_id": uploadID,
	})
}

// GetProcessingStatus returns the processing status of an upload
func (h *UploadHandler) GetProcessingStatus(c *gin.Context) {
	uploadID := c.Param("id")
	if uploadID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Upload ID is required",
			"code":  "MISSING_UPLOAD_ID",
		})
		return
	}

	ctx := context.Background()
	status, err := h.processingService.GetProcessingStatus(ctx, uploadID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get processing status",
			"code":  "DATABASE_ERROR",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": status,
	})
}