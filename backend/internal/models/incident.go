package models

import (
	"fmt"
	"strings"
	"time"
)

// Incident represents the core incident data structure
type Incident struct {
	ID                   string     `json:"id" db:"id"`
	UploadID            string     `json:"upload_id" db:"upload_id"`
	IncidentID          string     `json:"incident_id" db:"incident_id"`
	ReportDate          time.Time  `json:"report_date" db:"report_date"`
	ResolveDate         *time.Time `json:"resolve_date,omitempty" db:"resolve_date"`
	LastResolveDate     *time.Time `json:"last_resolve_date,omitempty" db:"last_resolve_date"`
	BriefDescription    string     `json:"brief_description" db:"brief_description"`
	Description         string     `json:"description" db:"description"`
	ApplicationName     string     `json:"application_name" db:"application_name"`
	ResolutionGroup     string     `json:"resolution_group" db:"resolution_group"`
	ResolvedPerson      string     `json:"resolved_person" db:"resolved_person"`
	Priority            string     `json:"priority" db:"priority"`
	
	// Additional fields
	Category            string     `json:"category,omitempty" db:"category"`
	Subcategory         string     `json:"subcategory,omitempty" db:"subcategory"`
	Impact              string     `json:"impact,omitempty" db:"impact"`
	Urgency             string     `json:"urgency,omitempty" db:"urgency"`
	Status              string     `json:"status,omitempty" db:"status"`
	CustomerAffected    string     `json:"customer_affected,omitempty" db:"customer_affected"`
	BusinessService     string     `json:"business_service,omitempty" db:"business_service"`
	RootCause           string     `json:"root_cause,omitempty" db:"root_cause"`
	ResolutionNotes     string     `json:"resolution_notes,omitempty" db:"resolution_notes"`
	
	// Derived fields
	SentimentScore      *float64   `json:"sentiment_score,omitempty" db:"sentiment_score"`
	SentimentLabel      string     `json:"sentiment_label,omitempty" db:"sentiment_label"`
	ResolutionTimeHours *int       `json:"resolution_time_hours,omitempty" db:"resolution_time_hours"`
	AutomationScore     *float64   `json:"automation_score,omitempty" db:"automation_score"`
	AutomationFeasible  *bool      `json:"automation_feasible,omitempty" db:"automation_feasible"`
	ITProcessGroup      string     `json:"it_process_group,omitempty" db:"it_process_group"`
	
	CreatedAt           time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt           time.Time  `json:"updated_at" db:"updated_at"`
}

// Upload represents file upload metadata
type Upload struct {
	ID               string    `json:"id" db:"id"`
	Filename         string    `json:"filename" db:"filename"`
	OriginalFilename string    `json:"original_filename" db:"original_filename"`
	Status           string    `json:"status" db:"status"`
	RecordCount      int       `json:"record_count" db:"record_count"`
	ProcessedCount   int       `json:"processed_count" db:"processed_count"`
	ErrorCount       int       `json:"error_count" db:"error_count"`
	Errors           []string  `json:"errors,omitempty" db:"errors"`
	CreatedAt        time.Time `json:"created_at" db:"created_at"`
	ProcessedAt      *time.Time `json:"processed_at,omitempty" db:"processed_at"`
}

// Constants for validation
const (
	// Upload status values
	UploadStatusUploaded   = "uploaded"
	UploadStatusProcessing = "processing"
	UploadStatusCompleted  = "completed"
	UploadStatusFailed     = "failed"

	// Priority values
	PriorityP1 = "P1"
	PriorityP2 = "P2"
	PriorityP3 = "P3"
	PriorityP4 = "P4"

	// Sentiment labels
	SentimentPositive = "positive"
	SentimentNegative = "negative"
	SentimentNeutral  = "neutral"
)

// Valid values for validation
var (
	ValidUploadStatuses = []string{UploadStatusUploaded, UploadStatusProcessing, UploadStatusCompleted, UploadStatusFailed}
	ValidPriorities     = []string{PriorityP1, PriorityP2, PriorityP3, PriorityP4}
	ValidSentiments     = []string{SentimentPositive, SentimentNegative, SentimentNeutral}
)

// ValidationError represents a field validation error
type ValidationError struct {
	Field   string `json:"field"`
	Value   string `json:"value"`
	Message string `json:"message"`
	Row     int    `json:"row,omitempty"`
}

