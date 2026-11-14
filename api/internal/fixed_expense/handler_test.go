//go:build integration

package fixedexpense

import (
	"context"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/luizhenrique-dev/guild-banker/api/internal/audit"
	"github.com/luizhenrique-dev/guild-banker/api/internal/auth"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
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

			if tt.authenticated {
				assert.Equal(t, actor, got)
				return
			}
			assert.Zero(t, got)
		})
	}
}

type identityStorage struct {
	Storage
	t      *testing.T
	called bool
}

func (s *identityStorage) Create(_ context.Context, expense *FixedExpense) error {
	s.called = true
	assert.EqualValues(s.t, 1, expense.GuildID)
	assert.Equal(s.t, "local@example.test", expense.CreatedBy)
	return nil
}
func TestHandlerUsesTrustedActor(t *testing.T) {
	storage := &identityStorage{t: t}
	handler := NewHandler(NewService(storage))
	router := gin.New()
	router.POST("/api/guilds/:guildID/fixed-expenses", handler.Create)
	testIdentityRequest(t, router, "POST", "/api/guilds/1/fixed-expenses", `{"name":"fixture","amount":"10","due_day":1,"category":"OTHER"}`, 201)
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

func (s *identityStorage) IsGuildMember(_ context.Context, guildID, userID int64) (bool, error) {
	assert.EqualValues(s.t, 42, userID)
	assert.EqualValues(s.t, 1, guildID)
	return true, nil
}

func TestHandlerGuildErrors(t *testing.T) {
	for _, test := range []struct {
		name       string
		membership bool
		err        error
		code       int
	}{
		{name: "forbidden", code: 403},
		{name: "technical", err: errors.New("SQL password confidential"), code: 500},
		{name: "allowed", membership: true, code: 400},
	} {
		t.Run(test.name, func(t *testing.T) {
			storage := &storageMock{member: &test.membership, membershipErr: test.err}
			handler := NewHandler(NewService(storage))
			router := gin.New()
			router.POST("/guilds/:guildID/fixed-expenses", handler.Create)
			req := httptest.NewRequest("POST", "/guilds/1/fixed-expenses", strings.NewReader(`{"name":"","amount":"1","due_day":1,"category":"OTHER"}`))
			req.Header.Set("Content-Type", "application/json")
			req = req.WithContext(auth.WithPrincipal(req.Context(), auth.Principal{Actor: validActor()}))
			response := httptest.NewRecorder()
			router.ServeHTTP(response, req)
			assert.Equal(t, test.code, response.Code)
			assert.NotContains(t, response.Body.String(), "confidential")
			if test.code == 500 {
				assert.JSONEq(t, `{"error":"internal server error"}`, response.Body.String())
			}
		})
	}
	for _, id := range []string{"0", "-1", "invalid", "9223372036854775808"} {
		t.Run("invalid guild "+id, func(t *testing.T) {
			handler := NewHandler(NewService(&storageMock{}))
			router := gin.New()
			router.GET("/guilds/:guildID/fixed-expenses", handler.ListActiveByGuild)
			response := httptest.NewRecorder()
			router.ServeHTTP(response, httptest.NewRequest("GET", "/guilds/"+id+"/fixed-expenses", nil))
			assert.Equal(t, 400, response.Code)
		})
	}
}

func TestHandlerCreatesResponsibleExpense(t *testing.T) {
	storage := &storageMock{}
	storage.On("Create", mock.Anything, mock.Anything).Return(nil)
	router := gin.New()
	router.POST("/guilds/:guildID/fixed-expenses", NewHandler(NewService(storage)).Create)
	req := httptest.NewRequest("POST", "/guilds/1/fixed-expenses", strings.NewReader(`{"name":"Rent","amount":"10","due_day":1,"category":"OTHER","responsible_user_id":2}`))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(auth.WithPrincipal(req.Context(), auth.Principal{Actor: validActor()}))
	out := httptest.NewRecorder()
	router.ServeHTTP(out, req)
	assert.Equal(t, 201, out.Code, out.Body.String())
	assert.Contains(t, out.Body.String(), `"responsible_user_id":2`)
	storage.AssertExpectations(t)
}
