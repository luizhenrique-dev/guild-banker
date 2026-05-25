package fixedexpense

import (
	"testing"
	"time"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

var (
	validName      = "Internet"
	validAmount    = decimal.NewFromFloat(99.90)
	validDueDay    = 10
	validCategory  = CategorySubscriptions
	validCreatedBy = "admin@example.com"
	validChangedBy = "foo@example.com"
)

func newValidFixedExpense(t *testing.T) *FixedExpense {
	t.Helper()
	fe, err := New(validName, validAmount, validDueDay, validCategory, validCreatedBy)
	assert.NoError(t, err)
	return fe
}

func TestFixedExpense_New(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		fe, err := New(validName, validAmount, validDueDay, validCategory, validCreatedBy)

		assert.NoError(t, err)
		assert.EqualValues(t, 0, fe.ID)
		assert.Equal(t, validName, fe.Name)
		assert.True(t, fe.Amount.Equal(validAmount))
		assert.Equal(t, validDueDay, fe.DueDay)
		assert.Equal(t, validCategory, fe.Category)
		assert.Equal(t, StatusActive, fe.Status)
		assert.Equal(t, validCreatedBy, fe.CreatedBy)
		assert.False(t, fe.CreatedAt.IsZero())
		assert.Nil(t, fe.UpdatedAt)
		assert.Nil(t, fe.UpdatedBy)
	})

	t.Run("error when name is empty", func(t *testing.T) {
		fe, err := New("", validAmount, validDueDay, validCategory, validCreatedBy)

		assert.Error(t, err)
		assert.Nil(t, fe)
		assert.EqualError(t, err, "name is required")
	})

	t.Run("error when amount is zero", func(t *testing.T) {
		fe, err := New(validName, decimal.Zero, validDueDay, validCategory, validCreatedBy)

		assert.Error(t, err)
		assert.Nil(t, fe)
		assert.EqualError(t, err, "amount must be greater than zero")
	})

	t.Run("error when amount is negative", func(t *testing.T) {
		fe, err := New(validName, decimal.NewFromFloat(-1.00), validDueDay, validCategory, validCreatedBy)

		assert.Error(t, err)
		assert.Nil(t, fe)
		assert.EqualError(t, err, "amount must be greater than zero")
	})

	t.Run("error when dueDay is less than 1", func(t *testing.T) {
		fe, err := New(validName, validAmount, 0, validCategory, validCreatedBy)

		assert.Error(t, err)
		assert.Nil(t, fe)
		assert.EqualError(t, err, "dueDay must be between 1 and 31")
	})

	t.Run("error when dueDay is greater than 31", func(t *testing.T) {
		fe, err := New(validName, validAmount, 32, validCategory, validCreatedBy)

		assert.Error(t, err)
		assert.Nil(t, fe)
		assert.EqualError(t, err, "dueDay must be between 1 and 31")
	})

	t.Run("error when category is invalid", func(t *testing.T) {
		fe, err := New(validName, validAmount, validDueDay, "INVALID_CATEGORY", validCreatedBy)

		assert.Error(t, err)
		assert.Nil(t, fe)
		assert.EqualError(t, err, "invalid category")
	})

	t.Run("error when createdBy is empty", func(t *testing.T) {
		fe, err := New(validName, validAmount, validDueDay, validCategory, "")

		assert.Error(t, err)
		assert.Nil(t, fe)
		assert.EqualError(t, err, "createdBy is required")
	})
}

