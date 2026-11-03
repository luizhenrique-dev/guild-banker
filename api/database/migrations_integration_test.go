//go:build integration && postgres

package database

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMigratePreservesApplicationPool(t *testing.T) {
	dsn := os.Getenv("USER_TEST_DATABASE_URL")
	require.NotEmpty(t, dsn, "USER_TEST_DATABASE_URL is required")
	parsed, err := url.Parse(dsn)
	require.NoError(t, err)
	db, err := sqlx.Connect("postgres", dsn)
	require.NoError(t, err)
	defer func() { require.NoError(t, db.Close()) }()
	schema := fmt.Sprintf("issue3_migrate_%d", time.Now().UnixNano())
	_, err = db.Exec("CREATE SCHEMA " + pq.QuoteIdentifier(schema))
	require.NoError(t, err)
	defer func() {
		_, err := db.Exec("DROP SCHEMA " + pq.QuoteIdentifier(schema) + " CASCADE")
		require.NoError(t, err)
	}()
	params := parsed.Query()
	params.Set("search_path", schema)
	parsed.RawQuery = params.Encode()
	app, err := sqlx.Connect("postgres", parsed.String())
	require.NoError(t, err)
	defer func() { require.NoError(t, app.Close()) }()
	migration, err := os.ReadFile("../migrations/000001_create_users_table.up.sql")
	require.NoError(t, err)
	temp := t.TempDir()
	require.NoError(t, os.Mkdir(filepath.Join(temp, "migrations"), 0700))
	require.NoError(t, os.WriteFile(filepath.Join(temp, "migrations", "000001_initial.up.sql"), migration, 0600))
	t.Chdir(temp)
	require.NoError(t, Migrate(app, parsed.Path[1:]))
	require.NoError(t, app.PingContext(context.Background()), "migration cleanup must not close the application pool")
	require.NoError(t, Migrate(app, parsed.Path[1:]), "no-change path must preserve pool too")
	require.NoError(t, app.Ping())
	assert.Zero(t, app.Stats().InUse)
}
