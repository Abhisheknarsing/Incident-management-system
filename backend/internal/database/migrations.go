package database

import (
	"context"
	"fmt"
	"log"
	"sort"
	"time"
)

// Migration represents a database migration
type Migration struct {
	Version     int
	Name        string
	UpQuery     string
	DownQuery   string
	AppliedAt   *time.Time
}

// MigrationManager handles database migrations
type MigrationManager struct {
	db *DB
}

// NewMigrationManager creates a new migration manager
func NewMigrationManager(db *DB) *MigrationManager {
	return &MigrationManager{db: db}
}

// GetMigrations returns all available migrations
func (mm *MigrationManager) GetMigrations() []Migration {
	return []Migration{
		{
			Version: 1,
			Name:    "create_uploads_table",
			UpQuery: `
				CREATE TABLE IF NOT EXISTS uploads (
					id VARCHAR PRIMARY KEY,
					filename VARCHAR NOT NULL,
					original_filename VARCHAR NOT NULL,
					status VARCHAR NOT NULL CHECK (status IN ('uploaded', 'processing', 'completed', 'failed')),
					record_count INTEGER DEFAULT 0,
					processed_count INTEGER DEFAULT 0,
					error_count INTEGER DEFAULT 0,
					errors TEXT[],
					created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
					processed_at TIMESTAMP
				)
			`,
			DownQuery: "DROP TABLE IF EXISTS uploads",
		},
		{
			Version: 2,
			Name:    "create_incidents_table",
			UpQuery: `
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
			`,
			DownQuery: "DROP TABLE IF EXISTS incidents",
		},
		{
			Version: 3,
			Name:    "create_indexes",
			UpQuery: `
				CREATE INDEX IF NOT EXISTS idx_incidents_upload_id ON incidents(upload_id);
				CREATE INDEX IF NOT EXISTS idx_incidents_report_date ON incidents(report_date);
				CREATE INDEX IF NOT EXISTS idx_incidents_priority ON incidents(priority);
				CREATE INDEX IF NOT EXISTS idx_incidents_application ON incidents(application_name);
				CREATE INDEX IF NOT EXISTS idx_incidents_status ON incidents(status);
				CREATE INDEX IF NOT EXISTS idx_incidents_resolution_group ON incidents(resolution_group);
				CREATE INDEX IF NOT EXISTS idx_incidents_sentiment_label ON incidents(sentiment_label);
				CREATE INDEX IF NOT EXISTS idx_incidents_it_process_group ON incidents(it_process_group);
				CREATE INDEX IF NOT EXISTS idx_uploads_status ON uploads(status);
				CREATE INDEX IF NOT EXISTS idx_uploads_created_at ON uploads(created_at);
			`,
			DownQuery: `
				DROP INDEX IF EXISTS idx_incidents_upload_id;
				DROP INDEX IF EXISTS idx_incidents_report_date;
				DROP INDEX IF EXISTS idx_incidents_priority;
				DROP INDEX IF EXISTS idx_incidents_application;
				DROP INDEX IF EXISTS idx_incidents_status;
				DROP INDEX IF EXISTS idx_incidents_resolution_group;
				DROP INDEX IF EXISTS idx_incidents_sentiment_label;
				DROP INDEX IF EXISTS idx_incidents_it_process_group;
				DROP INDEX IF EXISTS idx_uploads_status;
				DROP INDEX IF EXISTS idx_uploads_created_at;
			`,
		},
		{
			Version: 4,
			Name:    "create_analytics_views",
			UpQuery: `
				-- Daily incident timeline
				CREATE VIEW IF NOT EXISTS incident_timeline AS
				SELECT 
					DATE_TRUNC('day', report_date) as date,
					COUNT(*) as incident_count,
					COUNT(CASE WHEN priority = 'P1' THEN 1 END) as p1_count,
					COUNT(CASE WHEN priority = 'P2' THEN 1 END) as p2_count,
					COUNT(CASE WHEN priority = 'P3' THEN 1 END) as p3_count,
					COUNT(CASE WHEN priority = 'P4' THEN 1 END) as p4_count
				FROM incidents 
				GROUP BY DATE_TRUNC('day', report_date)
				ORDER BY date;

				-- Weekly incident timeline
				CREATE VIEW IF NOT EXISTS weekly_timeline AS
				SELECT 
					DATE_TRUNC('week', report_date) as week,
					COUNT(*) as incident_count,
					COUNT(CASE WHEN priority = 'P1' THEN 1 END) as p1_count,
					COUNT(CASE WHEN priority = 'P2' THEN 1 END) as p2_count,
					COUNT(CASE WHEN priority = 'P3' THEN 1 END) as p3_count,
					COUNT(CASE WHEN priority = 'P4' THEN 1 END) as p4_count
				FROM incidents 
				GROUP BY DATE_TRUNC('week', report_date)
				ORDER BY week;

				-- Resolution metrics by application and priority
				CREATE VIEW IF NOT EXISTS resolution_metrics AS
				SELECT 
					application_name,
					priority,
					AVG(resolution_time_hours) as avg_resolution_time,
					PERCENTILE_CONT(0.5) WITHIN GROUP (ORDER BY resolution_time_hours) as median_resolution_time,
					COUNT(*) as total_incidents,
					COUNT(CASE WHEN resolve_date IS NOT NULL THEN 1 END) as resolved_incidents
				FROM incidents 
				WHERE resolution_time_hours IS NOT NULL
				GROUP BY application_name, priority;

				-- Priority analysis
				CREATE VIEW IF NOT EXISTS priority_analysis AS
				SELECT 
					priority,
					COUNT(*) as count,
					ROUND(COUNT(*) * 100.0 / SUM(COUNT(*)) OVER (), 2) as percentage
				FROM incidents 
				GROUP BY priority
				ORDER BY priority;

				-- Sentiment summary
				CREATE VIEW IF NOT EXISTS sentiment_summary AS
				SELECT 
					sentiment_label,
					COUNT(*) as count,
					ROUND(AVG(sentiment_score), 3) as avg_score
				FROM incidents 
				WHERE sentiment_label IS NOT NULL
				GROUP BY sentiment_label;

				-- Automation opportunities
				CREATE VIEW IF NOT EXISTS automation_opportunities AS
				SELECT 
					it_process_group,
					COUNT(*) as incident_count,
					AVG(automation_score) as avg_automation_score,
					COUNT(CASE WHEN automation_feasible = true THEN 1 END) as automatable_count,
					ROUND(COUNT(CASE WHEN automation_feasible = true THEN 1 END) * 100.0 / COUNT(*), 2) as automation_percentage
				FROM incidents 
				WHERE it_process_group IS NOT NULL
				GROUP BY it_process_group
				ORDER BY automation_percentage DESC;
			`,
			DownQuery: `
				DROP VIEW IF EXISTS incident_timeline;
				DROP VIEW IF EXISTS weekly_timeline;
				DROP VIEW IF EXISTS resolution_metrics;
				DROP VIEW IF EXISTS priority_analysis;
				DROP VIEW IF EXISTS sentiment_summary;
				DROP VIEW IF EXISTS automation_opportunities;
			`,
		},
	}
}

