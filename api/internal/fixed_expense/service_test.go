package fixedexpense

import (
	"context"
	"errors"
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/luizhenrique-dev/guild-banker/api/internal/audit"
)

type storageMock struct {
	mock.Mock
}

func (m *storageMock) Create(ctx context.Context, fixedExpense *FixedExpense, userID int64) error {
	args := m.Called(ctx, fixedExpense, userID)
	return args.Error(0)
}

func (m *storageMock) GetByIDAndUser(ctx context.Context, id, userID int64) (*FixedExpense, error) {
	args := m.Called(ctx, id, userID)
	return args.Get(0).(*FixedExpense), args.Error(1)
}

func (m *storageMock) ListActiveByUser(ctx context.Context, userID int64) ([]*FixedExpense, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]*FixedExpense), args.Error(1)
}

func (m *storageMock) Update(ctx context.Context, fixedExpense *FixedExpense, userID int64) error {
	args := m.Called(ctx, fixedExpense, userID)
	return args.Error(0)
}

func newService(t *testing.T) (*Service, *storageMock) {
	t.Helper()
	s := &storageMock{}
	return NewService(s), s
}

func validActor() audit.Actor {
	return audit.Actor{UserID: 1, Email: "admin@example.com"}
}

func stubFixedExpense() *FixedExpense {
	fe, _ := New("Netflix", decimal.NewFromFloat(99.90), 10, CategorySubscriptions, "admin@example.com")
	fe.ID = 1
	return fe
}

func TestService_Create(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		svc, storage := newService(t)
		actor := validActor()
		storage.On("Create", mock.Anything, mock.AnythingOfType("*fixedexpense.FixedExpense"), actor.UserID).Return(nil)

		fe, err := svc.Create(context.Background(), "Netflix", decimal.NewFromFloat(99.90), 10, CategorySubscriptions, actor)

		assert.NoError(t, err)
		assert.NotNil(t, fe)
		assert.Equal(t, "Netflix", fe.Name)
		assert.True(t, fe.Amount.Equal(decimal.NewFromFloat(99.90)))
		assert.Equal(t, 10, fe.DueDay)
		assert.Equal(t, CategorySubscriptions, fe.Category)
		assert.Equal(t, StatusActive, fe.Status)
		storage.AssertExpectations(t)
	})

	t.Run("error when actor is invalid", func(t *testing.T) {
		svc, storage := newService(t)

		fe, err := svc.Create(context.Background(), "Netflix", decimal.NewFromFloat(99.90), 10, CategorySubscriptions, audit.Actor{})

		assert.Error(t, err)
		assert.Nil(t, fe)
		storage.AssertNotCalled(t, "Create")
	})

	t.Run("error on invalid fixed expense params", func(t *testing.T) {
		svc, storage := newService(t)
		actor := validActor()

		fe, err := svc.Create(context.Background(), "", decimal.NewFromFloat(99.90), 10, CategorySubscriptions, actor)

		assert.Error(t, err)
		assert.Nil(t, fe)
		storage.AssertNotCalled(t, "Create")
	})

	t.Run("error on storage", func(t *testing.T) {
		svc, storage := newService(t)
		actor := validActor()
		storage.On("Create", mock.Anything, mock.AnythingOfType("*fixedexpense.FixedExpense"), actor.UserID).Return(errors.New("db error"))

		fe, err := svc.Create(context.Background(), "Netflix", decimal.NewFromFloat(99.90), 10, CategorySubscriptions, actor)

		assert.Error(t, err)
		assert.Nil(t, fe)
		storage.AssertExpectations(t)
	})
}

func TestService_ListActiveByUser(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		svc, storage := newService(t)
		expected := []*FixedExpense{stubFixedExpense()}
		storage.On("ListActiveByUser", mock.Anything, int64(1)).Return(expected, nil)

		result, err := svc.ListActiveByUser(context.Background(), 1)

		assert.NoError(t, err)
		assert.Equal(t, expected, result)
		storage.AssertExpectations(t)
	})

	t.Run("error when userID is zero", func(t *testing.T) {
		svc, storage := newService(t)

		result, err := svc.ListActiveByUser(context.Background(), 0)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.EqualError(t, err, "list fixed expenses: userID is required")
		storage.AssertNotCalled(t, "ListActiveByUser")
	})

	t.Run("error on storage", func(t *testing.T) {
		svc, storage := newService(t)
		storage.On("ListActiveByUser", mock.Anything, int64(1)).Return(([]*FixedExpense)(nil), errors.New("db error"))

		result, err := svc.ListActiveByUser(context.Background(), 1)

		assert.Error(t, err)
		assert.Nil(t, result)
		storage.AssertExpectations(t)
	})
}

