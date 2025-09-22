package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"incident-management-system/internal/database"
	"incident-management-system/internal/models"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// createTestDBAnalytics creates a test database connection for analytics tests
func createTestDBAnalytics(t *testing.T) *sql.DB {
	config := &database.Config{
		DatabasePath: ":memory:",
	}

	dbWrapper, err := database.NewDB(config)
	require.NoError(t, err, "Failed to create test database")

	err = dbWrapper.InitializeDatabase()
	require.NoError(t, err, "Failed to initialize test database")

	t.Cleanup(func() {
		dbWrapper.Close()
	})

	return dbWrapper.GetConnection()
}

// createTestIncidents creates test incidents in the database
func createTestIncidents(t *testing.T, db *sql.DB, count int) {
	for i := 0; i < count; i++ {
		incident := models.Incident{
			ID:                 uuid.New().String(),
			UploadID:           "test-upload",
			IncidentID:         "INC" + uuid.New().String()[:8],
			ApplicationName:    "TestApp",
			ReportDate:         time.Now().Add(-time.Duration(i) * time.Hour),
			BriefDescription:   "Test incident " + string(rune(i+65)),
			Description:        "Test incident description",
			ResolutionGroup:    "TestGroup",
			ResolvedPerson:     "TestPerson",
			Priority:           "P3",
			Status:             "Closed",
			ResolutionNotes:    "Test resolution",
			SentimentLabel:     "positive",
			SentimentScore:     func() *float64 { s := 0.5; return &s }(),
			AutomationScore:    func() *float64 { s := 0.8; return &s }(),
			AutomationFeasible: func() *bool { b := true; return &b }(),
			ITProcessGroup:     "Infrastructure",
		}

		incident.SetDefaults()

		// Create a simple insert for testing
		query := `
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

		_, err := db.Exec(query,
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
			incident.SentimentLabel,
			incident.ResolutionTimeHours,
			incident.AutomationScore,
			incident.AutomationFeasible,
			incident.ITProcessGroup,
			incident.CreatedAt,
			incident.UpdatedAt,
		)
		require.NoError(t, err, "Failed to create test incident")
	}
}

func TestAnalyticsHandler_GetDailyTimeline(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	db := createTestDBAnalytics(t)
	createTestIncidents(t, db, 10)

	handler := NewAnalyticsHandler(db)

	// Create request
	req := httptest.NewRequest("GET", "/analytics/timeline/daily", nil)
	w := httptest.NewRecorder()

	// Create gin context
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	// Execute handler
	handler.GetDailyTimeline(c)

	// Check response
	assert.Equal(t, http.StatusOK, w.Code)
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	data, ok := response["data"].([]interface{})
	assert.True(t, ok, "Data should be an array")
	assert.Greater(t, len(data), 0, "Should return timeline data")
}

func TestAnalyticsHandler_GetWeeklyTimeline(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	db := createTestDBAnalytics(t)
	createTestIncidents(t, db, 10)

	handler := NewAnalyticsHandler(db)

	// Create request
	req := httptest.NewRequest("GET", "/analytics/timeline/weekly", nil)
	w := httptest.NewRecorder()

	// Create gin context
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	// Execute handler
	handler.GetWeeklyTimeline(c)

	// Check response
	assert.Equal(t, http.StatusOK, w.Code)
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	_, ok := response["data"].([]interface{})
	assert.True(t, ok, "Data should be an array")
	// Note: With only 10 incidents spanning a few hours, we might not have weekly data
	// This test is mainly to ensure the endpoint doesn't error
}

func TestAnalyticsHandler_GetTrendAnalysis(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	db := createTestDBAnalytics(t)
	createTestIncidents(t, db, 10)

	handler := NewAnalyticsHandler(db)

	// Test cases
	tests := []struct {
		name     string
		period   string
		hasError bool
	}{
		{
			name:     "daily trend analysis",
			period:   "daily",
			hasError: false,
		},
		{
			name:     "weekly trend analysis",
			period:   "weekly",
			hasError: false,
		},
		{
			name:     "invalid period",
			period:   "monthly",
			hasError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create request
			url := "/analytics/trends"
			if tt.period != "" {
				url = "/analytics/trends?period=" + tt.period
			}
			req := httptest.NewRequest("GET", url, nil)
			w := httptest.NewRecorder()

			// Create gin context
			c, _ := gin.CreateTestContext(w)
			c.Request = req

			// Execute handler
			handler.GetTrendAnalysis(c)

			// Check response
			if tt.hasError {
				assert.Equal(t, http.StatusBadRequest, w.Code)
			} else {
				assert.Equal(t, http.StatusOK, w.Code)
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)

				_, ok := response["data"].([]interface{})
				assert.True(t, ok, "Data should be an array")
				// Trend analysis might be empty with limited test data
			}
		})
	}
}

func TestAnalyticsHandler_GetTicketsPerDayMetrics(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	db := createTestDBAnalytics(t)
	createTestIncidents(t, db, 10)

	handler := NewAnalyticsHandler(db)

	// Create request
	req := httptest.NewRequest("GET", "/analytics/metrics/daily", nil)
	w := httptest.NewRecorder()

	// Create gin context
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	// Execute handler
	handler.GetTicketsPerDayMetrics(c)

	// Check response
	assert.Equal(t, http.StatusOK, w.Code)
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	_, ok := response["data"].(map[string]interface{})
	assert.True(t, ok, "Data should be an object")
	// Metrics might be empty with limited test data, but endpoint should not error
}

func TestAnalyticsHandler_GetPriorityAnalysis(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	db := createTestDBAnalytics(t)
	createTestIncidents(t, db, 10)

	handler := NewAnalyticsHandler(db)

	// Create request
	req := httptest.NewRequest("GET", "/analytics/priority", nil)
	w := httptest.NewRecorder()

	// Create gin context
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	// Execute handler
	handler.GetPriorityAnalysis(c)

	// Check response
	assert.Equal(t, http.StatusOK, w.Code)
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	_, ok := response["data"].([]interface{})
	assert.True(t, ok, "Data should be an array")
	// Priority analysis might be empty with limited test data, but endpoint should not error
}

func TestAnalyticsHandler_GetApplicationAnalysis(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	db := createTestDBAnalytics(t)
	createTestIncidents(t, db, 10)

	handler := NewAnalyticsHandler(db)

	// Create request
	req := httptest.NewRequest("GET", "/analytics/applications", nil)
	w := httptest.NewRecorder()

	// Create gin context
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	// Execute handler
	handler.GetApplicationAnalysis(c)

	// Check response
	assert.Equal(t, http.StatusOK, w.Code)
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	_, ok := response["data"].([]interface{})
	assert.True(t, ok, "Data should be an array")
	// Application analysis might be empty with limited test data, but endpoint should not error
}

func TestAnalyticsHandler_GetResolutionAnalysis(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	db := createTestDBAnalytics(t)
	createTestIncidents(t, db, 10)

	handler := NewAnalyticsHandler(db)

	// Create request
	req := httptest.NewRequest("GET", "/analytics/resolution", nil)
	w := httptest.NewRecorder()

	// Create gin context
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	// Execute handler
	handler.GetResolutionAnalysis(c)

	// Check response
	assert.Equal(t, http.StatusOK, w.Code)
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	_, ok := response["data"].(map[string]interface{})
	assert.True(t, ok, "Data should be an object")
	// Resolution analysis might be empty with limited test data, but endpoint should not error
}

func TestAnalyticsHandler_GetPerformanceMetrics(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	db := createTestDBAnalytics(t)
	createTestIncidents(t, db, 10)

	handler := NewAnalyticsHandler(db)

	// Create request
	req := httptest.NewRequest("GET", "/analytics/performance", nil)
	w := httptest.NewRecorder()

	// Create gin context
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	// Execute handler
	handler.GetPerformanceMetrics(c)

	// Check response
	assert.Equal(t, http.StatusOK, w.Code)
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	_, ok := response["data"].(map[string]interface{})
	assert.True(t, ok, "Data should be an object")
	// Performance metrics might be empty with limited test data, but endpoint should not error
}

func TestAnalyticsHandler_GetSentimentAnalysis(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	db := createTestDBAnalytics(t)
	createTestIncidents(t, db, 10)

	handler := NewAnalyticsHandler(db)

	// Create request
	req := httptest.NewRequest("GET", "/analytics/sentiment", nil)
	w := httptest.NewRecorder()

	// Create gin context
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	// Execute handler
	handler.GetSentimentAnalysis(c)

	// Check response
	assert.Equal(t, http.StatusOK, w.Code)
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	_, ok := response["data"].([]interface{})
	assert.True(t, ok, "Data should be an array")
	// Sentiment analysis might be empty with limited test data, but endpoint should not error
}

func TestAnalyticsHandler_GetAutomationAnalysis(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	db := createTestDBAnalytics(t)
	createTestIncidents(t, db, 10)

	handler := NewAnalyticsHandler(db)

	// Create request
	req := httptest.NewRequest("GET", "/analytics/automation", nil)
	w := httptest.NewRecorder()

	// Create gin context
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	// Execute handler
	handler.GetAutomationAnalysis(c)

	// Check response
	assert.Equal(t, http.StatusOK, w.Code)
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	_, ok := response["data"].([]interface{})
	assert.True(t, ok, "Data should be an array")
	// Automation analysis might be empty with limited test data, but endpoint should not error
}

func TestAnalyticsHandler_GetAnalyticsSummary(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	db := createTestDBAnalytics(t)
	createTestIncidents(t, db, 10)

	handler := NewAnalyticsHandler(db)

	// Create request
	req := httptest.NewRequest("GET", "/analytics/summary", nil)
	w := httptest.NewRecorder()

	// Create gin context
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	// Execute handler
	handler.GetAnalyticsSummary(c)

	// Check response
	assert.Equal(t, http.StatusOK, w.Code)
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	_, ok := response["data"].(map[string]interface{})
	assert.True(t, ok, "Data should be an object")
	// Summary should contain data even with limited test data
}
