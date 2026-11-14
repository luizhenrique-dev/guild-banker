//go:build integration && postgres

package fixedexpense

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/luizhenrique-dev/guild-banker/api/internal/auth"
	"github.com/luizhenrique-dev/guild-banker/api/internal/guild"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"
)

func TestResponsibleExpensePostgres(t *testing.T) {
	db := postgresTestDB(t)
	ctx := context.Background()
	_, err := db.Exec(`INSERT INTO user_account(id,external_id,name,email,created_by) VALUES
 (1,'a','A','a@test.local','fixture'),(2,'b','B','b@test.local','fixture'),(3,'c','C','c@test.local','fixture');
 INSERT INTO guild(id,name,display_name,created_by) VALUES (1,'g1','G1','fixture'),(2,'g2','G2','fixture');
 INSERT INTO guild_member(guild_id,user_id,invited_by) VALUES (1,1,1),(1,2,1),(2,1,1),(2,2,1),(2,3,1)`)
	require.NoError(t, err)
	storage := NewStorage(db)
	handler := NewHandler(NewService(storage))
	router := gin.New()
	router.POST("/guilds/:guildID/expenses", handler.Create)
	router.GET("/guilds/:guildID/expenses", handler.ListActiveByGuild)
	router.PATCH("/guilds/:guildID/expenses/:id", handler.Update)
	request := func(method, path, body string, status int) *httptest.ResponseRecorder {
		t.Helper()
		req := httptest.NewRequest(method, path, strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req = req.WithContext(auth.WithPrincipal(req.Context(), auth.Principal{Actor: validActor()}))
		response := httptest.NewRecorder()
		router.ServeHTTP(response, req)
		require.Equal(t, status, response.Code, response.Body.String())
		return response
	}
	create := func(guildID int, responsible string) fixedExpenseResponse {
		t.Helper()
		var expense fixedExpenseResponse
		payload := `{"name":"Rent","amount":"123.45","due_day":10,"category":"HOUSING"` + responsible + `}`
		response := request("POST", fmt.Sprintf("/guilds/%d/expenses", guildID), payload, 201)
		require.NoError(t, json.Unmarshal(response.Body.Bytes(), &expense))
		return expense
	}
	old := create(1, "")
	require.Nil(t, old.ResponsibleUserID)
	expense := create(1, `,"responsible_user_id":2`)
	require.EqualValues(t, 2, *expense.ResponsibleUserID)
	otherGuild := create(2, `,"responsible_user_id":2`)
	path := fmt.Sprintf("/guilds/1/expenses/%d", expense.ID)
	request("PATCH", path, `{"status":"PAUSED"}`, 200)
	persisted, err := storage.GetByIDAndGuild(ctx, expense.ID, 1)
	require.NoError(t, err)
	require.EqualValues(t, 2, *persisted.ResponsibleUserID)
	require.Equal(t, StatusPaused, persisted.Status)
	before := *persisted
	for _, responsible := range []string{"0", "-1", "3", "999", `"2"`, "1.5", "true", "9223372036854775808"} {
		request("PATCH", path, `{"amount":"999","responsible_user_id":`+responsible+`}`, 400)
		request("POST", "/guilds/1/expenses", `{"name":"Invalid","amount":"999","due_day":1,"category":"OTHER","responsible_user_id":`+responsible+`}`, 400)
		unchanged, err := storage.GetByIDAndGuild(ctx, expense.ID, 1)
		require.NoError(t, err)
		require.Equal(t, &before, unchanged)
	}
	request("PATCH", fmt.Sprintf("/guilds/2/expenses/%d", expense.ID), `{"responsible_user_id":3}`, 404)
	request("PATCH", path, `{"responsible_user_id":null}`, 200)
	persisted, err = storage.GetByIDAndGuild(ctx, expense.ID, 1)
	require.NoError(t, err)
	require.Nil(t, persisted.ResponsibleUserID)
	request("PATCH", path, `{"responsible_user_id":1}`, 200)
	request("PATCH", path, `{"responsible_user_id":2}`, 200)
	require.NoError(t, guild.NewService(guild.NewStorage(db)).RemoveUser(ctx, 1, 1, 2))
	persisted, err = storage.GetByIDAndGuild(ctx, expense.ID, 1)
	require.NoError(t, err)
	require.Nil(t, persisted.ResponsibleUserID)
	require.True(t, persisted.Amount.Equal(before.Amount))
	require.Equal(t, before.DueDay, persisted.DueDay)
	require.Equal(t, before.Category, persisted.Category)
	require.Equal(t, before.CreatedAt, persisted.CreatedAt)
	require.Equal(t, before.CreatedBy, persisted.CreatedBy)
	require.Equal(t, before.Status, persisted.Status)
	untouched, err := storage.GetByIDAndGuild(ctx, otherGuild.ID, 2)
	require.NoError(t, err)
	require.EqualValues(t, 2, *untouched.ResponsibleUserID)
	var rows []fixedExpenseResponse
	require.NoError(t, json.Unmarshal(request("GET", "/guilds/1/expenses?include_paused=true", "", 200).Body.Bytes(), &rows))
	require.Len(t, rows, 2)
	for _, row := range rows {
		require.Nil(t, row.ResponsibleUserID)
	}
}

func TestResponsibleMigrationPreservesExistingExpenses(t *testing.T) {
	db := postgresTestDB(t)
	down, err := os.ReadFile("../../migrations/000005_fixed_expense_responsible.down.sql")
	require.NoError(t, err)
	_, err = db.Exec(string(down))
	require.NoError(t, err)
	_, err = db.Exec(`INSERT INTO guild(id,name,display_name,created_by) VALUES (1,'legacy','Legacy','fixture');
 INSERT INTO fixed_expense(guild_id,name,amount,due_day,category,created_by) VALUES (1,'Legacy rent',123.45,10,'HOUSING','fixture')`)
	require.NoError(t, err)
	up, err := os.ReadFile("../../migrations/000005_fixed_expense_responsible.up.sql")
	require.NoError(t, err)
	_, err = db.Exec(string(up))
	require.NoError(t, err)
	expenses, err := NewStorage(db).ListActiveByGuild(context.Background(), 1)
	require.NoError(t, err)
	require.Len(t, expenses, 1)
	require.Nil(t, expenses[0].ResponsibleUserID)
	require.Equal(t, "Legacy rent", expenses[0].Name)
	require.True(t, expenses[0].Amount.Equal(decimal.RequireFromString("123.45")))
}

func TestResponsibleAssignmentConcurrentRemoval(t *testing.T) {
	db := postgresTestDB(t)
	_, err := db.Exec(`INSERT INTO user_account(id,external_id,name,email,created_by) VALUES (1,'a','A','a@test.local','fixture'),(2,'b','B','b@test.local','fixture');
 INSERT INTO guild(id,name,display_name,created_by) VALUES (1,'g1','G1','fixture');
 INSERT INTO guild_member(guild_id,user_id,invited_by) VALUES (1,1,1)`)
	require.NoError(t, err)
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	storage := NewStorage(db)
	for _, operation := range []string{"create", "update"} {
		for round := range 10 {
			_, err := db.Exec(`INSERT INTO guild_member(guild_id,user_id,invited_by) VALUES (1,2,1)`)
			require.NoError(t, err)
			expense, err := New(1, fmt.Sprintf("%s-%d", operation, round), decimal.NewFromInt(10), 1, CategoryOther, "fixture")
			require.NoError(t, err)
			if operation == "update" {
				require.NoError(t, storage.Create(ctx, expense))
			}
			expense.ResponsibleUserID = new(int64(2))
			start := make(chan struct{})
			saved := make(chan error, 1)
			removed := make(chan error, 1)
			go func() {
				<-start
				if operation == "create" {
					saved <- storage.Create(ctx, expense)
				} else {
					saved <- storage.Update(ctx, expense)
				}
			}()
			go func() {
				<-start
				removed <- guild.NewService(guild.NewStorage(db)).RemoveUser(ctx, 1, 1, 2)
			}()
			close(start)
			saveErr := <-saved
			removeErr := <-removed
			require.NoError(t, removeErr)
			if saveErr != nil {
				require.ErrorIs(t, saveErr, ErrInvalidInput)
			}
			if saveErr == nil || operation == "update" {
				persisted, err := storage.GetByIDAndGuild(ctx, expense.ID, 1)
				require.NoError(t, err)
				require.Nil(t, persisted.ResponsibleUserID)
				require.True(t, persisted.Amount.Equal(decimal.NewFromInt(10)))
			}
			var orphanCount int
			require.NoError(t, db.Get(&orphanCount, `SELECT count(*) FROM fixed_expense f LEFT JOIN guild_member m ON m.guild_id=f.guild_id AND m.user_id=f.responsible_user_id WHERE f.responsible_user_id IS NOT NULL AND m.user_id IS NULL`))
			require.Zero(t, orphanCount)
		}
	}
}
