package database

import (
	"testing"
	"time"
)

func TestNewDB(t *testing.T) {
	// Use a temporary database file for testing
	config := &Config{
		DatabasePath:    ":memory:", // In-memory database for testing
		MaxOpenConns:    5,
		MaxIdleConns:    2,
		ConnMaxLifetime: time.Minute,
		ConnMaxIdleTime: time.Second * 30,
	}

	db, err := NewDB(config)
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer db.Close()

	if !db.IsReady() {
		t.Error("Database should be ready after creation")
	}

	// Test health check
	if err := db.HealthCheck(); err != nil {
		t.Errorf("Health check failed: %v", err)
	}

	// Test connection stats
	stats := db.Stats()
	if stats.MaxOpenConnections != 5 {
		t.Errorf("Expected MaxOpenConnections to be 5, got %d", stats.MaxOpenConnections)
	}
}

func TestDatabaseInitialization(t *testing.T) {
	// Use a temporary database file for testing
	config := &Config{
		DatabasePath:    ":memory:",
		MaxOpenConns:    5,
		MaxIdleConns:    2,
		ConnMaxLifetime: time.Minute,
		ConnMaxIdleTime: time.Second * 30,
	}

	db, err := NewDB(config)
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer db.Close()

	// Initialize the database schema
	if err := db.InitializeDatabase(); err != nil {
		t.Fatalf("Failed to initialize database: %v", err)
	}

	// Check if schema exists
	exists, err := db.CheckSchemaExists()
	if err != nil {
		t.Fatalf("Failed to check schema existence: %v", err)
	}

	if !exists {
		t.Error("Schema should exist after initialization")
	}
}

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	if config.DatabasePath != "./data/incidents.db" {
		t.Errorf("Expected default database path to be './data/incidents.db', got %s", config.DatabasePath)
	}

	if config.MaxOpenConns != 25 {
		t.Errorf("Expected default MaxOpenConns to be 25, got %d", config.MaxOpenConns)
	}

	if config.MaxIdleConns != 5 {
		t.Errorf("Expected default MaxIdleConns to be 5, got %d", config.MaxIdleConns)
	}
}

func TestDatabaseReset(t *testing.T) {
	config := &Config{
		DatabasePath:    ":memory:",
		MaxOpenConns:    5,
		MaxIdleConns:    2,
		ConnMaxLifetime: time.Minute,
		ConnMaxIdleTime: time.Second * 30,
	}

	db, err := NewDB(config)
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer db.Close()

	// Initialize the database schema
	if err := db.InitializeDatabase(); err != nil {
		t.Fatalf("Failed to initialize database: %v", err)
	}

	// Reset the database
	if err := db.ResetDatabase(); err != nil {
		t.Fatalf("Failed to reset database: %v", err)
	}

	// Check if schema still exists after reset
	exists, err := db.CheckSchemaExists()
	if err != nil {
		t.Fatalf("Failed to check schema existence after reset: %v", err)
	}

	if !exists {
		t.Error("Schema should exist after reset")
	}
}