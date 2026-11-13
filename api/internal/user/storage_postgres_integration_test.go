//go:build integration && postgres

package user

import (
	"context"
	"database/sql"
	"fmt"
	"net/url"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"github.com/luizhenrique-dev/guild-banker/api/internal/emailaddress"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func postgresTestDB(t *testing.T) *sqlx.DB {
	t.Helper()
	dsn := os.Getenv("USER_TEST_DATABASE_URL")
	require.NotEmpty(t, dsn, "USER_TEST_DATABASE_URL must point to an isolated PostgreSQL 17+ test database")
	parsed, err := url.Parse(dsn)
	require.NoError(t, err)
	require.Contains(t, []string{"postgres", "postgresql"}, parsed.Scheme, "use a PostgreSQL URL")
	admin, err := sqlx.Connect("postgres", dsn)
	require.NoError(t, err, "connect to test database")
	t.Cleanup(func() { require.NoError(t, admin.Close()) })
	schema := fmt.Sprintf("issue3_%d", time.Now().UnixNano())
	_, err = admin.Exec("CREATE SCHEMA " + pq.QuoteIdentifier(schema))
	require.NoError(t, err)
	t.Cleanup(func() {
		_, err := admin.Exec("DROP SCHEMA " + pq.QuoteIdentifier(schema) + " CASCADE")
		require.NoError(t, err)
	})
	params := parsed.Query()
	params.Set("search_path", schema)
	parsed.RawQuery = params.Encode()
	db, err := sqlx.Connect("postgres", parsed.String())
	require.NoError(t, err)
	db.SetMaxOpenConns(16)
	db.SetMaxIdleConns(16)
	t.Cleanup(func() { require.NoError(t, db.Close()) })
	migration, err := os.ReadFile("../../migrations/000001_create_users_table.up.sql")
	require.NoError(t, err)
	_, err = db.Exec(string(migration))
	require.NoError(t, err)
	var version int
	require.NoError(t, db.Get(&version, "SELECT current_setting('server_version_num')::integer"))
	require.GreaterOrEqual(t, version, 170000)
	return db
}

func rowVersion(t *testing.T, db *sqlx.DB, id int64) string {
	t.Helper()
	var version string
	require.NoError(t, db.Get(&version, "SELECT xmin::text FROM user_account WHERE id=$1", id))
	return version
}

func TestStorageCanonicalEmailPostgres(t *testing.T) {
	db := postgresTestDB(t)
	storage := NewStorage(db)
	u, err := New("Member", ` MEMBER@example.test `, "member", "fixture")
	require.NoError(t, err)
	require.NoError(t, storage.Create(t.Context(), u))
	require.Equal(t, "member@example.test", u.Email)
	persisted, err := storage.GetByID(t.Context(), u.ID)
	require.NoError(t, err)
	require.Equal(t, u.Email, persisted.Email)

	u.Sync("Updated", `UPDATED@example.test`, "fixture")
	require.NoError(t, storage.Update(t.Context(), u))
	found, err := storage.GetByEmail(t.Context(), ` Updated@EXAMPLE.test `)
	require.NoError(t, err)
	require.Equal(t, u.ID, found.ID)
	require.Equal(t, "updated@example.test", found.Email)

	candidate, err := New("Other", `OTHER@example.test`, "other", "fixture")
	require.NoError(t, err)
	other, err := storage.CreateIfAbsent(t.Context(), candidate)
	require.NoError(t, err)
	require.Equal(t, "other@example.test", other.Email)
	other.Sync("Changed", ` CHANGED@example.test `, "fixture")
	changed, err := storage.UpdateProfile(t.Context(), other)
	require.NoError(t, err)
	require.Equal(t, "changed@example.test", changed.Email)
	version := rowVersion(t, db, changed.ID)
	changed, err = NewService(storage).EnsureAccount(t.Context(), changed.ExternalID, changed.Name, `CHANGED@example.test`)
	require.NoError(t, err)
	require.Equal(t, "changed@example.test", changed.Email)
	require.Equal(t, version, rowVersion(t, db, changed.ID), "equivalent claims must not rewrite the profile")

	_, err = NewService(storage).EnsureAccount(t.Context(), "duplicate", "Duplicate", `CHANGED@example.test`)
	require.ErrorIs(t, err, ErrEmailConflict)
	_, err = NewService(storage).EnsureAccount(t.Context(), "invalid", "Invalid", "private invalid email")
	require.ErrorIs(t, err, emailaddress.ErrInvalid)

	u.Email = "private invalid email"
	require.ErrorIs(t, storage.Update(t.Context(), u), emailaddress.ErrInvalid)
	found, err = storage.GetByID(t.Context(), u.ID)
	require.NoError(t, err)
	require.Equal(t, "updated@example.test", found.Email)
}

func TestNormalizeStoredEmailsPostgres(t *testing.T) {
	for _, tt := range []struct {
		name, second string
		wantErr      error
	}{
		{name: "normalizes and is idempotent", second: "other@example.test"},
		{name: "collision rolls back", second: "member@example.test", wantErr: ErrEmailConflict},
		{name: "keeps unsupported legacy syntax", second: ` "OTHER"@EXAMPLE.test `},
	} {
		t.Run(tt.name, func(t *testing.T) {
			db := postgresTestDB(t)
			var firstID int64
			require.NoError(t, db.Get(&firstID, `INSERT INTO user_account(external_id,name,email,created_by) VALUES ('first','First',' MEMBER@example.test ','fixture') RETURNING id`))
			_, err := db.Exec(`INSERT INTO user_account(external_id,name,email,created_by) VALUES ('second','Second',$1,'fixture')`, tt.second)
			require.NoError(t, err)
			version := rowVersion(t, db, firstID)
			err = NormalizeStoredEmails(t.Context(), db)
			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				require.NotContains(t, err.Error(), tt.second)
				require.Equal(t, version, rowVersion(t, db, firstID))
				return
			}
			require.NoError(t, err)
			u, err := NewStorage(db).GetByID(t.Context(), firstID)
			require.NoError(t, err)
			require.Equal(t, "member@example.test", u.Email)
			var secondEmail string
			require.NoError(t, db.Get(&secondEmail, "SELECT email FROM user_account WHERE external_id='second'"))
			require.Equal(t, strings.ToLower(strings.TrimSpace(tt.second)), secondEmail)
			version = rowVersion(t, db, firstID)
			require.NoError(t, NormalizeStoredEmails(t.Context(), db))
			require.Equal(t, version, rowVersion(t, db, firstID))
		})
	}
}

