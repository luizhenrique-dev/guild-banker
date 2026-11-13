//go:build integration && postgres

package webserver

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"github.com/luizhenrique-dev/guild-banker/api/internal/auth"
	"github.com/luizhenrique-dev/guild-banker/api/internal/infra/webserver/middleware"
	"github.com/luizhenrique-dev/guild-banker/api/internal/user"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type provisioningVerifier func(context.Context, string) (auth.Claims, error)

func (f provisioningVerifier) Verify(ctx context.Context, raw string) (auth.Claims, error) {
	return f(ctx, raw)
}

func TestProvisioningHTTPPostgres(t *testing.T) {
	dsn := os.Getenv("USER_TEST_DATABASE_URL")
	require.NotEmpty(t, dsn, "USER_TEST_DATABASE_URL is required")
	parsed, err := url.Parse(dsn)
	require.NoError(t, err)
	admin, err := sqlx.Connect("postgres", dsn)
	require.NoError(t, err)
	defer func() { require.NoError(t, admin.Close()) }()
	schema := fmt.Sprintf("issue3_http_%d", time.Now().UnixNano())
	_, err = admin.Exec("CREATE SCHEMA " + pq.QuoteIdentifier(schema))
	require.NoError(t, err)
	defer func() {
		_, err := admin.Exec("DROP SCHEMA " + pq.QuoteIdentifier(schema) + " CASCADE")
		require.NoError(t, err)
	}()
	params := parsed.Query()
	params.Set("search_path", schema)
	parsed.RawQuery = params.Encode()
	db, err := sqlx.Connect("postgres", parsed.String())
	require.NoError(t, err)
	defer func() { require.NoError(t, db.Close()) }()
	migrations, err := filepath.Glob("../../../migrations/*.up.sql")
	require.NoError(t, err)
	for _, path := range migrations {
		migration, err := os.ReadFile(path)
		require.NoError(t, err)
		_, err = db.Exec(string(migration))
		require.NoError(t, err)
	}
	storage := user.NewStorage(db)
	accounts := user.NewService(storage)
	claims := auth.Claims{UserID: "secret-sub", Name: "Secret Name", Email: "secret@example.test", EmailVerified: true}
	var verifyErr error
	verify := provisioningVerifier(func(ctx context.Context, raw string) (auth.Claims, error) {
		assert.Equal(t, "secret-token", raw)
		_, ok := ctx.Deadline()
		assert.True(t, ok)
		return claims, verifyErr
	})
	var logs bytes.Buffer
	router := gin.New()
	router.Use(middleware.Auth(auth.NewService(verify, accounts), slog.New(slog.NewJSONHandler(&logs, nil)), time.Second))
	var principal auth.Principal
	handled := 0
	router.GET("/api/probe", func(c *gin.Context) {
		handled++
		var ok bool
		principal, ok = auth.PrincipalFromContext(c.Request.Context())
		require.True(t, ok)
		c.Status(204)
	})
	request := func(status int, ctx context.Context) {
		t.Helper()
		logs.Reset()
		before := handled
		req := httptest.NewRequest("GET", "/api/probe", nil).WithContext(ctx)
		req.Header.Set("Authorization", "Bearer secret-token")
		req.Header.Set("X-User-ID", "999")
		req.Header.Set("X-User-Email", "spoof@example.test")
		out := httptest.NewRecorder()
		router.ServeHTTP(out, req)
		require.Equal(t, status, out.Code)
		if status == 204 {
			assert.Equal(t, before+1, handled)
		} else {
			assert.Equal(t, before, handled)
		}
		for _, sensitive := range []string{"secret-sub", "Secret Name", "secret@example.test", "secret-token", "user_account", "pq:"} {
			assert.NotContains(t, out.Body.String(), sensitive)
			assert.NotContains(t, logs.String(), sensitive)
		}
	}
	ctx := context.Background()
	request(204, ctx)
	local, err := storage.GetByExternalID(ctx, claims.UserID)
	require.NoError(t, err)
	assert.Equal(t, local.ID, principal.Actor.UserID)
	assert.Equal(t, local.Email, principal.Actor.Email)
	assert.Nil(t, local.UpdatedAt)
	var firstAccess time.Time
	require.NoError(t, db.Get(&firstAccess, "SELECT last_access_at FROM user_account WHERE id=$1", local.ID))
	request(204, ctx)
	var lastAccess time.Time
	require.NoError(t, db.Get(&lastAccess, "SELECT last_access_at FROM user_account WHERE id=$1", local.ID))
	assert.True(t, lastAccess.After(firstAccess))
	equal, err := storage.GetByID(ctx, local.ID)
	require.NoError(t, err)
	assert.Equal(t, local, equal)
	claims.Name = "Changed Secret"
	claims.Email = "changed@example.test"
	request(204, ctx)
	changed, err := storage.GetByID(ctx, local.ID)
	require.NoError(t, err)
	assert.Equal(t, claims.Name, changed.Name)
	assert.Equal(t, claims.Email, principal.Actor.Email)
	assert.Equal(t, local.ID, principal.Actor.UserID)
	assert.Equal(t, claims.Email, *changed.UpdatedBy)
	require.NoError(t, db.Get(&lastAccess, "SELECT last_access_at FROM user_account WHERE id=$1", local.ID))
	_, err = accounts.EnsureAccount(ctx, "other-sub", "Other", "other@example.test")
	require.NoError(t, err)
	claims.Email = "other@example.test"
	request(403, ctx)
	var rejectedAccess time.Time
	require.NoError(t, db.Get(&rejectedAccess, "SELECT last_access_at FROM user_account WHERE id=$1", local.ID))
	assert.Equal(t, lastAccess, rejectedAccess, "rejected requests must not record activity")
	unchanged, err := storage.GetByID(ctx, local.ID)
	require.NoError(t, err)
	assert.Equal(t, changed, unchanged)
	claims.UserID = "new-conflict"
	request(403, ctx)
	_, err = storage.GetByExternalID(ctx, claims.UserID)
	require.ErrorIs(t, err, user.ErrNotFound)
	claims.UserID = local.ExternalID
	claims.Email = changed.Email
	require.NoError(t, storage.Disable(ctx, local.ID, "admin@example.test"))
	request(403, ctx)
	disabled, err := storage.GetByID(ctx, local.ID)
	require.NoError(t, err)
	assert.Equal(t, changed.Name, disabled.Name)
	assert.False(t, disabled.Enabled)
	claims.UserID = "missing-profile"
	claims.Name = " "
	request(403, ctx)
	_, err = storage.GetByExternalID(ctx, claims.UserID)
	require.ErrorIs(t, err, user.ErrNotFound)
	claims.Name = "Secret Name"
	verifyErr = auth.ErrInvalidToken
	request(401, ctx)
	verifyErr = nil
	canceled, cancel := context.WithCancel(ctx)
	cancel()
	request(500, canceled)
	// An actual persistence failure must not release the handler or leak SQL details.
	require.NoError(t, db.Close())
	request(500, ctx)
}
