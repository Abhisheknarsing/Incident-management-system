package handlers

import (
	"database/sql"
	"net/http"
	"strconv"
	"strings"
	"time"

	"incident-management-system/internal/services"

	"github.com/gin-gonic/gin"
)

// AnalyticsHandler handles analytics and reporting endpoints
type AnalyticsHandler struct {
	analyticsService *services.AnalyticsService
}

// NewAnalyticsHandler creates a new analytics handler
func NewAnalyticsHandler(db *sql.DB) *AnalyticsHandler {
	return &AnalyticsHandler{
		analyticsService: services.NewAnalyticsService(db),
	}
}

// APIError represents a standardized API error response
type APIError struct {
	Code      string      `json:"code"`
	Message   string      `json:"message"`
	Details   interface{} `json:"details,omitempty"`
	Timestamp time.Time   `json:"timestamp"`
	RequestID string      `json:"request_id"`
}

// parseTimelineFilters parses query parameters into TimelineFilters
func parseTimelineFilters(c *gin.Context) (*services.TimelineFilters, error) {
	filters := &services.TimelineFilters{}

	// Parse start_date
	if startDateStr := c.Query("start_date"); startDateStr != "" {
		startDate, err := time.Parse("2006-01-02", startDateStr)
		if err != nil {
			return nil, err
		}
		filters.StartDate = &startDate
	}

	// Parse end_date
	if endDateStr := c.Query("end_date"); endDateStr != "" {
		endDate, err := time.Parse("2006-01-02", endDateStr)
		if err != nil {
			return nil, err
		}
		filters.EndDate = &endDate
	}

	// Parse priorities
	if prioritiesStr := c.Query("priorities"); prioritiesStr != "" {
		filters.Priorities = strings.Split(prioritiesStr, ",")
	}

	// Parse applications
	if applicationsStr := c.Query("applications"); applicationsStr != "" {
		filters.Applications = strings.Split(applicationsStr, ",")
	}

	// Parse statuses
	if statusesStr := c.Query("statuses"); statusesStr != "" {
		filters.Statuses = strings.Split(statusesStr, ",")
	}

	return filters, nil
}

// sendError sends a standardized error response
func sendError(c *gin.Context, code string, message string, statusCode int, details interface{}) {
	apiError := APIError{
		Code:      code,
		Message:   message,
		Details:   details,
		Timestamp: time.Now(),
		RequestID: c.GetString("request_id"),
	}
	c.JSON(statusCode, apiError)
}

// GetDailyTimeline handles GET /api/analytics/timeline/daily
func (h *AnalyticsHandler) GetDailyTimeline(c *gin.Context) {
	filters, err := parseTimelineFilters(c)
	if err != nil {
		sendError(c, "INVALID_DATE_FORMAT", "Invalid date format. Use YYYY-MM-DD", http.StatusBadRequest, err.Error())
		return
	}

	timeline, err := h.analyticsService.GetDailyTimeline(c.Request.Context(), filters)
	if err != nil {
		sendError(c, "DATABASE_ERROR", "Failed to retrieve daily timeline", http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":    timeline,
		"filters": filters,
		"count":   len(timeline),
	})
}

// GetWeeklyTimeline handles GET /api/analytics/timeline/weekly
func (h *AnalyticsHandler) GetWeeklyTimeline(c *gin.Context) {
	filters, err := parseTimelineFilters(c)
	if err != nil {
		sendError(c, "INVALID_DATE_FORMAT", "Invalid date format. Use YYYY-MM-DD", http.StatusBadRequest, err.Error())
		return
	}

	timeline, err := h.analyticsService.GetWeeklyTimeline(c.Request.Context(), filters)
	if err != nil {
		sendError(c, "DATABASE_ERROR", "Failed to retrieve weekly timeline", http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":    timeline,
		"filters": filters,
		"count":   len(timeline),
	})
}