func TestStorageEnsureAccountPostgres(t *testing.T) {
	db := postgresTestDB(t)
	storage := NewStorage(db)
	svc := NewService(storage)
	ctx := context.Background()
	u, err := svc.EnsureAccount(ctx, "opaque-sub", "Member", "member@example.test")
	require.NoError(t, err)
	require.Positive(t, u.ID)
	assert.True(t, u.Enabled)
	assert.Equal(t, "opaque-sub", u.ExternalID)
	assert.Equal(t, u.Email, u.CreatedBy)
	assert.Nil(t, u.UpdatedAt)
	assert.Nil(t, u.UpdatedBy)
	original := *u
	version := rowVersion(t, db, u.ID)
	equal, err := svc.EnsureAccount(ctx, u.ExternalID, u.Name, u.Email)
	require.NoError(t, err)
	assert.Equal(t, original, *equal)
	assert.Equal(t, version, rowVersion(t, db, u.ID))
	for _, profile := range [][2]string{{"New Name", u.Email}, {"New Name", "new@example.test"}, {"Both", "NEW@example.test"}} {
		u, err = svc.EnsureAccount(ctx, u.ExternalID, profile[0], profile[1])
		require.NoError(t, err)
		assert.Equal(t, original.ID, u.ID)
		assert.Equal(t, original.CreatedAt, u.CreatedAt)
		assert.Equal(t, original.CreatedBy, u.CreatedBy)
		assert.Equal(t, profile[0], u.Name)
		assert.Equal(t, strings.ToLower(profile[1]), u.Email)
		require.NotNil(t, u.UpdatedBy)
		assert.Equal(t, u.Email, *u.UpdatedBy)
	}
	other, err := svc.EnsureAccount(ctx, "other-sub", "Other", "other@example.test")
	require.NoError(t, err)
	before, err := storage.GetByID(ctx, u.ID)
	require.NoError(t, err)
	_, err = svc.EnsureAccount(ctx, "third-sub", "Intruder", other.Email)
	require.ErrorIs(t, err, ErrEmailConflict)
	_, err = svc.EnsureAccount(ctx, u.ExternalID, "Intruder", other.Email)
	require.ErrorIs(t, err, ErrEmailConflict)
	after, err := storage.GetByID(ctx, u.ID)
	require.NoError(t, err)
	assert.Equal(t, before, after)
	afterOther, err := storage.GetByID(ctx, other.ID)
	require.NoError(t, err)
	assert.Equal(t, other, afterOther)
	var count int
	require.NoError(t, db.Get(&count, "SELECT count(*) FROM user_account"))
	assert.Equal(t, 2, count)
	require.NoError(t, storage.Disable(ctx, u.ID, "admin@example.test"))
	disabled, err := storage.GetByID(ctx, u.ID)
	require.NoError(t, err)
	assert.Equal(t, "admin@example.test", *disabled.DisabledBy)
	assert.Equal(t, u.Email, *disabled.UpdatedBy)
	_, err = svc.EnsureAccount(ctx, u.ExternalID, "Blocked", "blocked@example.test")
	require.ErrorIs(t, err, ErrDisabled)
	unchanged, err := storage.GetByID(ctx, u.ID)
	require.NoError(t, err)
	assert.Equal(t, disabled, unchanged)
	canceled, cancel := context.WithCancel(ctx)
	cancel()
	_, err = svc.EnsureAccount(canceled, "canceled-sub", "Canceled", "cancel@example.test")
	require.ErrorIs(t, err, context.Canceled)
}

