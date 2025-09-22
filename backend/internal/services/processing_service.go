package services

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	"incident-management-system/internal/models"
	"incident-management-system/internal/storage"
)

// ProcessingService coordinates Excel file processing
type ProcessingService struct {
	db                 *sql.DB
	fileStore          *storage.FileStore
	excelParser        *ExcelParser
	incidentService    *IncidentService
	sentimentAnalyzer  SentimentAnalyzer
	automationAnalyzer AutomationAnalyzer
}

// NewProcessingService creates a new ProcessingService instance
func NewProcessingService(db *sql.DB, fileStore *storage.FileStore) *ProcessingService {
	return &ProcessingService{
		db:                 db,
		fileStore:          fileStore,
		excelParser:        NewExcelParser(),
		incidentService:    NewIncidentService(db),
		sentimentAnalyzer:  NewSimpleSentimentAnalyzer(),
		automationAnalyzer: NewSimpleAutomationAnalyzer(),
	}
}

// ProcessingProgress represents the progress of file processing
type ProcessingProgress struct {
	UploadID      string    `json:"upload_id"`
	Status        string    `json:"status"`
	TotalRows     int       `json:"total_rows"`
	ProcessedRows int       `json:"processed_rows"`
	ValidRows     int       `json:"valid_rows"`
	ErrorCount    int       `json:"error_count"`
	Errors        []string  `json:"errors"`
	StartTime     time.Time `json:"start_time"`
	EndTime       *time.Time `json:"end_time,omitempty"`
	Duration      string    `json:"duration,omitempty"`
}

// ProcessUpload processes an uploaded Excel file
func (s *ProcessingService) ProcessUpload(ctx context.Context, uploadID string) (*ProcessingProgress, error) {
	progress := &ProcessingProgress{
		UploadID:  uploadID,
		Status:    models.UploadStatusProcessing,
		StartTime: time.Now(),
		Errors:    make([]string, 0),
	}

	// Update upload status to processing
	if err := s.incidentService.UpdateUploadStatus(ctx, uploadID, models.UploadStatusProcessing, 0, 0, 0, nil); err != nil {
		return nil, fmt.Errorf("failed to update upload status to processing: %w", err)
	}

	// Get upload record to find the file
	upload, err := s.getUploadRecord(ctx, uploadID)
	if err != nil {
		s.markProcessingFailed(ctx, uploadID, []string{fmt.Sprintf("Failed to get upload record: %v", err)})
		return nil, fmt.Errorf("failed to get upload record: %w", err)
	}

	// Get file path
	filePath := s.fileStore.GetFilePath(upload.Filename)

	// Parse Excel file
	log.Printf("Starting to parse Excel file: %s", filePath)
	parseResult, err := s.excelParser.ParseExcelFile(filePath, uploadID)
	if err != nil {
		errorMsg := fmt.Sprintf("Failed to parse Excel file: %v", err)
		s.markProcessingFailed(ctx, uploadID, []string{errorMsg})
		return nil, fmt.Errorf("failed to parse Excel file: %w", err)
	}

	progress.TotalRows = parseResult.TotalRows
	progress.ValidRows = parseResult.ValidRows
	progress.ErrorCount = len(parseResult.Errors)

	log.Printf("Parsed Excel file: %d total rows, %d valid rows, %d errors", 
		parseResult.TotalRows, parseResult.ValidRows, len(parseResult.Errors))

	// Collect error messages
	errorMessages := make([]string, 0)
	for _, validationError := range parseResult.Errors {
		errorMessages = append(errorMessages, validationError.Error())
	}
	progress.Errors = errorMessages

	// If we have valid incidents, process them with analysis and then insert
	var insertResult *BatchInsertResult
	if len(parseResult.Incidents) > 0 {
		log.Printf("Processing %d incidents with analysis", len(parseResult.Incidents))
		
		// Process incidents with sentiment and automation analysis
		err = s.processIncidentsWithAnalysis(parseResult.Incidents)
		if err != nil {
			log.Printf("Warning: Analysis processing failed: %v", err)
			// Continue with insertion even if analysis fails
		}

		log.Printf("Inserting %d incidents into database", len(parseResult.Incidents))
		insertResult, err = s.incidentService.BatchInsertIncidents(ctx, parseResult.Incidents, uploadID)
		if err != nil {
			errorMsg := fmt.Sprintf("Failed to insert incidents: %v", err)
			s.markProcessingFailed(ctx, uploadID, append(errorMessages, errorMsg))
			return nil, fmt.Errorf("failed to insert incidents: %w", err)
		}

		progress.ProcessedRows = insertResult.InsertedCount

		// Add insertion errors to the error list
		for _, insertError := range insertResult.Errors {
			errorMessages = append(errorMessages, insertError.Error())
		}
		progress.Errors = errorMessages
		progress.ErrorCount = len(errorMessages)

		log.Printf("Inserted %d incidents successfully", insertResult.InsertedCount)
	}

	// Determine final status
	finalStatus := models.UploadStatusCompleted
	if progress.ProcessedRows == 0 && progress.ErrorCount > 0 {
		finalStatus = models.UploadStatusFailed
	}

	// Update final upload status
	err = s.incidentService.UpdateUploadStatus(ctx, uploadID, finalStatus, 
		progress.TotalRows, progress.ProcessedRows, progress.ErrorCount, errorMessages)
	if err != nil {
		log.Printf("Warning: Failed to update final upload status: %v", err)
	}

	// Set completion time and duration
	endTime := time.Now()
	progress.EndTime = &endTime
	progress.Status = finalStatus
	progress.Duration = endTime.Sub(progress.StartTime).String()

	log.Printf("Processing completed for upload %s: status=%s, processed=%d, errors=%d", 
		uploadID, finalStatus, progress.ProcessedRows, progress.ErrorCount)

	return progress, nil
}

