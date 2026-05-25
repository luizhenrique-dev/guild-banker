package fixedexpense

import (
	"context"
	"errors"
	"fmt"

	"github.com/shopspring/decimal"

	"github.com/luizhenrique-dev/guild-banker/api/internal/audit"
)

type Storage interface {
	Create(ctx context.Context, fixedExpense *FixedExpense, userID int64) error
	GetByIDAndUser(ctx context.Context, id, userID int64) (*FixedExpense, error)
	ListActiveByUser(ctx context.Context, userID int64) ([]*FixedExpense, error)
	Update(ctx context.Context, fixedExpense *FixedExpense, userID int64) error
}

type Service struct {
	storage Storage
}

type UpdateInput struct {
	Amount   *decimal.Decimal
	DueDay   *int
	Category *Category
	Status   *Status
}

func NewService(storage Storage) *Service {
	return &Service{storage: storage}
}

func (s *Service) Create(
	ctx context.Context,
	name string,
	amount decimal.Decimal,
	dueDay int,
	category Category,
	actor audit.Actor,
) (*FixedExpense, error) {
	if err := actor.Validate(); err != nil {
		return nil, fmt.Errorf("create fixed expense: %w", err)
	}

	fixedExpense, err := New(name, amount, dueDay, category, actor.Email)
	if err != nil {
		return nil, fmt.Errorf("create fixed expense: %w", err)
	}

	if err := s.storage.Create(ctx, fixedExpense, actor.UserID); err != nil {
		return nil, fmt.Errorf("create fixed expense: %w", err)
	}

	return fixedExpense, nil
}

func (s *Service) ListActiveByUser(ctx context.Context, userID int64) ([]*FixedExpense, error) {
	if userID == 0 {
		return nil, errors.New("list fixed expenses: userID is required")
	}

	fixedExpenses, err := s.storage.ListActiveByUser(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("list fixed expenses: %w", err)
	}

	return fixedExpenses, nil
}

func (s *Service) Update(ctx context.Context, id int64, input UpdateInput, actor audit.Actor) (*FixedExpense, error) {
	if err := actor.Validate(); err != nil {
		return nil, fmt.Errorf("update fixed expense: %w", err)
	}
	if id == 0 {
		return nil, errors.New("update fixed expense: id is required")
	}

	fixedExpense, err := s.storage.GetByIDAndUser(ctx, id, actor.UserID)
	if err != nil {
		return nil, fmt.Errorf("update fixed expense: get by id: %w", err)
	}

	amount := fixedExpense.Amount
	if input.Amount != nil {
		amount = *input.Amount
	}

	dueDay := fixedExpense.DueDay
	if input.DueDay != nil {
		dueDay = *input.DueDay
	}

	category := fixedExpense.Category
	if input.Category != nil {
		category = *input.Category
	}

	status := fixedExpense.Status
	if input.Status != nil {
		status = *input.Status
	}

	if err := fixedExpense.Update(amount, dueDay, category, status, actor.Email); err != nil {
		return nil, fmt.Errorf("update fixed expense: %w", err)
	}

	if err := s.storage.Update(ctx, fixedExpense, actor.UserID); err != nil {
		return nil, fmt.Errorf("update fixed expense: persist update: %w", err)
	}

	return fixedExpense, nil
}

func (s *Service) Deactivate(ctx context.Context, id int64, status Status, actor audit.Actor) error {
	if err := actor.Validate(); err != nil {
		return fmt.Errorf("deactivate fixed expense: %w", err)
	}
	if id == 0 {
		return errors.New("deactivate fixed expense: id is required")
	}

	fixedExpense, err := s.storage.GetByIDAndUser(ctx, id, actor.UserID)
	if err != nil {
		return fmt.Errorf("deactivate fixed expense: get by id: %w", err)
	}

	if err := fixedExpense.Deactivate(status, actor.Email); err != nil {
		return fmt.Errorf("deactivate fixed expense: %w", err)
	}

	if err := s.storage.Update(ctx, fixedExpense, actor.UserID); err != nil {
		return fmt.Errorf("deactivate fixed expense: persist update: %w", err)
	}

	return nil
}