func TestStorageEnsureAccountConcurrent(t *testing.T) {
	db := postgresTestDB(t)
	storage := NewStorage(db)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	const workers = 8
	start := make(chan struct{})
	results := make(chan *User, workers)
	failures := make(chan error, workers)
	var wg sync.WaitGroup
	for range workers {
		wg.Go(func() {
			<-start
			u, err := NewService(storage).EnsureAccount(ctx, "shared-sub", "Member", "member@example.test")
			results <- u
			failures <- err
		})
	}
	close(start)
	wg.Wait()
	close(results)
	close(failures)
	for err := range failures {
		require.NoError(t, err)
	}
	var id int64
	for u := range results {
		require.NotNil(t, u)
		if id == 0 {
			id = u.ID
		}
		assert.Equal(t, id, u.ID)
	}
	var count int
	require.NoError(t, db.Get(&count, "SELECT count(*) FROM user_account"))
	assert.Equal(t, 1, count)
	old, err := storage.GetByID(ctx, id)
	require.NoError(t, err)
	// Both candidates represent the same old snapshot, so the second SQL must suppress its write.
	first, second := *old, *old
	first.Sync("Changed", "changed@example.test", "changed@example.test")
	second.Sync(first.Name, first.Email, first.Email)
	first.UpdatedAt = new(time.Unix(1800000000, 0))
	second.UpdatedAt = new(time.Unix(1800000060, 0))
	start = make(chan struct{})
	updates := make(chan *User, 2)
	errors := make(chan error, 2)
	for _, candidate := range []*User{&first, &second} {
		wg.Go(func() { <-start; u, err := storage.UpdateProfile(ctx, candidate); updates <- u; errors <- err })
	}
	close(start)
	wg.Wait()
	require.NoError(t, <-errors)
	require.NoError(t, <-errors)
	one, two := <-updates, <-updates
	assert.Equal(t, one, two)
	version := rowVersion(t, db, id)
	_, err = storage.UpdateProfile(ctx, &second)
	require.NoError(t, err)
	assert.Equal(t, version, rowVersion(t, db, id))
	// A single write selected one candidate's timestamp; a redundant write would return distinct audits.
	require.NotNil(t, one.UpdatedAt)
	require.NotNil(t, two.UpdatedAt)
	assert.Equal(t, *one.UpdatedAt, *two.UpdatedAt)
}

func TestStorageProfileDisableRace(t *testing.T) {
	db := postgresTestDB(t)
	storage := NewStorage(db)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	u, err := NewService(storage).EnsureAccount(ctx, "race-sub", "Before", "before@example.test")
	require.NoError(t, err)
	candidate := *u
	candidate.Sync("After", "after@example.test", "after@example.test")
	tx, err := db.BeginTxx(ctx, &sql.TxOptions{})
	require.NoError(t, err)
	defer func() { _ = tx.Rollback() }()
	_, err = tx.ExecContext(ctx, "UPDATE user_account SET enabled=false, disabled_at=NOW(), disabled_by='admin@example.test' WHERE id=$1", u.ID)
	require.NoError(t, err)
	done := make(chan error, 1)
	var wg sync.WaitGroup
	wg.Go(func() { _, err := storage.UpdateProfile(ctx, &candidate); done <- err })
	// Wait until a separate backend is blocked on our uncommitted row, not a scheduler sleep.
	for {
		var waiting bool
		require.NoError(t, db.GetContext(ctx, &waiting, "SELECT EXISTS(SELECT 1 FROM pg_stat_activity WHERE datname=current_database() AND wait_event_type='Lock' AND query LIKE 'UPDATE user_account%')"))
		if waiting {
			break
		}
		require.NoError(t, ctx.Err())
	}
	require.NoError(t, tx.Commit())
	wg.Wait()
	require.ErrorIs(t, <-done, ErrDisabled)
	after, err := storage.GetByID(ctx, u.ID)
	require.NoError(t, err)
	assert.False(t, after.Enabled)
	assert.Equal(t, u.Name, after.Name)
	assert.Equal(t, u.Email, after.Email)
	assert.Nil(t, after.UpdatedBy)
	assert.Nil(t, after.UpdatedAt)
	assert.Equal(t, "admin@example.test", *after.DisabledBy)
}
