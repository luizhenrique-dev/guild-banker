//go:build integration

package keycloak

import (
	"context"
	"crypto"
	"encoding/base64"
	"encoding/json"
	"io"
	"log/slog"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/gin-gonic/gin"
	"github.com/go-jose/go-jose/v4"
	"github.com/luizhenrique-dev/guild-banker/api/internal/audit"
	"github.com/luizhenrique-dev/guild-banker/api/internal/auth"
	"github.com/luizhenrique-dev/guild-banker/api/internal/infra/webserver/middleware"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type keycloakMetadata struct {
	Version    string   `json:"keycloak_version"`
	Issuer     string   `json:"issuer"`
	Audience   string   `json:"audience"`
	KeyID      string   `json:"kid"`
	Subject    string   `json:"sub"`
	Username   string   `json:"preferred_username"`
	RealmRoles []string `json:"realm_roles"`
	IssuedAt   int64    `json:"issued_at"`
	ExpiresAt  int64    `json:"expires_at"`
	ValidAt    int64    `json:"valid_at"`
}

func loadKeycloakFixture(t *testing.T) (string, *oidc.StaticKeySet, keycloakMetadata) {
	t.Helper()
	read := func(name string) []byte {
		t.Helper()
		data, err := os.ReadFile(filepath.Join("testdata", "keycloak", name))
		require.NoError(t, err)
		return data
	}
	var meta keycloakMetadata
	require.NoError(t, json.Unmarshal(read("metadata.json"), &meta))
	require.Equal(t, "26.7.5", meta.Version)
	require.Equal(t, "guild-banker-api", meta.Audience)
	require.Greater(t, meta.ValidAt, meta.IssuedAt)
	require.Less(t, meta.ValidAt, meta.ExpiresAt)

	jwks := read("jwks.json")
	var publicJSON struct {
		Keys []map[string]json.RawMessage `json:"keys"`
	}
	require.NoError(t, json.Unmarshal(jwks, &publicJSON))
	for _, key := range publicJSON.Keys {
		for _, private := range []string{"d", "p", "q", "dp", "dq", "qi", "oth", "k"} {
			require.NotContains(t, key, private, "fixture must contain only public keys")
		}
	}
	var keys jose.JSONWebKeySet
	require.NoError(t, json.Unmarshal(jwks, &keys))
	keySet := &oidc.StaticKeySet{}
	for _, key := range keys.Keys {
		require.True(t, key.IsPublic())
		if key.KeyID == meta.KeyID {
			require.Equal(t, "RS256", key.Algorithm)
			require.Equal(t, "sig", key.Use)
			keySet.PublicKeys = append(keySet.PublicKeys, crypto.PublicKey(key.Key))
		}
	}
	require.Len(t, keySet.PublicKeys, 1)
	raw := strings.TrimSpace(string(read("access-token.jwt")))
	parts := strings.Split(raw, ".")
	require.Len(t, parts, 3)
	header, err := base64.RawURLEncoding.DecodeString(parts[0])
	require.NoError(t, err)
	var fields map[string]string
	require.NoError(t, json.Unmarshal(header, &fields))
	require.Equal(t, meta.KeyID, fields["kid"])
	require.Equal(t, "RS256", fields["alg"])
	return raw, keySet, meta
}

func TestVerifierKeycloakFixture(t *testing.T) {
	raw, keys, meta := loadKeycloakFixture(t)
	v := fixedVerifier(meta.Issuer, keys, time.Unix(meta.ValidAt, 0))
	claims, err := v.Verify(context.Background(), raw)
	require.NoError(t, err)
	assert.Equal(t, auth.Claims{
		UserID: meta.Subject, PreferredUsername: meta.Username, RealmRoles: meta.RealmRoles,
	}, claims)
	assert.True(t, claims.HasRole("member"))
	assert.False(t, claims.HasRole("admin"))

	v = fixedVerifier(meta.Issuer, keys, time.Unix(meta.ExpiresAt+1, 0))
	_, err = v.Verify(context.Background(), raw)
	require.ErrorIs(t, err, auth.ErrInvalidToken)
	var rejection *auth.RejectionError
	require.ErrorAs(t, err, &rejection)
	assert.Equal(t, auth.ReasonTokenExpired, rejection.Reason)
}

func TestAuthKeycloakFixture(t *testing.T) {
	raw, keys, meta := loadKeycloakFixture(t)
	for _, tt := range []struct {
		name   string
		now    int64
		status int
	}{
		{name: "valid", now: meta.ValidAt, status: 200},
		{name: "expired", now: meta.ExpiresAt + 1, status: 401},
	} {
		t.Run(tt.name, func(t *testing.T) {
			v := fixedVerifier(meta.Issuer, keys, time.Unix(tt.now, 0))
			users := &fixtureUsers{}
			router := gin.New()
			logger := slog.New(slog.NewJSONHandler(io.Discard, nil))
			router.Use(middleware.Auth(auth.NewService(v, users), logger, time.Second))
			handled := false
			router.GET("/api/probe", func(c *gin.Context) {
				handled = true
				principal, ok := auth.PrincipalFromContext(c.Request.Context())
				require.True(t, ok)
				assert.Equal(t, meta.Subject, principal.Claims.UserID)
				assert.Equal(t, meta.Username, principal.Claims.PreferredUsername)
				assert.Equal(t, meta.RealmRoles, principal.Claims.RealmRoles)
				assert.Equal(t, audit.Actor{UserID: 42, Email: "local@example.test"}, principal.Actor)
				c.Status(200)
			})
			req := httptest.NewRequest("GET", "/api/probe", nil)
			req.Header.Set("Authorization", "Bearer "+raw)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			assert.Equal(t, tt.status, w.Code)
			assert.Equal(t, tt.status == 200, handled)
			if tt.status == 200 {
				assert.Equal(t, 1, users.calls)
				assert.Equal(t, meta.Subject, users.sub)
			} else {
				assert.Zero(t, users.calls)
				assert.JSONEq(t, `{"error":"unauthorized"}`, w.Body.String())
			}
		})
	}
}
