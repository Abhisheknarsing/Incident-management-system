package services

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"
)

// AnalyticsService provides analytics and reporting functionality
type AnalyticsService struct {
	db *sql.DB
}

// NewAnalyticsService creates a new analytics service
func NewAnalyticsService(db *sql.DB) *AnalyticsService {
	return &AnalyticsService{
		db: db,
	}
}

// buildFilterConditions builds WHERE conditions and arguments for filters
func buildFilterConditions(filters *TimelineFilters, startArgIndex int) (string, []interface{}, int) {
	if filters == nil {
		return "", []interface{}{}, startArgIndex
	}

	var conditions []string
	var args []interface{}
	argIndex := startArgIndex

	if filters.StartDate != nil {
		conditions = append(conditions, fmt.Sprintf("report_date >= $%d", argIndex))
		args = append(args, *filters.StartDate)
		argIndex++
	}
	if filters.EndDate != nil {
		conditions = append(conditions, fmt.Sprintf("report_date <= $%d", argIndex))
		args = append(args, *filters.EndDate)
		argIndex++
	}
	if len(filters.Priorities) > 0 {
		placeholders := make([]string, len(filters.Priorities))
		for i, priority := range filters.Priorities {
			placeholders[i] = fmt.Sprintf("$%d", argIndex)
			args = append(args, priority)
			argIndex++
		}
		conditions = append(conditions, fmt.Sprintf("priority IN (%s)", strings.Join(placeholders, ",")))
	}
	if len(filters.Applications) > 0 {
		placeholders := make([]string, len(filters.Applications))
		for i, app := range filters.Applications {
			placeholders[i] = fmt.Sprintf("$%d", argIndex)
			args = append(args, app)
			argIndex++
		}
		conditions = append(conditions, fmt.Sprintf("application_name IN (%s)", strings.Join(placeholders, ",")))
	}
	if len(filters.Statuses) > 0 {
		placeholders := make([]string, len(filters.Statuses))
		for i, status := range filters.Statuses {
			placeholders[i] = fmt.Sprintf("$%d", argIndex)
			args = append(args, status)
			argIndex++
		}
		conditions = append(conditions, fmt.Sprintf("status IN (%s)", strings.Join(placeholders, ",")))
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = " AND " + strings.Join(conditions, " AND ")
	}

	return whereClause, args, argIndex
}

// TimelineData represents incident timeline data
type TimelineData struct {
	Date         string `json:"date"`
	IncidentCount int    `json:"incident_count"`
	P1Count      int    `json:"p1_count"`
	P2Count      int    `json:"p2_count"`
	P3Count      int    `json:"p3_count"`
	P4Count      int    `json:"p4_count"`
}

// TrendAnalysis represents trend analysis data
type TrendAnalysis struct {
	Period       string  `json:"period"`
	IncidentCount int     `json:"incident_count"`
	GrowthRate   float64 `json:"growth_rate"`
	Trend        string  `json:"trend"` // "increasing", "decreasing", "stable"
}

// PriorityAnalysis represents priority distribution analysis
type PriorityAnalysis struct {
	Priority   string  `json:"priority"`
	Count      int     `json:"count"`
	Percentage float64 `json:"percentage"`
}

// ApplicationAnalysis represents application-wise incident analysis
type ApplicationAnalysis struct {
	ApplicationName     string  `json:"application_name"`
	IncidentCount       int     `json:"incident_count"`
	AvgResolutionTime   float64 `json:"avg_resolution_time"`
	MedianResolutionTime float64 `json:"median_resolution_time"`
	ResolvedIncidents   int     `json:"resolved_incidents"`
	Trend               string  `json:"trend"`
}

// ResolutionMetrics represents resolution analysis metrics
type ResolutionMetrics struct {
	AvgResolutionTime    float64 `json:"avg_resolution_time"`
	MedianResolutionTime float64 `json:"median_resolution_time"`
	TotalIncidents       int     `json:"total_incidents"`
	ResolvedIncidents    int     `json:"resolved_incidents"`
	ResolutionRate       float64 `json:"resolution_rate"`
}

// SentimentAnalysis represents sentiment analysis aggregation
type SentimentAnalysis struct {
	SentimentLabel string  `json:"sentiment_label"`
	Count          int     `json:"count"`
	Percentage     float64 `json:"percentage"`
	AvgScore       float64 `json:"avg_score"`
}