// InitializeMigrationTable creates the migration tracking table
func (mm *MigrationManager) InitializeMigrationTable() error {
	conn := mm.db.GetConnection()
	if conn == nil {
		return fmt.Errorf("database connection not available")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	query := `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version INTEGER PRIMARY KEY,
			name VARCHAR NOT NULL,
			applied_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`

	_, err := conn.ExecContext(ctx, query)
	if err != nil {
		return WrapError("initialize_migration_table", err)
	}

	log.Println("Migration table initialized")
	return nil
}

// GetAppliedMigrations returns all applied migrations
func (mm *MigrationManager) GetAppliedMigrations() ([]Migration, error) {
	conn := mm.db.GetConnection()
	if conn == nil {
		return nil, fmt.Errorf("database connection not available")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	query := `
		SELECT version, name, applied_at 
		FROM schema_migrations 
		ORDER BY version
	`

	rows, err := conn.QueryContext(ctx, query)
	if err != nil {
		return nil, WrapError("get_applied_migrations", err)
	}
	defer rows.Close()

	var migrations []Migration
	for rows.Next() {
		var migration Migration
		err := rows.Scan(&migration.Version, &migration.Name, &migration.AppliedAt)
		if err != nil {
			return nil, WrapError("scan_migration", err)
		}
		migrations = append(migrations, migration)
	}

	return migrations, nil
}

// GetPendingMigrations returns migrations that haven't been applied
func (mm *MigrationManager) GetPendingMigrations() ([]Migration, error) {
	allMigrations := mm.GetMigrations()
	appliedMigrations, err := mm.GetAppliedMigrations()
	if err != nil {
		return nil, err
	}

	// Create a map of applied migration versions
	appliedVersions := make(map[int]bool)
	for _, migration := range appliedMigrations {
		appliedVersions[migration.Version] = true
	}

	// Find pending migrations
	var pendingMigrations []Migration
	for _, migration := range allMigrations {
		if !appliedVersions[migration.Version] {
			pendingMigrations = append(pendingMigrations, migration)
		}
	}

	// Sort by version
	sort.Slice(pendingMigrations, func(i, j int) bool {
		return pendingMigrations[i].Version < pendingMigrations[j].Version
	})

	return pendingMigrations, nil
}

// ApplyMigration applies a single migration
func (mm *MigrationManager) ApplyMigration(migration Migration) error {
	conn := mm.db.GetConnection()
	if conn == nil {
		return fmt.Errorf("database connection not available")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	tx, err := conn.BeginTx(ctx, nil)
	if err != nil {
		return WrapError("begin_migration_transaction", err)
	}
	defer tx.Rollback()

	// Apply the migration
	_, err = tx.ExecContext(ctx, migration.UpQuery)
	if err != nil {
		return WrapError(fmt.Sprintf("apply_migration_%d", migration.Version), err)
	}

	// Record the migration
	recordQuery := `
		INSERT INTO schema_migrations (version, name, applied_at) 
		VALUES (?, ?, CURRENT_TIMESTAMP)
	`
	_, err = tx.ExecContext(ctx, recordQuery, migration.Version, migration.Name)
	if err != nil {
		return WrapError("record_migration", err)
	}

	if err := tx.Commit(); err != nil {
		return WrapError("commit_migration", err)
	}

	log.Printf("Applied migration %d: %s", migration.Version, migration.Name)
	return nil
}

// RollbackMigration rolls back a single migration
func (mm *MigrationManager) RollbackMigration(migration Migration) error {
	conn := mm.db.GetConnection()
	if conn == nil {
		return fmt.Errorf("database connection not available")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	tx, err := conn.BeginTx(ctx, nil)
	if err != nil {
		return WrapError("begin_rollback_transaction", err)
	}
	defer tx.Rollback()

	// Apply the rollback
	_, err = tx.ExecContext(ctx, migration.DownQuery)
	if err != nil {
		return WrapError(fmt.Sprintf("rollback_migration_%d", migration.Version), err)
	}

	// Remove the migration record
	removeQuery := `DELETE FROM schema_migrations WHERE version = ?`
	_, err = tx.ExecContext(ctx, removeQuery, migration.Version)
	if err != nil {
		return WrapError("remove_migration_record", err)
	}

	if err := tx.Commit(); err != nil {
		return WrapError("commit_rollback", err)
	}

	log.Printf("Rolled back migration %d: %s", migration.Version, migration.Name)
	return nil
}

// MigrateUp applies all pending migrations
func (mm *MigrationManager) MigrateUp() error {
	// Initialize migration table first
	if err := mm.InitializeMigrationTable(); err != nil {
		return err
	}

	pendingMigrations, err := mm.GetPendingMigrations()
	if err != nil {
		return err
	}

	if len(pendingMigrations) == 0 {
		log.Println("No pending migrations to apply")
		return nil
	}

	log.Printf("Applying %d pending migrations", len(pendingMigrations))

	for _, migration := range pendingMigrations {
		if err := mm.ApplyMigration(migration); err != nil {
			return fmt.Errorf("failed to apply migration %d: %w", migration.Version, err)
		}
	}

	log.Println("All migrations applied successfully")
	return nil
}

// MigrateDown rolls back migrations to a specific version
func (mm *MigrationManager) MigrateDown(targetVersion int) error {
	appliedMigrations, err := mm.GetAppliedMigrations()
	if err != nil {
		return err
	}

	// Find migrations to rollback (in reverse order)
	var migrationsToRollback []Migration
	allMigrations := mm.GetMigrations()
	migrationMap := make(map[int]Migration)
	for _, migration := range allMigrations {
		migrationMap[migration.Version] = migration
	}

	for i := len(appliedMigrations) - 1; i >= 0; i-- {
		migration := appliedMigrations[i]
		if migration.Version > targetVersion {
			if fullMigration, exists := migrationMap[migration.Version]; exists {
				migrationsToRollback = append(migrationsToRollback, fullMigration)
			}
		}
	}

	if len(migrationsToRollback) == 0 {
		log.Printf("No migrations to rollback to version %d", targetVersion)
		return nil
	}

	log.Printf("Rolling back %d migrations to version %d", len(migrationsToRollback), targetVersion)

	for _, migration := range migrationsToRollback {
		if err := mm.RollbackMigration(migration); err != nil {
			return fmt.Errorf("failed to rollback migration %d: %w", migration.Version, err)
		}
	}

	log.Printf("Successfully rolled back to version %d", targetVersion)
	return nil
}

// GetMigrationStatus returns the current migration status
func (mm *MigrationManager) GetMigrationStatus() (map[string]interface{}, error) {
	appliedMigrations, err := mm.GetAppliedMigrations()
	if err != nil {
		return nil, err
	}

	pendingMigrations, err := mm.GetPendingMigrations()
	if err != nil {
		return nil, err
	}

	var currentVersion int
	if len(appliedMigrations) > 0 {
		currentVersion = appliedMigrations[len(appliedMigrations)-1].Version
	}

	status := map[string]interface{}{
		"current_version":     currentVersion,
		"applied_count":       len(appliedMigrations),
		"pending_count":       len(pendingMigrations),
		"applied_migrations":  appliedMigrations,
		"pending_migrations":  pendingMigrations,
	}

	return status, nil
}