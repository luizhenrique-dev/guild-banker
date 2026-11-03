package database

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

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
func Migrate(db *sqlx.DB, dbName string) (err error) {
	migrationsPath, err := resolveMigrationsPath()
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	conn, err := db.Conn(ctx)
	if err != nil {
		return fmt.Errorf("migrations: acquire connection: %w", err)
	}
	// The application owns the pool; the migration driver owns only this connection.
	driver, err := postgres.WithConnection(ctx, conn, &postgres.Config{DatabaseName: dbName})
	if err != nil {
		return fmt.Errorf("migrations: initialize driver: %w", errors.Join(err, conn.Close()))
	}
	m, err := migrate.NewWithDatabaseInstance(migrationsPath, dbName, driver)
	if err != nil {
		return fmt.Errorf("migrations: initialize: %w", errors.Join(err, driver.Close()))
	}
	defer func() {
		sourceErr, driverErr := m.Close()
		if closeErr := errors.Join(sourceErr, driverErr); closeErr != nil {
			err = errors.Join(err, fmt.Errorf("migrations: close: %w", closeErr))
		}
	}()
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
