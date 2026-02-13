package migrate

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"

	"github.com/microservices-go/shared/errors"
	"github.com/microservices-go/shared/logger"
)

// Migrator handles database migrations
type Migrator struct {
	db     *sql.DB
	logger *logger.Logger
}

// NewMigrator creates a new migrator instance
func NewMigrator(db *sql.DB, logger *logger.Logger) *Migrator {
	return &Migrator{
		db:     db,
		logger: logger,
	}
}

// Run executes all pending migrations up to the latest version
func (m *Migrator) Run(migrationsPath string) error {
	return m.Up(migrationsPath)
}

// Up runs all pending migrations
func (m *Migrator) Up(migrationsPath string) error {
	m.logger.Infof("Running database migrations from: %s", migrationsPath)

	// Ensure absolute path
	absPath, err := filepath.Abs(migrationsPath)
	if err != nil {
		return errors.Wrap(err, errors.ErrInternalServer, "Failed to resolve migrations path")
	}

	// Check if path exists
	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		return errors.New(errors.ErrNotFound, fmt.Sprintf("Migrations path not found: %s", absPath))
	}

	// Create postgres instance
	driver, err := postgres.WithInstance(m.db, &postgres.Config{})
	if err != nil {
		return errors.Wrap(err, errors.ErrDatabaseError, "Failed to create migration driver")
	}

	// Create migrate instance
	migrator, err := migrate.NewWithDatabaseInstance(
		fmt.Sprintf("file://%s", absPath),
		"postgres",
		driver,
	)
	if err != nil {
		return errors.Wrap(err, errors.ErrDatabaseError, "Failed to create migrator")
	}

	// Run migrations
	if err := migrator.Up(); err != nil {
		if err == migrate.ErrNoChange {
			m.logger.Info("No migrations to run - database is up to date")
			return nil
		}
		return errors.Wrap(err, errors.ErrDatabaseError, "Failed to run migrations")
	}

	// Get current version
	version, dirty, err := migrator.Version()
	if err != nil && err != migrate.ErrNilVersion {
		m.logger.WithError(err).Warn("Failed to get migration version")
	}

	m.logger.Infof("Database migrations completed successfully - version: %d, dirty: %v", version, dirty)

	return nil
}

// Down rolls back one migration
func (m *Migrator) Down(migrationsPath string) error {
	m.logger.Infof("Rolling back one migration from: %s", migrationsPath)

	absPath, err := filepath.Abs(migrationsPath)
	if err != nil {
		return errors.Wrap(err, errors.ErrInternalServer, "Failed to resolve migrations path")
	}

	driver, err := postgres.WithInstance(m.db, &postgres.Config{})
	if err != nil {
		return errors.Wrap(err, errors.ErrDatabaseError, "Failed to create migration driver")
	}

	migrator, err := migrate.NewWithDatabaseInstance(
		fmt.Sprintf("file://%s", absPath),
		"postgres",
		driver,
	)
	if err != nil {
		return errors.Wrap(err, errors.ErrDatabaseError, "Failed to create migrator")
	}

	if err := migrator.Steps(-1); err != nil {
		if err == migrate.ErrNoChange {
			m.logger.Info("No migrations to rollback")
			return nil
		}
		return errors.Wrap(err, errors.ErrDatabaseError, "Failed to rollback migration")
	}

	m.logger.Info("Migration rolled back successfully")
	return nil
}

// Version returns current migration version
func (m *Migrator) Version(migrationsPath string) (version uint, dirty bool, err error) {
	absPath, err := filepath.Abs(migrationsPath)
	if err != nil {
		return 0, false, errors.Wrap(err, errors.ErrInternalServer, "Failed to resolve migrations path")
	}

	driver, err := postgres.WithInstance(m.db, &postgres.Config{})
	if err != nil {
		return 0, false, errors.Wrap(err, errors.ErrDatabaseError, "Failed to create migration driver")
	}

	migrator, err := migrate.NewWithDatabaseInstance(
		fmt.Sprintf("file://%s", absPath),
		"postgres",
		driver,
	)
	if err != nil {
		return 0, false, errors.Wrap(err, errors.ErrDatabaseError, "Failed to create migrator")
	}

	return migrator.Version()
}

// Status prints current migration status
func (m *Migrator) Status(migrationsPath string) error {
	version, dirty, err := m.Version(migrationsPath)
	if err != nil {
		if err == migrate.ErrNilVersion {
			m.logger.Info("No migrations have been run yet")
			return nil
		}
		return err
	}

	m.logger.Infof("Current migration status - version: %d, dirty: %v", version, dirty)

	return nil
}

// Force sets migration version without running migrations (use with caution)
func (m *Migrator) Force(migrationsPath string, version int) error {
	m.logger.Warnf("Forcing migration version to: %d", version)

	absPath, err := filepath.Abs(migrationsPath)
	if err != nil {
		return errors.Wrap(err, errors.ErrInternalServer, "Failed to resolve migrations path")
	}

	driver, err := postgres.WithInstance(m.db, &postgres.Config{})
	if err != nil {
		return errors.Wrap(err, errors.ErrDatabaseError, "Failed to create migration driver")
	}

	migrator, err := migrate.NewWithDatabaseInstance(
		fmt.Sprintf("file://%s", absPath),
		"postgres",
		driver,
	)
	if err != nil {
		return errors.Wrap(err, errors.ErrDatabaseError, "Failed to create migrator")
	}

	if err := migrator.Force(version); err != nil {
		return errors.Wrap(err, errors.ErrDatabaseError, "Failed to force migration version")
	}

	m.logger.Infof("Migration version forced successfully to: %d", version)
	return nil
}
