package services

import (
	"context"
	"math"
	"testing"
	"time"

	"incident-management-system/internal/database"
	"incident-management-system/internal/models"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAnalyticsService_GetDailyTimeline(t *testing.T) {
	// Setup test database
	dbConfig := &database.Config{
		DatabasePath: ":memory:",
	}
	db, err := database.NewDB(dbConfig)
	require.NoError(t, err)
	defer db.Close()

	// Initialize database schema
	err = db.InitializeDatabase()
	require.NoError(t, err)

	analyticsService := NewAnalyticsService(db.GetConnection())

	// Create test data
	uploadID := uuid.New().String()
	testIncidents := []models.Incident{
		{
			ID:               uuid.New().String(),
			UploadID:         uploadID,
			IncidentID:       "INC001",
			ReportDate:       time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			BriefDescription: "Test incident 1",
			ApplicationName:  "App1",
			ResolutionGroup:  "Group1",
			ResolvedPerson:   "Person1",
			Priority:         "P1",
		},
		{
			ID:               uuid.New().String(),
			UploadID:         uploadID,
			IncidentID:       "INC002",
			ReportDate:       time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			BriefDescription: "Test incident 2",
			ApplicationName:  "App2",
			ResolutionGroup:  "Group2",
			ResolvedPerson:   "Person2",
			Priority:         "P2",
		},
		{
			ID:               uuid.New().String(),
			UploadID:         uploadID,
			IncidentID:       "INC003",
			ReportDate:       time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC),
			BriefDescription: "Test incident 3",
			ApplicationName:  "App1",
			ResolutionGroup:  "Group1",
			ResolvedPerson:   "Person1",
			Priority:         "P1",
		},
	}

	// Insert test data
	for _, incident := range testIncidents {
		incident.SetDefaults()
		query := `
			INSERT INTO incidents (
				id, upload_id, incident_id, report_date, brief_description,
				application_name, resolution_group, resolved_person, priority,
				created_at, updated_at
			) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		`
		_, err := db.GetConnection().Exec(query,
			incident.ID, incident.UploadID, incident.IncidentID, incident.ReportDate,
			incident.BriefDescription, incident.ApplicationName, incident.ResolutionGroup,
			incident.ResolvedPerson, incident.Priority, incident.CreatedAt, incident.UpdatedAt,
		)
		require.NoError(t, err)
	}

	// Test GetDailyTimeline without filters
	timeline, err := analyticsService.GetDailyTimeline(context.Background(), nil)
	require.NoError(t, err)
	assert.Len(t, timeline, 2) // Two different days

	// Check first day data
	day1 := timeline[0]
	assert.Equal(t, "2024-01-01", day1.Date)
	assert.Equal(t, 2, day1.IncidentCount)
	assert.Equal(t, 1, day1.P1Count)
	assert.Equal(t, 1, day1.P2Count)
	assert.Equal(t, 0, day1.P3Count)
	assert.Equal(t, 0, day1.P4Count)

	// Check second day data
	day2 := timeline[1]
	assert.Equal(t, "2024-01-02", day2.Date)
	assert.Equal(t, 1, day2.IncidentCount)
	assert.Equal(t, 1, day2.P1Count)
	assert.Equal(t, 0, day2.P2Count)

	// Test with date filters
	startDate := time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC)
	filters := &TimelineFilters{
		StartDate: &startDate,
	}
	
	filteredTimeline, err := analyticsService.GetDailyTimeline(context.Background(), filters)
	require.NoError(t, err)
	assert.Len(t, filteredTimeline, 1)
	assert.Equal(t, "2024-01-02", filteredTimeline[0].Date)

	// Test with priority filters
	priorityFilters := &TimelineFilters{
		Priorities: []string{"P1"},
	}
	
	priorityTimeline, err := analyticsService.GetDailyTimeline(context.Background(), priorityFilters)
	require.NoError(t, err)
	assert.Len(t, priorityTimeline, 2)
	
	// Both days should have P1 incidents
	for _, day := range priorityTimeline {
		assert.Equal(t, day.P1Count, day.IncidentCount) // Only P1 incidents
		assert.Equal(t, 0, day.P2Count)
		assert.Equal(t, 0, day.P3Count)
		assert.Equal(t, 0, day.P4Count)
	}
}