// AutomationAnalysis represents automation opportunities analysis
type AutomationAnalysis struct {
	ITProcessGroup      string  `json:"it_process_group"`
	IncidentCount       int     `json:"incident_count"`
	AvgAutomationScore  float64 `json:"avg_automation_score"`
	AutomatableCount    int     `json:"automatable_count"`
	AutomationPercentage float64 `json:"automation_percentage"`
}

// AnalyticsSummary represents comprehensive analytics summary
type AnalyticsSummary struct {
	TotalIncidents      int                   `json:"total_incidents"`
	ResolvedIncidents   int                   `json:"resolved_incidents"`
	ResolutionRate      float64               `json:"resolution_rate"`
	AvgResolutionTime   float64               `json:"avg_resolution_time"`
	PriorityBreakdown   []PriorityAnalysis    `json:"priority_breakdown"`
	SentimentBreakdown  []SentimentAnalysis   `json:"sentiment_breakdown"`
	AutomationSummary   []AutomationAnalysis  `json:"automation_summary"`
	TopApplications     []ApplicationAnalysis `json:"top_applications"`
}

// TimelineFilters represents filters for timeline queries
type TimelineFilters struct {
	StartDate    *time.Time `json:"start_date,omitempty"`
	EndDate      *time.Time `json:"end_date,omitempty"`
	Priorities   []string   `json:"priorities,omitempty"`
	Applications []string   `json:"applications,omitempty"`
	Statuses     []string   `json:"statuses,omitempty"`
}

// GetDailyTimeline returns daily incident timeline data with optional filters
func (s *AnalyticsService) GetDailyTimeline(ctx context.Context, filters *TimelineFilters) ([]TimelineData, error) {
	query := `
		SELECT 
			DATE_TRUNC('day', report_date) as date,
			COUNT(*) as incident_count,
			COUNT(CASE WHEN priority = 'P1' THEN 1 END) as p1_count,
			COUNT(CASE WHEN priority = 'P2' THEN 1 END) as p2_count,
			COUNT(CASE WHEN priority = 'P3' THEN 1 END) as p3_count,
			COUNT(CASE WHEN priority = 'P4' THEN 1 END) as p4_count
		FROM incidents 
		WHERE 1=1`

	// Apply filters
	whereClause, args, _ := buildFilterConditions(filters, 1)
	query += whereClause
	query += " GROUP BY DATE_TRUNC('day', report_date) ORDER BY date"

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query daily timeline: %w", err)
	}
	defer rows.Close()

	var timeline []TimelineData
	for rows.Next() {
		var data TimelineData
		var date time.Time
		
		err := rows.Scan(
			&date,
			&data.IncidentCount,
			&data.P1Count,
			&data.P2Count,
			&data.P3Count,
			&data.P4Count,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan timeline row: %w", err)
		}
		
		data.Date = date.Format("2006-01-02")
		timeline = append(timeline, data)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating timeline rows: %w", err)
	}

	return timeline, nil
}

// GetWeeklyTimeline returns weekly incident timeline data with optional filters
func (s *AnalyticsService) GetWeeklyTimeline(ctx context.Context, filters *TimelineFilters) ([]TimelineData, error) {
	query := `
		SELECT 
			DATE_TRUNC('week', report_date) as week,
			COUNT(*) as incident_count,
			COUNT(CASE WHEN priority = 'P1' THEN 1 END) as p1_count,
			COUNT(CASE WHEN priority = 'P2' THEN 1 END) as p2_count,
			COUNT(CASE WHEN priority = 'P3' THEN 1 END) as p3_count,
			COUNT(CASE WHEN priority = 'P4' THEN 1 END) as p4_count
		FROM incidents 
		WHERE 1=1`

	// Apply filters
	whereClause, args, _ := buildFilterConditions(filters, 1)
	query += whereClause
	query += " GROUP BY DATE_TRUNC('week', report_date) ORDER BY week"

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query weekly timeline: %w", err)
	}
	defer rows.Close()

	var timeline []TimelineData
	for rows.Next() {
		var data TimelineData
		var week time.Time
		
		err := rows.Scan(
			&week,
			&data.IncidentCount,
			&data.P1Count,
			&data.P2Count,
			&data.P3Count,
			&data.P4Count,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan weekly timeline row: %w", err)
		}
		
		data.Date = week.Format("2006-01-02")
		timeline = append(timeline, data)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating weekly timeline rows: %w", err)
	}

	return timeline, nil
}