func TestService_Update(t *testing.T) {
	newAmount := decimal.NewFromFloat(199.90)
	newDueDay := 15
	newCategory := CategoryHealth
	newStatus := StatusActive

	t.Run("success with all fields updated", func(t *testing.T) {
		svc, storage := newService(t)
		actor := validActor()
		existing := stubFixedExpense()
		storage.On("GetByIDAndUser", mock.Anything, int64(1), actor.UserID).Return(existing, nil)
		storage.On("Update", mock.Anything, existing, actor.UserID).Return(nil)

		input := UpdateInput{
			Amount:   &newAmount,
			DueDay:   &newDueDay,
			Category: &newCategory,
			Status:   &newStatus,
		}

		fe, err := svc.Update(context.Background(), 1, input, actor)

		assert.NoError(t, err)
		assert.NotNil(t, fe)
		assert.True(t, fe.Amount.Equal(newAmount))
		assert.Equal(t, newDueDay, fe.DueDay)
		assert.Equal(t, newCategory, fe.Category)
		assert.Equal(t, newStatus, fe.Status)
		storage.AssertExpectations(t)
	})

	t.Run("success keeping existing fields when input is nil", func(t *testing.T) {
		svc, storage := newService(t)
		actor := validActor()
		existing := stubFixedExpense()
		storage.On("GetByIDAndUser", mock.Anything, int64(1), actor.UserID).Return(existing, nil)
		storage.On("Update", mock.Anything, existing, actor.UserID).Return(nil)

		fe, err := svc.Update(context.Background(), 1, UpdateInput{}, actor)

		assert.NoError(t, err)
		assert.NotNil(t, fe)
		assert.True(t, fe.Amount.Equal(existing.Amount))
		assert.Equal(t, existing.DueDay, fe.DueDay)
		assert.Equal(t, existing.Category, fe.Category)
		assert.Equal(t, existing.Status, fe.Status)
		storage.AssertExpectations(t)
	})

	t.Run("error when actor is invalid", func(t *testing.T) {
		svc, storage := newService(t)

		fe, err := svc.Update(context.Background(), 1, UpdateInput{}, audit.Actor{})

		assert.Error(t, err)
		assert.Nil(t, fe)
		storage.AssertNotCalled(t, "GetByIDAndUser")
		storage.AssertNotCalled(t, "Update")
	})

	t.Run("error when id is zero", func(t *testing.T) {
		svc, storage := newService(t)
		actor := validActor()

		fe, err := svc.Update(context.Background(), 0, UpdateInput{}, actor)

		assert.Error(t, err)
		assert.Nil(t, fe)
		assert.EqualError(t, err, "update fixed expense: id is required")
		storage.AssertNotCalled(t, "GetByIDAndUser")
		storage.AssertNotCalled(t, "Update")
	})

	t.Run("error on get by id", func(t *testing.T) {
		svc, storage := newService(t)
		actor := validActor()
		storage.On("GetByIDAndUser", mock.Anything, int64(1), actor.UserID).Return((*FixedExpense)(nil), errors.New("db error"))

		fe, err := svc.Update(context.Background(), 1, UpdateInput{}, actor)

		assert.Error(t, err)
		assert.Nil(t, fe)
		storage.AssertNotCalled(t, "Update")
	})

	t.Run("error when update produces invalid fixed expense", func(t *testing.T) {
		svc, storage := newService(t)
		actor := validActor()
		existing := stubFixedExpense()
		invalidAmount := decimal.NewFromFloat(-1.00)
		storage.On("GetByIDAndUser", mock.Anything, int64(1), actor.UserID).Return(existing, nil)

		fe, err := svc.Update(context.Background(), 1, UpdateInput{Amount: &invalidAmount}, actor)

		assert.Error(t, err)
		assert.Nil(t, fe)
		storage.AssertNotCalled(t, "Update")
	})

	t.Run("error on storage update", func(t *testing.T) {
		svc, storage := newService(t)
		actor := validActor()
		existing := stubFixedExpense()
		storage.On("GetByIDAndUser", mock.Anything, int64(1), actor.UserID).Return(existing, nil)
		storage.On("Update", mock.Anything, existing, actor.UserID).Return(errors.New("db error"))

		fe, err := svc.Update(context.Background(), 1, UpdateInput{}, actor)

		assert.Error(t, err)
		assert.Nil(t, fe)
		storage.AssertExpectations(t)
	})
}

