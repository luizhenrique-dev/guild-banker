//go:build integration

package keycloak

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/go-jose/go-jose/v4"
	"github.com/luizhenrique-dev/guild-banker/api/internal/auth"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type oidcFixture struct {
	server   *httptest.Server
	mu       sync.Mutex
	keys     []jose.JSONWebKey
	status   int
	requests atomic.Int64
}

func newOIDCFixture(t *testing.T, key jose.JSONWebKey) *oidcFixture {
	t.Helper()
	f := &oidcFixture{keys: []jose.JSONWebKey{key}, status: http.StatusOK}
	f.server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/.well-known/openid-configuration":
			err := json.NewEncoder(w).Encode(map[string]any{
				"issuer": f.server.URL, "jwks_uri": f.server.URL + "/jwks",
				"id_token_signing_alg_values_supported": []string{"RS256", "RS512"},
			})
			assert.NoError(t, err)
		case "/jwks":
			f.requests.Add(1)
			f.mu.Lock()
			defer f.mu.Unlock()
			w.WriteHeader(f.status)
			assert.NoError(t, json.NewEncoder(w).Encode(jose.JSONWebKeySet{Keys: f.keys}))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	t.Cleanup(f.server.Close)
	return f
}

func TestProviderDiscovery(t *testing.T) {
	for _, tt := range []struct {
		name              string
		status            int
		mismatch, timeout bool
	}{
		{name: "success", status: http.StatusOK},
		{name: "issuer mismatch", status: http.StatusOK, mismatch: true},
		{name: "HTTP failure", status: http.StatusServiceUnavailable},
		{name: "timeout", timeout: true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			var server *httptest.Server
			server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if tt.timeout {
					<-r.Context().Done()
					return
				}
				issuer := server.URL
				if tt.mismatch {
					issuer = "https://foreign.example.test"
				}
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.status)
				assert.NoError(t, json.NewEncoder(w).Encode(map[string]string{
					"issuer": issuer, "jwks_uri": server.URL + "/jwks",
				}))
			}))
			defer server.Close()
			v, err := NewVerifier(context.Background(), server.URL, "guild-banker-api", 100*time.Millisecond)
			if tt.name == "success" {
				require.NoError(t, err)
				require.NotNil(t, v)
				return
			}
			require.Error(t, err)
			assert.Nil(t, v)
		})
	}
}

func TestVerifierJWKSRotation(t *testing.T) {
	keyA, keyB, keyC := testKey(t), testKey(t), testKey(t)
	fixture := newOIDCFixture(t, jose.JSONWebKey{Key: &keyA.PublicKey, KeyID: "A", Algorithm: "RS256", Use: "sig"})
	v, err := NewVerifier(context.Background(), fixture.server.URL, "guild-banker-api", time.Second)
	require.NoError(t, err)
	claims := testClaims(fixture.server.URL, time.Now())
	rawA := signToken(t, claims, keyA, "A", jose.RS256)
	_, err = v.Verify(context.Background(), rawA)
	require.NoError(t, err)
	_, err = v.Verify(context.Background(), rawA)
	require.NoError(t, err)
	assert.EqualValues(t, 1, fixture.requests.Load(), "known kid should use cache")
	fixture.mu.Lock()
	fixture.keys = []jose.JSONWebKey{{Key: &keyB.PublicKey, KeyID: "B", Algorithm: "RS256", Use: "sig"}}
	fixture.mu.Unlock()
	_, err = v.Verify(context.Background(), signToken(t, claims, keyB, "B", jose.RS256))
	require.NoError(t, err)
	assert.EqualValues(t, 2, fixture.requests.Load())
	_, err = v.Verify(context.Background(), signToken(t, claims, keyC, "C", jose.RS256))
	require.ErrorIs(t, err, auth.ErrInvalidToken)
	fixture.mu.Lock()
	fixture.status = http.StatusServiceUnavailable
	fixture.mu.Unlock()
	_, err = v.Verify(context.Background(), signToken(t, claims, keyB, "B", jose.RS256))
	require.NoError(t, err, "a trusted cached key remains usable while JWKS is unavailable")
	_, err = v.Verify(context.Background(), signToken(t, claims, keyC, "C", jose.RS256))
	require.ErrorIs(t, err, auth.ErrInvalidToken)
}

func TestVerifierCancellation(t *testing.T) {
	key := testKey(t)
	started := make(chan struct{})
	finished := make(chan struct{})
	var server *httptest.Server
	server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/jwks" {
			close(started)
			<-r.Context().Done()
			close(finished)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		assert.NoError(t, json.NewEncoder(w).Encode(map[string]string{
			"issuer": server.URL, "jwks_uri": server.URL + "/jwks",
		}))
	}))
	defer server.Close()
	v, err := NewVerifier(context.Background(), server.URL, "guild-banker-api", 200*time.Millisecond)
	require.NoError(t, err)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	result := make(chan error, 1)
	raw := signToken(t, testClaims(server.URL, time.Now()), key, "A", jose.RS256)
	go func() {
		_, err := v.Verify(ctx, raw)
		result <- err
	}()
	select {
	case <-started:
	case <-time.After(2 * time.Second):
		t.Fatal("JWKS request not started")
	}
	cancel()
	select {
	case err := <-result:
		require.ErrorIs(t, err, auth.ErrInvalidToken)
	case <-time.After(2 * time.Second):
		t.Fatal("request cancellation not propagated")
	}
	select {
	case <-finished:
	case <-time.After(2 * time.Second):
		t.Fatal("HTTP timeout did not bound shared JWKS fetch")
	}
}

func TestDiscoveryCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := NewVerifier(ctx, "http://127.0.0.1:1", "guild-banker-api", time.Second)
	require.ErrorIs(t, err, context.Canceled)
}
