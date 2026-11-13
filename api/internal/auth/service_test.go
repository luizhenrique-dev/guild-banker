package auth

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/luizhenrique-dev/guild-banker/api/internal/audit"
	"github.com/luizhenrique-dev/guild-banker/api/internal/emailaddress"
	"github.com/luizhenrique-dev/guild-banker/api/internal/user"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type verifierFunc func(context.Context, string) (Claims, error)

func (f verifierFunc) Verify(ctx context.Context, raw string) (Claims, error) { return f(ctx, raw) }

type userLookupFunc func(context.Context, string, string, string) (*user.User, error)

func (f userLookupFunc) EnsureAccount(ctx context.Context, sub, name, email string) (*user.User, error) {
	return f(ctx, sub, name, email)
}

func (f userLookupFunc) RecordAccess(context.Context, int64) error { return nil }

type accessRecorder struct {
	userLookupFunc
	record func(context.Context, int64) error
}

func (r accessRecorder) RecordAccess(ctx context.Context, id int64) error { return r.record(ctx, id) }

func TestServiceRecordsAuthenticatedAccess(t *testing.T) {
	for _, disabled := range []bool{false, true} {
		t.Run(fmt.Sprint("disabled=", disabled), func(t *testing.T) {
			calls := 0
			cause := errors.New("access persistence failed")
			users := accessRecorder{
				userLookupFunc: func(context.Context, string, string, string) (*user.User, error) {
					return &user.User{ID: 42, Enabled: !disabled}, nil
				},
				record: func(ctx context.Context, id int64) error {
					calls++
					assert.Equal(t, t.Context(), ctx)
					assert.EqualValues(t, 42, id)
					return cause
				},
			}
			verify := verifierFunc(func(context.Context, string) (Claims, error) {
				return Claims{Name: "Member", Email: "member@example.test", EmailVerified: true}, nil
			})
			principal, err := NewService(verify, users).Authenticate(t.Context(), "token")
			assert.Empty(t, principal)
			if disabled {
				assert.ErrorIs(t, err, ErrForbidden)
				assert.Zero(t, calls)
			} else {
				assert.ErrorIs(t, err, cause)
				assert.Equal(t, 1, calls)
			}
		})
	}
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
		{name: "email conflict", lookupErr: user.ErrEmailConflict, wantErr: ErrForbidden},
		{name: "invalid email profile", lookupErr: emailaddress.ErrInvalid, wantErr: ErrForbidden},
		{name: "disabled resolution", lookupErr: user.ErrDisabled, wantErr: ErrForbidden},
		{name: "persistence canceled", lookupErr: context.Canceled, wantErr: context.Canceled},
		{name: "disabled user", disabled: true, wantErr: ErrForbidden},
		{name: "technical failure", lookupErr: dbErr, wantErr: dbErr},
	} {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
			defer cancel()
			claims := Claims{Name: "Member", Email: "member@example.test", EmailVerified: true, UserID: "non-numeric-external-sub", PreferredUsername: "untrusted-username", RealmRoles: []string{"admin"}}
			order := []string{}
			verify := verifierFunc(func(got context.Context, raw string) (Claims, error) {
				assert.Equal(t, ctx, got)
				assert.Equal(t, "raw-token", raw)
				order = append(order, "verify")
				return claims, tt.verifyErr
			})
			lookup := userLookupFunc(func(got context.Context, sub, name, email string) (*user.User, error) {
				assert.Equal(t, ctx, got)
				assert.Equal(t, claims.UserID, sub)
				assert.Equal(t, claims.Name, name)
				assert.Equal(t, claims.Email, email)
				order = append(order, "lookup")
				return &user.User{ID: 42, Email: "local@example.test", Enabled: !tt.disabled}, tt.lookupErr
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
	lookup := userLookupFunc(func(context.Context, string, string, string) (*user.User, error) {
		t.Fatal("cancelled verification must not resolve a user")
		return nil, nil
	})
	_, err := NewService(verify, lookup).Authenticate(ctx, "raw")
	require.ErrorIs(t, err, context.Canceled)
}

func TestServiceAuthenticateRequiresProfile(t *testing.T) {
	for _, profile := range []Claims{
		{Name: "", Email: "member@example.test", EmailVerified: true},
		{Name: " ", Email: "member@example.test", EmailVerified: true},
		{Name: "Member", Email: "", EmailVerified: true},
		{Name: "Member", Email: "\t", EmailVerified: true},
		{Name: "Member", Email: "member@example.test"},
	} {
		calls := 0
		verify := verifierFunc(func(context.Context, string) (Claims, error) { return profile, nil })
		users := userLookupFunc(func(context.Context, string, string, string) (*user.User, error) {
			calls++
			return &user.User{ID: 42, Enabled: true}, nil
		})
		got, err := NewService(verify, users).Authenticate(context.Background(), "raw")
		require.ErrorIs(t, err, ErrForbidden)
		assert.Empty(t, got)
		assert.Zero(t, calls)
	}
}