// GetTrendAnalysis calculates trend analysis for incident data
func (s *AnalyticsService) GetTrendAnalysis(ctx context.Context, period string, filters *TimelineFilters) ([]TrendAnalysis, error) {
	var timelineData []TimelineData
	var err error

	// Get timeline data based on period
	switch period {
	case "daily":
		timelineData, err = s.GetDailyTimeline(ctx, filters)
	case "weekly":
		timelineData, err = s.GetWeeklyTimeline(ctx, filters)
	default:
		return nil, fmt.Errorf("unsupported period: %s", period)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get timeline data for trend analysis: %w", err)
	}

	if len(timelineData) < 2 {
		return []TrendAnalysis{}, nil
	}

	var trends []TrendAnalysis
	for i := 1; i < len(timelineData); i++ {
		current := timelineData[i]
		previous := timelineData[i-1]
		
		var growthRate float64
		if previous.IncidentCount > 0 {
			growthRate = float64(current.IncidentCount-previous.IncidentCount) / float64(previous.IncidentCount) * 100
		}

		trend := "stable"
		if growthRate > 5 {
			trend = "increasing"
		} else if growthRate < -5 {
			trend = "decreasing"
		}

		trends = append(trends, TrendAnalysis{
			Period:        current.Date,
			IncidentCount: current.IncidentCount,
			GrowthRate:    growthRate,
			Trend:         trend,
		})
	}

	return trends, nil
}

// GetTicketsPerDayMetrics returns metrics for tickets per day
func (s *AnalyticsService) GetTicketsPerDayMetrics(ctx context.Context, filters *TimelineFilters) (map[string]interface{}, error) {
	query := `
		SELECT 
			SUM(daily_count) as total_incidents,
			AVG(daily_count) as avg_per_day,
			MAX(daily_count) as max_per_day,
			MIN(daily_count) as min_per_day,
			PERCENTILE_CONT(0.5) WITHIN GROUP (ORDER BY daily_count) as median_per_day
		FROM (
			SELECT 
				DATE_TRUNC('day', report_date) as date,
				COUNT(*) as daily_count
			FROM incidents 
			WHERE 1=1`

	// Apply filters
	whereClause, args, _ := buildFilterConditions(filters, 1)
	query += whereClause
	query += " GROUP BY DATE_TRUNC('day', report_date)) daily_stats"

	var totalIncidents int
	var avgPerDay, maxPerDay, minPerDay, medianPerDay float64

	err := s.db.QueryRowContext(ctx, query, args...).Scan(
		&totalIncidents,
		&avgPerDay,
		&maxPerDay,
		&minPerDay,
		&medianPerDay,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query tickets per day metrics: %w", err)
	}

	return map[string]interface{}{
		"total_incidents": totalIncidents,
		"avg_per_day":     avgPerDay,
		"max_per_day":     maxPerDay,
		"min_per_day":     minPerDay,
		"median_per_day":  medianPerDay,
	}, nil
}

// GetTicketsPerWeekMetrics returns metrics for tickets per week
func (s *AnalyticsService) GetTicketsPerWeekMetrics(ctx context.Context, filters *TimelineFilters) (map[string]interface{}, error) {
	query := `
		SELECT 
			SUM(weekly_count) as total_incidents,
			AVG(weekly_count) as avg_per_week,
			MAX(weekly_count) as max_per_week,
			MIN(weekly_count) as min_per_week,
			PERCENTILE_CONT(0.5) WITHIN GROUP (ORDER BY weekly_count) as median_per_week
		FROM (
			SELECT 
				DATE_TRUNC('week', report_date) as week,
				COUNT(*) as weekly_count
			FROM incidents 
			WHERE 1=1`

	// Apply filters
	whereClause, args, _ := buildFilterConditions(filters, 1)
	query += whereClause
	query += " GROUP BY DATE_TRUNC('week', report_date)) weekly_stats"

	var totalIncidents int
	var avgPerWeek, maxPerWeek, minPerWeek, medianPerWeek float64

	err := s.db.QueryRowContext(ctx, query, args...).Scan(
		&totalIncidents,
		&avgPerWeek,
		&maxPerWeek,
		&minPerWeek,
		&medianPerWeek,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query tickets per week metrics: %w", err)
	}

	return map[string]interface{}{
		"total_incidents": totalIncidents,
		"avg_per_week":    avgPerWeek,
		"max_per_week":    maxPerWeek,
		"min_per_week":    minPerWeek,
		"median_per_week": medianPerWeek,
	}, nil
}

