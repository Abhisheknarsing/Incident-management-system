package database

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"
)

// InitializeDatabase creates the database directory and initializes the schema
func (db *DB) InitializeDatabase() error {
	// Create database directory if it doesn't exist
	if err := db.createDatabaseDirectory(); err != nil {
		return fmt.Errorf("failed to create database directory: %w", err)
	}

	// Initialize schema
	if err := db.initializeSchema(); err != nil {
		return fmt.Errorf("failed to initialize database schema: %w", err)
	}

	log.Println("Database initialized successfully")
	return nil
}

// createDatabaseDirectory creates the directory for the database file
func (db *DB) createDatabaseDirectory() error {
	dbDir := filepath.Dir(db.dbPath)
	if dbDir == "." {
		return nil // Current directory, no need to create
	}

	if err := os.MkdirAll(dbDir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", dbDir, err)
	}

	log.Printf("Database directory created: %s", dbDir)
	return nil
}

// initializeSchema creates the database schema if it doesn't exist
func (db *DB) initializeSchema() error {
	conn := db.GetConnection()
	if conn == nil {
		return fmt.Errorf("database connection not available")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Start transaction for schema creation
	tx, err := conn.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Create uploads table
	if err := db.createUploadsTable(ctx, tx); err != nil {
		return fmt.Errorf("failed to create uploads table: %w", err)
	}

	// Create incidents table
	if err := db.createIncidentsTable(ctx, tx); err != nil {
		return fmt.Errorf("failed to create incidents table: %w", err)
	}

	// Create indexes
	if err := db.createIndexes(ctx, tx); err != nil {
		return fmt.Errorf("failed to create indexes: %w", err)
	}

	// Create analytics views
	if err := db.createAnalyticsViews(ctx, tx); err != nil {
		return fmt.Errorf("failed to create analytics views: %w", err)
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit schema transaction: %w", err)
	}

	log.Println("Database schema initialized successfully")
	return nil
}

// CheckSchemaExists checks if the database schema exists
func (db *DB) CheckSchemaExists() (bool, error) {
	conn := db.GetConnection()
	if conn == nil {
		return false, fmt.Errorf("database connection not available")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Check if uploads table exists
	query := `
		SELECT COUNT(*) 
		FROM information_schema.tables 
		WHERE table_name = 'uploads'
	`

	var count int
	err := conn.QueryRowContext(ctx, query).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("failed to check schema existence: %w", err)
	}

	return count > 0, nil
}

// ResetDatabase drops all tables and recreates the schema (for testing/development)
func (db *DB) ResetDatabase() error {
	conn := db.GetConnection()
	if conn == nil {
		return fmt.Errorf("database connection not available")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	tx, err := conn.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Drop views first (due to dependencies)
	dropViews := []string{
		"DROP VIEW IF EXISTS incident_timeline",
		"DROP VIEW IF EXISTS weekly_timeline", 
		"DROP VIEW IF EXISTS resolution_metrics",
		"DROP VIEW IF EXISTS priority_analysis",
		"DROP VIEW IF EXISTS sentiment_summary",
		"DROP VIEW IF EXISTS automation_opportunities",
	}

	for _, query := range dropViews {
		if _, err := tx.ExecContext(ctx, query); err != nil {
			return fmt.Errorf("failed to drop view: %w", err)
		}
	}

	// Drop tables
	dropTables := []string{
		"DROP TABLE IF EXISTS incidents",
		"DROP TABLE IF EXISTS uploads",
	}

	for _, query := range dropTables {
		if _, err := tx.ExecContext(ctx, query); err != nil {
			return fmt.Errorf("failed to drop table: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit drop transaction: %w", err)
	}

	// Reinitialize schema
	return db.initializeSchema()
}