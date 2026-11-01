package keycloak

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/luizhenrique-dev/guild-banker/api/internal/auth"
)

type Verifier struct {
	verifier *oidc.IDTokenVerifier
}

func NewVerifier(ctx context.Context, issuer, audience string, timeout time.Duration) (*Verifier, error) {
	if timeout <= 0 || issuer == "" || audience == "" {
		return nil, errors.New("keycloak: invalid verifier configuration")
	}
	client := &http.Client{Timeout: timeout}
	ctx = oidc.ClientContext(ctx, client)
	provider, err := oidc.NewProvider(ctx, issuer)
	if err != nil {
		return nil, fmt.Errorf("keycloak discovery: %w", err)
	}
	return &Verifier{verifier: provider.VerifierContext(ctx, verifierConfig(audience))}, nil
}

func verifierConfig(audience string) *oidc.Config {
	return &oidc.Config{ClientID: audience, SupportedSigningAlgs: []string{oidc.RS256}}
}

type tokenClaims struct {
	Type              string      `json:"typ"`
	PreferredUsername claimString `json:"preferred_username"`
	RealmAccess       struct {
		Roles []claimString `json:"roles"`
	} `json:"realm_access"`
}

// claimString rejects JSON null, including null elements in the roles array.
type claimString string

func (s *claimString) UnmarshalJSON(data []byte) error {
	var value *string
	if err := json.Unmarshal(data, &value); err != nil {
		return fmt.Errorf("decode string claim: %w", err)
	}
	if value == nil {
		return errors.New("string claim must not be null")
	}
	*s = claimString(*value)
	return nil
}

func (v *Verifier) Verify(ctx context.Context, raw string) (auth.Claims, error) {
	token, err := v.verifier.Verify(ctx, raw)
	if err != nil {
		reason := auth.ReasonInvalidToken
		if _, ok := errors.AsType[*oidc.TokenExpiredError](err); ok {
			reason = auth.ReasonTokenExpired
		}
		return auth.Claims{}, &auth.RejectionError{Reason: reason}
	}
	var claims tokenClaims
	if err := token.Claims(&claims); err != nil {
		return auth.Claims{}, &auth.RejectionError{Reason: auth.ReasonInvalidToken}
	}
	if claims.Type != "Bearer" {
		return auth.Claims{}, &auth.RejectionError{Reason: auth.ReasonInvalidTokenType}
	}
	if token.Subject == "" {
		return auth.Claims{}, &auth.RejectionError{Reason: auth.ReasonMissingSubject}
	}
	roles := make([]string, 0, len(claims.RealmAccess.Roles))
	for _, role := range claims.RealmAccess.Roles {
		roles = append(roles, string(role))
	}
	return auth.Claims{
		UserID:            token.Subject,
		PreferredUsername: string(claims.PreferredUsername),
		RealmRoles:        roles,
	}, nil
}