// GetPriorityAnalysis returns priority distribution analysis with optional filters
func (s *AnalyticsService) GetPriorityAnalysis(ctx context.Context, filters *TimelineFilters) ([]PriorityAnalysis, error) {
	query := `
		SELECT 
			priority,
			COUNT(*) as count,
			ROUND(COUNT(*) * 100.0 / SUM(COUNT(*)) OVER (), 2) as percentage
		FROM incidents 
		WHERE 1=1`

	// Apply filters
	whereClause, args, _ := buildFilterConditions(filters, 1)
	query += whereClause
	query += " GROUP BY priority ORDER BY priority"

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query priority analysis: %w", err)
	}
	defer rows.Close()

	var analysis []PriorityAnalysis
	for rows.Next() {
		var data PriorityAnalysis
		
		err := rows.Scan(
			&data.Priority,
			&data.Count,
			&data.Percentage,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan priority analysis row: %w", err)
		}
		
		analysis = append(analysis, data)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating priority analysis rows: %w", err)
	}

	return analysis, nil
}

// GetApplicationAnalysis returns application-wise incident breakdown with optional filters
func (s *AnalyticsService) GetApplicationAnalysis(ctx context.Context, filters *TimelineFilters) ([]ApplicationAnalysis, error) {
	query := `
		SELECT 
			application_name,
			COUNT(*) as incident_count,
			AVG(resolution_time_hours) as avg_resolution_time,
			PERCENTILE_CONT(0.5) WITHIN GROUP (ORDER BY resolution_time_hours) as median_resolution_time,
			COUNT(CASE WHEN resolve_date IS NOT NULL THEN 1 END) as resolved_incidents
		FROM incidents 
		WHERE 1=1`

	// Apply filters
	whereClause, args, _ := buildFilterConditions(filters, 1)
	query += whereClause
	query += " GROUP BY application_name ORDER BY incident_count DESC"

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query application analysis: %w", err)
	}
	defer rows.Close()

	var analysis []ApplicationAnalysis
	for rows.Next() {
		var data ApplicationAnalysis
		var avgResolutionTime, medianResolutionTime sql.NullFloat64
		
		err := rows.Scan(
			&data.ApplicationName,
			&data.IncidentCount,
			&avgResolutionTime,
			&medianResolutionTime,
			&data.ResolvedIncidents,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan application analysis row: %w", err)
		}
		
		if avgResolutionTime.Valid {
			data.AvgResolutionTime = avgResolutionTime.Float64
		}
		if medianResolutionTime.Valid {
			data.MedianResolutionTime = medianResolutionTime.Float64
		}
		
		// Calculate trend (simplified - could be enhanced with historical data)
		data.Trend = "stable"
		if data.IncidentCount > 10 {
			data.Trend = "increasing"
		} else if data.IncidentCount < 5 {
			data.Trend = "decreasing"
		}
		
		analysis = append(analysis, data)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating application analysis rows: %w", err)
	}

	return analysis, nil
}

// GetResolutionAnalysis returns resolution analysis with average times and metrics
func (s *AnalyticsService) GetResolutionAnalysis(ctx context.Context, filters *TimelineFilters) (*ResolutionMetrics, error) {
	query := `
		SELECT 
			COUNT(*) as total_incidents,
			COUNT(CASE WHEN resolve_date IS NOT NULL THEN 1 END) as resolved_incidents,
			AVG(resolution_time_hours) as avg_resolution_time,
			PERCENTILE_CONT(0.5) WITHIN GROUP (ORDER BY resolution_time_hours) as median_resolution_time
		FROM incidents 
		WHERE 1=1`

	// Apply filters
	whereClause, args, _ := buildFilterConditions(filters, 1)
	query += whereClause

	var metrics ResolutionMetrics
	var avgResolutionTime, medianResolutionTime sql.NullFloat64

	err := s.db.QueryRowContext(ctx, query, args...).Scan(
		&metrics.TotalIncidents,
		&metrics.ResolvedIncidents,
		&avgResolutionTime,
		&medianResolutionTime,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query resolution analysis: %w", err)
	}

	if avgResolutionTime.Valid {
		metrics.AvgResolutionTime = avgResolutionTime.Float64
	}
	if medianResolutionTime.Valid {
		metrics.MedianResolutionTime = medianResolutionTime.Float64
	}

	// Calculate resolution rate
	if metrics.TotalIncidents > 0 {
		metrics.ResolutionRate = float64(metrics.ResolvedIncidents) / float64(metrics.TotalIncidents) * 100
	}

	return &metrics, nil
}