func (e ValidationError) Error() string {
	if e.Row > 0 {
		return fmt.Sprintf("row %d, field '%s': %s (value: '%s')", e.Row, e.Field, e.Message, e.Value)
	}
	return fmt.Sprintf("field '%s': %s (value: '%s')", e.Field, e.Message, e.Value)
}

// ValidationErrors represents multiple validation errors
type ValidationErrors []ValidationError

func (e ValidationErrors) Error() string {
	if len(e) == 0 {
		return "no validation errors"
	}
	if len(e) == 1 {
		return e[0].Error()
	}
	return fmt.Sprintf("%d validation errors: %s (and %d more)", len(e), e[0].Error(), len(e)-1)
}

// Validate validates the incident data
func (i *Incident) Validate() error {
	var errors ValidationErrors

	// Required fields validation
	if strings.TrimSpace(i.IncidentID) == "" {
		errors = append(errors, ValidationError{
			Field:   "incident_id",
			Value:   i.IncidentID,
			Message: "incident ID is required",
		})
	}

	if strings.TrimSpace(i.BriefDescription) == "" {
		errors = append(errors, ValidationError{
			Field:   "brief_description",
			Value:   i.BriefDescription,
			Message: "brief description is required",
		})
	}

	if strings.TrimSpace(i.ApplicationName) == "" {
		errors = append(errors, ValidationError{
			Field:   "application_name",
			Value:   i.ApplicationName,
			Message: "application name is required",
		})
	}

	if strings.TrimSpace(i.ResolutionGroup) == "" {
		errors = append(errors, ValidationError{
			Field:   "resolution_group",
			Value:   i.ResolutionGroup,
			Message: "resolution group is required",
		})
	}

	if strings.TrimSpace(i.ResolvedPerson) == "" {
		errors = append(errors, ValidationError{
			Field:   "resolved_person",
			Value:   i.ResolvedPerson,
			Message: "resolved person is required",
		})
	}

	if strings.TrimSpace(i.Priority) == "" {
		errors = append(errors, ValidationError{
			Field:   "priority",
			Value:   i.Priority,
			Message: "priority is required",
		})
	}

	// Priority validation
	if i.Priority != "" && !isValidPriority(i.Priority) {
		errors = append(errors, ValidationError{
			Field:   "priority",
			Value:   i.Priority,
			Message: fmt.Sprintf("priority must be one of: %s", strings.Join(ValidPriorities, ", ")),
		})
	}

	// Date validation
	if i.ResolveDate != nil && i.ResolveDate.Before(i.ReportDate) {
		errors = append(errors, ValidationError{
			Field:   "resolve_date",
			Value:   i.ResolveDate.Format("2006-01-02"),
			Message: "resolve date cannot be before report date",
		})
	}

	// Sentiment validation
	if i.SentimentLabel != "" && !isValidSentiment(i.SentimentLabel) {
		errors = append(errors, ValidationError{
			Field:   "sentiment_label",
			Value:   i.SentimentLabel,
			Message: fmt.Sprintf("sentiment label must be one of: %s", strings.Join(ValidSentiments, ", ")),
		})
	}

	// Sentiment score validation
	if i.SentimentScore != nil && (*i.SentimentScore < -1.0 || *i.SentimentScore > 1.0) {
		errors = append(errors, ValidationError{
			Field:   "sentiment_score",
			Value:   fmt.Sprintf("%.3f", *i.SentimentScore),
			Message: "sentiment score must be between -1.0 and 1.0",
		})
	}

	// Automation score validation
	if i.AutomationScore != nil && (*i.AutomationScore < 0.0 || *i.AutomationScore > 1.0) {
		errors = append(errors, ValidationError{
			Field:   "automation_score",
			Value:   fmt.Sprintf("%.3f", *i.AutomationScore),
			Message: "automation score must be between 0.0 and 1.0",
		})
	}

	if len(errors) > 0 {
		return errors
	}

	return nil
}

// ValidateForRow validates the incident data with row context for Excel processing
func (i *Incident) ValidateForRow(row int) error {
	err := i.Validate()
	if err == nil {
		return nil
	}

	// Add row context to validation errors
	if validationErrors, ok := err.(ValidationErrors); ok {
		for j := range validationErrors {
			validationErrors[j].Row = row
		}
		return validationErrors
	}

	return err
}

