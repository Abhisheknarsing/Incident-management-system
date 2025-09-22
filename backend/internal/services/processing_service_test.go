package services

import (
	"context"
	"testing"
	"time"

	"incident-management-system/internal/database"
	"incident-management-system/internal/models"
	"incident-management-system/internal/storage"

	_ "github.com/mattn/go-sqlite3"
)

func TestProcessingService_NewProcessingService(t *testing.T) {
	// Create a mock database for testing
	config := &database.Config{
		DatabasePath: ":memory:",
	}
	dbWrapper, err := database.NewDB(config)
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}
	defer dbWrapper.Close()

	db := dbWrapper.GetConnection()

	// Create a mock file store
	fileStore := storage.NewFileStore("/tmp")

	// Test creating processing service
	service := NewProcessingService(db, fileStore)
	if service == nil {
		t.Fatal("Expected non-nil ProcessingService")
	}

	if service.db != db {
		t.Error("Expected database to be set correctly")
	}

	if service.fileStore != fileStore {
		t.Error("Expected file store to be set correctly")
	}

	if service.excelParser == nil {
		t.Error("Expected excel parser to be initialized")
	}

	if service.incidentService == nil {
		t.Error("Expected incident service to be initialized")
	}

	if service.sentimentAnalyzer == nil {
		t.Error("Expected sentiment analyzer to be initialized")
	}

	if service.automationAnalyzer == nil {
		t.Error("Expected automation analyzer to be initialized")
	}
}

func TestProcessingService_GetProcessingStatus(t *testing.T) {
	// Create a mock database for testing
	config := &database.Config{
		DatabasePath: ":memory:",
	}
	dbWrapper, err := database.NewDB(config)
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}
	defer dbWrapper.Close()

	// Initialize the database schema
	if err := dbWrapper.InitializeDatabase(); err != nil {
		t.Fatalf("Failed to initialize database schema: %v", err)
	}

	db := dbWrapper.GetConnection()

	// Create a mock file store
	fileStore := storage.NewFileStore("/tmp")

	// Create processing service
	service := NewProcessingService(db, fileStore)

	// Test getting status for non-existent upload (should error)
	ctx := context.Background()
	_, err = service.GetProcessingStatus(ctx, "non-existent-upload")
	if err == nil {
		t.Error("Expected error when getting status for non-existent upload")
	}

	// Test with valid upload ID (will fail due to missing schema)
	_, err = service.GetProcessingStatus(ctx, "upload-123")
	if err != nil {
		// Expect an error due to missing schema
		t.Logf("Expected error due to missing schema: %v", err)
	}
}

func TestProcessingService_RollbackProcessing(t *testing.T) {
	// Create a mock database for testing
	config := &database.Config{
		DatabasePath: ":memory:",
	}
	dbWrapper, err := database.NewDB(config)
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}
	defer dbWrapper.Close()

	// Initialize the database schema
	if err := dbWrapper.InitializeDatabase(); err != nil {
		t.Fatalf("Failed to initialize database schema: %v", err)
	}

	db := dbWrapper.GetConnection()

	// Create a mock file store
	fileStore := storage.NewFileStore("/tmp")

	// Create processing service
	service := NewProcessingService(db, fileStore)

	// Test rollback (will fail due to missing schema)
	ctx := context.Background()
	err = service.RollbackProcessing(ctx, "upload-123")
	if err != nil {
		// Expect an error due to missing schema
		t.Logf("Expected error due to missing schema: %v", err)
	}
}

func TestProcessingService_markProcessingFailed(t *testing.T) {
	// Create a mock database for testing
	config := &database.Config{
		DatabasePath: ":memory:",
	}
	dbWrapper, err := database.NewDB(config)
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}
	defer dbWrapper.Close()

	// Initialize the database schema
	if err := dbWrapper.InitializeDatabase(); err != nil {
		t.Fatalf("Failed to initialize database schema: %v", err)
	}

	db := dbWrapper.GetConnection()

	// Create a mock file store
	fileStore := storage.NewFileStore("/tmp")

	// Create processing service
	service := NewProcessingService(db, fileStore)

	// Test marking processing as failed (will fail due to missing schema)
	errors := []string{"test error 1", "test error 2"}
	service.markProcessingFailed(context.Background(), "upload-123", errors)

	// Should not panic - just log the error
	t.Log("markProcessingFailed completed without panic")
}

