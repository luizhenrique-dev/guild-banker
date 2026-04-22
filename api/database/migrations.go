package database

import (
	"errors"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jmoiron/sqlx"
)

const migrationsPath = "file://migrations"

// Migrate executes all pending migrations (Up).
func Migrate(db *sqlx.DB, dbName string) error {
	driver, err := postgres.WithInstance(db.DB, &postgres.Config{
		DatabaseName: dbName,
	})
	if err != nil {
		return fmt.Errorf("migrations: failed to create driver: %w", err)
	}

	m, err := migrate.NewWithDatabaseInstance(migrationsPath, dbName, driver)
	if err != nil {
		return fmt.Errorf("migrations: failed to initialize: %w", err)
	}
	defer func(m *migrate.Migrate) {
		err, _ := m.Close()
		if err != nil {

		}
	}(m)

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("migrations: failed to run: %w", err)
	}

	return nil
}