func TestAnalyticsService_GetWeeklyTimeline(t *testing.T) {
	// Setup test database
	dbConfig := &database.Config{
		DatabasePath: ":memory:",
	}
	db, err := database.NewDB(dbConfig)
	require.NoError(t, err)
	defer db.Close()

	// Initialize database schema
	err = db.InitializeDatabase()
	require.NoError(t, err)

	analyticsService := NewAnalyticsService(db.GetConnection())

	// Create test data spanning multiple weeks
	uploadID := uuid.New().String()
	testIncidents := []models.Incident{
		{
			ID:               uuid.New().String(),
			UploadID:         uploadID,
			IncidentID:       "INC001",
			ReportDate:       time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC), // Week 1
			BriefDescription: "Test incident 1",
			ApplicationName:  "App1",
			ResolutionGroup:  "Group1",
			ResolvedPerson:   "Person1",
			Priority:         "P1",
		},
		{
			ID:               uuid.New().String(),
			UploadID:         uploadID,
			IncidentID:       "INC002",
			ReportDate:       time.Date(2024, 1, 8, 0, 0, 0, 0, time.UTC), // Week 2
			BriefDescription: "Test incident 2",
			ApplicationName:  "App2",
			ResolutionGroup:  "Group2",
			ResolvedPerson:   "Person2",
			Priority:         "P2",
		},
	}

	// Insert test data
	for _, incident := range testIncidents {
		incident.SetDefaults()
		query := `
			INSERT INTO incidents (
				id, upload_id, incident_id, report_date, brief_description,
				application_name, resolution_group, resolved_person, priority,
				created_at, updated_at
			) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		`
		_, err := db.GetConnection().Exec(query,
			incident.ID, incident.UploadID, incident.IncidentID, incident.ReportDate,
			incident.BriefDescription, incident.ApplicationName, incident.ResolutionGroup,
			incident.ResolvedPerson, incident.Priority, incident.CreatedAt, incident.UpdatedAt,
		)
		require.NoError(t, err)
	}

	// Test GetWeeklyTimeline
	timeline, err := analyticsService.GetWeeklyTimeline(context.Background(), nil)
	require.NoError(t, err)
	assert.Len(t, timeline, 2) // Two different weeks

	// Each week should have one incident
	for _, week := range timeline {
		assert.Equal(t, 1, week.IncidentCount)
	}
}

func TestAnalyticsService_GetTrendAnalysis(t *testing.T) {
	// Setup test database
	dbConfig := &database.Config{
		DatabasePath: ":memory:",
	}
	db, err := database.NewDB(dbConfig)
	require.NoError(t, err)
	defer db.Close()

	// Initialize database schema
	err = db.InitializeDatabase()
	require.NoError(t, err)

	analyticsService := NewAnalyticsService(db.GetConnection())

	// Create test data with increasing trend
	uploadID := uuid.New().String()
	testIncidents := []models.Incident{
		{
			ID:               uuid.New().String(),
			UploadID:         uploadID,
			IncidentID:       "INC001",
			ReportDate:       time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			BriefDescription: "Test incident 1",
			ApplicationName:  "App1",
			ResolutionGroup:  "Group1",
			ResolvedPerson:   "Person1",
			Priority:         "P1",
		},
		{
			ID:               uuid.New().String(),
			UploadID:         uploadID,
			IncidentID:       "INC002",
			ReportDate:       time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC),
			BriefDescription: "Test incident 2",
			ApplicationName:  "App2",
			ResolutionGroup:  "Group2",
			ResolvedPerson:   "Person2",
			Priority:         "P2",
		},
		{
			ID:               uuid.New().String(),
			UploadID:         uploadID,
			IncidentID:       "INC003",
			ReportDate:       time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC),
			BriefDescription: "Test incident 3",
			ApplicationName:  "App1",
			ResolutionGroup:  "Group1",
			ResolvedPerson:   "Person1",
			Priority:         "P1",
		},
		{
			ID:               uuid.New().String(),
			UploadID:         uploadID,
			IncidentID:       "INC004",
			ReportDate:       time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC),
			BriefDescription: "Test incident 4",
			ApplicationName:  "App2",
			ResolutionGroup:  "Group2",
			ResolvedPerson:   "Person2",
			Priority:         "P3",
		},
	}

	// Insert test data
	for _, incident := range testIncidents {
		incident.SetDefaults()
		query := `
			INSERT INTO incidents (
				id, upload_id, incident_id, report_date, brief_description,
				application_name, resolution_group, resolved_person, priority,
				created_at, updated_at
			) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		`
		_, err := db.GetConnection().Exec(query,
			incident.ID, incident.UploadID, incident.IncidentID, incident.ReportDate,
			incident.BriefDescription, incident.ApplicationName, incident.ResolutionGroup,
			incident.ResolvedPerson, incident.Priority, incident.CreatedAt, incident.UpdatedAt,
		)
		require.NoError(t, err)
	}

	// Test GetTrendAnalysis
	trends, err := analyticsService.GetTrendAnalysis(context.Background(), "daily", nil)
	require.NoError(t, err)
	assert.Len(t, trends, 1) // One trend point (day 2 compared to day 1)

	trend := trends[0]
	assert.Equal(t, "2024-01-02", trend.Period)
	assert.Equal(t, 3, trend.IncidentCount)
	assert.Equal(t, 200.0, trend.GrowthRate) // 200% increase from 1 to 3
	assert.Equal(t, "increasing", trend.Trend)
}

