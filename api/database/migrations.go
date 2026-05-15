package database

import (
	"errors"
	"fmt"
	"os"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jmoiron/sqlx"
)

const (
	localMigrationsPath     = "migrations"
	workspaceMigrationsPath = "api/migrations"
)

// Migrate executes all pending migrations (Up).
func Migrate(db *sqlx.DB, dbName string) error {
	driver, err := postgres.WithInstance(db.DB, &postgres.Config{
		DatabaseName: dbName,
	})
	if err != nil {
		return fmt.Errorf("migrations: failed to create driver: %w", err)
	}

	migrationsPath, err := resolveMigrationsPath()
	if err != nil {
		return err
	}

	m, err := migrate.NewWithDatabaseInstance(migrationsPath, dbName, driver)
	if err != nil {
		return fmt.Errorf("migrations: failed to initialize: %w", err)
	}
	defer func(m *migrate.Migrate) {
		closeErr, _ := m.Close()
		if closeErr != nil && err == nil {
			err = fmt.Errorf("migrations: failed to close: %w", closeErr)
		}
	}(m)

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("migrations: failed to run: %w", err)
	}

	return nil
}

func resolveMigrationsPath() (string, error) {
	if _, err := os.Stat(localMigrationsPath); err == nil {
		return "file://" + localMigrationsPath, nil
	}

	if _, err := os.Stat(workspaceMigrationsPath); err == nil {
		return "file://" + workspaceMigrationsPath, nil
	}

	return "", fmt.Errorf(
		"migrations: directory not found (expected %q or %q)",
		localMigrationsPath,
		workspaceMigrationsPath,
	)
}