// RollbackProcessing rolls back a failed processing operation
func (s *ProcessingService) RollbackProcessing(ctx context.Context, uploadID string) error {
	log.Printf("Rolling back processing for upload %s", uploadID)

	// Delete any inserted incidents
	if err := s.incidentService.DeleteIncidentsByUpload(ctx, uploadID); err != nil {
		log.Printf("Warning: Failed to delete incidents during rollback: %v", err)
	}

	// Reset upload status
	err := s.incidentService.UpdateUploadStatus(ctx, uploadID, models.UploadStatusUploaded, 0, 0, 0, nil)
	if err != nil {
		return fmt.Errorf("failed to reset upload status during rollback: %w", err)
	}

	log.Printf("Rollback completed for upload %s", uploadID)
	return nil
}

// GetProcessingStatus returns the current processing status of an upload
func (s *ProcessingService) GetProcessingStatus(ctx context.Context, uploadID string) (*ProcessingProgress, error) {
	upload, err := s.getUploadRecord(ctx, uploadID)
	if err != nil {
		return nil, fmt.Errorf("failed to get upload record: %w", err)
	}

	progress := &ProcessingProgress{
		UploadID:      upload.ID,
		Status:        upload.Status,
		TotalRows:     upload.RecordCount,
		ProcessedRows: upload.ProcessedCount,
		ErrorCount:    upload.ErrorCount,
		Errors:        upload.Errors,
	}

	// Calculate duration if processing is complete
	if upload.ProcessedAt != nil {
		duration := upload.ProcessedAt.Sub(upload.CreatedAt)
		progress.Duration = duration.String()
		progress.EndTime = upload.ProcessedAt
	}

	return progress, nil
}

// markProcessingFailed marks an upload as failed with error messages
func (s *ProcessingService) markProcessingFailed(ctx context.Context, uploadID string, errors []string) {
	err := s.incidentService.UpdateUploadStatus(ctx, uploadID, models.UploadStatusFailed, 0, 0, len(errors), errors)
	if err != nil {
		log.Printf("Failed to mark upload %s as failed: %v", uploadID, err)
	}
}

// getUploadRecord retrieves an upload record from the database
func (s *ProcessingService) getUploadRecord(ctx context.Context, uploadID string) (*models.Upload, error) {
	query := `
		SELECT id, filename, original_filename, status, record_count, 
			   processed_count, error_count, errors, created_at, processed_at
		FROM uploads 
		WHERE id = ?
	`

	var upload models.Upload
	var errorsJSON string

	err := s.db.QueryRowContext(ctx, query, uploadID).Scan(
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

// processIncidentsWithAnalysis processes incidents with sentiment and automation analysis
func (s *ProcessingService) processIncidentsWithAnalysis(incidents []models.Incident) error {
	log.Printf("Starting analysis processing for %d incidents", len(incidents))

	for i := range incidents {
		// Calculate resolution time if not already calculated
		incidents[i].CalculateResolutionTime()

		// Perform sentiment analysis
		if s.sentimentAnalyzer != nil {
			sentimentResult, err := s.sentimentAnalyzer.AnalyzeSentiment(
				incidents[i].BriefDescription + " " + incidents[i].Description)
			if err != nil {
				log.Printf("Warning: Sentiment analysis failed for incident %s: %v", 
					incidents[i].IncidentID, err)
			} else {
				incidents[i].SentimentScore = &sentimentResult.Score
				incidents[i].SentimentLabel = sentimentResult.Label
			}
		}

		// Perform automation analysis
		if s.automationAnalyzer != nil {
			automationResult, err := s.automationAnalyzer.AnalyzeAutomation(&incidents[i])
			if err != nil {
				log.Printf("Warning: Automation analysis failed for incident %s: %v", 
					incidents[i].IncidentID, err)
			} else {
				incidents[i].AutomationScore = &automationResult.Score
				incidents[i].AutomationFeasible = &automationResult.Feasible
				incidents[i].ITProcessGroup = automationResult.ITProcessGroup
			}
		}
	}

	log.Printf("Completed analysis processing for %d incidents", len(incidents))
	return nil
}