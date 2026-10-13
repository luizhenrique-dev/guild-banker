package auth

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/luizhenrique-dev/guild-banker/api/internal/audit"
	"github.com/luizhenrique-dev/guild-banker/api/internal/user"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type verifierFunc func(context.Context, string) (Claims, error)

func (f verifierFunc) Verify(ctx context.Context, raw string) (Claims, error) { return f(ctx, raw) }

type userLookupFunc func(context.Context, string) (*user.User, error)

func (f userLookupFunc) GetByExternalID(ctx context.Context, sub string) (*user.User, error) {
	return f(ctx, sub)
}

func TestServiceAuthenticate(t *testing.T) {
	dbErr := errors.New("database unavailable with sensitive detail")
	for _, tt := range []struct {
		name                          string
		verifyErr, lookupErr, wantErr error
		disabled                      bool
	}{
		{name: "active"},
		{name: "invalid token", verifyErr: &RejectionError{Reason: ReasonInvalidToken}, wantErr: ErrInvalidToken},
		{name: "missing user", lookupErr: fmt.Errorf("get user: %w", user.ErrNotFound), wantErr: ErrForbidden},
		{name: "disabled user", disabled: true, wantErr: ErrForbidden},
		{name: "technical failure", lookupErr: dbErr, wantErr: dbErr},
	} {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
			defer cancel()
			claims := Claims{UserID: "non-numeric-external-sub", PreferredUsername: "untrusted-username", RealmRoles: []string{"admin"}}
			order := []string{}
			verify := verifierFunc(func(got context.Context, raw string) (Claims, error) {
				assert.Equal(t, ctx, got)
				assert.Equal(t, "raw-token", raw)
				order = append(order, "verify")
				return claims, tt.verifyErr
			})
			lookup := userLookupFunc(func(got context.Context, sub string) (*user.User, error) {
				assert.Equal(t, ctx, got)
				assert.Equal(t, claims.UserID, sub)
				order = append(order, "lookup")
				return &user.User{ID: 42, Email: "local@example.test", DisableEntry: audit.DisableEntry{Enabled: !tt.disabled}}, tt.lookupErr
			})
			principal, err := NewService(verify, lookup).Authenticate(ctx, "raw-token")
			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				assert.Empty(t, principal)
			} else {
				require.NoError(t, err)
				assert.Equal(t, claims, principal.Claims)
				assert.Equal(t, audit.Actor{UserID: 42, Email: "local@example.test"}, principal.Actor)
			}
			if tt.verifyErr != nil {
				assert.Equal(t, []string{"verify"}, order)
				return
			}
			assert.Equal(t, []string{"verify", "lookup"}, order)
		})
	}
}

func TestServiceAuthenticateContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	verify := verifierFunc(func(got context.Context, _ string) (Claims, error) {
		assert.ErrorIs(t, got.Err(), context.Canceled)
		return Claims{}, context.Canceled
	})
	lookup := userLookupFunc(func(context.Context, string) (*user.User, error) {
		t.Fatal("cancelled verification must not resolve a user")
		return nil, nil
	})
	_, err := NewService(verify, lookup).Authenticate(ctx, "raw")
	require.ErrorIs(t, err, context.Canceled)
}
