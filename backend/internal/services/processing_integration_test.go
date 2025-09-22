package services

import (
	"testing"
	"time"

	"incident-management-system/internal/database"
	"incident-management-system/internal/models"
	"incident-management-system/internal/storage"
)

func TestProcessingService_ProcessIncidentsWithAnalysis(t *testing.T) {
	// Create test database
	config := &database.Config{
		DatabasePath:    ":memory:",
		MaxOpenConns:    5,
		MaxIdleConns:    2,
		ConnMaxLifetime: time.Minute,
		ConnMaxIdleTime: time.Second * 30,
	}

	dbWrapper, err := database.NewDB(config)
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer dbWrapper.Close()

	if err := dbWrapper.InitializeDatabase(); err != nil {
		t.Fatalf("Failed to initialize database: %v", err)
	}

	db := dbWrapper.GetConnection()

	// Create test file store
	tempDir := t.TempDir()
	fileStore := storage.NewFileStore(tempDir)

	// Create processing service
	service := NewProcessingService(db, fileStore)

	// Create test incidents
	incidents := []models.Incident{
		{
			ID:               "test-1",
			UploadID:         "upload-1",
			IncidentID:       "INC001",
			ReportDate:       time.Now().Add(-2 * time.Hour),
			ResolveDate:      func() *time.Time { t := time.Now(); return &t }(),
			BriefDescription: "Server needs restart",
			Description:      "Application server requires restart to resolve memory issue",
			ApplicationName:  "Web Server",
			ResolutionGroup:  "Infrastructure Team",
			ResolvedPerson:   "John Doe",
			Priority:         "P2",
			ResolutionNotes:  "Automated restart script executed successfully",
		},
		{
			ID:               "test-2",
			UploadID:         "upload-1",
			IncidentID:       "INC002",
			ReportDate:       time.Now().Add(-48 * time.Hour),
			ResolveDate:      func() *time.Time { t := time.Now(); return &t }(),
			BriefDescription: "Complex application error requiring investigation",
			Description:      "Users reporting intermittent errors that require detailed analysis and troubleshooting",
			ApplicationName:  "Custom Business App",
			ResolutionGroup:  "Application Support",
			ResolvedPerson:   "Jane Smith",
			Priority:         "P3",
			ResolutionNotes:  "Required manual investigation and custom code changes",
		},
	}

	// Set defaults for incidents
	for i := range incidents {
		incidents[i].SetDefaults()
	}

	// Process incidents with analysis
	err = service.processIncidentsWithAnalysis(incidents)
	if err != nil {
		t.Fatalf("Failed to process incidents with analysis: %v", err)
	}

	// Verify first incident (should have high automation potential)
	incident1 := incidents[0]
	
	// Check resolution time was calculated
	if incident1.ResolutionTimeHours == nil {
		t.Errorf("Expected resolution time to be calculated for incident 1")
	} else if *incident1.ResolutionTimeHours != 2 {
		t.Errorf("Expected resolution time of 2 hours, got %d", *incident1.ResolutionTimeHours)
	}

	// Check sentiment analysis was performed
	if incident1.SentimentScore == nil {
		t.Errorf("Expected sentiment score to be set for incident 1")
	}
	if incident1.SentimentLabel == "" {
		t.Errorf("Expected sentiment label to be set for incident 1")
	}

	// Check automation analysis was performed
	if incident1.AutomationScore == nil {
		t.Errorf("Expected automation score to be set for incident 1")
	} else {
		if *incident1.AutomationScore < 0.0 || *incident1.AutomationScore > 1.0 {
			t.Errorf("Automation score %.3f is outside valid range [0.0, 1.0]", *incident1.AutomationScore)
		}
	}

	if incident1.AutomationFeasible == nil {
		t.Errorf("Expected automation feasible to be set for incident 1")
	}

	if incident1.ITProcessGroup == "" {
		t.Errorf("Expected IT process group to be set for incident 1")
	} else if incident1.ITProcessGroup != "Infrastructure" {
		t.Errorf("Expected IT process group 'Infrastructure', got '%s'", incident1.ITProcessGroup)
	}

	// Verify second incident (should have lower automation potential)
	incident2 := incidents[1]
	
	// Check resolution time was calculated
	if incident2.ResolutionTimeHours == nil {
		t.Errorf("Expected resolution time to be calculated for incident 2")
	} else if *incident2.ResolutionTimeHours != 48 {
		t.Errorf("Expected resolution time of 48 hours, got %d", *incident2.ResolutionTimeHours)
	}

	// Check automation analysis shows lower potential
	if incident2.AutomationScore != nil && incident1.AutomationScore != nil {
		if *incident2.AutomationScore >= *incident1.AutomationScore {
			t.Errorf("Expected incident 2 to have lower automation score than incident 1")
		}
	}

	// Check IT process group classification
	if incident2.ITProcessGroup == "" {
		t.Errorf("Expected IT process group to be set for incident 2")
	}
}

