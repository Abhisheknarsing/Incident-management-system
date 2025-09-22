package database

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"sync"
	"time"

	_ "github.com/marcboeker/go-duckdb"
)

// DB represents the database connection pool
type DB struct {
	conn     *sql.DB
	mu       sync.RWMutex
	isReady  bool
	dbPath   string
}

// Config holds database configuration
type Config struct {
	DatabasePath    string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
	ConnMaxIdleTime time.Duration
}

// DefaultConfig returns default database configuration
func DefaultConfig() *Config {
	return &Config{
		DatabasePath:    "./data/incidents.db",
		MaxOpenConns:    25,
		MaxIdleConns:    5,
		ConnMaxLifetime: time.Hour,
		ConnMaxIdleTime: time.Minute * 10,
	}
}

// NewDB creates a new database connection with the given configuration
func NewDB(config *Config) (*DB, error) {
	if config == nil {
		config = DefaultConfig()
	}

	db := &DB{
		dbPath: config.DatabasePath,
	}

	if err := db.connect(config); err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	return db, nil
}

// connect establishes the database connection with connection pooling
func (db *DB) connect(config *Config) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	conn, err := sql.Open("duckdb", config.DatabasePath)
	if err != nil {
		return fmt.Errorf("failed to open database connection: %w", err)
	}

	// Configure connection pool
	conn.SetMaxOpenConns(config.MaxOpenConns)
	conn.SetMaxIdleConns(config.MaxIdleConns)
	conn.SetConnMaxLifetime(config.ConnMaxLifetime)
	conn.SetConnMaxIdleTime(config.ConnMaxIdleTime)

	// Test the connection
	if err := conn.Ping(); err != nil {
		conn.Close()
		return fmt.Errorf("failed to ping database: %w", err)
	}

	db.conn = conn
	db.isReady = true

	log.Printf("Database connection established: %s", config.DatabasePath)
	return nil
}

// GetConnection returns the database connection
func (db *DB) GetConnection() *sql.DB {
	db.mu.RLock()
	defer db.mu.RUnlock()
	return db.conn
}

// IsReady returns true if the database connection is ready
func (db *DB) IsReady() bool {
	db.mu.RLock()
	defer db.mu.RUnlock()
	return db.isReady
}

// Close closes the database connection
func (db *DB) Close() error {
	db.mu.Lock()
	defer db.mu.Unlock()

	if db.conn != nil {
		err := db.conn.Close()
		db.conn = nil
		db.isReady = false
		log.Printf("Database connection closed: %s", db.dbPath)
		return err
	}
	return nil
}

// HealthCheck performs a health check on the database connection
func (db *DB) HealthCheck() error {
	db.mu.RLock()
	conn := db.conn
	ready := db.isReady
	db.mu.RUnlock()

	if !ready || conn == nil {
		return fmt.Errorf("database connection not ready")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := conn.PingContext(ctx); err != nil {
		return fmt.Errorf("database health check failed: %w", err)
	}

	return nil
}

// Stats returns database connection statistics
func (db *DB) Stats() sql.DBStats {
	db.mu.RLock()
	defer db.mu.RUnlock()

	if db.conn == nil {
		return sql.DBStats{}
	}

	return db.conn.Stats()
}