func TestAnalyticsService_GetTicketsPerDayMetrics(t *testing.T) {
	// Setup test database
	dbConfig := &database.Config{
		DatabasePath: ":memory:",
	}
	db, err := database.NewDB(dbConfig)
	require.NoError(t, err)
	defer db.Close()

	// Initialize database schema
	err = db.InitializeDatabase()
	require.NoError(t, err)

	analyticsService := NewAnalyticsService(db.GetConnection())

	// Create test data
	uploadID := uuid.New().String()
	testIncidents := []models.Incident{
		{
			ID:               uuid.New().String(),
			UploadID:         uploadID,
			IncidentID:       "INC001",
			ReportDate:       time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			BriefDescription: "Test incident 1",
			ApplicationName:  "App1",
			ResolutionGroup:  "Group1",
			ResolvedPerson:   "Person1",
			Priority:         "P1",
		},
		{
			ID:               uuid.New().String(),
			UploadID:         uploadID,
			IncidentID:       "INC002",
			ReportDate:       time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC),
			BriefDescription: "Test incident 2",
			ApplicationName:  "App2",
			ResolutionGroup:  "Group2",
			ResolvedPerson:   "Person2",
			Priority:         "P2",
		},
		{
			ID:               uuid.New().String(),
			UploadID:         uploadID,
			IncidentID:       "INC003",
			ReportDate:       time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC),
			BriefDescription: "Test incident 3",
			ApplicationName:  "App1",
			ResolutionGroup:  "Group1",
			ResolvedPerson:   "Person1",
			Priority:         "P1",
		},
	}

	// Insert test data
	for _, incident := range testIncidents {
		incident.SetDefaults()
		query := `
			INSERT INTO incidents (
				id, upload_id, incident_id, report_date, brief_description,
				application_name, resolution_group, resolved_person, priority,
				created_at, updated_at
			) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		`
		_, err := db.GetConnection().Exec(query,
			incident.ID, incident.UploadID, incident.IncidentID, incident.ReportDate,
			incident.BriefDescription, incident.ApplicationName, incident.ResolutionGroup,
			incident.ResolvedPerson, incident.Priority, incident.CreatedAt, incident.UpdatedAt,
		)
		require.NoError(t, err)
	}

	// Test GetTicketsPerDayMetrics
	metrics, err := analyticsService.GetTicketsPerDayMetrics(context.Background(), nil)
	require.NoError(t, err)

	assert.Equal(t, 3, metrics["total_incidents"])
	assert.Equal(t, 1.5, metrics["avg_per_day"])  // (1 + 2) / 2 = 1.5
	assert.Equal(t, 2.0, metrics["max_per_day"])  // Max incidents in a day
	assert.Equal(t, 1.0, metrics["min_per_day"])  // Min incidents in a day
	assert.Equal(t, 1.5, metrics["median_per_day"]) // Median of [1, 2]
}

