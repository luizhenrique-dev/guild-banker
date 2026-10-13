//go:build integration

package webserver

import (
	"io"
	"log/slog"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/luizhenrique-dev/guild-banker/api/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestServerCORS(t *testing.T) {
	for _, tt := range []struct {
		name, method, path, origin, requestedMethod string
		status                                      int
		allowOrigin                                 bool
	}{
		{name: "UI preflight GET", method: "OPTIONS", path: "/api/v1/guilds", origin: "http://localhost:5173", requestedMethod: "GET", status: 204, allowOrigin: true},
		{name: "UI preflight POST", method: "OPTIONS", path: "/api/v1/guilds", origin: "http://127.0.0.1:5173", requestedMethod: "POST", status: 204, allowOrigin: true},
		{name: "UI effective GET", method: "GET", path: "/api/v1/guilds", origin: "http://localhost:5173", status: 401, allowOrigin: true},
		{name: "UI effective POST", method: "POST", path: "/api/v1/guilds", origin: "http://127.0.0.1:5173", status: 401, allowOrigin: true},
		{name: "foreign effective", method: "GET", path: "/api/v1/guilds", origin: "https://foreign.example.test", status: 403},
		{name: "foreign preflight", method: "OPTIONS", path: "/api/v1/guilds", origin: "https://foreign.example.test", requestedMethod: "GET", status: 403},
		{name: "preview not configured", method: "GET", path: "/api/v1/guilds", origin: "http://localhost:4173", status: 403},
		{name: "no Origin", method: "GET", path: "/api/v1/guilds", status: 401},
		{name: "OPTIONS without preflight", method: "OPTIONS", path: "/api/v1/guilds", origin: "http://localhost:5173", status: 401},
		{name: "public health", method: "GET", path: "/health", origin: "http://localhost:5173", status: 200, allowOrigin: true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			authenticator := &testAuthenticator{}
			cfg := &config.Config{
				Keycloak:           config.KeycloakConfig{Timeout: 1},
				CORSAllowedOrigins: []string{"http://localhost:5173", "http://127.0.0.1:5173"},
			}
			server := NewServer(cfg, nil, authenticator, slog.New(slog.NewJSONHandler(io.Discard, nil)))
			req := httptest.NewRequest(tt.method, tt.path, nil)
			if tt.origin != "" {
				req.Header.Set("Origin", tt.origin)
			}
			if tt.requestedMethod != "" {
				req.Header.Set("Access-Control-Request-Method", tt.requestedMethod)
				req.Header.Set("Access-Control-Request-Headers", "Authorization, Content-Type")
			}
			w := httptest.NewRecorder()
			server.router.ServeHTTP(w, req)
			assert.Equal(t, tt.status, w.Code)
			if tt.allowOrigin {
				assert.Equal(t, tt.origin, w.Header().Get("Access-Control-Allow-Origin"))
			} else {
				assert.Empty(t, w.Header().Get("Access-Control-Allow-Origin"))
			}
			assert.Empty(t, w.Header().Get("Access-Control-Allow-Credentials"))
			if tt.status == 204 {
				headers := strings.ToLower(w.Header().Get("Access-Control-Allow-Headers"))
				assert.Contains(t, headers, "authorization")
				assert.Contains(t, headers, "content-type")
				assert.NotContains(t, headers, "x-user")
			}
			if tt.status == 401 {
				assert.JSONEq(t, `{"error":"unauthorized"}`, w.Body.String())
			}
			assert.Zero(t, authenticator.calls)
		})
	}
}

func TestServerCORSLoadedOrigins(t *testing.T) {
	t.Chdir(t.TempDir())
	for _, key := range []string{"DB_HOST", "DB_PORT", "DB_USER", "DB_PASSWORD", "DB_NAME"} {
		t.Setenv(key, "test")
	}
	t.Setenv("KEYCLOAK_ISSUER", "http://identity.example.test/realms/fixture")
	t.Setenv("API_AUDIENCE", "fixture-api")
	t.Setenv("KEYCLOAK_TIMEOUT_SECONDS", "1")
	t.Setenv("CORS_ALLOWED_ORIGINS", "HTTP://localhost:5173")
	cfg, err := config.Load()
	require.NoError(t, err)
	var server *Server
	require.NotPanics(t, func() {
		server = NewServer(cfg, nil, &testAuthenticator{}, slog.New(slog.NewJSONHandler(io.Discard, nil)))
	})
	req := httptest.NewRequest("GET", "/api/not-found", nil)
	req.Header.Set("Origin", "http://localhost:5173")
	w := httptest.NewRecorder()
	server.router.ServeHTTP(w, req)
	assert.Equal(t, 401, w.Code)
	assert.Equal(t, "http://localhost:5173", w.Header().Get("Access-Control-Allow-Origin"))
}
