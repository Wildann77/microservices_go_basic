package database

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/microservices-go/shared/logger"
)

// RunMigrations executes database migrations on startup.
// It assumes the migrations are stored in .sql files in the migrationsPath directory.
func RunMigrations(db *sql.DB, databaseName string, migrationsPath string) error {
	log := logger.New("database-migration")

	// Verify migrations path exists (simple check or defer to migrate lib)
	if migrationsPath == "" {
		migrationsPath = "migrations"
	}

	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("could not create postgres driver: %w", err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://"+migrationsPath,
		databaseName,
		driver,
	)
	if err != nil {
		return fmt.Errorf("could not create migrate instance: %w", err)
	}

	log.Info(fmt.Sprintf("Running migrations for %s from %s...", databaseName, migrationsPath))

	if err := m.Up(); err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			log.Info("No new migrations to run")
			return nil
		}
		return fmt.Errorf("could not run up migrations: %w", err)
	}

	log.Info("Migrations completed successfully")
	return nil
}
