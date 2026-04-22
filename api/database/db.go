package database

import (
	"context"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"

	"github.com/luizhenrique-dev/guild-banker/api/config"
)

const driver = "postgres"

func New(ctx context.Context, cfg config.DBConfig) (*sqlx.DB, error) {
	db, err := sqlx.Open(driver, cfg.DSN())
	if err != nil {
		return nil, fmt.Errorf("database: failed to open connection: %w", err)
	}

	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetMaxIdleConns(cfg.MaxIdleConns)
	db.SetConnMaxLifetime(time.Duration(cfg.ConnMaxLifetimeMinutes) * time.Minute)

	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("database: ping failed: %w", err)
	}

	return db, nil
}
