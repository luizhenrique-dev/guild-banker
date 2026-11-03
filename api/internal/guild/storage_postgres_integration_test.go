//go:build integration && postgres

package guild

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"github.com/stretchr/testify/require"
)

func TestStorageCreatePostgres(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	dsn := os.Getenv("GUILD_TEST_DATABASE_URL")
	require.NotEmpty(t, dsn, "GUILD_TEST_DATABASE_URL must point to an isolated PostgreSQL 17+ test database")
	parsed, err := url.Parse(dsn)
	require.True(t, err == nil, "invalid test database URL")
	require.Contains(t, []string{"postgres", "postgresql"}, parsed.Scheme)
	admin, err := sqlx.ConnectContext(ctx, "postgres", dsn)
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, admin.Close()) })
	schema := fmt.Sprintf("issue4_review_%d", time.Now().UnixNano())
	_, err = admin.ExecContext(ctx, "CREATE SCHEMA "+pq.QuoteIdentifier(schema))
	require.NoError(t, err)
	defer func() {
		_, cleanupErr := admin.ExecContext(context.Background(), "DROP SCHEMA "+pq.QuoteIdentifier(schema)+" CASCADE")
		require.NoError(t, cleanupErr)
	}()
	params := parsed.Query()
	params.Set("search_path", schema)
	parsed.RawQuery = params.Encode()
	db, err := sqlx.ConnectContext(ctx, "postgres", parsed.String())
	require.NoError(t, err)
	defer func() { require.NoError(t, db.Close()) }()
	migration, err := os.ReadFile("../../migrations/000001_create_users_table.up.sql")
	require.NoError(t, err)
	_, err = db.ExecContext(ctx, string(migration))
	require.NoError(t, err)
	var id int64
	err = db.GetContext(ctx, &id, `INSERT INTO user_account(external_id,name,email,created_by) VALUES ('review-probe','Review','review@example.test','review') RETURNING id`)
	require.NoError(t, err)
	g, err := New("review-probe", "Review probe", "review@example.test")
	require.NoError(t, err)
	err = NewStorage(db).Create(ctx, g, id)
	require.NoError(t, err, "guild creation must persist both the guild and its creator membership")
	require.Positive(t, g.ID)
	storage := NewStorage(db)
	persisted, err := storage.GetByID(ctx, g.ID)
	require.NoError(t, err)
	require.Equal(t, g.Name, persisted.Name)
	member, err := storage.IsMember(ctx, g.ID, id)
	require.NoError(t, err)
	require.True(t, member)
	failed, err := New("rollback-probe", "Rollback probe", "review@example.test")
	require.NoError(t, err)
	require.Error(t, storage.Create(ctx, failed, id+1000))
	exists, err := storage.NameExists(ctx, failed.Name, 0)
	require.NoError(t, err)
	require.False(t, exists, "membership failure must roll back guild creation")
}
