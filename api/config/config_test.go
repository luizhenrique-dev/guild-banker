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
}

func TestLoad(t *testing.T) {
	t.Run("process environment without file", func(t *testing.T) {
		configEnvironment(t)
		cfg, err := Load()
		require.NoError(t, err)
		assert.Equal(t, "fixture-api", cfg.Keycloak.Audience)
		assert.Equal(t, "http://localhost:8080/realms/fixture", cfg.Keycloak.Issuer)
		assert.Equal(t, 5, cfg.Keycloak.Timeout)
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
