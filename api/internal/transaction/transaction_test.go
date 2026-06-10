package transaction

import (
	"testing"
	"time"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	t.Run("defaults", func(t *testing.T) {
		occurredAt := time.Now().Add(-time.Hour)
		transaction, err := New(
			TypeExpense,
			"Grocery",
			decimal.NewFromFloat(100.50),
			CategoryPersonal,
			VisibilityPublic,
			occurredAt,
			10,
			20,
			"test@example.com",
		)

		require.NoError(t, err)
		require.Equal(t, StatusActive, transaction.Status)
		require.Equal(t, SourceManual, transaction.Source)
	})

	t.Run("invalid amount", func(t *testing.T) {
		_, err := New(
			TypeExpense,
			"Grocery",
			decimal.Zero,
			CategoryPersonal,
			VisibilityPublic,
			time.Now(),
			10,
			20,
			"test@example.com",
		)

		require.EqualError(t, err, "amount must be greater than zero")
	})
}

func TestTransaction_SetVisibility(t *testing.T) {
	transaction, err := New(
		TypeExpense,
		"Grocery",
		decimal.NewFromInt(50),
		CategoryPersonal,
		VisibilityPublic,
		time.Now(),
		10,
		20,
		"test@example.com",
	)
	require.NoError(t, err)

	err = transaction.SetVisibility(VisibilityPrivate, "test@example.com")
	require.NoError(t, err)
	require.Equal(t, VisibilityPrivate, transaction.Visibility)
	require.NotNil(t, transaction.UpdatedAt)
}
