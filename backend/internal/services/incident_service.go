package services

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"incident-management-system/internal/models"
)

// IncidentService handles incident data operations
type IncidentService struct {
	db *sql.DB
}

// NewIncidentService creates a new IncidentService instance
func NewIncidentService(db *sql.DB) *IncidentService {
	return &IncidentService{
		db: db,
	}
}

// BatchInsertResult represents the result of a batch insert operation
type BatchInsertResult struct {
	InsertedCount int                      `json:"inserted_count"`
	Errors        []models.ValidationError `json:"errors"`
	Success       bool                     `json:"success"`
}

// BatchInsertIncidents inserts multiple incidents in a single transaction
func (s *IncidentService) BatchInsertIncidents(ctx context.Context, incidents []models.Incident, uploadID string) (*BatchInsertResult, error) {
	if len(incidents) == 0 {
		return &BatchInsertResult{
			InsertedCount: 0,
			Errors:        []models.ValidationError{},
			Success:       true,
		}, nil
	}

	// Start transaction
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}

	// Ensure rollback on error
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	result := &BatchInsertResult{
		InsertedCount: 0,
		Errors:        make([]models.ValidationError, 0),
		Success:       false,
	}

	// Prepare insert statement
	insertQuery := `
		INSERT INTO incidents (
			id, upload_id, incident_id, report_date, resolve_date, last_resolve_date,
			brief_description, description, application_name, resolution_group, 
			resolved_person, priority, category, subcategory, impact, urgency, 
			status, customer_affected, business_service, root_cause, resolution_notes,
			sentiment_score, sentiment_label, resolution_time_hours, automation_score,
			automation_feasible, it_process_group, created_at, updated_at
		) VALUES (
			?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 
			?, ?, ?, ?, ?, ?, ?, ?
		)
	`

	stmt, err := tx.PrepareContext(ctx, insertQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare insert statement: %w", err)
	}
	defer stmt.Close()

	// Check for duplicate incident IDs within the upload
	duplicateMap := make(map[string]bool)
	
	// Insert incidents one by one to handle individual errors
	for i, incident := range incidents {
		// Check for duplicates within this batch
		if duplicateMap[incident.IncidentID] {
			result.Errors = append(result.Errors, models.ValidationError{
				Field:   "incident_id",
				Value:   incident.IncidentID,
				Message: "duplicate incident ID within upload",
				Row:     i + 2, // Excel row number (1-based + header)
			})
			continue
		}
		duplicateMap[incident.IncidentID] = true

		// Check for existing incident ID in database
		exists, err := s.checkIncidentExists(ctx, tx, incident.IncidentID, uploadID)
		if err != nil {
			result.Errors = append(result.Errors, models.ValidationError{
				Field:   "incident_id",
				Value:   incident.IncidentID,
				Message: fmt.Sprintf("database error checking duplicate: %v", err),
				Row:     i + 2,
			})
			continue
		}

		if exists {
			result.Errors = append(result.Errors, models.ValidationError{
				Field:   "incident_id",
				Value:   incident.IncidentID,
				Message: "incident ID already exists in this upload",
				Row:     i + 2,
			})
			continue
		}

		// Execute insert
		// Convert empty strings to nil for optional fields
		var sentimentLabel interface{}
		if incident.SentimentLabel == "" {
			sentimentLabel = nil
		} else {
			sentimentLabel = incident.SentimentLabel
		}

		_, err = stmt.ExecContext(ctx,
			incident.ID,
			incident.UploadID,
			incident.IncidentID,
			incident.ReportDate,
			incident.ResolveDate,
			incident.LastResolveDate,
			incident.BriefDescription,
			incident.Description,
			incident.ApplicationName,
			incident.ResolutionGroup,
			incident.ResolvedPerson,
			incident.Priority,
			incident.Category,
			incident.Subcategory,
			incident.Impact,
			incident.Urgency,
			incident.Status,
			incident.CustomerAffected,
			incident.BusinessService,
			incident.RootCause,
			incident.ResolutionNotes,
			incident.SentimentScore,
			sentimentLabel,
			incident.ResolutionTimeHours,
			incident.AutomationScore,
			incident.AutomationFeasible,
			incident.ITProcessGroup,
			incident.CreatedAt,
			incident.UpdatedAt,
		)

		if err != nil {
			// Handle constraint violations and other database errors
			errorMsg := err.Error()
			if strings.Contains(errorMsg, "UNIQUE constraint failed") {
				result.Errors = append(result.Errors, models.ValidationError{
					Field:   "incident_id",
					Value:   incident.IncidentID,
					Message: "incident ID already exists",
					Row:     i + 2,
				})
			} else if strings.Contains(errorMsg, "CHECK constraint failed") {
				result.Errors = append(result.Errors, models.ValidationError{
					Field:   "general",
					Value:   "",
					Message: "data validation failed: " + errorMsg,
					Row:     i + 2,
				})
			} else {
				result.Errors = append(result.Errors, models.ValidationError{
					Field:   "general",
					Value:   "",
					Message: "database error: " + errorMsg,
					Row:     i + 2,
				})
			}
			continue
		}

		result.InsertedCount++
	}

	// Commit transaction if we have any successful inserts
	if result.InsertedCount > 0 {
		if err = tx.Commit(); err != nil {
			return nil, fmt.Errorf("failed to commit transaction: %w", err)
		}
		result.Success = true
	} else {
		// Rollback if no successful inserts
		tx.Rollback()
		result.Success = false
	}

	return result, nil
}