// GetTrendAnalysis handles GET /api/analytics/trends
func (h *AnalyticsHandler) GetTrendAnalysis(c *gin.Context) {
	period := c.DefaultQuery("period", "daily")
	if period != "daily" && period != "weekly" {
		sendError(c, "INVALID_PARAMETER", "Period must be 'daily' or 'weekly'", http.StatusBadRequest, nil)
		return
	}

	filters, err := parseTimelineFilters(c)
	if err != nil {
		sendError(c, "INVALID_DATE_FORMAT", "Invalid date format. Use YYYY-MM-DD", http.StatusBadRequest, err.Error())
		return
	}

	trends, err := h.analyticsService.GetTrendAnalysis(c.Request.Context(), period, filters)
	if err != nil {
		sendError(c, "DATABASE_ERROR", "Failed to retrieve trend analysis", http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":    trends,
		"period":  period,
		"filters": filters,
		"count":   len(trends),
	})
}

// GetTicketsPerDayMetrics handles GET /api/analytics/metrics/daily
func (h *AnalyticsHandler) GetTicketsPerDayMetrics(c *gin.Context) {
	filters, err := parseTimelineFilters(c)
	if err != nil {
		sendError(c, "INVALID_DATE_FORMAT", "Invalid date format. Use YYYY-MM-DD", http.StatusBadRequest, err.Error())
		return
	}

	metrics, err := h.analyticsService.GetTicketsPerDayMetrics(c.Request.Context(), filters)
	if err != nil {
		sendError(c, "DATABASE_ERROR", "Failed to retrieve daily metrics", http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":    metrics,
		"filters": filters,
	})
}

// GetTicketsPerWeekMetrics handles GET /api/analytics/metrics/weekly
func (h *AnalyticsHandler) GetTicketsPerWeekMetrics(c *gin.Context) {
	filters, err := parseTimelineFilters(c)
	if err != nil {
		sendError(c, "INVALID_DATE_FORMAT", "Invalid date format. Use YYYY-MM-DD", http.StatusBadRequest, err.Error())
		return
	}

	metrics, err := h.analyticsService.GetTicketsPerWeekMetrics(c.Request.Context(), filters)
	if err != nil {
		sendError(c, "DATABASE_ERROR", "Failed to retrieve weekly metrics", http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":    metrics,
		"filters": filters,
	})
}

// GetTimelineOverview handles GET /api/analytics/timeline/overview
func (h *AnalyticsHandler) GetTimelineOverview(c *gin.Context) {
	filters, err := parseTimelineFilters(c)
	if err != nil {
		sendError(c, "INVALID_DATE_FORMAT", "Invalid date format. Use YYYY-MM-DD", http.StatusBadRequest, err.Error())
		return
	}

	// Get both daily and weekly data
	dailyTimeline, err := h.analyticsService.GetDailyTimeline(c.Request.Context(), filters)
	if err != nil {
		sendError(c, "DATABASE_ERROR", "Failed to retrieve daily timeline", http.StatusInternalServerError, err.Error())
		return
	}

	weeklyTimeline, err := h.analyticsService.GetWeeklyTimeline(c.Request.Context(), filters)
	if err != nil {
		sendError(c, "DATABASE_ERROR", "Failed to retrieve weekly timeline", http.StatusInternalServerError, err.Error())
		return
	}

	// Get metrics
	dailyMetrics, err := h.analyticsService.GetTicketsPerDayMetrics(c.Request.Context(), filters)
	if err != nil {
		sendError(c, "DATABASE_ERROR", "Failed to retrieve daily metrics", http.StatusInternalServerError, err.Error())
		return
	}

	weeklyMetrics, err := h.analyticsService.GetTicketsPerWeekMetrics(c.Request.Context(), filters)
	if err != nil {
		sendError(c, "DATABASE_ERROR", "Failed to retrieve weekly metrics", http.StatusInternalServerError, err.Error())
		return
	}

	// Get trend analysis
	dailyTrends, err := h.analyticsService.GetTrendAnalysis(c.Request.Context(), "daily", filters)
	if err != nil {
		sendError(c, "DATABASE_ERROR", "Failed to retrieve daily trends", http.StatusInternalServerError, err.Error())
		return
	}

	weeklyTrends, err := h.analyticsService.GetTrendAnalysis(c.Request.Context(), "weekly", filters)
	if err != nil {
		sendError(c, "DATABASE_ERROR", "Failed to retrieve weekly trends", http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"daily": gin.H{
			"timeline": dailyTimeline,
			"metrics":  dailyMetrics,
			"trends":   dailyTrends,
		},
		"weekly": gin.H{
			"timeline": weeklyTimeline,
			"metrics":  weeklyMetrics,
			"trends":   weeklyTrends,
		},
		"filters": filters,
	})
}