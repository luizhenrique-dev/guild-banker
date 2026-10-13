//go:build integration

package transaction

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/luizhenrique-dev/guild-banker/api/internal/audit"
	"github.com/luizhenrique-dev/guild-banker/api/internal/auth"
	"github.com/stretchr/testify/assert"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestAuthenticatedHandlerIdentity(t *testing.T) {
	for _, tt := range []struct {
		name          string
		authenticated bool
	}{
		{name: "principal overrides conflicting headers", authenticated: true},
		{name: "headers cannot establish identity"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			c, _ := gin.CreateTestContext(httptest.NewRecorder())
			c.Request = httptest.NewRequest("GET", "/api/probe", nil)
			c.Request.Header.Set("X-User-ID", "999")
			c.Request.Header.Set("X-User-Email", "spoofed@example.test")
			actor := audit.Actor{UserID: 42, Email: "local@example.test"}
			if tt.authenticated {
				c.Request = c.Request.WithContext(auth.WithPrincipal(c.Request.Context(), auth.Principal{Claims: auth.Claims{UserID: "external-sub"}, Actor: actor}))
			}
			got, ok := getRequesterActor(c)
			assert.Equal(t, tt.authenticated, ok)
			id, ok := getRequesterUserID(c)
			assert.Equal(t, tt.authenticated, ok)
			if tt.authenticated {
				assert.Equal(t, actor, got)
				assert.EqualValues(t, 42, id)
				return
			}
			assert.Zero(t, got)
			assert.Zero(t, id)
		})
	}
}

type identityStorage struct {
	Storage
	t      *testing.T
	called bool
}

func (s *identityStorage) IsGuildMember(_ context.Context, guildID, userID int64) (bool, error) {
	assert.EqualValues(s.t, 1, guildID)
	assert.EqualValues(s.t, 42, userID)
	return true, nil
}
func (s *identityStorage) Create(_ context.Context, transaction *Transaction) error {
	s.called = true
	assert.EqualValues(s.t, 42, transaction.UserAccountID)
	assert.Equal(s.t, "local@example.test", transaction.CreatedBy)
	return nil
}
func TestHandlerUsesTrustedActor(t *testing.T) {
	storage := &identityStorage{t: t}
	handler := NewHandler(NewService(storage))
	router := gin.New()
	router.POST("/api/guilds/:guildID/transactions", handler.Create)
	testIdentityRequest(t, router, "POST", "/api/guilds/1/transactions", `{"type":"EXPENSE","description":"fixture","amount":"10","category":"OTHER","occurredAt":"2026-01-01T00:00:00Z"}`, 201)
	assert.True(t, storage.called)
}

func testIdentityRequest(t *testing.T, router *gin.Engine, method, path, body string, status int) {
	t.Helper()
	for _, authenticated := range []bool{true, false} {
		req := httptest.NewRequest(method, path, strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-User-ID", "999")
		req.Header.Set("X-User-Email", "spoofed@example.test")
		if authenticated {
			req = req.WithContext(auth.WithPrincipal(req.Context(), auth.Principal{Actor: audit.Actor{UserID: 42, Email: "local@example.test"}}))
		}
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		want := 401
		if authenticated {
			want = status
		}
		assert.Equal(t, want, w.Code, w.Body.String())
		if !authenticated {
			assert.JSONEq(t, `{"error":"unauthorized"}`, w.Body.String())
		}
	}
}
