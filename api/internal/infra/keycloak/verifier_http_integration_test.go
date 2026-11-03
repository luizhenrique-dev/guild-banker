//go:build integration

package keycloak

import (
	"bytes"
	"context"
	"crypto"
	"encoding/base64"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/gin-gonic/gin"
	"github.com/go-jose/go-jose/v4"
	"github.com/luizhenrique-dev/guild-banker/api/internal/audit"
	"github.com/luizhenrique-dev/guild-banker/api/internal/auth"
	"github.com/luizhenrique-dev/guild-banker/api/internal/infra/webserver/middleware"
	"github.com/luizhenrique-dev/guild-banker/api/internal/user"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fixtureUsers struct {
	calls    int
	sub      string
	err      error
	disabled bool
}

func (u *fixtureUsers) EnsureAccount(_ context.Context, sub, _, _ string) (*user.User, error) {
	u.calls++
	u.sub = sub
	return &user.User{ID: 42, Email: "local@example.test", Enabled: !u.disabled}, u.err
}

func TestAuthAcceptance(t *testing.T) {
	now := time.Unix(1800000000, 0)
	issuer := "https://identity.example.test/realms/fixture"
	key, foreign := testKey(t), testKey(t)
	v := fixedVerifier(issuer, &oidc.StaticKeySet{PublicKeys: []crypto.PublicKey{&key.PublicKey}}, now)
	for _, tt := range []struct {
		name                                          string
		change                                        func(map[string]any)
		status                                        int
		alg                                           jose.SignatureAlgorithm
		foreign, tamper, none, absent, disabled, role bool
		lookupErr                                     error
	}{
		{name: "CA1 absent", absent: true, status: 401},
		{name: "CA2 valid", status: 200},
		{name: "CA3 foreign key", foreign: true, status: 401},
		{name: "CA3 wrong issuer", change: func(c map[string]any) { c["iss"] = "https://foreign.example.test" }, status: 401},
		{name: "CA3 tampered", tamper: true, status: 401},
		{name: "CA4 expired", change: func(c map[string]any) { c["exp"] = now.Add(-time.Second).Unix() }, status: 401},
		{name: "CA5 other client", change: func(c map[string]any) { c["aud"] = "other-client" }, status: 401},
		{name: "CA5 absent aud with azp", change: func(c map[string]any) { delete(c, "aud"); c["azp"] = "guild-banker-api" }, status: 401},
		{name: "CA5 audience array", change: func(c map[string]any) { c["aud"] = []string{"other", "guild-banker-api"} }, status: 200},
		{name: "CA6 without role", role: true, status: 403},
		{name: "CA6 with role", role: true, change: func(c map[string]any) { c["realm_access"] = map[string]any{"roles": []string{"admin"}} }, status: 200},
		{name: "CA8 none", none: true, status: 401},
		{name: "CA8 HS256", alg: jose.HS256, status: 401},
		{name: "CA8 RS512", alg: jose.RS512, status: 401},
		{name: "email conflict", lookupErr: user.ErrEmailConflict, status: 403},
		{name: "CA10 disabled local user", disabled: true, status: 403},
		{name: "CA10 lookup failure", lookupErr: errors.New("raw lookup failure fixture-sub"), status: 500},
		{name: "CA14 ID token", change: func(c map[string]any) { c["typ"] = "ID" }, status: 401},
		{name: "CA14 refresh", change: func(c map[string]any) { c["typ"] = "Refresh" }, status: 401},
		{name: "CA14 missing sub", change: func(c map[string]any) { delete(c, "sub") }, status: 401},
		{name: "CA14 no roles", change: func(c map[string]any) { delete(c, "realm_access") }, status: 200},
	} {
		t.Run(tt.name, func(t *testing.T) {
			claims := testClaims(issuer, now)
			if tt.change != nil {
				tt.change(claims)
			}
			var signingKey any = key
			alg := tt.alg
			if alg == "" {
				alg = jose.RS256
			}
			if tt.foreign {
				signingKey = foreign
			}
			if alg == jose.HS256 {
				signingKey = []byte(strings.Repeat("x", 32))
			}
			raw := signToken(t, claims, signingKey, "A", alg)
			parts := strings.Split(raw, ".")
			if tt.none {
				raw = base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"none"}`)) + "." + parts[1] + "."
			}
			if tt.tamper {
				claims["sub"] = "attacker"
				payload, err := json.Marshal(claims)
				require.NoError(t, err)
				raw = parts[0] + "." + base64.RawURLEncoding.EncodeToString(payload) + "." + parts[2]
			}
			users := &fixtureUsers{err: tt.lookupErr, disabled: tt.disabled}
			var logs bytes.Buffer
			logger := slog.New(slog.NewJSONHandler(&logs, nil))
			router := gin.New()
			router.Use(middleware.Auth(auth.NewService(v, users), logger, time.Second))
			handled := false
			handler := func(c *gin.Context) {
				handled = true
				principal, ok := auth.PrincipalFromContext(c.Request.Context())
				require.True(t, ok)
				assert.Equal(t, "fixture-sub", principal.Claims.UserID)
				assert.Equal(t, "fixture-user", principal.Claims.PreferredUsername)
				assert.Equal(t, audit.Actor{UserID: 42, Email: "local@example.test"}, principal.Actor)
				assert.Equal(t, "999", c.GetHeader("X-User-ID"), "headers must not be rewritten")
				c.Status(200)
			}
			router.GET("/api/probe", handler)
			router.GET("/api/admin", middleware.RequireRole("admin", logger), handler)
			path := "/api/probe"
			if tt.role {
				path = "/api/admin"
			}
			req := httptest.NewRequest("GET", path, nil)
			req.Header.Set("X-User-ID", "999")
			req.Header.Set("X-User-Email", "spoofed@example.test")
			if !tt.absent {
				req.Header.Set("Authorization", "Bearer "+raw)
			}
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			assert.Equal(t, tt.status, w.Code)
			assert.Equal(t, tt.status == 200, handled)
			if tt.status == 401 {
				assert.Zero(t, users.calls)
				assert.JSONEq(t, `{"error":"unauthorized"}`, w.Body.String())
				return
			}
			assert.Equal(t, 1, users.calls)
			assert.Equal(t, "fixture-sub", users.sub)
			if tt.status == 403 {
				assert.JSONEq(t, `{"error":"forbidden"}`, w.Body.String())
			}
			if tt.status == 500 {
				assert.JSONEq(t, `{"error":"internal server error"}`, w.Body.String())
			}
		})
	}
}

func TestAuthHTTPRotation(t *testing.T) {
	keyA, keyB, keyC := testKey(t), testKey(t), testKey(t)
	f := newOIDCFixture(t, jose.JSONWebKey{Key: &keyA.PublicKey, KeyID: "A", Use: "sig", Algorithm: "RS256"})
	now := time.Unix(1800000000, 0)
	ctx := oidc.ClientContext(context.Background(), &http.Client{Timeout: time.Second})
	v := fixedVerifier(f.server.URL, oidc.NewRemoteKeySet(ctx, f.server.URL+"/jwks"), now)
	users := &fixtureUsers{}
	var logs bytes.Buffer
	router := gin.New()
	router.Use(middleware.Auth(auth.NewService(v, users), slog.New(slog.NewJSONHandler(&logs, nil)), time.Second))
	router.GET("/api/probe", func(c *gin.Context) { c.Status(200) })
	request := func(raw string, status int) {
		t.Helper()
		req := httptest.NewRequest("GET", "/api/probe", nil)
		req.Header.Set("Authorization", "Bearer "+raw)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		require.Equal(t, status, w.Code)
	}
	claims := testClaims(f.server.URL, now)
	rawA := signToken(t, claims, keyA, "A", jose.RS256)
	request(rawA, 200)
	request(rawA, 200)
	assert.EqualValues(t, 1, f.requests.Load())
	f.mu.Lock()
	f.keys = []jose.JSONWebKey{{Key: &keyB.PublicKey, KeyID: "B", Use: "sig", Algorithm: "RS256"}}
	f.mu.Unlock()
	request(signToken(t, claims, keyB, "B", jose.RS256), 200)
	assert.EqualValues(t, 2, f.requests.Load())
	request(signToken(t, claims, keyC, "C", jose.RS256), 401)
	assert.Equal(t, 3, users.calls)
}
