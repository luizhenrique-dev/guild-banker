package keycloak

import (
	"context"
	"crypto"
	"crypto/hmac"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/base64"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/go-jose/go-jose/v4"
	"github.com/luizhenrique-dev/guild-banker/api/internal/auth"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testKey(t *testing.T) *rsa.PrivateKey {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	return key
}

func testClaims(issuer string, now time.Time) map[string]any {
	return map[string]any{
		"iss": issuer, "aud": "guild-banker-api", "exp": now.Add(time.Minute).Unix(),
		"sub": "fixture-sub", "typ": "Bearer", "preferred_username": "fixture-user",
		"realm_access": map[string]any{"roles": []string{"member"}},
	}
}

func signToken(t *testing.T, claims map[string]any, key any, kid string, alg jose.SignatureAlgorithm) string {
	t.Helper()
	header, err := json.Marshal(map[string]string{"alg": string(alg), "kid": kid, "typ": "JWT"})
	require.NoError(t, err)
	payload, err := json.Marshal(claims)
	require.NoError(t, err)
	input := base64.RawURLEncoding.EncodeToString(header) + "." + base64.RawURLEncoding.EncodeToString(payload)
	var signature []byte
	switch alg {
	case jose.RS256:
		digest := sha256.Sum256([]byte(input))
		signature, err = rsa.SignPKCS1v15(rand.Reader, key.(*rsa.PrivateKey), crypto.SHA256, digest[:])
	case jose.RS512:
		digest := sha512.Sum512([]byte(input))
		signature, err = rsa.SignPKCS1v15(rand.Reader, key.(*rsa.PrivateKey), crypto.SHA512, digest[:])
	case jose.HS256:
		mac := hmac.New(sha256.New, key.([]byte))
		_, err = mac.Write([]byte(input))
		signature = mac.Sum(nil)
	default:
		t.Fatalf("unsupported fixture algorithm: %s", alg)
	}
	require.NoError(t, err)
	return input + "." + base64.RawURLEncoding.EncodeToString(signature)
}

func fixedVerifier(issuer string, keySet oidc.KeySet, now time.Time) *Verifier {
	cfg := verifierConfig("guild-banker-api")
	cfg.Now = func() time.Time { return now }
	return &Verifier{verifier: oidc.NewVerifier(issuer, keySet, cfg)}
}

func TestVerifierVerify(t *testing.T) {
	now := time.Unix(1800000000, 0)
	issuer := "https://identity.example.test/realms/fixture"
	key := testKey(t)
	foreignKey := testKey(t)
	v := fixedVerifier(issuer, &oidc.StaticKeySet{PublicKeys: []crypto.PublicKey{&key.PublicKey}}, now)
	for _, tt := range []struct {
		name      string
		change    func(map[string]any)
		alg       jose.SignatureAlgorithm
		foreign   bool
		tamper    bool
		none      bool
		wantValid bool
		reason    auth.Reason
	}{
		{name: "valid", wantValid: true},
		{name: "array audience", change: func(c map[string]any) { c["aud"] = []string{"other", "guild-banker-api"} }, wantValid: true},
		{name: "no roles", change: func(c map[string]any) { delete(c, "realm_access") }, wantValid: true},
		{name: "no username", change: func(c map[string]any) { delete(c, "preferred_username") }, wantValid: true},
		{name: "foreign key", foreign: true},
		{name: "tampered payload", tamper: true},
		{name: "wrong issuer", change: func(c map[string]any) { c["iss"] = "https://foreign.example.test" }},
		{name: "expired", change: func(c map[string]any) { c["exp"] = now.Add(-time.Second).Unix() }, reason: auth.ReasonTokenExpired},
		{name: "missing exp", change: func(c map[string]any) { delete(c, "exp") }, reason: auth.ReasonTokenExpired},
		{name: "wrong audience", change: func(c map[string]any) { c["aud"] = "other-client" }},
		{name: "missing audience", change: func(c map[string]any) { delete(c, "aud") }},
		{name: "azp cannot replace aud", change: func(c map[string]any) { delete(c, "aud"); c["azp"] = "guild-banker-api" }},
		{name: "none", none: true},
		{name: "HS256", alg: jose.HS256},
		{name: "RS512", alg: jose.RS512},
		{name: "ID token", change: func(c map[string]any) { c["typ"] = "ID" }, reason: auth.ReasonInvalidTokenType},
		{name: "refresh token", change: func(c map[string]any) { c["typ"] = "Refresh" }, reason: auth.ReasonInvalidTokenType},
		{name: "missing type", change: func(c map[string]any) { delete(c, "typ") }, reason: auth.ReasonInvalidTokenType},
		{name: "missing subject", change: func(c map[string]any) { delete(c, "sub") }, reason: auth.ReasonMissingSubject},
		{name: "empty subject", change: func(c map[string]any) { c["sub"] = "" }, reason: auth.ReasonMissingSubject},
		{name: "wrong roles type", change: func(c map[string]any) { c["realm_access"] = map[string]any{"roles": "admin"} }},
		{name: "wrong username type", change: func(c map[string]any) { c["preferred_username"] = 123 }},
		{name: "null username", change: func(c map[string]any) { c["preferred_username"] = nil }},
		{name: "null role element", change: func(c map[string]any) { c["realm_access"] = map[string]any{"roles": []any{"member", nil}} }},
	} {
		t.Run(tt.name, func(t *testing.T) {
			claims := testClaims(issuer, now)
			if tt.change != nil {
				tt.change(claims)
			}
			alg := tt.alg
			if alg == "" {
				alg = jose.RS256
			}
			var signingKey any = key
			if tt.foreign {
				signingKey = foreignKey
			}
			if alg == jose.HS256 {
				signingKey = []byte(strings.Repeat("x", 32))
			}
			raw := signToken(t, claims, signingKey, "A", alg)
			if tt.none {
				parts := strings.Split(raw, ".")
				raw = base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"none"}`)) + "." + parts[1] + "."
			}
			if tt.tamper {
				parts := strings.Split(raw, ".")
				claims["sub"] = "attacker"
				payload, err := json.Marshal(claims)
				require.NoError(t, err)
				raw = parts[0] + "." + base64.RawURLEncoding.EncodeToString(payload) + "." + parts[2]
			}
			got, err := v.Verify(context.Background(), raw)
			if tt.wantValid {
				require.NoError(t, err)
				assert.Equal(t, "fixture-sub", got.UserID)
				if tt.name == "valid" {
					assert.Equal(t, "fixture-user", got.PreferredUsername)
					assert.Equal(t, []string{"member"}, got.RealmRoles)
				}
				return
			}
			require.ErrorIs(t, err, auth.ErrInvalidToken)
			assert.Empty(t, got)
			var rejection *auth.RejectionError
			require.ErrorAs(t, err, &rejection)
			assert.Equal(t, tt.reason, rejection.Reason)
			assert.NotContains(t, err.Error(), raw)
			assert.NotContains(t, err.Error(), "fixture-sub")
		})
	}
}
