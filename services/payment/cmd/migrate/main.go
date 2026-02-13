package main

import (
	"flag"
	"fmt"
	"os"

	"database/sql"
	_ "github.com/lib/pq"
	"github.com/microservices-go/shared/config"
	"github.com/microservices-go/shared/logger"
	"github.com/microservices-go/shared/migrate"
)

func main() {
	var (
		action = flag.String("action", "up", "Migration action: up, down, status, version")
		path   = flag.String("path", "./migrations", "Path to migrations directory")
		force  = flag.Int("force", -1, "Force specific version (use with caution)")
	)
	flag.Parse()

	// Initialize logger
	log := logger.New("payment-migrate")

	// Load database config
	dbConfig := config.LoadDatabaseConfig("payment")

	// Open database connection
	db, err := sql.Open("postgres", dbConfig.DSN())
	if err != nil {
		log.Fatal("Failed to connect to database: " + err.Error())
	}
	defer db.Close()

	// Test connection
	if err := db.Ping(); err != nil {
		log.Fatal("Failed to ping database: " + err.Error())
	}

	// Create migrator
	migrator := migrate.NewMigrator(db, log.WithField("component", "migrator"))

	// Execute action
	switch *action {
	case "up":
		log.Info("Running migrations up...")
		if err := migrator.Up(*path); err != nil {
			log.Fatal("Migration failed: " + err.Error())
		}
		log.Info("Migrations completed successfully")

	case "down":
		log.Info("Rolling back one migration...")
		if err := migrator.Down(*path); err != nil {
			log.Fatal("Rollback failed: " + err.Error())
		}
		log.Info("Rollback completed successfully")

	case "status":
		log.Info("Checking migration status...")
		if err := migrator.Status(*path); err != nil {
			log.Fatal("Failed to get status: " + err.Error())
		}

	case "version":
		version, dirty, err := migrator.Version(*path)
		if err != nil {
			if err.Error() == "no migration" {
				fmt.Println("No migrations have been run yet")
				os.Exit(0)
			}
			log.Fatal("Failed to get version: " + err.Error())
		}
		fmt.Printf("Current version: %d, Dirty: %v\n", version, dirty)

	case "force":
		if *force < 0 {
			fmt.Println("Usage: migrate -action=force -force=<version>")
			os.Exit(1)
		}
		log.Infof("Forcing migration version to %d...", *force)
		if err := migrator.Force(*path, *force); err != nil {
			log.Fatal("Force failed: " + err.Error())
		}
		log.Info("Force completed successfully")

	default:
		fmt.Printf("Unknown action: %s\n", *action)
		fmt.Println("Usage: migrate -action=[up|down|status|version|force]")
		os.Exit(1)
	}
}
