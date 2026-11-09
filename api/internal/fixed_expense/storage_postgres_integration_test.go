//go:build integration && postgres

package fixedexpense

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"github.com/luizhenrique-dev/guild-banker/api/internal/audit"
	"github.com/luizhenrique-dev/guild-banker/api/internal/auth"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"
)

func postgresTestDB(t *testing.T) *sqlx.DB {
	t.Helper()
	dsn := os.Getenv("FIXED_EXPENSE_TEST_DATABASE_URL")
	require.NotEmpty(t, dsn, "FIXED_EXPENSE_TEST_DATABASE_URL must point to an isolated PostgreSQL 17+ test database")
	parsed, err := url.Parse(dsn)
	require.NoError(t, err)
	require.Contains(t, []string{"postgres", "postgresql"}, parsed.Scheme, "use a PostgreSQL URL")
	admin, err := sqlx.Connect("postgres", dsn)
	require.NoError(t, err, "connect to test database")
	t.Cleanup(func() { require.NoError(t, admin.Close()) })
	schema := fmt.Sprintf("issue4_%d", time.Now().UnixNano())
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

func TestFixedExpenseGuildSchema(t *testing.T) {
	db := postgresTestDB(t)
	_, err := db.Exec(`INSERT INTO guild(id,name,display_name,created_by) VALUES (1,'test','Test','fixture')`)
	require.NoError(t, err)
	_, err = db.Exec(`INSERT INTO fixed_expense(guild_id,name,amount,due_day,category,status,created_by) VALUES (1,'Expense',10,1,'OTHER','ACTIVE','fixture')`)
	require.NoError(t, err)
	for _, guild := range []any{nil, int64(999)} {
		_, err := db.Exec(`INSERT INTO fixed_expense(guild_id,name,amount,due_day,category,created_by) VALUES ($1,'Invalid',10,1,'OTHER','fixture')`, guild)
		require.Error(t, err)
		var databaseError *pq.Error
		require.ErrorAs(t, err, &databaseError)
		if guild == nil {
			require.Equal(t, pq.ErrorCode("23502"), databaseError.Code)
		} else {
			require.Equal(t, pq.ErrorCode("23503"), databaseError.Code)
		}
	}
}

func TestFixedExpenseSharingPostgres(t *testing.T) {
	db := postgresTestDB(t)
	ctx := context.Background()
	_, err := db.Exec(`
 INSERT INTO user_account(id,external_id,name,email,created_by) VALUES
 (1,'a','Member A','a@example.test','fixture'),(2,'b','Member B','b@example.test','fixture'),(3,'c','Other','c@example.test','fixture');
 INSERT INTO guild(id,name,display_name,created_by) VALUES (1,'g1','G1','fixture'),(2,'g2','G2','fixture');
 INSERT INTO guild_member(guild_id,user_id,invited_by) VALUES (1,1,1),(1,2,1),(2,2,2),(2,3,2);`)
	require.NoError(t, err)
	storage := NewStorage(db)
	service := NewService(storage)
	handler := NewHandler(service)
	router := gin.New()
	root := "/guilds/:guildID/fixed-expenses"
	router.POST(root, handler.Create)
	router.GET(root, handler.ListActiveByGuild)
	router.PATCH(root+"/:id", handler.Update)
	router.PATCH(root+"/:id/deactivate", handler.Deactivate)
	a := audit.Actor{UserID: 1, Email: "a@example.test"}
	b := audit.Actor{UserID: 2, Email: "b@example.test"}
	outsider := audit.Actor{UserID: 3, Email: "c@example.test"}
	request := func(actor audit.Actor, method, path, body string, status int) *httptest.ResponseRecorder {
		t.Helper()
		req := httptest.NewRequest(method, path, strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-User-ID", "999")
		req.Header.Set("X-User-Email", "spoofed@example.test")
		req = req.WithContext(auth.WithPrincipal(req.Context(), auth.Principal{Actor: actor}))
		response := httptest.NewRecorder()
		router.ServeHTTP(response, req)
		require.Equal(t, status, response.Code, response.Body.String())
		return response
	}
	const payload = `{"name":"Rent","amount":"123.45","due_day":10,"category":"HOUSING","guild_id":2,"user_id":999}`
	created := request(a, "POST", "/guilds/1/fixed-expenses", payload, 201)
	var response fixedExpenseResponse
	require.NoError(t, json.Unmarshal(created.Body.Bytes(), &response))
	require.EqualValues(t, 1, response.GuildID)
	require.Positive(t, response.ID)
	path := fmt.Sprintf("/guilds/1/fixed-expenses/%d", response.ID)
	wrongPath := fmt.Sprintf("/guilds/2/fixed-expenses/%d", response.ID)
	before, err := storage.GetByIDAndGuild(ctx, response.ID, 1)
	require.NoError(t, err)
	require.Equal(t, a.Email, before.CreatedBy)
	require.True(t, before.Amount.Equal(decimal.RequireFromString("123.45")))
	require.Contains(t, request(b, "GET", "/guilds/1/fixed-expenses", "", 200).Body.String(), fmt.Sprintf(`"id":%d`, response.ID))
	require.JSONEq(t, `[]`, request(b, "GET", "/guilds/2/fixed-expenses", "", 200).Body.String())
	for _, actor := range []audit.Actor{outsider} {
		request(actor, "POST", "/guilds/1/fixed-expenses", payload, 403)
		request(actor, "GET", "/guilds/1/fixed-expenses", "", 403)
		request(actor, "PATCH", path, `{"amount":"999"}`, 403)
		request(actor, "PATCH", path+"/deactivate", `{"status":"PAUSED"}`, 403)
	}
	request(b, "PATCH", wrongPath, `{"amount":"999"}`, 404)
	request(b, "PATCH", wrongPath+"/deactivate", `{"status":"CANCELLED"}`, 404)
	copy := *before
	copy.GuildID = 2
	copy.Amount = decimal.NewFromInt(999)
	require.ErrorIs(t, storage.Update(ctx, &copy), ErrFixedExpenseNotFound)
	unchanged, err := storage.GetByIDAndGuild(ctx, response.ID, 1)
	require.NoError(t, err)
	require.Equal(t, before, unchanged)
	request(b, "PATCH", path, `{"amount":"200.50","guild_id":2,"name":"Forbidden rename"}`, 200)
	edited, err := storage.GetByIDAndGuild(ctx, response.ID, 1)
	require.NoError(t, err)
	require.Equal(t, before.CreatedAt, edited.CreatedAt)
	require.Equal(t, before.CreatedBy, edited.CreatedBy)
	require.Equal(t, before.Name, edited.Name)
	require.EqualValues(t, 1, edited.GuildID)
	require.True(t, edited.Amount.Equal(decimal.RequireFromString("200.50")))
	require.NotNil(t, edited.UpdatedAt)
	require.NotNil(t, edited.UpdatedBy)
	require.Equal(t, b.Email, *edited.UpdatedBy)
	_, err = db.Exec(`DELETE FROM guild_member WHERE guild_id=1 AND user_id=1`)
	require.NoError(t, err)
	request(a, "POST", "/guilds/1/fixed-expenses", payload, 403)
	request(a, "GET", "/guilds/1/fixed-expenses", "", 403)
	request(a, "PATCH", path, `{"amount":"999"}`, 403)
	request(a, "PATCH", path+"/deactivate", `{"status":"PAUSED"}`, 403)
	request(b, "PATCH", path+"/deactivate", `{"status":"CANCELLED"}`, 204)
	paused, err := storage.GetByIDAndGuild(ctx, response.ID, 1)
	require.NoError(t, err)
	require.Equal(t, StatusCancelled, paused.Status)
	require.Equal(t, a.Email, paused.CreatedBy)
	require.Equal(t, before.CreatedAt, paused.CreatedAt)
	require.Equal(t, b.Email, *paused.UpdatedBy)
	require.JSONEq(t, `[]`, request(b, "GET", "/guilds/1/fixed-expenses", "", 200).Body.String())
	for _, expense := range []struct {
		name string
		day  int
	}{{"Zulu", 5}, {"Alpha", 5}, {"Late", 20}} {
		_, err := service.Create(ctx, 1, expense.name, decimal.NewFromInt(1), expense.day, CategoryOther, b)
		require.NoError(t, err)
	}
	list, err := service.ListActiveByGuild(ctx, 1, b)
	require.NoError(t, err)
	require.Len(t, list, 3)
	require.Equal(t, []string{"Alpha", "Zulu", "Late"}, []string{list[0].Name, list[1].Name, list[2].Name})
	canceled, cancel := context.WithCancel(ctx)
	cancel()
	_, err = service.ListActiveByGuild(canceled, 1, b)
	require.ErrorIs(t, err, context.Canceled)
}