// GetPerformanceMetrics returns performance metrics calculation utilities
func (s *AnalyticsService) GetPerformanceMetrics(ctx context.Context, filters *TimelineFilters) (map[string]interface{}, error) {
	// Get resolution analysis
	resolutionMetrics, err := s.GetResolutionAnalysis(ctx, filters)
	if err != nil {
		return nil, fmt.Errorf("failed to get resolution metrics: %w", err)
	}

	// Get priority analysis
	priorityAnalysis, err := s.GetPriorityAnalysis(ctx, filters)
	if err != nil {
		return nil, fmt.Errorf("failed to get priority analysis: %w", err)
	}

	// Get application analysis
	applicationAnalysis, err := s.GetApplicationAnalysis(ctx, filters)
	if err != nil {
		return nil, fmt.Errorf("failed to get application analysis: %w", err)
	}

	// Calculate additional metrics
	var p1Count, p2Count, p3Count, p4Count int
	for _, priority := range priorityAnalysis {
		switch priority.Priority {
		case "P1":
			p1Count = priority.Count
		case "P2":
			p2Count = priority.Count
		case "P3":
			p3Count = priority.Count
		case "P4":
			p4Count = priority.Count
		}
	}

	// Calculate top applications by incident count
	topApplications := make([]ApplicationAnalysis, 0)
	if len(applicationAnalysis) > 0 {
		limit := 5
		if len(applicationAnalysis) < limit {
			limit = len(applicationAnalysis)
		}
		topApplications = applicationAnalysis[:limit]
	}

	return map[string]interface{}{
		"resolution_metrics": resolutionMetrics,
		"priority_breakdown": map[string]interface{}{
			"p1_count": p1Count,
			"p2_count": p2Count,
			"p3_count": p3Count,
			"p4_count": p4Count,
			"total":    p1Count + p2Count + p3Count + p4Count,
		},
		"top_applications":     topApplications,
		"total_applications":   len(applicationAnalysis),
		"priority_distribution": priorityAnalysis,
	}, nil
}

// GetSentimentAnalysis returns sentiment analysis aggregation with optional filters
func (s *AnalyticsService) GetSentimentAnalysis(ctx context.Context, filters *TimelineFilters) ([]SentimentAnalysis, error) {
	query := `
		SELECT 
			sentiment_label,
			COUNT(*) as count,
			ROUND(COUNT(*) * 100.0 / SUM(COUNT(*)) OVER (), 2) as percentage,
			ROUND(AVG(sentiment_score), 3) as avg_score
		FROM incidents 
		WHERE sentiment_label IS NOT NULL`

	// Apply filters
	whereClause, args, _ := buildFilterConditions(filters, 1)
	query += whereClause
	query += " GROUP BY sentiment_label ORDER BY count DESC"

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query sentiment analysis: %w", err)
	}
	defer rows.Close()

	var analysis []SentimentAnalysis
	for rows.Next() {
		var data SentimentAnalysis
		var avgScore sql.NullFloat64
		
		err := rows.Scan(
			&data.SentimentLabel,
			&data.Count,
			&data.Percentage,
			&avgScore,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan sentiment analysis row: %w", err)
		}
		
		if avgScore.Valid {
			data.AvgScore = avgScore.Float64
		}
		
		analysis = append(analysis, data)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating sentiment analysis rows: %w", err)
	}

	return analysis, nil
}

// GetAutomationAnalysis returns automation opportunities analysis with optional filters
func (s *AnalyticsService) GetAutomationAnalysis(ctx context.Context, filters *TimelineFilters) ([]AutomationAnalysis, error) {
	query := `
		SELECT 
			it_process_group,
			COUNT(*) as incident_count,
			AVG(automation_score) as avg_automation_score,
			COUNT(CASE WHEN automation_feasible = true THEN 1 END) as automatable_count,
			ROUND(COUNT(CASE WHEN automation_feasible = true THEN 1 END) * 100.0 / COUNT(*), 2) as automation_percentage
		FROM incidents 
		WHERE it_process_group IS NOT NULL`

	// Apply filters
	whereClause, args, _ := buildFilterConditions(filters, 1)
	query += whereClause
	query += " GROUP BY it_process_group ORDER BY automation_percentage DESC"

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query automation analysis: %w", err)
	}
	defer rows.Close()

	var analysis []AutomationAnalysis
	for rows.Next() {
		var data AutomationAnalysis
		var avgAutomationScore sql.NullFloat64
		
		err := rows.Scan(
			&data.ITProcessGroup,
			&data.IncidentCount,
			&avgAutomationScore,
			&data.AutomatableCount,
			&data.AutomationPercentage,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan automation analysis row: %w", err)
		}
		
		if avgAutomationScore.Valid {
			data.AvgAutomationScore = avgAutomationScore.Float64
		}
		
		analysis = append(analysis, data)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating automation analysis rows: %w", err)
	}

	return analysis, nil
}

