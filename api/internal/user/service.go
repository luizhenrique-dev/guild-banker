package user

import (
	"context"
	"errors"
	"fmt"
)

type Service struct {
	storage Storage
}

func NewService(storage Storage) *Service {
	return &Service{storage: storage}
}

func (s *Service) Create(ctx context.Context, name, email, externalID, createdBy string) (*User, error) {
	u, err := New(name, email, externalID, createdBy)
	if err != nil {
		return nil, fmt.Errorf("create user: %w", err)
	}

	if err := s.storage.Create(ctx, u); err != nil {
		return nil, fmt.Errorf("create user: %w", err)
	}

	return u, nil
}

func (s *Service) Update(ctx context.Context, id int64, name, email, by string) (*User, error) {
	if id == 0 {
		return nil, errors.New("update user: id is required")
	}
	u, err := s.storage.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get user by id for update: %w", err)
	}

	u.Sync(name, email, by)

	if err := s.storage.Update(ctx, u); err != nil {
		return nil, fmt.Errorf("update user: %w", err)
	}

	return u, nil
}

func (s *Service) GetByID(ctx context.Context, id int64) (*User, error) {
	if id == 0 {
		return nil, errors.New("get user by id: id is required")
	}
	u, err := s.storage.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get user by id: %w", err)
	}

	return u, nil
}

func (s *Service) GetByExternalID(ctx context.Context, externalID string) (*User, error) {
	if externalID == "" {
		return nil, errors.New("get user by external id: externalID is required")
	}
	u, err := s.storage.GetByExternalID(ctx, externalID)
	if err != nil {
		return nil, fmt.Errorf("get user by external id: %w", err)
	}

	return u, nil
}

func (s *Service) GetByEmail(ctx context.Context, email string) (*User, error) {
	if email == "" {
		return nil, errors.New("get user by email: email is required")
	}
	u, err := s.storage.GetByEmail(ctx, email)
	if err != nil {
		return nil, fmt.Errorf("get user by email: %w", err)
	}

	return u, nil
}

func (s *Service) Enable(ctx context.Context, id int64) error {
	if id == 0 {
		return errors.New("enable user: id is required")
	}
	if err := s.storage.Enable(ctx, id); err != nil {
		return fmt.Errorf("enable user: %w", err)
	}

	return nil
}

func (s *Service) Disable(ctx context.Context, id int64, disabledBy string) error {
	if id == 0 {
		return errors.New("disable user: id is required")
	}
	if disabledBy == "" {
		return errors.New("disable user: disabledBy is required")
	}
	if err := s.storage.Disable(ctx, id, disabledBy); err != nil {
		return fmt.Errorf("disable user: %w", err)
	}

	return nil
}
