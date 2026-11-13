package guild

import (
	"context"
	"strings"
	"testing"

	"github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestCreateGeneratesGuildName(t *testing.T) {
	for _, tt := range []struct{ displayName, expected string }{
		{"Família Silva", "familia-silva"},
		{"  Fami\u0301lia & Amigos!  ", "familia-amigos"},
		{"東京 💰", "guild"},
		{strings.Repeat("Á", 70), strings.Repeat("a", 70)},
	} {
		t.Run(tt.expected, func(t *testing.T) {
			svc, storage := newService(t)
			storage.On("NameExists", mock.Anything, mock.Anything, int64(0)).Return(false, nil).Maybe()
			storage.On("Create", mock.Anything, mock.Anything, int64(1)).Return(nil)
			g, err := svc.Create(context.Background(), tt.displayName, "description", validActor())
			require.NoError(t, err)
			assert.Equal(t, tt.expected, g.Name)
			assert.Equal(t, strings.TrimSpace(tt.displayName), g.DisplayName)
		})
	}
}

func TestGuildDisplayNameLimit(t *testing.T) {
	_, err := New("valid", strings.Repeat("á", 71), "actor@example.test")
	require.Error(t, err)
}

func TestCreateRetriesOnlyNameConflicts(t *testing.T) {
	svc, storage := newService(t)
	storage.On("NameExists", mock.Anything, mock.Anything, int64(0)).Return(false, nil).Maybe()
	storage.On("Create", mock.Anything, mock.Anything, int64(1)).Return(ErrGuildNameAlreadyUsed).Once()
	storage.On("Create", mock.Anything, mock.Anything, int64(1)).Return(nil).Once()
	g, err := svc.Create(context.Background(), strings.Repeat("á", 70), "", validActor())
	require.NoError(t, err)
	assert.Len(t, g.Name, 70)
	assert.Regexp(t, `^a{61}-[a-f0-9]{8}$`, g.Name)
	storage.AssertExpectations(t)
}

func TestNameConflictClassification(t *testing.T) {
	require.True(t, isNameConflict(&pq.Error{Code: "23505", Constraint: "guild_name_key"}))
	require.True(t, isNameConflict(&pq.Error{Code: "23505", Constraint: "idx_guild_name_lower"}))
	require.False(t, isNameConflict(&pq.Error{Code: "23505", Constraint: "guild_member_pkey"}))
	require.False(t, isNameConflict(&pq.Error{Code: "23503", Constraint: "guild_name_key"}))
}

func TestNameGenerationAttemptsAreBounded(t *testing.T) {
	svc, storage := newService(t)
	storage.On("Create", mock.Anything, mock.Anything, int64(1)).Return(ErrGuildNameAlreadyUsed).Times(6)
	_, err := svc.Create(context.Background(), "Repeated", "", validActor())
	require.ErrorIs(t, err, ErrNameGenerationUnavailable)
	storage.AssertExpectations(t)
}