// GetITProcessAutomationReporting returns IT process automation reporting utilities
func (s *AnalyticsService) GetITProcessAutomationReporting(ctx context.Context, filters *TimelineFilters) (map[string]interface{}, error) {
	// Get automation analysis
	automationAnalysis, err := s.GetAutomationAnalysis(ctx, filters)
	if err != nil {
		return nil, fmt.Errorf("failed to get automation analysis: %w", err)
	}

	// Calculate overall automation metrics
	var totalIncidents, totalAutomatable int
	var totalAutomationScore float64
	processGroups := make(map[string]AutomationAnalysis)

	for _, analysis := range automationAnalysis {
		totalIncidents += analysis.IncidentCount
		totalAutomatable += analysis.AutomatableCount
		totalAutomationScore += analysis.AvgAutomationScore * float64(analysis.IncidentCount)
		processGroups[analysis.ITProcessGroup] = analysis
	}

	var overallAutomationScore float64
	if totalIncidents > 0 {
		overallAutomationScore = totalAutomationScore / float64(totalIncidents)
	}

	var overallAutomationPercentage float64
	if totalIncidents > 0 {
		overallAutomationPercentage = float64(totalAutomatable) / float64(totalIncidents) * 100
	}

	// Get top automation opportunities
	topOpportunities := make([]AutomationAnalysis, 0)
	if len(automationAnalysis) > 0 {
		limit := 5
		if len(automationAnalysis) < limit {
			limit = len(automationAnalysis)
		}
		topOpportunities = automationAnalysis[:limit]
	}

	return map[string]interface{}{
		"overall_metrics": map[string]interface{}{
			"total_incidents":             totalIncidents,
			"total_automatable":           totalAutomatable,
			"overall_automation_score":    overallAutomationScore,
			"overall_automation_percentage": overallAutomationPercentage,
		},
		"process_groups":      processGroups,
		"top_opportunities":   topOpportunities,
		"detailed_analysis":   automationAnalysis,
		"total_process_groups": len(processGroups),
	}, nil
}

// GetAnalyticsSummary returns comprehensive analytics summary endpoint
func (s *AnalyticsService) GetAnalyticsSummary(ctx context.Context, filters *TimelineFilters) (*AnalyticsSummary, error) {
	// Get resolution metrics
	resolutionMetrics, err := s.GetResolutionAnalysis(ctx, filters)
	if err != nil {
		return nil, fmt.Errorf("failed to get resolution metrics: %w", err)
	}

	// Get priority analysis
	priorityAnalysis, err := s.GetPriorityAnalysis(ctx, filters)
	if err != nil {
		return nil, fmt.Errorf("failed to get priority analysis: %w", err)
	}

	// Get sentiment analysis
	sentimentAnalysis, err := s.GetSentimentAnalysis(ctx, filters)
	if err != nil {
		return nil, fmt.Errorf("failed to get sentiment analysis: %w", err)
	}

	// Get automation analysis
	automationAnalysis, err := s.GetAutomationAnalysis(ctx, filters)
	if err != nil {
		return nil, fmt.Errorf("failed to get automation analysis: %w", err)
	}

	// Get top applications
	applicationAnalysis, err := s.GetApplicationAnalysis(ctx, filters)
	if err != nil {
		return nil, fmt.Errorf("failed to get application analysis: %w", err)
	}

	// Get top 5 applications
	topApplications := make([]ApplicationAnalysis, 0)
	if len(applicationAnalysis) > 0 {
		limit := 5
		if len(applicationAnalysis) < limit {
			limit = len(applicationAnalysis)
		}
		topApplications = applicationAnalysis[:limit]
	}

	summary := &AnalyticsSummary{
		TotalIncidents:     resolutionMetrics.TotalIncidents,
		ResolvedIncidents:  resolutionMetrics.ResolvedIncidents,
		ResolutionRate:     resolutionMetrics.ResolutionRate,
		AvgResolutionTime:  resolutionMetrics.AvgResolutionTime,
		PriorityBreakdown:  priorityAnalysis,
		SentimentBreakdown: sentimentAnalysis,
		AutomationSummary:  automationAnalysis,
		TopApplications:    topApplications,
	}

	return summary, nil
}