func TestAnalyticsService_GetPriorityAnalysis(t *testing.T) {
	// Setup test database
	dbConfig := &database.Config{
		DatabasePath: ":memory:",
	}
	db, err := database.NewDB(dbConfig)
	require.NoError(t, err)
	defer db.Close()

	// Initialize database schema
	err = db.InitializeDatabase()
	require.NoError(t, err)

	analyticsService := NewAnalyticsService(db.GetConnection())

	// Create test data with different priorities
	uploadID := uuid.New().String()
	testIncidents := []models.Incident{
		{
			ID:               uuid.New().String(),
			UploadID:         uploadID,
			IncidentID:       "INC001",
			ReportDate:       time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			BriefDescription: "Test incident 1",
			ApplicationName:  "App1",
			ResolutionGroup:  "Group1",
			ResolvedPerson:   "Person1",
			Priority:         "P1",
		},
		{
			ID:               uuid.New().String(),
			UploadID:         uploadID,
			IncidentID:       "INC002",
			ReportDate:       time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			BriefDescription: "Test incident 2",
			ApplicationName:  "App2",
			ResolutionGroup:  "Group2",
			ResolvedPerson:   "Person2",
			Priority:         "P1",
		},
		{
			ID:               uuid.New().String(),
			UploadID:         uploadID,
			IncidentID:       "INC003",
			ReportDate:       time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC),
			BriefDescription: "Test incident 3",
			ApplicationName:  "App1",
			ResolutionGroup:  "Group1",
			ResolvedPerson:   "Person1",
			Priority:         "P2",
		},
	}

	// Insert test data
	for _, incident := range testIncidents {
		incident.SetDefaults()
		query := `
			INSERT INTO incidents (
				id, upload_id, incident_id, report_date, brief_description,
				application_name, resolution_group, resolved_person, priority,
				created_at, updated_at
			) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		`
		_, err := db.GetConnection().Exec(query,
			incident.ID, incident.UploadID, incident.IncidentID, incident.ReportDate,
			incident.BriefDescription, incident.ApplicationName, incident.ResolutionGroup,
			incident.ResolvedPerson, incident.Priority, incident.CreatedAt, incident.UpdatedAt,
		)
		require.NoError(t, err)
	}

	// Test GetPriorityAnalysis
	analysis, err := analyticsService.GetPriorityAnalysis(context.Background(), nil)
	require.NoError(t, err)
	assert.Len(t, analysis, 2) // P1 and P2

	// Check P1 analysis
	var p1Analysis, p2Analysis *PriorityAnalysis
	for i := range analysis {
		if analysis[i].Priority == "P1" {
			p1Analysis = &analysis[i]
		} else if analysis[i].Priority == "P2" {
			p2Analysis = &analysis[i]
		}
	}

	require.NotNil(t, p1Analysis)
	assert.Equal(t, "P1", p1Analysis.Priority)
	assert.Equal(t, 2, p1Analysis.Count)
	assert.Equal(t, 66.67, p1Analysis.Percentage) // 2/3 * 100 = 66.67

	require.NotNil(t, p2Analysis)
	assert.Equal(t, "P2", p2Analysis.Priority)
	assert.Equal(t, 1, p2Analysis.Count)
	assert.Equal(t, 33.33, p2Analysis.Percentage) // 1/3 * 100 = 33.33
}

