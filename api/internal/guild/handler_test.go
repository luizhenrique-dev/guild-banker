//go:build integration

package guild

import (
	"context"
	"encoding/json"
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

func (s *identityStorage) NameExists(context.Context, string, int64) (bool, error) { return false, nil }
func (s *identityStorage) Create(_ context.Context, guild *Guild, userID int64) error {
	s.called = true
	assert.EqualValues(s.t, 42, userID)
	assert.Equal(s.t, "local@example.test", guild.CreatedBy)
	return nil
}
func TestHandlerUsesTrustedActor(t *testing.T) {
	storage := &identityStorage{t: t}
	handler := NewHandler(NewService(storage))
	router := gin.New()
	router.POST("/api/guilds", handler.Create)
	testIdentityRequest(t, router, "POST", "/api/guilds", `{"name":"fixture-guild","display_name":"Fixture"}`, 201)
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

func TestNameGenerationExhaustionReturnsUnavailable(t *testing.T) {
	service, storage := newService(t)
	storage.On("Create", mock.Anything, mock.Anything, int64(42)).Return(ErrGuildNameAlreadyUsed).Times(6)
	handler := NewHandler(service)
	router := gin.New()
	router.POST("/api/guilds", handler.Create)
	testIdentityRequest(t, router, "POST", "/api/guilds", `{"display_name":"Repeated"}`, 503)
	storage.AssertExpectations(t)
}

func TestInvitationErrorCodes(t *testing.T) {
	for _, tt := range []struct {
		name   string
		err    error
		status int
		code   string
	}{
		{name: "pending", err: ErrInvitationAlreadyPending, status: 409, code: "invitation_already_pending"},
		{name: "member", err: ErrGuildMemberAlreadySet, status: 409, code: "guild_member_already_set"},
		{name: "invalid", err: ErrInvalidInvitation, status: 400},
		{name: "technical", err: errors.New("private database details"), status: 500},
	} {
		t.Run(tt.name, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(recorder)
			(&Handler{}).handleInvitationError(c, tt.err)
			assert.Equal(t, tt.status, recorder.Code)
			payload := map[string]string{}
			assert.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &payload))
			assert.Equal(t, tt.code, payload["code"])
			assert.NotContains(t, recorder.Body.String(), "private database details")
		})
	}
}
