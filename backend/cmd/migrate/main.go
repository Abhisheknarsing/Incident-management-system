package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"

	"incident-management-system/internal/database"
)

func main() {
	var (
		dbPath    = flag.String("db", "./data/incidents.db", "Database file path")
		command   = flag.String("cmd", "up", "Migration command: up, down, status, reset")
		version   = flag.String("version", "", "Target version for down migration")
		help      = flag.Bool("help", false, "Show help")
	)
	flag.Parse()

	if *help {
		showHelp()
		return
	}

	// Create database connection
	config := &database.Config{
		DatabasePath: *dbPath,
	}

	db, err := database.NewDB(config)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	mm := database.NewMigrationManager(db)

	switch *command {
	case "up":
		if err := mm.MigrateUp(); err != nil {
			log.Fatalf("Migration failed: %v", err)
		}
		fmt.Println("Migrations completed successfully")

	case "down":
		if *version == "" {
			log.Fatal("Version is required for down migration. Use -version=N")
		}
		targetVersion, err := strconv.Atoi(*version)
		if err != nil {
			log.Fatalf("Invalid version number: %v", err)
		}
		if err := mm.MigrateDown(targetVersion); err != nil {
			log.Fatalf("Rollback failed: %v", err)
		}
		fmt.Printf("Rollback to version %d completed successfully\n", targetVersion)

	case "status":
		status, err := mm.GetMigrationStatus()
		if err != nil {
			log.Fatalf("Failed to get migration status: %v", err)
		}
		
		jsonData, err := json.MarshalIndent(status, "", "  ")
		if err != nil {
			log.Fatalf("Failed to format status: %v", err)
		}
		fmt.Println(string(jsonData))

	case "reset":
		fmt.Print("This will drop all tables and recreate the schema. Are you sure? (y/N): ")
		var response string
		fmt.Scanln(&response)
		if response != "y" && response != "Y" {
			fmt.Println("Reset cancelled")
			return
		}
		
		if err := db.ResetDatabase(); err != nil {
			log.Fatalf("Reset failed: %v", err)
		}
		fmt.Println("Database reset completed successfully")

	default:
		fmt.Printf("Unknown command: %s\n", *command)
		showHelp()
		os.Exit(1)
	}
}

func showHelp() {
	fmt.Println("Database Migration Tool")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  migrate [options]")
	fmt.Println()
	fmt.Println("Options:")
	fmt.Println("  -db string")
	fmt.Println("        Database file path (default \"./data/incidents.db\")")
	fmt.Println("  -cmd string")
	fmt.Println("        Migration command: up, down, status, reset (default \"up\")")
	fmt.Println("  -version string")
	fmt.Println("        Target version for down migration")
	fmt.Println("  -help")
	fmt.Println("        Show this help message")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  up      Apply all pending migrations")
	fmt.Println("  down    Rollback to specified version")
	fmt.Println("  status  Show current migration status")
	fmt.Println("  reset   Drop all tables and recreate schema")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  migrate -cmd=up")
	fmt.Println("  migrate -cmd=down -version=2")
	fmt.Println("  migrate -cmd=status")
	fmt.Println("  migrate -cmd=reset")
}