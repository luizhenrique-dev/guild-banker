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
	member            *bool
	membershipErr     error
	membershipCtx     context.Context
	membershipGuildID int64
	membershipUserID  int64
}

func (m *storageMock) ListByGuild(ctx context.Context, guildID int64) ([]*FixedExpense, error) {
	args := m.Called(ctx, guildID)
	return args.Get(0).([]*FixedExpense), args.Error(1)
}

func (m *storageMock) Delete(ctx context.Context, id, guildID int64) error {
	return m.Called(ctx, id, guildID).Error(0)
}

func (m *storageMock) Create(ctx context.Context, fixedExpense *FixedExpense) error {
	args := m.Called(ctx, fixedExpense)
	return args.Error(0)
}

func (m *storageMock) GetByIDAndGuild(ctx context.Context, id, userID int64) (*FixedExpense, error) {
	args := m.Called(ctx, id, userID)
	return args.Get(0).(*FixedExpense), args.Error(1)
}

func (m *storageMock) ListActiveByGuild(ctx context.Context, userID int64) ([]*FixedExpense, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]*FixedExpense), args.Error(1)
}

func (m *storageMock) Update(ctx context.Context, fixedExpense *FixedExpense) error {
	args := m.Called(ctx, fixedExpense)
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
	fe, _ := New(1, "Netflix", decimal.NewFromFloat(99.90), 10, CategorySubscriptions, "admin@example.com")
	fe.ID = 1
	return fe
}

func TestService_Create(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		svc, storage := newService(t)
		actor := validActor()
		storage.On("Create", mock.Anything, mock.AnythingOfType("*fixedexpense.FixedExpense")).Return(nil)

		fe, err := svc.Create(context.Background(), 1, "Netflix", decimal.NewFromFloat(99.90), 10, CategorySubscriptions, actor)

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

		fe, err := svc.Create(context.Background(), 1, "Netflix", decimal.NewFromFloat(99.90), 10, CategorySubscriptions, audit.Actor{})

		assert.Error(t, err)
		assert.Nil(t, fe)
		storage.AssertNotCalled(t, "Create")
	})

	t.Run("error on invalid fixed expense params", func(t *testing.T) {
		svc, storage := newService(t)
		actor := validActor()

		fe, err := svc.Create(context.Background(), 1, "", decimal.NewFromFloat(99.90), 10, CategorySubscriptions, actor)

		assert.Error(t, err)
		assert.Nil(t, fe)
		storage.AssertNotCalled(t, "Create")
	})

	t.Run("error on storage", func(t *testing.T) {
		svc, storage := newService(t)
		actor := validActor()
		storage.On("Create", mock.Anything, mock.AnythingOfType("*fixedexpense.FixedExpense")).Return(errors.New("db error"))

		fe, err := svc.Create(context.Background(), 1, "Netflix", decimal.NewFromFloat(99.90), 10, CategorySubscriptions, actor)

		assert.Error(t, err)
		assert.Nil(t, fe)
		storage.AssertExpectations(t)
	})
}