func TestService_Deactivate(t *testing.T) {
	t.Run("success with status paused", func(t *testing.T) {
		svc, storage := newService(t)
		actor := validActor()
		existing := stubFixedExpense()
		storage.On("GetByIDAndUser", mock.Anything, int64(1), actor.UserID).Return(existing, nil)
		storage.On("Update", mock.Anything, existing, actor.UserID).Return(nil)

		err := svc.Deactivate(context.Background(), 1, StatusPaused, actor)

		assert.NoError(t, err)
		assert.Equal(t, StatusPaused, existing.Status)
		storage.AssertExpectations(t)
	})

	t.Run("success with status cancelled", func(t *testing.T) {
		svc, storage := newService(t)
		actor := validActor()
		existing := stubFixedExpense()
		storage.On("GetByIDAndUser", mock.Anything, int64(1), actor.UserID).Return(existing, nil)
		storage.On("Update", mock.Anything, existing, actor.UserID).Return(nil)

		err := svc.Deactivate(context.Background(), 1, StatusCancelled, actor)

		assert.NoError(t, err)
		assert.Equal(t, StatusCancelled, existing.Status)
		storage.AssertExpectations(t)
	})

	t.Run("error when actor is invalid", func(t *testing.T) {
		svc, storage := newService(t)

		err := svc.Deactivate(context.Background(), 1, StatusPaused, audit.Actor{})

		assert.Error(t, err)
		storage.AssertNotCalled(t, "GetByIDAndUser")
		storage.AssertNotCalled(t, "Update")
	})

	t.Run("error when id is zero", func(t *testing.T) {
		svc, storage := newService(t)
		actor := validActor()

		err := svc.Deactivate(context.Background(), 0, StatusPaused, actor)

		assert.Error(t, err)
		assert.EqualError(t, err, "deactivate fixed expense: id is required")
		storage.AssertNotCalled(t, "GetByIDAndUser")
		storage.AssertNotCalled(t, "Update")
	})

	t.Run("error on get by id", func(t *testing.T) {
		svc, storage := newService(t)
		actor := validActor()
		storage.On("GetByIDAndUser", mock.Anything, int64(1), actor.UserID).Return((*FixedExpense)(nil), errors.New("db error"))

		err := svc.Deactivate(context.Background(), 1, StatusPaused, actor)

		assert.Error(t, err)
		storage.AssertNotCalled(t, "Update")
	})

	t.Run("error when status is invalid for deactivation", func(t *testing.T) {
		svc, storage := newService(t)
		actor := validActor()
		existing := stubFixedExpense()
		storage.On("GetByIDAndUser", mock.Anything, int64(1), actor.UserID).Return(existing, nil)

		err := svc.Deactivate(context.Background(), 1, StatusActive, actor)

		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrInvalidFixedExpenseStatus)
		storage.AssertNotCalled(t, "Update")
	})

	t.Run("error on storage update", func(t *testing.T) {
		svc, storage := newService(t)
		actor := validActor()
		existing := stubFixedExpense()
		storage.On("GetByIDAndUser", mock.Anything, int64(1), actor.UserID).Return(existing, nil)
		storage.On("Update", mock.Anything, existing, actor.UserID).Return(errors.New("db error"))

		err := svc.Deactivate(context.Background(), 1, StatusPaused, actor)

		assert.Error(t, err)
		storage.AssertExpectations(t)
	})
}
