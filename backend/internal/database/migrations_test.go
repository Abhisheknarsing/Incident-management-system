package database

import (
	"testing"
	"time"
)

func TestMigrationManager(t *testing.T) {
	// Create in-memory database for testing
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

	mm := NewMigrationManager(db)

	// Test initialization
	if err := mm.InitializeMigrationTable(); err != nil {
		t.Fatalf("Failed to initialize migration table: %v", err)
	}

	// Test getting pending migrations (should be all migrations initially)
	pendingMigrations, err := mm.GetPendingMigrations()
	if err != nil {
		t.Fatalf("Failed to get pending migrations: %v", err)
	}

	allMigrations := mm.GetMigrations()
	if len(pendingMigrations) != len(allMigrations) {
		t.Errorf("Expected %d pending migrations, got %d", len(allMigrations), len(pendingMigrations))
	}

	// Test applying all migrations
	if err := mm.MigrateUp(); err != nil {
		t.Fatalf("Failed to migrate up: %v", err)
	}

	// Test that no migrations are pending after migration
	pendingMigrations, err = mm.GetPendingMigrations()
	if err != nil {
		t.Fatalf("Failed to get pending migrations after migration: %v", err)
	}

	if len(pendingMigrations) != 0 {
		t.Errorf("Expected 0 pending migrations after migration, got %d", len(pendingMigrations))
	}

	// Test getting applied migrations
	appliedMigrations, err := mm.GetAppliedMigrations()
	if err != nil {
		t.Fatalf("Failed to get applied migrations: %v", err)
	}

	if len(appliedMigrations) != len(allMigrations) {
		t.Errorf("Expected %d applied migrations, got %d", len(allMigrations), len(appliedMigrations))
	}

	// Test migration status
	status, err := mm.GetMigrationStatus()
	if err != nil {
		t.Fatalf("Failed to get migration status: %v", err)
	}

	currentVersion, ok := status["current_version"].(int)
	if !ok || currentVersion == 0 {
		t.Error("Current version should be set after migrations")
	}

	appliedCount, ok := status["applied_count"].(int)
	if !ok || appliedCount != len(allMigrations) {
		t.Errorf("Expected applied count to be %d, got %d", len(allMigrations), appliedCount)
	}

	pendingCount, ok := status["pending_count"].(int)
	if !ok || pendingCount != 0 {
		t.Errorf("Expected pending count to be 0, got %d", pendingCount)
	}
}

func TestMigrationRollback(t *testing.T) {
	// Create in-memory database for testing
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

	mm := NewMigrationManager(db)

	// Apply all migrations first
	if err := mm.MigrateUp(); err != nil {
		t.Fatalf("Failed to migrate up: %v", err)
	}

	// Test rollback to version 2
	if err := mm.MigrateDown(2); err != nil {
		t.Fatalf("Failed to migrate down: %v", err)
	}

	// Check that only migrations 1 and 2 are applied
	appliedMigrations, err := mm.GetAppliedMigrations()
	if err != nil {
		t.Fatalf("Failed to get applied migrations: %v", err)
	}

	if len(appliedMigrations) != 2 {
		t.Errorf("Expected 2 applied migrations after rollback, got %d", len(appliedMigrations))
	}

	// Check that the highest version is 2
	if appliedMigrations[len(appliedMigrations)-1].Version != 2 {
		t.Errorf("Expected highest version to be 2, got %d", appliedMigrations[len(appliedMigrations)-1].Version)
	}

	// Test that pending migrations exist
	pendingMigrations, err := mm.GetPendingMigrations()
	if err != nil {
		t.Fatalf("Failed to get pending migrations: %v", err)
	}

	allMigrations := mm.GetMigrations()
	expectedPending := len(allMigrations) - 2
	if len(pendingMigrations) != expectedPending {
		t.Errorf("Expected %d pending migrations after rollback, got %d", expectedPending, len(pendingMigrations))
	}
}

func TestIndividualMigrationOperations(t *testing.T) {
	// Create in-memory database for testing
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

	mm := NewMigrationManager(db)

	// Initialize migration table
	if err := mm.InitializeMigrationTable(); err != nil {
		t.Fatalf("Failed to initialize migration table: %v", err)
	}

	// Get first migration
	allMigrations := mm.GetMigrations()
	if len(allMigrations) == 0 {
		t.Fatal("No migrations available for testing")
	}

	firstMigration := allMigrations[0]

	// Apply first migration
	if err := mm.ApplyMigration(firstMigration); err != nil {
		t.Fatalf("Failed to apply migration: %v", err)
	}

	// Check that it's recorded as applied
	appliedMigrations, err := mm.GetAppliedMigrations()
	if err != nil {
		t.Fatalf("Failed to get applied migrations: %v", err)
	}

	if len(appliedMigrations) != 1 {
		t.Errorf("Expected 1 applied migration, got %d", len(appliedMigrations))
	}

	if appliedMigrations[0].Version != firstMigration.Version {
		t.Errorf("Expected applied migration version %d, got %d", firstMigration.Version, appliedMigrations[0].Version)
	}

	// Rollback the migration
	if err := mm.RollbackMigration(firstMigration); err != nil {
		t.Fatalf("Failed to rollback migration: %v", err)
	}

	// Check that it's no longer applied
	appliedMigrations, err = mm.GetAppliedMigrations()
	if err != nil {
		t.Fatalf("Failed to get applied migrations after rollback: %v", err)
	}

	if len(appliedMigrations) != 0 {
		t.Errorf("Expected 0 applied migrations after rollback, got %d", len(appliedMigrations))
	}
}