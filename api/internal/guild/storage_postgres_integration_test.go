//go:build integration && postgres

package guild

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/luizhenrique-dev/guild-banker/api/internal/audit"
	"github.com/luizhenrique-dev/guild-banker/api/internal/auth"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"github.com/stretchr/testify/require"
)

func guildTestDB(t *testing.T) *sqlx.DB {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	t.Cleanup(cancel)
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
	t.Cleanup(func() {
		_, cleanupErr := admin.ExecContext(context.Background(), "DROP SCHEMA "+pq.QuoteIdentifier(schema)+" CASCADE")
		require.NoError(t, cleanupErr)
	})
	params := parsed.Query()
	params.Set("search_path", schema)
	parsed.RawQuery = params.Encode()
	db, err := sqlx.ConnectContext(ctx, "postgres", parsed.String())
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, db.Close()) })
	migrations, err := filepath.Glob("../../migrations/*.up.sql")
	require.NoError(t, err)
	for _, migrationPath := range migrations {
		migration, err := os.ReadFile(migrationPath)
		require.NoError(t, err)
		_, err = db.ExecContext(ctx, string(migration))
		require.NoError(t, err)
	}
	return db
}

func TestStorageCreatePostgres(t *testing.T) {
	db := guildTestDB(t)
	ctx := context.Background()
	var err error
	var id int64
	err = db.GetContext(ctx, &id, `INSERT INTO user_account(external_id,name,email,created_by) VALUES ('review-probe','Review','review@example.test','review') RETURNING id`)
	require.NoError(t, err)
	g, err := New("review-probe", "Review probe", "review@example.test")
	require.NoError(t, err)
	g.Description = "Shared finances"
	err = NewStorage(db).Create(ctx, g, id)
	require.NoError(t, err, "guild creation must persist both the guild and its creator membership")
	require.Positive(t, g.ID)
	storage := NewStorage(db)
	persisted, err := storage.GetByID(ctx, g.ID)
	require.NoError(t, err)
	require.Equal(t, g.Name, persisted.Name)
	require.Equal(t, g.Description, persisted.Description)
	updated, err := NewService(storage).Update(ctx, g.ID, UpdateInput{DisplayName: "Repeated visual name"}, audit.Actor{UserID: id, Email: "review@example.test"})
	require.NoError(t, err)
	require.Equal(t, "review-probe", updated.Name)
	require.Equal(t, "Shared finances", updated.Description)
	empty := ""
	updated, err = NewService(storage).Update(ctx, g.ID, UpdateInput{DisplayName: "Repeated visual name", Description: &empty}, audit.Actor{UserID: id, Email: "review@example.test"})
	require.NoError(t, err)
	require.Empty(t, updated.Description)
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

func TestConcurrentGuildCreationPostgres(t *testing.T) {
	db := guildTestDB(t)
	var userID int64
	require.NoError(t, db.Get(&userID, `INSERT INTO user_account(external_id,name,email,created_by) VALUES ('creator','Creator','creator@example.test','fixture') RETURNING id`))
	service := NewService(NewStorage(db))
	actor := audit.Actor{UserID: userID, Email: "creator@example.test"}
	var wg sync.WaitGroup
	results := make(chan *Guild, 6)
	failures := make(chan error, 6)
	for range 6 {
		wg.Go(func() {
			g, err := service.Create(context.Background(), strings.Repeat("Á", 70), "Optional description", actor)
			if err != nil {
				failures <- err
				return
			}
			results <- g
		})
	}
	wg.Wait()
	close(failures)
	close(results)
	for err := range failures {
		require.NoError(t, err)
	}
	names := map[string]bool{}
	for g := range results {
		require.LessOrEqual(t, len(g.Name), 70)
		require.False(t, names[g.Name])
		names[g.Name] = true
	}
	require.Len(t, names, 6)
	var count int
	require.NoError(t, db.Get(&count, "SELECT COUNT(*) FROM guild_member WHERE user_id = $1", userID))
	require.Equal(t, 6, count)
}

func TestGuildMetadataHTTPPostgres(t *testing.T) {
	db := guildTestDB(t)
	var userID int64
	require.NoError(t, db.Get(&userID, `INSERT INTO user_account(external_id,name,email,created_by) VALUES ('http-creator','Creator','creator@example.test','fixture') RETURNING id`))
	actor := audit.Actor{UserID: userID, Email: "creator@example.test"}
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Request = c.Request.WithContext(auth.WithPrincipal(c.Request.Context(), auth.Principal{Actor: actor}))
	})
	handler := NewHandler(NewService(NewStorage(db)))
	router.POST("/guilds", handler.Create)
	router.PUT("/guilds/:guildID", handler.Update)
	request := func(method, path, body string, status int) map[string]any {
		t.Helper()
		req := httptest.NewRequest(method, path, strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		recorder := httptest.NewRecorder()
		router.ServeHTTP(recorder, req)
		require.Equal(t, status, recorder.Code, recorder.Body.String())
		response := map[string]any{}
		require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &response))
		return response
	}
	created := request("POST", "/guilds", `{"name":"spoofed","display_name":"Família Silva","description":"Casa"}`, 201)
	require.Equal(t, "familia-silva", created["name"])
	require.Equal(t, "Casa", created["description"])
	path := fmt.Sprintf("/guilds/%.0f", created["id"])
	updated := request("PUT", path, `{"name":"must-not-change","display_name":"Outro visual"}`, 200)
	require.Equal(t, "familia-silva", updated["name"])
	require.Equal(t, "Casa", updated["description"])
	cleared := request("PUT", path, `{"display_name":"Outro visual","description":""}`, 200)
	require.Empty(t, cleared["description"])
	request("PUT", path, `{"name":"legacy-field-only"}`, 400)
	request("POST", "/guilds", `{"display_name":"`+strings.Repeat("á", 71)+`"}`, 400)
}
