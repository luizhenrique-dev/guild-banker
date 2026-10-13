package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func configEnvironment(t *testing.T) {
	t.Helper()
	t.Chdir(t.TempDir())
	for _, key := range []string{"DB_HOST", "DB_PORT", "DB_USER", "DB_PASSWORD", "DB_NAME"} {
		t.Setenv(key, "test")
	}
	t.Setenv("KEYCLOAK_ISSUER", "http://localhost:8080/realms/fixture")
	t.Setenv("API_AUDIENCE", "fixture-api")
	t.Setenv("KEYCLOAK_TIMEOUT_SECONDS", "")
	t.Setenv("KEYCLOAK_BASE_URL", "")
	t.Setenv("KEYCLOAK_REALM", "")
	t.Setenv("CORS_ALLOWED_ORIGINS", "http://localhost:5173,http://127.0.0.1:5173")
}

func TestLoadCORSOrigins(t *testing.T) {
	for _, tt := range []struct {
		name    string
		origins string
		valid   bool
	}{
		{name: "UI list", origins: " http://localhost:5173 , http://127.0.0.1:5173 ", valid: true},
		{name: "HTTPS", origins: "https://ui.example.test", valid: true},
		{name: "uppercase scheme", origins: "HTTP://localhost:5173", valid: true},
		{name: "IPv6", origins: "http://[::1]:5173", valid: true},
		{name: "missing"},
		{name: "blank", origins: " "},
		{name: "wildcard", origins: "*"},
		{name: "host wildcard", origins: "https://*.example.test"},
		{name: "empty list element", origins: "http://localhost:5173,"},
		{name: "no scheme", origins: "localhost:5173"},
		{name: "invalid scheme", origins: "ftp://localhost:5173"},
		{name: "no host", origins: "https://"},
		{name: "path", origins: "https://ui.example.test/app"},
		{name: "root path", origins: "https://ui.example.test/"},
		{name: "query", origins: "https://ui.example.test?secret=query"},
		{name: "empty query", origins: "https://ui.example.test?"},
		{name: "fragment", origins: "https://ui.example.test#fragment"},
		{name: "empty fragment", origins: "https://ui.example.test#"},
		{name: "unbracketed IPv6", origins: "http://::1"},
		{name: "credentials", origins: "https://user:secret@ui.example.test"},
		{name: "non-numeric port", origins: "https://ui.example.test:abc"},
		{name: "empty port", origins: "https://ui.example.test:"},
		{name: "zero port", origins: "https://ui.example.test:0"},
		{name: "overflow port", origins: "https://ui.example.test:65536"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			configEnvironment(t)
			t.Setenv("CORS_ALLOWED_ORIGINS", tt.origins)
			cfg, err := Load()
			if tt.valid {
				require.NoError(t, err)
				if tt.name == "UI list" {
					assert.Equal(t, []string{"http://localhost:5173", "http://127.0.0.1:5173"}, cfg.CORSAllowedOrigins)
				}
				if tt.name == "uppercase scheme" {
					assert.Equal(t, []string{"http://localhost:5173"}, cfg.CORSAllowedOrigins)
				}
				return
			}
			require.Error(t, err)
			assert.EqualError(t, err, "config: CORS_ALLOWED_ORIGINS must contain only complete HTTP(S) origins")
		})
	}
}

func TestLoad(t *testing.T) {
	t.Run("process environment without file", func(t *testing.T) {
		configEnvironment(t)
		cfg, err := Load()
		require.NoError(t, err)
		assert.Equal(t, "fixture-api", cfg.Keycloak.Audience)
		assert.Equal(t, "http://localhost:8080/realms/fixture", cfg.Keycloak.Issuer)
		assert.Equal(t, 5, cfg.Keycloak.Timeout)
		assert.Equal(t, []string{"http://localhost:5173", "http://127.0.0.1:5173"}, cfg.CORSAllowedOrigins)
	})
	t.Run("process wins over file", func(t *testing.T) {
		configEnvironment(t)
		require.NoError(t, os.Mkdir("api", 0700))
		require.NoError(t, os.WriteFile(filepath.Join("api", "config.env"), []byte("API_AUDIENCE=file-api\n"), 0600))
		cfg, err := Load()
		require.NoError(t, err)
		assert.Equal(t, "fixture-api", cfg.Keycloak.Audience)
	})
	t.Run("file read error", func(t *testing.T) {
		configEnvironment(t)
		require.NoError(t, os.MkdirAll(filepath.Join("api", "config.env"), 0700))
		_, err := Load()
		require.Error(t, err)
	})
	for _, tt := range []struct{ name, key, value string }{
		{name: "missing issuer", key: "KEYCLOAK_ISSUER"},
		{name: "invalid issuer", key: "KEYCLOAK_ISSUER", value: "realm"},
		{name: "issuer query", key: "KEYCLOAK_ISSUER", value: "http://host/realm?x=1"},
		{name: "missing audience", key: "API_AUDIENCE"},
		{name: "blank audience", key: "API_AUDIENCE", value: " "},
		{name: "zero timeout", key: "KEYCLOAK_TIMEOUT_SECONDS", value: "0"},
		{name: "negative timeout", key: "KEYCLOAK_TIMEOUT_SECONDS", value: "-1"},
		{name: "invalid timeout", key: "KEYCLOAK_TIMEOUT_SECONDS", value: "abc"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			configEnvironment(t)
			t.Setenv(tt.key, tt.value)
			require.NotPanics(t, func() { _, err := Load(); require.Error(t, err) })
		})
	}
}
