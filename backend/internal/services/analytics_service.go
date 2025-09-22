package services

import (
	"context"
	"database/sql"
	"fmt"
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

	args := []interface{}{}
	argIndex := 1

	// Apply filters
	if filters != nil {
		if filters.StartDate != nil {
			query += fmt.Sprintf(" AND report_date >= $%d", argIndex)
			args = append(args, *filters.StartDate)
			argIndex++
		}
		if filters.EndDate != nil {
			query += fmt.Sprintf(" AND report_date <= $%d", argIndex)
			args = append(args, *filters.EndDate)
			argIndex++
		}
		if len(filters.Priorities) > 0 {
			query += fmt.Sprintf(" AND priority = ANY($%d)", argIndex)
			args = append(args, filters.Priorities)
			argIndex++
		}
		if len(filters.Applications) > 0 {
			query += fmt.Sprintf(" AND application_name = ANY($%d)", argIndex)
			args = append(args, filters.Applications)
			argIndex++
		}
		if len(filters.Statuses) > 0 {
			query += fmt.Sprintf(" AND status = ANY($%d)", argIndex)
			args = append(args, filters.Statuses)
			argIndex++
		}
	}

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

	args := []interface{}{}
	argIndex := 1

	// Apply filters
	if filters != nil {
		if filters.StartDate != nil {
			query += fmt.Sprintf(" AND report_date >= $%d", argIndex)
			args = append(args, *filters.StartDate)
			argIndex++
		}
		if filters.EndDate != nil {
			query += fmt.Sprintf(" AND report_date <= $%d", argIndex)
			args = append(args, *filters.EndDate)
			argIndex++
		}
		if len(filters.Priorities) > 0 {
			query += fmt.Sprintf(" AND priority = ANY($%d)", argIndex)
			args = append(args, filters.Priorities)
			argIndex++
		}
		if len(filters.Applications) > 0 {
			query += fmt.Sprintf(" AND application_name = ANY($%d)", argIndex)
			args = append(args, filters.Applications)
			argIndex++
		}
		if len(filters.Statuses) > 0 {
			query += fmt.Sprintf(" AND status = ANY($%d)", argIndex)
			args = append(args, filters.Statuses)
			argIndex++
		}
	}

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
			COUNT(*) as total_incidents,
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

	args := []interface{}{}
	argIndex := 1

	// Apply filters
	if filters != nil {
		if filters.StartDate != nil {
			query += fmt.Sprintf(" AND report_date >= $%d", argIndex)
			args = append(args, *filters.StartDate)
			argIndex++
		}
		if filters.EndDate != nil {
			query += fmt.Sprintf(" AND report_date <= $%d", argIndex)
			args = append(args, *filters.EndDate)
			argIndex++
		}
		if len(filters.Priorities) > 0 {
			query += fmt.Sprintf(" AND priority = ANY($%d)", argIndex)
			args = append(args, filters.Priorities)
			argIndex++
		}
		if len(filters.Applications) > 0 {
			query += fmt.Sprintf(" AND application_name = ANY($%d)", argIndex)
			args = append(args, filters.Applications)
			argIndex++
		}
		if len(filters.Statuses) > 0 {
			query += fmt.Sprintf(" AND status = ANY($%d)", argIndex)
			args = append(args, filters.Statuses)
			argIndex++
		}
	}

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
			COUNT(*) as total_incidents,
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

	args := []interface{}{}
	argIndex := 1

	// Apply filters
	if filters != nil {
		if filters.StartDate != nil {
			query += fmt.Sprintf(" AND report_date >= $%d", argIndex)
			args = append(args, *filters.StartDate)
			argIndex++
		}
		if filters.EndDate != nil {
			query += fmt.Sprintf(" AND report_date <= $%d", argIndex)
			args = append(args, *filters.EndDate)
			argIndex++
		}
		if len(filters.Priorities) > 0 {
			query += fmt.Sprintf(" AND priority = ANY($%d)", argIndex)
			args = append(args, filters.Priorities)
			argIndex++
		}
		if len(filters.Applications) > 0 {
			query += fmt.Sprintf(" AND application_name = ANY($%d)", argIndex)
			args = append(args, filters.Applications)
			argIndex++
		}
		if len(filters.Statuses) > 0 {
			query += fmt.Sprintf(" AND status = ANY($%d)", argIndex)
			args = append(args, filters.Statuses)
			argIndex++
		}
	}

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