func TestAnalyticsService_GetApplicationAnalysis(t *testing.T) {
	// Setup test database
	dbConfig := &database.Config{
		DatabasePath: ":memory:",
	}
	db, err := database.NewDB(dbConfig)
	require.NoError(t, err)
	defer db.Close()

	// Initialize database schema
	err = db.InitializeDatabase()
	require.NoError(t, err)

	analyticsService := NewAnalyticsService(db.GetConnection())

	// Create test data with different applications and resolution times
	uploadID := uuid.New().String()
	resolveDate1 := time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC)
	resolveDate2 := time.Date(2024, 1, 3, 0, 0, 0, 0, time.UTC)
	
	testIncidents := []models.Incident{
		{
			ID:               uuid.New().String(),
			UploadID:         uploadID,
			IncidentID:       "INC001",
			ReportDate:       time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			ResolveDate:      &resolveDate1,
			BriefDescription: "Test incident 1",
			ApplicationName:  "App1",
			ResolutionGroup:  "Group1",
			ResolvedPerson:   "Person1",
			Priority:         "P1",
		},
		{
			ID:               uuid.New().String(),
			UploadID:         uploadID,
			IncidentID:       "INC002",
			ReportDate:       time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			ResolveDate:      &resolveDate2,
			BriefDescription: "Test incident 2",
			ApplicationName:  "App1",
			ResolutionGroup:  "Group1",
			ResolvedPerson:   "Person1",
			Priority:         "P2",
		},
		{
			ID:               uuid.New().String(),
			UploadID:         uploadID,
			IncidentID:       "INC003",
			ReportDate:       time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC),
			BriefDescription: "Test incident 3",
			ApplicationName:  "App2",
			ResolutionGroup:  "Group2",
			ResolvedPerson:   "Person2",
			Priority:         "P1",
		},
	}

	// Insert test data and calculate resolution times
	for _, incident := range testIncidents {
		incident.SetDefaults()
		incident.CalculateResolutionTime()
		
		query := `
			INSERT INTO incidents (
				id, upload_id, incident_id, report_date, resolve_date, brief_description,
				application_name, resolution_group, resolved_person, priority,
				resolution_time_hours, created_at, updated_at
			) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		`
		_, err := db.GetConnection().Exec(query,
			incident.ID, incident.UploadID, incident.IncidentID, incident.ReportDate,
			incident.ResolveDate, incident.BriefDescription, incident.ApplicationName,
			incident.ResolutionGroup, incident.ResolvedPerson, incident.Priority,
			incident.ResolutionTimeHours, incident.CreatedAt, incident.UpdatedAt,
		)
		require.NoError(t, err)
	}

	// Test GetApplicationAnalysis
	analysis, err := analyticsService.GetApplicationAnalysis(context.Background(), nil)
	require.NoError(t, err)
	assert.Len(t, analysis, 2) // App1 and App2

	// Check App1 analysis (should be first due to higher incident count)
	app1Analysis := analysis[0]
	assert.Equal(t, "App1", app1Analysis.ApplicationName)
	assert.Equal(t, 2, app1Analysis.IncidentCount)
	assert.Equal(t, 2, app1Analysis.ResolvedIncidents)
	assert.Greater(t, app1Analysis.AvgResolutionTime, 0.0)

	// Check App2 analysis
	app2Analysis := analysis[1]
	assert.Equal(t, "App2", app2Analysis.ApplicationName)
	assert.Equal(t, 1, app2Analysis.IncidentCount)
	assert.Equal(t, 0, app2Analysis.ResolvedIncidents) // No resolve date set
}

func TestAnalyticsService_GetResolutionAnalysis(t *testing.T) {
	// Setup test database
	dbConfig := &database.Config{
		DatabasePath: ":memory:",
	}
	db, err := database.NewDB(dbConfig)
	require.NoError(t, err)
	defer db.Close()

	// Initialize database schema
	err = db.InitializeDatabase()
	require.NoError(t, err)

	analyticsService := NewAnalyticsService(db.GetConnection())

	// Create test data with resolution times
	uploadID := uuid.New().String()
	resolveDate1 := time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC)
	resolveDate2 := time.Date(2024, 1, 3, 0, 0, 0, 0, time.UTC)
	
	testIncidents := []models.Incident{
		{
			ID:               uuid.New().String(),
			UploadID:         uploadID,
			IncidentID:       "INC001",
			ReportDate:       time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			ResolveDate:      &resolveDate1,
			BriefDescription: "Test incident 1",
			ApplicationName:  "App1",
			ResolutionGroup:  "Group1",
			ResolvedPerson:   "Person1",
			Priority:         "P1",
		},
		{
			ID:               uuid.New().String(),
			UploadID:         uploadID,
			IncidentID:       "INC002",
			ReportDate:       time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			ResolveDate:      &resolveDate2,
			BriefDescription: "Test incident 2",
			ApplicationName:  "App2",
			ResolutionGroup:  "Group2",
			ResolvedPerson:   "Person2",
			Priority:         "P2",
		},
		{
			ID:               uuid.New().String(),
			UploadID:         uploadID,
			IncidentID:       "INC003",
			ReportDate:       time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC),
			BriefDescription: "Test incident 3",
			ApplicationName:  "App1",
			ResolutionGroup:  "Group1",
			ResolvedPerson:   "Person1",
			Priority:         "P1",
		},
	}

	// Insert test data and calculate resolution times
	for _, incident := range testIncidents {
		incident.SetDefaults()
		incident.CalculateResolutionTime()
		
		query := `
			INSERT INTO incidents (
				id, upload_id, incident_id, report_date, resolve_date, brief_description,
				application_name, resolution_group, resolved_person, priority,
				resolution_time_hours, created_at, updated_at
			) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		`
		_, err := db.GetConnection().Exec(query,
			incident.ID, incident.UploadID, incident.IncidentID, incident.ReportDate,
			incident.ResolveDate, incident.BriefDescription, incident.ApplicationName,
			incident.ResolutionGroup, incident.ResolvedPerson, incident.Priority,
			incident.ResolutionTimeHours, incident.CreatedAt, incident.UpdatedAt,
		)
		require.NoError(t, err)
	}

	// Test GetResolutionAnalysis
	metrics, err := analyticsService.GetResolutionAnalysis(context.Background(), nil)
	require.NoError(t, err)

	assert.Equal(t, 3, metrics.TotalIncidents)
	assert.Equal(t, 2, metrics.ResolvedIncidents)
	assert.Equal(t, 66.67, math.Round(metrics.ResolutionRate*100)/100) // 2/3 * 100 = 66.67
	assert.Greater(t, metrics.AvgResolutionTime, 0.0)
	assert.Greater(t, metrics.MedianResolutionTime, 0.0)
}