func TestService_ListActiveByGuild(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		svc, storage := newService(t)
		expected := []*FixedExpense{stubFixedExpense()}
		storage.On("ListActiveByGuild", mock.Anything, int64(1)).Return(expected, nil)

		result, err := svc.ListActiveByGuild(context.Background(), 1, validActor())

		assert.NoError(t, err)
		assert.Equal(t, expected, result)
		storage.AssertExpectations(t)
	})

	t.Run("error when userID is zero", func(t *testing.T) {
		svc, storage := newService(t)

		result, err := svc.ListActiveByGuild(context.Background(), 0, validActor())

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.ErrorIs(t, err, ErrInvalidInput)
		storage.AssertNotCalled(t, "ListActiveByGuild")
	})

	t.Run("error on storage", func(t *testing.T) {
		svc, storage := newService(t)
		storage.On("ListActiveByGuild", mock.Anything, int64(1)).Return(([]*FixedExpense)(nil), errors.New("db error"))

		result, err := svc.ListActiveByGuild(context.Background(), 1, validActor())

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
		storage.On("GetByIDAndGuild", mock.Anything, int64(1), int64(1)).Return(existing, nil)
		storage.On("Update", mock.Anything, existing).Return(nil)

		input := UpdateInput{
			Amount:   &newAmount,
			DueDay:   &newDueDay,
			Category: &newCategory,
			Status:   &newStatus,
		}

		fe, err := svc.Update(context.Background(), 1, 1, input, actor)

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
		storage.On("GetByIDAndGuild", mock.Anything, int64(1), int64(1)).Return(existing, nil)
		storage.On("Update", mock.Anything, existing).Return(nil)

		fe, err := svc.Update(context.Background(), 1, 1, UpdateInput{}, actor)

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

		fe, err := svc.Update(context.Background(), 1, 1, UpdateInput{}, audit.Actor{})

		assert.Error(t, err)
		assert.Nil(t, fe)
		storage.AssertNotCalled(t, "GetByIDAndGuild")
		storage.AssertNotCalled(t, "Update")
	})

	t.Run("error when id is zero", func(t *testing.T) {
		svc, storage := newService(t)
		actor := validActor()

		fe, err := svc.Update(context.Background(), 1, 0, UpdateInput{}, actor)

		assert.Error(t, err)
		assert.Nil(t, fe)
		assert.ErrorIs(t, err, ErrInvalidInput)
		storage.AssertNotCalled(t, "GetByIDAndGuild")
		storage.AssertNotCalled(t, "Update")
	})

	t.Run("error on get by id", func(t *testing.T) {
		svc, storage := newService(t)
		actor := validActor()
		storage.On("GetByIDAndGuild", mock.Anything, int64(1), int64(1)).Return((*FixedExpense)(nil), errors.New("db error"))

		fe, err := svc.Update(context.Background(), 1, 1, UpdateInput{}, actor)

		assert.Error(t, err)
		assert.Nil(t, fe)
		storage.AssertNotCalled(t, "Update")
	})

	t.Run("error when update produces invalid fixed expense", func(t *testing.T) {
		svc, storage := newService(t)
		actor := validActor()
		existing := stubFixedExpense()
		invalidAmount := decimal.NewFromFloat(-1.00)
		storage.On("GetByIDAndGuild", mock.Anything, int64(1), int64(1)).Return(existing, nil)

		fe, err := svc.Update(context.Background(), 1, 1, UpdateInput{Amount: &invalidAmount}, actor)

		assert.Error(t, err)
		assert.Nil(t, fe)
		storage.AssertNotCalled(t, "Update")
	})

	t.Run("error on storage update", func(t *testing.T) {
		svc, storage := newService(t)
		actor := validActor()
		existing := stubFixedExpense()
		storage.On("GetByIDAndGuild", mock.Anything, int64(1), int64(1)).Return(existing, nil)
		storage.On("Update", mock.Anything, existing).Return(errors.New("db error"))

		fe, err := svc.Update(context.Background(), 1, 1, UpdateInput{}, actor)

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
		storage.On("GetByIDAndGuild", mock.Anything, int64(1), int64(1)).Return(existing, nil)
		storage.On("Update", mock.Anything, existing).Return(nil)

		err := svc.Deactivate(context.Background(), 1, 1, StatusPaused, actor)

		assert.NoError(t, err)
		assert.Equal(t, StatusPaused, existing.Status)
		storage.AssertExpectations(t)
	})

	t.Run("success with status cancelled", func(t *testing.T) {
		svc, storage := newService(t)
		actor := validActor()
		existing := stubFixedExpense()
		storage.On("GetByIDAndGuild", mock.Anything, int64(1), int64(1)).Return(existing, nil)
		storage.On("Update", mock.Anything, existing).Return(nil)

		err := svc.Deactivate(context.Background(), 1, 1, StatusCancelled, actor)

		assert.NoError(t, err)
		assert.Equal(t, StatusCancelled, existing.Status)
		storage.AssertExpectations(t)
	})

	t.Run("error when actor is invalid", func(t *testing.T) {
		svc, storage := newService(t)

		err := svc.Deactivate(context.Background(), 1, 1, StatusPaused, audit.Actor{})

		assert.Error(t, err)
		storage.AssertNotCalled(t, "GetByIDAndGuild")
		storage.AssertNotCalled(t, "Update")
	})

	t.Run("error when id is zero", func(t *testing.T) {
		svc, storage := newService(t)
		actor := validActor()

		err := svc.Deactivate(context.Background(), 1, 0, StatusPaused, actor)

		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrInvalidInput)
		storage.AssertNotCalled(t, "GetByIDAndGuild")
		storage.AssertNotCalled(t, "Update")
	})

	t.Run("error on get by id", func(t *testing.T) {
		svc, storage := newService(t)
		actor := validActor()
		storage.On("GetByIDAndGuild", mock.Anything, int64(1), int64(1)).Return((*FixedExpense)(nil), errors.New("db error"))

		err := svc.Deactivate(context.Background(), 1, 1, StatusPaused, actor)

		assert.Error(t, err)
		storage.AssertNotCalled(t, "Update")
	})

	t.Run("error when status is invalid for deactivation", func(t *testing.T) {
		svc, storage := newService(t)
		actor := validActor()
		existing := stubFixedExpense()
		storage.On("GetByIDAndGuild", mock.Anything, int64(1), int64(1)).Return(existing, nil)

		err := svc.Deactivate(context.Background(), 1, 1, StatusActive, actor)

		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrInvalidFixedExpenseStatus)
		storage.AssertNotCalled(t, "Update")
	})

	t.Run("error on storage update", func(t *testing.T) {
		svc, storage := newService(t)
		actor := validActor()
		existing := stubFixedExpense()
		storage.On("GetByIDAndGuild", mock.Anything, int64(1), int64(1)).Return(existing, nil)
		storage.On("Update", mock.Anything, existing).Return(errors.New("db error"))

		err := svc.Deactivate(context.Background(), 1, 1, StatusPaused, actor)

		assert.Error(t, err)
		storage.AssertExpectations(t)
	})
}