// checkIncidentExists checks if an incident ID already exists for the given upload
func (s *IncidentService) checkIncidentExists(ctx context.Context, tx *sql.Tx, incidentID, uploadID string) (bool, error) {
	query := "SELECT COUNT(*) FROM incidents WHERE incident_id = ? AND upload_id = ?"
	
	var count int
	err := tx.QueryRowContext(ctx, query, incidentID, uploadID).Scan(&count)
	if err != nil {
		return false, err
	}
	
	return count > 0, nil
}

// UpdateUploadStatus updates the status and statistics of an upload
func (s *IncidentService) UpdateUploadStatus(ctx context.Context, uploadID string, status string, recordCount, processedCount, errorCount int, errors []string) error {
	// Convert errors to JSON string (simplified for now)
	errorsJSON := "[]"
	if len(errors) > 0 {
		// In production, use proper JSON marshaling
		errorsJSON = fmt.Sprintf(`["%s"]`, strings.Join(errors, `", "`))
	}

	// Debug: Check if record exists
	var existingCount int
	checkQuery := "SELECT COUNT(*) FROM uploads WHERE id = ?"
	err := s.db.QueryRowContext(ctx, checkQuery, uploadID).Scan(&existingCount)
	if err != nil {
		return fmt.Errorf("failed to check existing upload: %w", err)
	}
	
	if existingCount == 0 {
		return fmt.Errorf("upload record not found: %s", uploadID)
	}

	// Update without processed_at first
	query := `
		UPDATE uploads 
		SET status = ?, record_count = ?, processed_count = ?, error_count = ?, errors = ?
		WHERE id = ?
	`

	result, err := s.db.ExecContext(ctx, query, status, recordCount, processedCount, errorCount, errorsJSON, uploadID)
	if err != nil {
		return fmt.Errorf("failed to update upload status (uploadID=%s, status=%s): %w", uploadID, status, err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("no rows updated for upload %s", uploadID)
	}

	// Update processed_at separately if not processing
	if status != models.UploadStatusProcessing {
		processedAtQuery := "UPDATE uploads SET processed_at = ? WHERE id = ?"
		_, err = s.db.ExecContext(ctx, processedAtQuery, time.Now(), uploadID)
		if err != nil {
			return fmt.Errorf("failed to update processed_at: %w", err)
		}
	}

	return nil
}

// GetIncidentsByUpload retrieves all incidents for a specific upload
func (s *IncidentService) GetIncidentsByUpload(ctx context.Context, uploadID string) ([]models.Incident, error) {
	query := `
		SELECT id, upload_id, incident_id, report_date, resolve_date, last_resolve_date,
			   brief_description, description, application_name, resolution_group,
			   resolved_person, priority, category, subcategory, impact, urgency,
			   status, customer_affected, business_service, root_cause, resolution_notes,
			   sentiment_score, sentiment_label, resolution_time_hours, automation_score,
			   automation_feasible, it_process_group, created_at, updated_at
		FROM incidents 
		WHERE upload_id = ?
		ORDER BY created_at ASC
	`

	rows, err := s.db.QueryContext(ctx, query, uploadID)
	if err != nil {
		return nil, fmt.Errorf("failed to query incidents: %w", err)
	}
	defer rows.Close()

	var incidents []models.Incident
	for rows.Next() {
		var incident models.Incident
		
		err := rows.Scan(
			&incident.ID,
			&incident.UploadID,
			&incident.IncidentID,
			&incident.ReportDate,
			&incident.ResolveDate,
			&incident.LastResolveDate,
			&incident.BriefDescription,
			&incident.Description,
			&incident.ApplicationName,
			&incident.ResolutionGroup,
			&incident.ResolvedPerson,
			&incident.Priority,
			&incident.Category,
			&incident.Subcategory,
			&incident.Impact,
			&incident.Urgency,
			&incident.Status,
			&incident.CustomerAffected,
			&incident.BusinessService,
			&incident.RootCause,
			&incident.ResolutionNotes,
			&incident.SentimentScore,
			&incident.SentimentLabel,
			&incident.ResolutionTimeHours,
			&incident.AutomationScore,
			&incident.AutomationFeasible,
			&incident.ITProcessGroup,
			&incident.CreatedAt,
			&incident.UpdatedAt,
		)
		
		if err != nil {
			return nil, fmt.Errorf("failed to scan incident: %w", err)
		}
		
		incidents = append(incidents, incident)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating incidents: %w", err)
	}

	return incidents, nil
}

// DeleteIncidentsByUpload deletes all incidents for a specific upload (for rollback)
func (s *IncidentService) DeleteIncidentsByUpload(ctx context.Context, uploadID string) error {
	query := "DELETE FROM incidents WHERE upload_id = ?"
	
	_, err := s.db.ExecContext(ctx, query, uploadID)
	if err != nil {
		return fmt.Errorf("failed to delete incidents for upload %s: %w", uploadID, err)
	}

	return nil
}

// GetIncidentCount returns the total number of incidents for an upload
func (s *IncidentService) GetIncidentCount(ctx context.Context, uploadID string) (int, error) {
	query := "SELECT COUNT(*) FROM incidents WHERE upload_id = ?"
	
	var count int
	err := s.db.QueryRowContext(ctx, query, uploadID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to get incident count: %w", err)
	}

	return count, nil
}