func TestAnalyticsService_GetSentimentAnalysis(t *testing.T) {
	// Setup test database
	dbConfig := &database.Config{
		DatabasePath: ":memory:",
	}
	db, err := database.NewDB(dbConfig)
	require.NoError(t, err)
	defer db.Close()

	// Initialize database schema
	err = db.InitializeDatabase()
	require.NoError(t, err)

	analyticsService := NewAnalyticsService(db.GetConnection())

	// Create test data with sentiment analysis
	uploadID := uuid.New().String()
	sentimentScore1 := 0.8
	sentimentScore2 := -0.5
	sentimentScore3 := 0.1
	
	testIncidents := []models.Incident{
		{
			ID:               uuid.New().String(),
			UploadID:         uploadID,
			IncidentID:       "INC001",
			ReportDate:       time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			BriefDescription: "Test incident 1",
			ApplicationName:  "App1",
			ResolutionGroup:  "Group1",
			ResolvedPerson:   "Person1",
			Priority:         "P1",
			SentimentScore:   &sentimentScore1,
			SentimentLabel:   "positive",
		},
		{
			ID:               uuid.New().String(),
			UploadID:         uploadID,
			IncidentID:       "INC002",
			ReportDate:       time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			BriefDescription: "Test incident 2",
			ApplicationName:  "App2",
			ResolutionGroup:  "Group2",
			ResolvedPerson:   "Person2",
			Priority:         "P2",
			SentimentScore:   &sentimentScore2,
			SentimentLabel:   "negative",
		},
		{
			ID:               uuid.New().String(),
			UploadID:         uploadID,
			IncidentID:       "INC003",
			ReportDate:       time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC),
			BriefDescription: "Test incident 3",
			ApplicationName:  "App1",
			ResolutionGroup:  "Group1",
			ResolvedPerson:   "Person1",
			Priority:         "P1",
			SentimentScore:   &sentimentScore3,
			SentimentLabel:   "neutral",
		},
	}

	// Insert test data
	for _, incident := range testIncidents {
		incident.SetDefaults()
		query := `
			INSERT INTO incidents (
				id, upload_id, incident_id, report_date, brief_description,
				application_name, resolution_group, resolved_person, priority,
				sentiment_score, sentiment_label, created_at, updated_at
			) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		`
		_, err := db.GetConnection().Exec(query,
			incident.ID, incident.UploadID, incident.IncidentID, incident.ReportDate,
			incident.BriefDescription, incident.ApplicationName, incident.ResolutionGroup,
			incident.ResolvedPerson, incident.Priority, incident.SentimentScore,
			incident.SentimentLabel, incident.CreatedAt, incident.UpdatedAt,
		)
		require.NoError(t, err)
	}

	// Test GetSentimentAnalysis
	analysis, err := analyticsService.GetSentimentAnalysis(context.Background(), nil)
	require.NoError(t, err)
	assert.Len(t, analysis, 3) // positive, negative, neutral

	// Check that we have all sentiment types
	sentimentMap := make(map[string]SentimentAnalysis)
	for _, sentiment := range analysis {
		sentimentMap[sentiment.SentimentLabel] = sentiment
	}

	assert.Contains(t, sentimentMap, "positive")
	assert.Contains(t, sentimentMap, "negative")
	assert.Contains(t, sentimentMap, "neutral")

	// Check positive sentiment
	positive := sentimentMap["positive"]
	assert.Equal(t, 1, positive.Count)
	assert.Equal(t, 0.8, positive.AvgScore)

	// Check negative sentiment
	negative := sentimentMap["negative"]
	assert.Equal(t, 1, negative.Count)
	assert.Equal(t, -0.5, negative.AvgScore)

	// Check neutral sentiment
	neutral := sentimentMap["neutral"]
	assert.Equal(t, 1, neutral.Count)
	assert.Equal(t, 0.1, neutral.AvgScore)
}

