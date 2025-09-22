package models

import (
	"testing"
	"time"
)

func TestIncidentValidation(t *testing.T) {
	// Test valid incident
	validIncident := &Incident{
		IncidentID:       "INC001",
		ReportDate:       time.Now(),
		BriefDescription: "Test incident",
		ApplicationName:  "Test App",
		ResolutionGroup:  "IT Support",
		ResolvedPerson:   "John Doe",
		Priority:         PriorityP1,
	}

	if err := validIncident.Validate(); err != nil {
		t.Errorf("Valid incident should not have validation errors: %v", err)
	}

	// Test missing required fields
	invalidIncident := &Incident{}
	err := invalidIncident.Validate()
	if err == nil {
		t.Error("Invalid incident should have validation errors")
	}

	validationErrors, ok := err.(ValidationErrors)
	if !ok {
		t.Error("Error should be ValidationErrors type")
	}

	expectedFields := []string{"incident_id", "brief_description", "application_name", "resolution_group", "resolved_person", "priority"}
	if len(validationErrors) != len(expectedFields) {
		t.Errorf("Expected %d validation errors, got %d", len(expectedFields), len(validationErrors))
	}

	// Test invalid priority
	invalidPriorityIncident := &Incident{
		IncidentID:       "INC001",
		ReportDate:       time.Now(),
		BriefDescription: "Test incident",
		ApplicationName:  "Test App",
		ResolutionGroup:  "IT Support",
		ResolvedPerson:   "John Doe",
		Priority:         "INVALID",
	}

	err = invalidPriorityIncident.Validate()
	if err == nil {
		t.Error("Incident with invalid priority should have validation errors")
	}

	// Test invalid date
	now := time.Now()
	yesterday := now.AddDate(0, 0, -1)
	invalidDateIncident := &Incident{
		IncidentID:       "INC001",
		ReportDate:       now,
		ResolveDate:      &yesterday,
		BriefDescription: "Test incident",
		ApplicationName:  "Test App",
		ResolutionGroup:  "IT Support",
		ResolvedPerson:   "John Doe",
		Priority:         PriorityP1,
	}

	err = invalidDateIncident.Validate()
	if err == nil {
		t.Error("Incident with resolve date before report date should have validation errors")
	}
}

func TestUploadValidation(t *testing.T) {
	// Test valid upload
	validUpload := &Upload{
		Filename:         "test.xlsx",
		OriginalFilename: "original_test.xlsx",
		Status:           UploadStatusUploaded,
		RecordCount:      100,
		ProcessedCount:   50,
		ErrorCount:       0,
	}

	if err := validUpload.Validate(); err != nil {
		t.Errorf("Valid upload should not have validation errors: %v", err)
	}

	// Test missing required fields
	invalidUpload := &Upload{}
	err := invalidUpload.Validate()
	if err == nil {
		t.Error("Invalid upload should have validation errors")
	}

	// Test invalid status
	invalidStatusUpload := &Upload{
		Filename:         "test.xlsx",
		OriginalFilename: "original_test.xlsx",
		Status:           "INVALID_STATUS",
	}

	err = invalidStatusUpload.Validate()
	if err == nil {
		t.Error("Upload with invalid status should have validation errors")
	}

	// Test invalid counts
	invalidCountUpload := &Upload{
		Filename:         "test.xlsx",
		OriginalFilename: "original_test.xlsx",
		Status:           UploadStatusUploaded,
		RecordCount:      100,
		ProcessedCount:   150, // More than record count
	}

	err = invalidCountUpload.Validate()
	if err == nil {
		t.Error("Upload with processed count > record count should have validation errors")
	}
}

func TestIncidentCalculateResolutionTime(t *testing.T) {
	reportTime := time.Date(2023, 1, 1, 10, 0, 0, 0, time.UTC)
	resolveTime := time.Date(2023, 1, 1, 14, 30, 0, 0, time.UTC) // 4.5 hours later

	incident := &Incident{
		ReportDate:  reportTime,
		ResolveDate: &resolveTime,
	}

	incident.CalculateResolutionTime()

	if incident.ResolutionTimeHours == nil {
		t.Error("Resolution time hours should be calculated")
	}

	if *incident.ResolutionTimeHours != 4 {
		t.Errorf("Expected resolution time to be 4 hours, got %d", *incident.ResolutionTimeHours)
	}
}

func TestUploadMethods(t *testing.T) {
	upload := &Upload{
		Status: UploadStatusCompleted,
	}

	if !upload.IsCompleted() {
		t.Error("Upload should be completed")
	}

	if upload.IsFailed() {
		t.Error("Upload should not be failed")
	}

	if upload.IsProcessing() {
		t.Error("Upload should not be processing")
	}

	// Test adding errors
	upload.AddError("Test error 1")
	upload.AddError("Test error 2")

	if len(upload.Errors) != 2 {
		t.Errorf("Expected 2 errors, got %d", len(upload.Errors))
	}

	if upload.ErrorCount != 2 {
		t.Errorf("Expected error count to be 2, got %d", upload.ErrorCount)
	}

	// Test clearing errors
	upload.ClearErrors()

	if len(upload.Errors) != 0 {
		t.Errorf("Expected 0 errors after clearing, got %d", len(upload.Errors))
	}

	if upload.ErrorCount != 0 {
		t.Errorf("Expected error count to be 0 after clearing, got %d", upload.ErrorCount)
	}
}

func TestValidationErrorForRow(t *testing.T) {
	invalidIncident := &Incident{
		Priority: "INVALID",
	}

	err := invalidIncident.ValidateForRow(5)
	if err == nil {
		t.Error("Should have validation errors")
	}

	validationErrors, ok := err.(ValidationErrors)
	if !ok {
		t.Error("Error should be ValidationErrors type")
	}

	// Check that row numbers are set
	for _, validationError := range validationErrors {
		if validationError.Row != 5 {
			t.Errorf("Expected row number to be 5, got %d", validationError.Row)
		}
	}
}

func TestSetDefaults(t *testing.T) {
	incident := &Incident{}
	incident.SetDefaults()

	if incident.CreatedAt.IsZero() {
		t.Error("CreatedAt should be set to current time")
	}

	if incident.UpdatedAt.IsZero() {
		t.Error("UpdatedAt should be set to current time")
	}

	upload := &Upload{}
	upload.SetDefaults()

	if upload.CreatedAt.IsZero() {
		t.Error("CreatedAt should be set to current time")
	}

	if upload.Status != UploadStatusUploaded {
		t.Errorf("Expected status to be %s, got %s", UploadStatusUploaded, upload.Status)
	}
}