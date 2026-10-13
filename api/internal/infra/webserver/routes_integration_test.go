//go:build integration

package webserver

import (
	"bytes"
	"context"
	"log/slog"
	"net/http/httptest"
	"testing"

	"github.com/luizhenrique-dev/guild-banker/api/config"
	"github.com/luizhenrique-dev/guild-banker/api/internal/auth"
	"github.com/stretchr/testify/assert"
)

type testAuthenticator struct{ calls int }

func (s *testAuthenticator) Authenticate(context.Context, string) (auth.Principal, error) {
	s.calls++
	return auth.Principal{}, nil
}

func TestAPIPrefixProtection(t *testing.T) {
	var logs bytes.Buffer
	authenticator := &testAuthenticator{}
	cfg := &config.Config{
		Keycloak:           config.KeycloakConfig{Timeout: 1},
		CORSAllowedOrigins: []string{"http://localhost:5173"},
	}
	server := NewServer(cfg, nil, authenticator, slog.New(slog.NewJSONHandler(&logs, nil)))
	for _, tt := range []struct {
		path   string
		token  bool
		status int
	}{
		{path: "/health", status: 200}, {path: "/apix", status: 404},
		{path: "/api", status: 401}, {path: "/api/", status: 401}, {path: "/api/not-found", status: 401},
		{path: "/api/v1/guilds", status: 401}, {path: "/api/v1/fixed-expenses", status: 401},
		{path: "/api/v1/guilds/1/transactions", status: 401}, {path: "/api/v1/guilds/1/imports/1", status: 401},
		{path: "/api/not-found", token: true, status: 404},
	} {
		t.Run(tt.path+map[bool]string{false: " without token", true: " with token"}[tt.token], func(t *testing.T) {
			before := authenticator.calls
			req := httptest.NewRequest("GET", tt.path, nil)
			if tt.token {
				req.Header.Set("Authorization", "Bearer token")
			}
			w := httptest.NewRecorder()
			server.router.ServeHTTP(w, req)
			assert.Equal(t, tt.status, w.Code)
			wantCalls := 0
			if tt.token {
				wantCalls = 1
			}
			assert.Equal(t, wantCalls, authenticator.calls-before)
		})
	}

	t.Run("plain OPTIONS is protected", func(t *testing.T) {
		req := httptest.NewRequest("OPTIONS", "/api/not-found", nil)
		req.Header.Set("Origin", "http://localhost:5173")
		w := httptest.NewRecorder()
		server.router.ServeHTTP(w, req)
		assert.Equal(t, 401, w.Code)
	})
	for _, method := range []string{"GET", "POST"} {
		t.Run("trailing slash "+method, func(t *testing.T) {
			req := httptest.NewRequest(method, "/api/v1/guilds/", nil)
			w := httptest.NewRecorder()
			server.router.ServeHTTP(w, req)
			assert.Equal(t, 401, w.Code)
		})
	}
}