// Validate validates the upload data
func (u *Upload) Validate() error {
	var errors ValidationErrors

	// Required fields validation
	if strings.TrimSpace(u.Filename) == "" {
		errors = append(errors, ValidationError{
			Field:   "filename",
			Value:   u.Filename,
			Message: "filename is required",
		})
	}

	if strings.TrimSpace(u.OriginalFilename) == "" {
		errors = append(errors, ValidationError{
			Field:   "original_filename",
			Value:   u.OriginalFilename,
			Message: "original filename is required",
		})
	}

	if strings.TrimSpace(u.Status) == "" {
		errors = append(errors, ValidationError{
			Field:   "status",
			Value:   u.Status,
			Message: "status is required",
		})
	}

	// Status validation
	if u.Status != "" && !isValidUploadStatus(u.Status) {
		errors = append(errors, ValidationError{
			Field:   "status",
			Value:   u.Status,
			Message: fmt.Sprintf("status must be one of: %s", strings.Join(ValidUploadStatuses, ", ")),
		})
	}

	// Count validation
	if u.RecordCount < 0 {
		errors = append(errors, ValidationError{
			Field:   "record_count",
			Value:   fmt.Sprintf("%d", u.RecordCount),
			Message: "record count cannot be negative",
		})
	}

	if u.ProcessedCount < 0 {
		errors = append(errors, ValidationError{
			Field:   "processed_count",
			Value:   fmt.Sprintf("%d", u.ProcessedCount),
			Message: "processed count cannot be negative",
		})
	}

	if u.ErrorCount < 0 {
		errors = append(errors, ValidationError{
			Field:   "error_count",
			Value:   fmt.Sprintf("%d", u.ErrorCount),
			Message: "error count cannot be negative",
		})
	}

	if u.ProcessedCount > u.RecordCount {
		errors = append(errors, ValidationError{
			Field:   "processed_count",
			Value:   fmt.Sprintf("%d", u.ProcessedCount),
			Message: "processed count cannot exceed record count",
		})
	}

	if len(errors) > 0 {
		return errors
	}

	return nil
}

// Helper functions for validation
func isValidPriority(priority string) bool {
	for _, valid := range ValidPriorities {
		if priority == valid {
			return true
		}
	}
	return false
}

func isValidUploadStatus(status string) bool {
	for _, valid := range ValidUploadStatuses {
		if status == valid {
			return true
		}
	}
	return false
}

func isValidSentiment(sentiment string) bool {
	for _, valid := range ValidSentiments {
		if sentiment == valid {
			return true
		}
	}
	return false
}

// CalculateResolutionTime calculates resolution time in hours
func (i *Incident) CalculateResolutionTime() {
	if i.ResolveDate != nil {
		duration := i.ResolveDate.Sub(i.ReportDate)
		hours := int(duration.Hours())
		i.ResolutionTimeHours = &hours
	}
}

// SetDefaults sets default values for the incident
func (i *Incident) SetDefaults() {
	now := time.Now()
	if i.CreatedAt.IsZero() {
		i.CreatedAt = now
	}
	if i.UpdatedAt.IsZero() {
		i.UpdatedAt = now
	}
}

// SetDefaults sets default values for the upload
func (u *Upload) SetDefaults() {
	if u.CreatedAt.IsZero() {
		u.CreatedAt = time.Now()
	}
	if u.Status == "" {
		u.Status = UploadStatusUploaded
	}
}

// IsCompleted returns true if the upload is completed
func (u *Upload) IsCompleted() bool {
	return u.Status == UploadStatusCompleted
}

// IsFailed returns true if the upload failed
func (u *Upload) IsFailed() bool {
	return u.Status == UploadStatusFailed
}

// IsProcessing returns true if the upload is being processed
func (u *Upload) IsProcessing() bool {
	return u.Status == UploadStatusProcessing
}

// AddError adds an error to the upload
func (u *Upload) AddError(err string) {
	if u.Errors == nil {
		u.Errors = make([]string, 0)
	}
	u.Errors = append(u.Errors, err)
	u.ErrorCount = len(u.Errors)
}

// ClearErrors clears all errors from the upload
func (u *Upload) ClearErrors() {
	u.Errors = nil
	u.ErrorCount = 0
}