func (m *storageMock) IsGuildMember(ctx context.Context, guildID, userID int64) (bool, error) {
	m.membershipCtx = ctx
	m.membershipGuildID = guildID
	m.membershipUserID = userID
	if m.membershipErr != nil {
		return false, m.membershipErr
	}
	if m.member != nil {
		return *m.member, nil
	}
	return true, nil
}

func TestServiceGuildAuthorization(t *testing.T) {
	for _, operation := range []string{"create", "list", "update", "deactivate"} {
		for _, failure := range []string{"non member", "technical failure", "canceled context", "invalid guild", "invalid actor"} {
			t.Run(operation+" "+failure, func(t *testing.T) {
				svc, storage := newService(t)
				member := false
				storage.member = &member
				ctx := context.Background()
				guildID := int64(2)
				actor := validActor()
				want := ErrRequesterIsNotMember
				switch failure {
				case "technical failure":
					want = errors.New("database unavailable")
					storage.membershipErr = want
				case "canceled context":
					canceled, cancel := context.WithCancel(ctx)
					cancel()
					ctx = canceled
					want = context.Canceled
					storage.membershipErr = want
				case "invalid guild":
					guildID = -1
					want = ErrInvalidInput
				case "invalid actor":
					actor = audit.Actor{}
					want = ErrInvalidInput
				}
				var err error
				switch operation {
				case "create":
					_, err = svc.Create(ctx, guildID, "Expense", decimal.NewFromInt(10), 1, CategoryOther, actor)
				case "list":
					_, err = svc.ListActiveByGuild(ctx, guildID, actor)
				case "update":
					_, err = svc.Update(ctx, guildID, 1, UpdateInput{}, actor)
				case "deactivate":
					err = svc.Deactivate(ctx, guildID, 1, StatusPaused, actor)
				}
				assert.ErrorIs(t, err, want)
				assert.Empty(t, storage.Calls, "no expense operation may execute when authorization fails")
				if failure != "invalid guild" && failure != "invalid actor" {
					assert.Equal(t, ctx, storage.membershipCtx)
					assert.Equal(t, guildID, storage.membershipGuildID)
					assert.Equal(t, actor.UserID, storage.membershipUserID)
				}
			})
		}
	}
}

type responsibleStorage struct {
	*storageMock
	responsibleMember bool
	responsibleErr    error
}

func (s *responsibleStorage) IsGuildMember(_ context.Context, guildID, userID int64) (bool, error) {
	if userID == validActor().UserID {
		return true, nil
	}
	return s.responsibleMember && guildID == 1, s.responsibleErr
}

func TestResponsibleAssignmentValidation(t *testing.T) {
	technical := errors.New("membership database unavailable")
	for _, tt := range []struct {
		name       string
		id         int64
		member     bool
		storageErr error
		want       error
	}{
		{name: "member", id: 2, member: true},
		{name: "non member", id: 2, want: ErrInvalidInput},
		{name: "zero", want: ErrInvalidInput},
		{name: "negative", id: -1, want: ErrInvalidInput},
		{name: "technical", id: 2, storageErr: technical, want: technical},
	} {
		for _, operation := range []string{"create", "update"} {
			t.Run(operation+"/"+tt.name, func(t *testing.T) {
				storage := &responsibleStorage{storageMock: &storageMock{}, responsibleMember: tt.member, responsibleErr: tt.storageErr}
				service := NewService(storage)
				expense := stubFixedExpense()
				if operation == "update" {
					storage.On("GetByIDAndGuild", mock.Anything, int64(1), int64(1)).Return(expense, nil)
				}
				if tt.want == nil {
					storage.On(map[string]string{"create": "Create", "update": "Update"}[operation], mock.Anything, mock.MatchedBy(func(expense *FixedExpense) bool {
						return expense.ResponsibleUserID != nil && *expense.ResponsibleUserID == tt.id
					})).Return(nil)
				}
				var err error
				if operation == "create" {
					_, err = service.CreateWithResponsible(context.Background(), 1, "Rent", decimal.NewFromInt(10), 1, CategoryOther, &tt.id, validActor())
				} else {
					_, err = service.Update(context.Background(), 1, 1, UpdateInput{ResponsibleUserID: &tt.id, SetResponsible: true}, validActor())
				}
				if tt.want == nil {
					assert.NoError(t, err)
				} else {
					assert.ErrorIs(t, err, tt.want)
					storage.AssertNotCalled(t, "Create")
					storage.AssertNotCalled(t, "Update")
				}
				storage.AssertExpectations(t)
			})
		}
	}
}