func TestFixedExpense_Update(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		fe := newValidFixedExpense(t)
		assert.Nil(t, fe.UpdatedAt)

		newAmount := decimal.NewFromFloat(199.90)
		err := fe.Update(newAmount, 15, CategoryHealth, StatusActive, validChangedBy)

		assert.NoError(t, err)
		assert.True(t, fe.Amount.Equal(newAmount))
		assert.Equal(t, 15, fe.DueDay)
		assert.Equal(t, CategoryHealth, fe.Category)
		assert.Equal(t, StatusActive, fe.Status)
		assert.NotNil(t, fe.UpdatedAt)
		assert.NotNil(t, fe.UpdatedBy)
		assert.Equal(t, validChangedBy, *fe.UpdatedBy)
		assert.WithinDuration(t, time.Now(), *fe.UpdatedAt, time.Second)
	})

	t.Run("error when changedBy is empty", func(t *testing.T) {
		fe := newValidFixedExpense(t)

		err := fe.Update(validAmount, validDueDay, validCategory, StatusActive, "")

		assert.Error(t, err)
		assert.EqualError(t, err, "changedBy is required")
	})

	t.Run("error when amount is zero", func(t *testing.T) {
		fe := newValidFixedExpense(t)

		err := fe.Update(decimal.Zero, validDueDay, validCategory, StatusActive, validChangedBy)

		assert.Error(t, err)
		assert.EqualError(t, err, "amount must be greater than zero")
	})

	t.Run("error when amount is negative", func(t *testing.T) {
		fe := newValidFixedExpense(t)

		err := fe.Update(decimal.NewFromFloat(-50.00), validDueDay, validCategory, StatusActive, validChangedBy)

		assert.Error(t, err)
		assert.EqualError(t, err, "amount must be greater than zero")
	})

	t.Run("error when dueDay is out of range", func(t *testing.T) {
		fe := newValidFixedExpense(t)

		err := fe.Update(validAmount, 0, validCategory, StatusActive, validChangedBy)

		assert.Error(t, err)
		assert.EqualError(t, err, "dueDay must be between 1 and 31")
	})

	t.Run("error when category is invalid", func(t *testing.T) {
		fe := newValidFixedExpense(t)

		err := fe.Update(validAmount, validDueDay, "INVALID_CATEGORY", StatusActive, validChangedBy)

		assert.Error(t, err)
		assert.EqualError(t, err, "invalid category")
	})

	t.Run("error when status is invalid", func(t *testing.T) {
		fe := newValidFixedExpense(t)

		err := fe.Update(validAmount, validDueDay, validCategory, "INVALID_STATUS", validChangedBy)

		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrInvalidFixedExpenseStatus)
	})

	t.Run("entry is not updated when validation fails", func(t *testing.T) {
		fe := newValidFixedExpense(t)

		err := fe.Update(decimal.Zero, validDueDay, validCategory, StatusActive, validChangedBy)

		assert.Error(t, err)
		assert.Nil(t, fe.UpdatedAt)
		assert.Nil(t, fe.UpdatedBy)
	})
}

func TestFixedExpense_Deactivate(t *testing.T) {
	t.Run("success with status paused", func(t *testing.T) {
		fe := newValidFixedExpense(t)

		err := fe.Deactivate(StatusPaused, validChangedBy)

		assert.NoError(t, err)
		assert.Equal(t, StatusPaused, fe.Status)
		assert.NotNil(t, fe.UpdatedAt)
		assert.NotNil(t, fe.UpdatedBy)
		assert.Equal(t, validChangedBy, *fe.UpdatedBy)
		assert.WithinDuration(t, time.Now(), *fe.UpdatedAt, time.Second)
	})

	t.Run("success with status cancelled", func(t *testing.T) {
		fe := newValidFixedExpense(t)

		err := fe.Deactivate(StatusCancelled, validChangedBy)

		assert.NoError(t, err)
		assert.Equal(t, StatusCancelled, fe.Status)
		assert.NotNil(t, fe.UpdatedAt)
		assert.NotNil(t, fe.UpdatedBy)
		assert.Equal(t, validChangedBy, *fe.UpdatedBy)
	})

	t.Run("error when status is active", func(t *testing.T) {
		fe := newValidFixedExpense(t)

		err := fe.Deactivate(StatusActive, validChangedBy)

		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrInvalidFixedExpenseStatus)
		assert.Equal(t, StatusActive, fe.Status)
	})

	t.Run("error when status is invalid", func(t *testing.T) {
		fe := newValidFixedExpense(t)

		err := fe.Deactivate("INVALID_STATUS", validChangedBy)

		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrInvalidFixedExpenseStatus)
	})

	t.Run("entry is not updated when status is invalid", func(t *testing.T) {
		fe := newValidFixedExpense(t)

		err := fe.Deactivate(StatusActive, validChangedBy)

		assert.Error(t, err)
		assert.Nil(t, fe.UpdatedAt)
		assert.Nil(t, fe.UpdatedBy)
	})
}

func TestCategory_IsValid(t *testing.T) {
	validCategories := []Category{
		CategoryHousing,
		CategorySubscriptions,
		CategoryInsurance,
		CategoryEducation,
		CategoryTransportation,
		CategoryHealth,
		CategoryPersonal,
		CategoryTaxes,
		CategoryOther,
	}

	for _, category := range validCategories {
		t.Run(string(category), func(t *testing.T) {
			assert.True(t, category.IsValid())
		})
	}

	t.Run("invalid category", func(t *testing.T) {
		assert.False(t, Category("INVALID_CATEGORY").IsValid())
	})
}

func TestStatus_IsValid(t *testing.T) {
	validStatuses := []Status{
		StatusActive,
		StatusPaused,
		StatusCancelled,
	}

	for _, status := range validStatuses {
		t.Run(string(status), func(t *testing.T) {
			assert.True(t, status.IsValid())
		})
	}

	t.Run("invalid status", func(t *testing.T) {
		assert.False(t, Status("INVALID_STATUS").IsValid())
	})
}
