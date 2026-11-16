//go:build integration && postgres

package fixedexpense

import (
	"encoding/json"
	"fmt"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/luizhenrique-dev/guild-banker/api/internal/auth"
	"github.com/stretchr/testify/require"
)

func TestPaymentMethodPostgresRoundTrip(t *testing.T) {
	db := postgresTestDB(t)
	_, err := db.Exec(`INSERT INTO user_account(id,external_id,name,email,created_by) VALUES (1,'owner','Owner','owner@example.test','fixture');
 INSERT INTO guild(id,name,display_name,created_by) VALUES (1,'g1','G1','fixture'),(2,'g2','G2','fixture');
 INSERT INTO guild_member(guild_id,user_id,invited_by) VALUES (1,1,1),(2,1,1)`)
	require.NoError(t, err)
	handler := NewHandler(NewService(NewStorage(db)))
	router := gin.New()
	router.POST("/guilds/:guildID/expenses", handler.Create)
	router.GET("/guilds/:guildID/expenses", handler.ListActiveByGuild)
	router.PATCH("/guilds/:guildID/expenses/:id", handler.Update)
	request := func(method, path, body string, status int) []byte {
		t.Helper()
		req := httptest.NewRequest(method, path, strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req = req.WithContext(auth.WithPrincipal(req.Context(), auth.Principal{Actor: validActor()}))
		out := httptest.NewRecorder()
		router.ServeHTTP(out, req)
		require.Equal(t, status, out.Code, out.Body.String())
		return out.Body.Bytes()
	}
	const base = `"name":"Rent","amount":"123.45","due_day":10,"category":"HOUSING"`
	for _, method := range []PaymentMethod{PaymentMethodCreditCard, PaymentMethodDebitCard, PaymentMethodPIX, PaymentMethodAutomaticDebit, PaymentMethodVoucher} {
		t.Run(string(method), func(t *testing.T) {
			var created fixedExpenseResponse
			require.NoError(t, json.Unmarshal(request("POST", "/guilds/1/expenses", `{`+base+`,"payment_method":"`+string(method)+`"}`, 201), &created))
			require.Equal(t, method, created.PaymentMethod)
			persisted, err := NewStorage(db).GetByIDAndGuild(t.Context(), created.ID, 1)
			require.NoError(t, err)
			require.Equal(t, method, persisted.PaymentMethod)
			path := fmt.Sprintf("/guilds/1/expenses/%d", created.ID)

			for _, updatedMethod := range []PaymentMethod{PaymentMethodCreditCard, PaymentMethodDebitCard, PaymentMethodPIX, PaymentMethodAutomaticDebit, PaymentMethodVoucher} {
				request("PATCH", path, `{"payment_method":"`+string(updatedMethod)+`"}`, 200)
				persisted, err := NewStorage(db).GetByIDAndGuild(t.Context(), created.ID, 1)
				require.NoError(t, err)
				require.Equal(t, updatedMethod, persisted.PaymentMethod)
			}
			request("PATCH", path, `{"payment_method":"PIX"}`, 200)
			request("PATCH", path, `{"status":"PAUSED"}`, 200)
			before, err := NewStorage(db).GetByIDAndGuild(t.Context(), created.ID, 1)
			require.NoError(t, err)
			require.Equal(t, PaymentMethodPIX, before.PaymentMethod)
			for _, invalid := range []string{`null`, `""`, `"CASH"`, `42`, `true`, `[]`, `{}`} {
				request("POST", "/guilds/1/expenses", `{`+base+`,"payment_method":`+invalid+`}`, 400)
				request("PATCH", path, `{"amount":"999","payment_method":`+invalid+`}`, 400)
				after, err := NewStorage(db).GetByIDAndGuild(t.Context(), created.ID, 1)
				require.NoError(t, err)
				require.Equal(t, before, after)
			}
			request("PATCH", fmt.Sprintf("/guilds/2/expenses/%d", created.ID), `{"payment_method":"VOUCHER"}`, 404)
			request("PATCH", path, `{"payment_method":"VOUCHER","status":"ACTIVE"}`, 200)
			var listed []fixedExpenseResponse
			require.NoError(t, json.Unmarshal(request("GET", "/guilds/1/expenses", "", 200), &listed))
			found := false
			for _, expense := range listed {
				if expense.ID == created.ID {
					require.Equal(t, PaymentMethodVoucher, expense.PaymentMethod)
					found = true
				}
			}
			require.True(t, found)
		})
	}
	request("POST", "/guilds/1/expenses", `{`+base+`}`, 400)
	var count int
	require.NoError(t, db.Get(&count, `SELECT count(*) FROM fixed_expense`))
	require.Equal(t, 5, count)
}

func TestPaymentMethodMigration(t *testing.T) {
	db := postgresTestDB(t)
	down, err := os.ReadFile("../../migrations/000006_fixed_expense_payment_method.down.sql")
	require.NoError(t, err)
	_, err = db.Exec(string(down))
	require.NoError(t, err)
	_, err = db.Exec(`INSERT INTO user_account(id,external_id,name,email,created_by) VALUES (1,'a','A','a@test.local','fixture');
 INSERT INTO guild(id,name,display_name,created_by) VALUES (1,'g1','G1','fixture'),(2,'g2','G2','fixture');
 INSERT INTO guild_member(guild_id,user_id,invited_by) VALUES (1,1,1),(2,1,1);
 INSERT INTO fixed_expense(guild_id,name,amount,due_day,category,status,responsible_user_id,created_by,updated_at,updated_by)
 SELECT g.id, 'Legacy ' || s.status, 123.45, 10, 'HOUSING', s.status, 1, 'fixture', CURRENT_TIMESTAMP, 'editor'
 FROM guild g CROSS JOIN (VALUES ('ACTIVE'),('PAUSED'),('CANCELLED')) s(status)`)
	require.NoError(t, err)
	var before string
	require.NoError(t, db.Get(&before, `SELECT jsonb_agg(to_jsonb(f) ORDER BY id)::text FROM fixed_expense f`))
	up, err := os.ReadFile("../../migrations/000006_fixed_expense_payment_method.up.sql")
	require.NoError(t, err)
	_, err = db.Exec(string(up))
	require.NoError(t, err)
	var after string
	require.NoError(t, db.Get(&after, `SELECT jsonb_agg(to_jsonb(f) - 'payment_method' ORDER BY id)::text FROM fixed_expense f`))
	require.JSONEq(t, before, after)
	var creditCount int
	require.NoError(t, db.Get(&creditCount, `SELECT count(*) FROM fixed_expense WHERE payment_method='CREDIT_CARD'`))
	require.Equal(t, 6, creditCount)
	for _, invalid := range []any{nil, "", "CASH"} {
		_, err := db.Exec(`UPDATE fixed_expense SET payment_method=$1`, invalid)
		require.Error(t, err)
	}
	_, err = db.Exec(`INSERT INTO fixed_expense(guild_id,name,amount,due_day,category,created_by) VALUES (1,'Missing',1,1,'OTHER','fixture')`)
	require.Error(t, err, "new writes must not have an automatic payment default")
	_, err = db.Exec(string(down))
	require.NoError(t, err)
	require.NoError(t, db.Get(&after, `SELECT jsonb_agg(to_jsonb(f) ORDER BY id)::text FROM fixed_expense f`))
	require.JSONEq(t, before, after)
}
