package auth

import (
	"context"
	"testing"
	"time"

	"github.com/luizhenrique-dev/guild-banker/api/internal/audit"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPrincipalFromContext(t *testing.T) {
	ctx := context.Background()
	_, ok := PrincipalFromContext(ctx)
	assert.False(t, ok)
	_, ok = ClaimsFromContext(ctx)
	assert.False(t, ok)
	want := Principal{
		Claims: Claims{UserID: "external-sub", PreferredUsername: "fixture", RealmRoles: []string{"member"}},
		Actor:  audit.Actor{UserID: 42, Email: "fixture@example.test"},
	}
	ctx = WithPrincipal(ctx, want)
	got, ok := PrincipalFromContext(ctx)
	require.True(t, ok)
	assert.Equal(t, want, got)
	claims, ok := ClaimsFromContext(ctx)
	require.True(t, ok)
	assert.Equal(t, want.Claims, claims)
}

func TestWithPrincipalPreservesContext(t *testing.T) {
	type parentKey struct{}
	parent := context.WithValue(context.Background(), parentKey{}, "parent value")
	deadline := time.Now().Add(time.Minute)
	parent, cancel := context.WithDeadline(parent, deadline)
	defer cancel()
	ctx := WithPrincipal(parent, Principal{})
	got, ok := ctx.Deadline()
	require.True(t, ok)
	assert.Equal(t, deadline, got)
	assert.Equal(t, "parent value", ctx.Value(parentKey{}))
	cancel()
	assert.ErrorIs(t, ctx.Err(), context.Canceled)
}
