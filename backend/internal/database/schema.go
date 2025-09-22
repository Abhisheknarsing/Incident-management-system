package database

import (
	"context"
	"database/sql"
)

// createUploadsTable creates the uploads table
func (db *DB) createUploadsTable(ctx context.Context, tx *sql.Tx) error {
	query := `
		CREATE TABLE IF NOT EXISTS uploads (
			id VARCHAR PRIMARY KEY,
			filename VARCHAR NOT NULL,
			original_filename VARCHAR NOT NULL,
			status VARCHAR NOT NULL CHECK (status IN ('uploaded', 'processing', 'completed', 'failed')),
			record_count INTEGER DEFAULT 0,
			processed_count INTEGER DEFAULT 0,
			error_count INTEGER DEFAULT 0,
			errors TEXT,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			processed_at TIMESTAMP
		)
	`

	_, err := tx.ExecContext(ctx, query)
	return err
}

// createIncidentsTable creates the incidents table
func (db *DB) createIncidentsTable(ctx context.Context, tx *sql.Tx) error {
	query := `
		CREATE TABLE IF NOT EXISTS incidents (
			id VARCHAR PRIMARY KEY,
			upload_id VARCHAR NOT NULL,
			incident_id VARCHAR NOT NULL,
			report_date DATE NOT NULL,
			resolve_date DATE,
			last_resolve_date DATE,
			brief_description TEXT NOT NULL,
			description TEXT,
			application_name VARCHAR NOT NULL,
			resolution_group VARCHAR NOT NULL,
			resolved_person VARCHAR NOT NULL,
			priority VARCHAR NOT NULL CHECK (priority IN ('P1', 'P2', 'P3', 'P4')),
			
			-- Additional supported fields
			category VARCHAR,
			subcategory VARCHAR,
			impact VARCHAR,
			urgency VARCHAR,
			status VARCHAR,
			customer_affected VARCHAR,
			business_service VARCHAR,
			root_cause TEXT,
			resolution_notes TEXT,
			
			-- Derived fields from processing
			sentiment_score FLOAT,
			sentiment_label VARCHAR CHECK (sentiment_label IN ('positive', 'negative', 'neutral')),
			resolution_time_hours INTEGER,
			automation_score FLOAT,
			automation_feasible BOOLEAN,
			it_process_group VARCHAR,
			
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			
			-- Constraints for data validation
			CONSTRAINT valid_dates CHECK (resolve_date >= report_date OR resolve_date IS NULL),
			CONSTRAINT unique_incident_per_upload UNIQUE (upload_id, incident_id)
		)
	`

	_, err := tx.ExecContext(ctx, query)
	return err
}

// createIndexes creates performance indexes
func (db *DB) createIndexes(ctx context.Context, tx *sql.Tx) error {
	indexes := []string{
		"CREATE INDEX IF NOT EXISTS idx_incidents_upload_id ON incidents(upload_id)",
		"CREATE INDEX IF NOT EXISTS idx_incidents_report_date ON incidents(report_date)",
		"CREATE INDEX IF NOT EXISTS idx_incidents_priority ON incidents(priority)",
		"CREATE INDEX IF NOT EXISTS idx_incidents_application ON incidents(application_name)",
		"CREATE INDEX IF NOT EXISTS idx_incidents_status ON incidents(status)",
		"CREATE INDEX IF NOT EXISTS idx_incidents_resolution_group ON incidents(resolution_group)",
		"CREATE INDEX IF NOT EXISTS idx_incidents_sentiment_label ON incidents(sentiment_label)",
		"CREATE INDEX IF NOT EXISTS idx_incidents_it_process_group ON incidents(it_process_group)",
		"CREATE INDEX IF NOT EXISTS idx_uploads_status ON uploads(status)",
		"CREATE INDEX IF NOT EXISTS idx_uploads_created_at ON uploads(created_at)",
	}

	for _, indexQuery := range indexes {
		if _, err := tx.ExecContext(ctx, indexQuery); err != nil {
			return err
		}
	}

	return nil
}

// createAnalyticsViews creates pre-computed analytics views for dashboard performance
func (db *DB) createAnalyticsViews(ctx context.Context, tx *sql.Tx) error {
	views := []string{
		// Daily incident timeline
		`CREATE VIEW IF NOT EXISTS incident_timeline AS
		SELECT 
			DATE_TRUNC('day', report_date) as date,
			COUNT(*) as incident_count,
			COUNT(CASE WHEN priority = 'P1' THEN 1 END) as p1_count,
			COUNT(CASE WHEN priority = 'P2' THEN 1 END) as p2_count,
			COUNT(CASE WHEN priority = 'P3' THEN 1 END) as p3_count,
			COUNT(CASE WHEN priority = 'P4' THEN 1 END) as p4_count
		FROM incidents 
		GROUP BY DATE_TRUNC('day', report_date)
		ORDER BY date`,

		// Weekly incident timeline
		`CREATE VIEW IF NOT EXISTS weekly_timeline AS
		SELECT 
			DATE_TRUNC('week', report_date) as week,
			COUNT(*) as incident_count,
			COUNT(CASE WHEN priority = 'P1' THEN 1 END) as p1_count,
			COUNT(CASE WHEN priority = 'P2' THEN 1 END) as p2_count,
			COUNT(CASE WHEN priority = 'P3' THEN 1 END) as p3_count,
			COUNT(CASE WHEN priority = 'P4' THEN 1 END) as p4_count
		FROM incidents 
		GROUP BY DATE_TRUNC('week', report_date)
		ORDER BY week`,

		// Resolution metrics by application and priority
		`CREATE VIEW IF NOT EXISTS resolution_metrics AS
		SELECT 
			application_name,
			priority,
			AVG(resolution_time_hours) as avg_resolution_time,
			PERCENTILE_CONT(0.5) WITHIN GROUP (ORDER BY resolution_time_hours) as median_resolution_time,
			COUNT(*) as total_incidents,
			COUNT(CASE WHEN resolve_date IS NOT NULL THEN 1 END) as resolved_incidents
		FROM incidents 
		WHERE resolution_time_hours IS NOT NULL
		GROUP BY application_name, priority`,

		// Priority analysis
		`CREATE VIEW IF NOT EXISTS priority_analysis AS
		SELECT 
			priority,
			COUNT(*) as count,
			ROUND(COUNT(*) * 100.0 / SUM(COUNT(*)) OVER (), 2) as percentage
		FROM incidents 
		GROUP BY priority
		ORDER BY priority`,

		// Sentiment summary
		`CREATE VIEW IF NOT EXISTS sentiment_summary AS
		SELECT 
			sentiment_label,
			COUNT(*) as count,
			ROUND(AVG(sentiment_score), 3) as avg_score
		FROM incidents 
		WHERE sentiment_label IS NOT NULL
		GROUP BY sentiment_label`,

		// Automation opportunities
		`CREATE VIEW IF NOT EXISTS automation_opportunities AS
		SELECT 
			it_process_group,
			COUNT(*) as incident_count,
			AVG(automation_score) as avg_automation_score,
			COUNT(CASE WHEN automation_feasible = true THEN 1 END) as automatable_count,
			ROUND(COUNT(CASE WHEN automation_feasible = true THEN 1 END) * 100.0 / COUNT(*), 2) as automation_percentage
		FROM incidents 
		WHERE it_process_group IS NOT NULL
		GROUP BY it_process_group
		ORDER BY automation_percentage DESC`,
	}

	for _, viewQuery := range views {
		if _, err := tx.ExecContext(ctx, viewQuery); err != nil {
			return err
		}
	}

	return nil
}