func TestAnalyticsService_GetAutomationAnalysis(t *testing.T) {
	// Setup test database
	dbConfig := &database.Config{
		DatabasePath: ":memory:",
	}
	db, err := database.NewDB(dbConfig)
	require.NoError(t, err)
	defer db.Close()

	// Initialize database schema
	err = db.InitializeDatabase()
	require.NoError(t, err)

	analyticsService := NewAnalyticsService(db.GetConnection())

	// Create test data with automation analysis
	uploadID := uuid.New().String()
	automationScore1 := 0.9
	automationScore2 := 0.3
	automationScore3 := 0.7
	automationFeasible1 := true
	automationFeasible2 := false
	automationFeasible3 := true
	
	testIncidents := []models.Incident{
		{
			ID:                 uuid.New().String(),
			UploadID:           uploadID,
			IncidentID:         "INC001",
			ReportDate:         time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			BriefDescription:   "Test incident 1",
			ApplicationName:    "App1",
			ResolutionGroup:    "Group1",
			ResolvedPerson:     "Person1",
			Priority:           "P1",
			AutomationScore:    &automationScore1,
			AutomationFeasible: &automationFeasible1,
			ITProcessGroup:     "Infrastructure",
		},
		{
			ID:                 uuid.New().String(),
			UploadID:           uploadID,
			IncidentID:         "INC002",
			ReportDate:         time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			BriefDescription:   "Test incident 2",
			ApplicationName:    "App2",
			ResolutionGroup:    "Group2",
			ResolvedPerson:     "Person2",
			Priority:           "P2",
			AutomationScore:    &automationScore2,
			AutomationFeasible: &automationFeasible2,
			ITProcessGroup:     "Application Support",
		},
		{
			ID:                 uuid.New().String(),
			UploadID:           uploadID,
			IncidentID:         "INC003",
			ReportDate:         time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC),
			BriefDescription:   "Test incident 3",
			ApplicationName:    "App1",
			ResolutionGroup:    "Group1",
			ResolvedPerson:     "Person1",
			Priority:           "P1",
			AutomationScore:    &automationScore3,
			AutomationFeasible: &automationFeasible3,
			ITProcessGroup:     "Infrastructure",
		},
	}

	// Insert test data
	for _, incident := range testIncidents {
		incident.SetDefaults()
		query := `
			INSERT INTO incidents (
				id, upload_id, incident_id, report_date, brief_description,
				application_name, resolution_group, resolved_person, priority,
				automation_score, automation_feasible, it_process_group,
				created_at, updated_at
			) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		`
		_, err := db.GetConnection().Exec(query,
			incident.ID, incident.UploadID, incident.IncidentID, incident.ReportDate,
			incident.BriefDescription, incident.ApplicationName, incident.ResolutionGroup,
			incident.ResolvedPerson, incident.Priority, incident.AutomationScore,
			incident.AutomationFeasible, incident.ITProcessGroup, incident.CreatedAt, incident.UpdatedAt,
		)
		require.NoError(t, err)
	}

	// Test GetAutomationAnalysis
	analysis, err := analyticsService.GetAutomationAnalysis(context.Background(), nil)
	require.NoError(t, err)
	assert.Len(t, analysis, 2) // Infrastructure and Application Support

	// Check Infrastructure group (should be first due to higher automation percentage)
	infrastructure := analysis[0]
	assert.Equal(t, "Infrastructure", infrastructure.ITProcessGroup)
	assert.Equal(t, 2, infrastructure.IncidentCount)
	assert.Equal(t, 2, infrastructure.AutomatableCount)
	assert.Equal(t, 100.0, infrastructure.AutomationPercentage)
	assert.InDelta(t, 0.8, infrastructure.AvgAutomationScore, 0.01) // (0.9 + 0.7) / 2

	// Check Application Support group
	appSupport := analysis[1]
	assert.Equal(t, "Application Support", appSupport.ITProcessGroup)
	assert.Equal(t, 1, appSupport.IncidentCount)
	assert.Equal(t, 0, appSupport.AutomatableCount)
	assert.Equal(t, 0.0, appSupport.AutomationPercentage)
	assert.InDelta(t, 0.3, appSupport.AvgAutomationScore, 0.01)
}

