package fixedexpense

import (
	"context"
	"fmt"

	"github.com/shopspring/decimal"

	"github.com/luizhenrique-dev/guild-banker/api/internal/audit"
)

type Storage interface {
	Delete(ctx context.Context, id, guildID int64) error
	ListByGuild(ctx context.Context, guildID int64) ([]*FixedExpense, error)
	IsGuildMember(ctx context.Context, guildID, userID int64) (bool, error)
	Create(ctx context.Context, fixedExpense *FixedExpense) error
	GetByIDAndGuild(ctx context.Context, id, guildID int64) (*FixedExpense, error)
	ListActiveByGuild(ctx context.Context, guildID int64) ([]*FixedExpense, error)
	Update(ctx context.Context, fixedExpense *FixedExpense) error
}

type Service struct {
	storage Storage
}

type UpdateInput struct {
	ResponsibleUserID *int64
	SetResponsible    bool
	Amount            *decimal.Decimal
	DueDay            *int
	Category          *Category
	Status            *Status
}

func NewService(storage Storage) *Service {
	return &Service{storage: storage}
}

func (s *Service) Create(
	ctx context.Context,
	guildID int64,
	name string,
	amount decimal.Decimal,
	dueDay int,
	category Category,
	actor audit.Actor,
) (*FixedExpense, error) {
	return s.CreateWithResponsible(ctx, guildID, name, amount, dueDay, category, nil, actor)
}

func (s *Service) CreateWithResponsible(
	ctx context.Context, guildID int64, name string, amount decimal.Decimal,
	dueDay int, category Category, responsibleUserID *int64, actor audit.Actor,
) (*FixedExpense, error) {
	if err := s.authorizeGuildMember(ctx, guildID, actor); err != nil {
		return nil, fmt.Errorf("create fixed expense: %w", err)
	}

	if err := s.validateResponsible(ctx, guildID, responsibleUserID); err != nil {
		return nil, fmt.Errorf("create fixed expense: %w", err)
	}
	fixedExpense, err := New(guildID, name, amount, dueDay, category, actor.Email)
	if err != nil {
		return nil, fmt.Errorf("create fixed expense: %w", invalidInput(err))
	}

	fixedExpense.ResponsibleUserID = responsibleUserID
	if err := s.storage.Create(ctx, fixedExpense); err != nil {
		return nil, fmt.Errorf("create fixed expense: %w", err)
	}

	return fixedExpense, nil
}

func (s *Service) ListActiveByGuild(ctx context.Context, guildID int64, actor audit.Actor) ([]*FixedExpense, error) {
	if err := s.authorizeGuildMember(ctx, guildID, actor); err != nil {
		return nil, fmt.Errorf("list fixed expenses: %w", err)
	}

	fixedExpenses, err := s.storage.ListActiveByGuild(ctx, guildID)
	if err != nil {
		return nil, fmt.Errorf("list fixed expenses: %w", err)
	}

	return fixedExpenses, nil
}

func (s *Service) ListByGuild(ctx context.Context, guildID int64, actor audit.Actor) ([]*FixedExpense, error) {
	if err := s.authorizeGuildMember(ctx, guildID, actor); err != nil {
		return nil, fmt.Errorf("list fixed expenses: %w", err)
	}
	expenses, err := s.storage.ListByGuild(ctx, guildID)
	if err != nil {
		return nil, fmt.Errorf("list fixed expenses: %w", err)
	}
	return expenses, nil
}

func (s *Service) Update(ctx context.Context, guildID, id int64, input UpdateInput, actor audit.Actor) (*FixedExpense, error) {
	if err := s.authorizeGuildMember(ctx, guildID, actor); err != nil {
		return nil, fmt.Errorf("update fixed expense: %w", err)
	}
	if id <= 0 {
		return nil, fmt.Errorf("%w: id must be positive", ErrInvalidInput)
	}

	fixedExpense, err := s.storage.GetByIDAndGuild(ctx, id, guildID)
	if err != nil {
		return nil, fmt.Errorf("update fixed expense: get by id: %w", err)
	}

	if input.SetResponsible {
		if err := s.validateResponsible(ctx, guildID, input.ResponsibleUserID); err != nil {
			return nil, fmt.Errorf("update fixed expense: %w", err)
		}
		fixedExpense.ResponsibleUserID = input.ResponsibleUserID
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
		return nil, fmt.Errorf("update fixed expense: %w", invalidInput(err))
	}

	if err := s.storage.Update(ctx, fixedExpense); err != nil {
		return nil, fmt.Errorf("update fixed expense: persist update: %w", err)
	}

	return fixedExpense, nil
}

func (s *Service) Deactivate(ctx context.Context, guildID, id int64, status Status, actor audit.Actor) error {
	if err := s.authorizeGuildMember(ctx, guildID, actor); err != nil {
		return fmt.Errorf("deactivate fixed expense: %w", err)
	}
	if id <= 0 {
		return fmt.Errorf("%w: id must be positive", ErrInvalidInput)
	}

	fixedExpense, err := s.storage.GetByIDAndGuild(ctx, id, guildID)
	if err != nil {
		return fmt.Errorf("deactivate fixed expense: get by id: %w", err)
	}

	if err := fixedExpense.Deactivate(status, actor.Email); err != nil {
		return fmt.Errorf("deactivate fixed expense: %w", invalidInput(err))
	}

	if err := s.storage.Update(ctx, fixedExpense); err != nil {
		return fmt.Errorf("deactivate fixed expense: persist update: %w", err)
	}

	return nil
}

func (s *Service) Delete(ctx context.Context, guildID, id int64, actor audit.Actor) error {
	if err := s.authorizeGuildMember(ctx, guildID, actor); err != nil {
		return fmt.Errorf("delete fixed expense: %w", err)
	}
	if id <= 0 {
		return fmt.Errorf("%w: id must be positive", ErrInvalidInput)
	}
	if err := s.storage.Delete(ctx, id, guildID); err != nil {
		return fmt.Errorf("delete fixed expense: %w", err)
	}
	return nil
}

func invalidInput(err error) error { return fmt.Errorf("%w: %w", ErrInvalidInput, err) }

func (s *Service) authorizeGuildMember(ctx context.Context, guildID int64, actor audit.Actor) error {
	if err := actor.Validate(); err != nil {
		return invalidInput(err)
	}
	if guildID <= 0 {
		return fmt.Errorf("%w: guildID must be positive", ErrInvalidInput)
	}
	isMember, err := s.storage.IsGuildMember(ctx, guildID, actor.UserID)
	if err != nil {
		return fmt.Errorf("check guild membership: %w", err)
	}
	if !isMember {
		return ErrRequesterIsNotMember
	}
	return nil
}

func (s *Service) validateResponsible(ctx context.Context, guildID int64, userID *int64) error {
	if userID == nil {
		return nil
	}
	if *userID <= 0 {
		return fmt.Errorf("%w: responsible user id must be positive", ErrInvalidInput)
	}
	isMember, err := s.storage.IsGuildMember(ctx, guildID, *userID)
	if err != nil {
		return fmt.Errorf("check responsible membership: %w", err)
	}
	if !isMember {
		return fmt.Errorf("%w: responsible user is not a guild member", ErrInvalidInput)
	}
	return nil
}