func TestProcessingService_ProcessIncidentsWithAnalysis_ErrorHandling(t *testing.T) {
	// Create test database
	config := &database.Config{
		DatabasePath:    ":memory:",
		MaxOpenConns:    5,
		MaxIdleConns:    2,
		ConnMaxLifetime: time.Minute,
		ConnMaxIdleTime: time.Second * 30,
	}

	dbWrapper, err := database.NewDB(config)
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer dbWrapper.Close()

	if err := dbWrapper.InitializeDatabase(); err != nil {
		t.Fatalf("Failed to initialize database: %v", err)
	}

	db := dbWrapper.GetConnection()

	// Create test file store
	tempDir := t.TempDir()
	fileStore := storage.NewFileStore(tempDir)

	// Create processing service
	service := NewProcessingService(db, fileStore)

	// Create test incidents with minimal data
	incidents := []models.Incident{
		{
			ID:               "test-1",
			UploadID:         "upload-1",
			IncidentID:       "INC001",
			ReportDate:       time.Now(),
			BriefDescription: "Test incident",
			ApplicationName:  "Test App",
			ResolutionGroup:  "Test Team",
			ResolvedPerson:   "Test Person",
			Priority:         "P3",
		},
	}

	// Set defaults for incidents
	for i := range incidents {
		incidents[i].SetDefaults()
	}

	// Process incidents with analysis - should not fail even with minimal data
	err = service.processIncidentsWithAnalysis(incidents)
	if err != nil {
		t.Fatalf("Processing should not fail with minimal data: %v", err)
	}

	// Verify analysis was attempted
	incident := incidents[0]
	
	// Should have some analysis results even with minimal data
	if incident.ITProcessGroup == "" {
		t.Errorf("Expected IT process group to be set even with minimal data")
	}
}

func TestProcessingService_ProcessIncidentsWithAnalysis_NilAnalyzers(t *testing.T) {
	// Create test database
	config := &database.Config{
		DatabasePath:    ":memory:",
		MaxOpenConns:    5,
		MaxIdleConns:    2,
		ConnMaxLifetime: time.Minute,
		ConnMaxIdleTime: time.Second * 30,
	}

	dbWrapper, err := database.NewDB(config)
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer dbWrapper.Close()

	if err := dbWrapper.InitializeDatabase(); err != nil {
		t.Fatalf("Failed to initialize database: %v", err)
	}

	db := dbWrapper.GetConnection()

	// Create test file store
	tempDir := t.TempDir()
	fileStore := storage.NewFileStore(tempDir)

	// Create processing service with nil analyzers
	service := &ProcessingService{
		db:                 db,
		fileStore:          fileStore,
		excelParser:        NewExcelParser(),
		incidentService:    NewIncidentService(db),
		sentimentAnalyzer:  nil, // Nil analyzer
		automationAnalyzer: nil, // Nil analyzer
	}

	// Create test incidents
	incidents := []models.Incident{
		{
			ID:               "test-1",
			UploadID:         "upload-1",
			IncidentID:       "INC001",
			ReportDate:       time.Now().Add(-2 * time.Hour),
			ResolveDate:      func() *time.Time { t := time.Now(); return &t }(),
			BriefDescription: "Test incident",
			ApplicationName:  "Test App",
			ResolutionGroup:  "Test Team",
			ResolvedPerson:   "Test Person",
			Priority:         "P3",
		},
	}

	// Set defaults for incidents
	for i := range incidents {
		incidents[i].SetDefaults()
	}

	// Process incidents with nil analyzers - should not fail
	err = service.processIncidentsWithAnalysis(incidents)
	if err != nil {
		t.Fatalf("Processing should not fail with nil analyzers: %v", err)
	}

	// Verify resolution time was still calculated
	incident := incidents[0]
	if incident.ResolutionTimeHours == nil {
		t.Errorf("Expected resolution time to be calculated even with nil analyzers")
	}

	// Verify analysis fields are not set
	if incident.SentimentScore != nil {
		t.Errorf("Expected sentiment score to be nil with nil analyzer")
	}
	if incident.AutomationScore != nil {
		t.Errorf("Expected automation score to be nil with nil analyzer")
	}
}