func TestAnalyticsService_GetAnalyticsSummary(t *testing.T) {
	// Setup test database
	dbConfig := &database.Config{
		DatabasePath: ":memory:",
	}
	db, err := database.NewDB(dbConfig)
	require.NoError(t, err)
	defer db.Close()

	// Initialize database schema
	err = db.InitializeDatabase()
	require.NoError(t, err)

	analyticsService := NewAnalyticsService(db.GetConnection())

	// Create comprehensive test data
	uploadID := uuid.New().String()
	resolveDate := time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC)
	sentimentScore := 0.5
	automationScore := 0.8
	automationFeasible := true
	
	testIncident := models.Incident{
		ID:                 uuid.New().String(),
		UploadID:           uploadID,
		IncidentID:         "INC001",
		ReportDate:         time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		ResolveDate:        &resolveDate,
		BriefDescription:   "Test incident 1",
		ApplicationName:    "App1",
		ResolutionGroup:    "Group1",
		ResolvedPerson:     "Person1",
		Priority:           "P1",
		SentimentScore:     &sentimentScore,
		SentimentLabel:     "positive",
		AutomationScore:    &automationScore,
		AutomationFeasible: &automationFeasible,
		ITProcessGroup:     "Infrastructure",
	}

	// Insert test data
	testIncident.SetDefaults()
	testIncident.CalculateResolutionTime()
	
	query := `
		INSERT INTO incidents (
			id, upload_id, incident_id, report_date, resolve_date, brief_description,
			application_name, resolution_group, resolved_person, priority,
			sentiment_score, sentiment_label, automation_score, automation_feasible,
			it_process_group, resolution_time_hours, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	_, err = db.GetConnection().Exec(query,
		testIncident.ID, testIncident.UploadID, testIncident.IncidentID, testIncident.ReportDate,
		testIncident.ResolveDate, testIncident.BriefDescription, testIncident.ApplicationName,
		testIncident.ResolutionGroup, testIncident.ResolvedPerson, testIncident.Priority,
		testIncident.SentimentScore, testIncident.SentimentLabel, testIncident.AutomationScore,
		testIncident.AutomationFeasible, testIncident.ITProcessGroup, testIncident.ResolutionTimeHours,
		testIncident.CreatedAt, testIncident.UpdatedAt,
	)
	require.NoError(t, err)

	// Test GetAnalyticsSummary
	summary, err := analyticsService.GetAnalyticsSummary(context.Background(), nil)
	require.NoError(t, err)

	assert.Equal(t, 1, summary.TotalIncidents)
	assert.Equal(t, 1, summary.ResolvedIncidents)
	assert.Equal(t, 100.0, summary.ResolutionRate)
	assert.Greater(t, summary.AvgResolutionTime, 0.0)

	assert.Len(t, summary.PriorityBreakdown, 1)
	assert.Equal(t, "P1", summary.PriorityBreakdown[0].Priority)

	assert.Len(t, summary.SentimentBreakdown, 1)
	assert.Equal(t, "positive", summary.SentimentBreakdown[0].SentimentLabel)

	assert.Len(t, summary.AutomationSummary, 1)
	assert.Equal(t, "Infrastructure", summary.AutomationSummary[0].ITProcessGroup)

	assert.Len(t, summary.TopApplications, 1)
	assert.Equal(t, "App1", summary.TopApplications[0].ApplicationName)
}