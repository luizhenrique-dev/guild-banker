//go:build integration

package main

import (
	"bytes"
	"context"
	"github.com/luizhenrique-dev/guild-banker/api/config"
	"github.com/stretchr/testify/require"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestBootstrapDiscoveryFailure(t *testing.T) {
	remote := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusServiceUnavailable) }))
	defer remote.Close()
	var logs bytes.Buffer
	for _, tt := range []struct {
		name, issuer, audience string
		timeout                int
	}{
		{name: "discovery unavailable", issuer: remote.URL, audience: "fixture-api", timeout: 1},
		{name: "missing configuration", audience: "fixture-api", timeout: 1},
		{name: "invalid timeout", issuer: remote.URL, audience: "fixture-api"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.Config{Keycloak: config.KeycloakConfig{Issuer: tt.issuer, Audience: tt.audience, Timeout: tt.timeout}}
			server, err := newServer(context.Background(), cfg, nil, slog.New(slog.NewJSONHandler(&logs, nil)))
			require.Error(t, err)
			require.Nil(t, server)
		})
	}
}