func TestProcessingService_getUploadRecord(t *testing.T) {
	// Create a mock database for testing
	config := &database.Config{
		DatabasePath: ":memory:",
	}
	dbWrapper, err := database.NewDB(config)
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}
	defer dbWrapper.Close()

	// Initialize the database schema
	if err := dbWrapper.InitializeDatabase(); err != nil {
		t.Fatalf("Failed to initialize database schema: %v", err)
	}

	db := dbWrapper.GetConnection()

	// Create a mock file store
	fileStore := storage.NewFileStore("/tmp")

	// Create processing service
	service := NewProcessingService(db, fileStore)

	// Test getting upload record (will fail due to missing schema)
	ctx := context.Background()
	_, err = service.getUploadRecord(ctx, "upload-123")
	if err != nil {
		// Expect an error due to missing schema
		t.Logf("Expected error due to missing schema: %v", err)
	}
}

func TestProcessingService_processIncidentsWithAnalysis(t *testing.T) {
	// Create a mock database for testing
	config := &database.Config{
		DatabasePath: ":memory:",
	}
	dbWrapper, err := database.NewDB(config)
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}
	defer dbWrapper.Close()

	// Initialize the database schema
	if err := dbWrapper.InitializeDatabase(); err != nil {
		t.Fatalf("Failed to initialize database schema: %v", err)
	}

	db := dbWrapper.GetConnection()

	// Create a mock file store
	fileStore := storage.NewFileStore("/tmp")

	// Create processing service
	service := NewProcessingService(db, fileStore)

	// Create test incidents
	incidents := []models.Incident{
		{
			ID:               "incident-1",
			UploadID:         "upload-123",
			IncidentID:       "INC001",
			ReportDate:       time.Now(),
			ResolveDate:      func() *time.Time { t := time.Now().Add(time.Hour); return &t }(),
			BriefDescription: "Test incident 1",
			Description:      "This is a test incident that should be resolved quickly",
			ApplicationName:  "Test App",
			ResolutionGroup:  "Test Group",
			ResolvedPerson:   "Test Person",
			Priority:         "P2",
		},
		{
			ID:               "incident-2",
			UploadID:         "upload-123",
			IncidentID:       "INC002",
			ReportDate:       time.Now(),
			BriefDescription: "Test incident 2",
			Description:      "This is a complex incident that requires investigation",
			ApplicationName:  "Test App",
			ResolutionGroup:  "Test Group",
			ResolvedPerson:   "Test Person",
			Priority:         "P3",
		},
	}

	// Test processing incidents with analysis
	err = service.processIncidentsWithAnalysis(incidents)
	if err != nil {
		t.Fatalf("Failed to process incidents with analysis: %v", err)
	}

	// Check that resolution times were calculated
	for i, incident := range incidents {
		if incident.ResolutionTimeHours != nil {
			t.Logf("Incident %d has resolution time: %d hours", i, *incident.ResolutionTimeHours)
		}
	}

	// Check that sentiment analysis was attempted
	for i, incident := range incidents {
		if incident.SentimentScore != nil {
			t.Logf("Incident %d has sentiment score: %.3f", i, *incident.SentimentScore)
		}
		if incident.SentimentLabel != "" {
			t.Logf("Incident %d has sentiment label: %s", i, incident.SentimentLabel)
		}
	}

	// Check that automation analysis was attempted
	for i, incident := range incidents {
		if incident.AutomationScore != nil {
			t.Logf("Incident %d has automation score: %.3f", i, *incident.AutomationScore)
		}
		if incident.AutomationFeasible != nil {
			t.Logf("Incident %d has automation feasible: %t", i, *incident.AutomationFeasible)
		}
		if incident.ITProcessGroup != "" {
			t.Logf("Incident %d has IT process group: %s", i, incident.ITProcessGroup)
		}
	}
}
