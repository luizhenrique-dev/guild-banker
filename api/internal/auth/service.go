package auth

import (
	"context"
	"errors"
	"fmt"

	"github.com/luizhenrique-dev/guild-banker/api/internal/audit"
	"github.com/luizhenrique-dev/guild-banker/api/internal/user"
)

type tokenVerifier interface {
	Verify(ctx context.Context, raw string) (Claims, error)
}

type userLookup interface {
	GetByExternalID(ctx context.Context, externalID string) (*user.User, error)
}

type Service struct {
	verifier tokenVerifier
	users    userLookup
}

func NewService(verifier tokenVerifier, users userLookup) *Service {
	return &Service{verifier: verifier, users: users}
}

func (s *Service) Authenticate(ctx context.Context, raw string) (Principal, error) {
	claims, err := s.verifier.Verify(ctx, raw)
	if err != nil {
		return Principal{}, fmt.Errorf("verify access token: %w", err)
	}
	u, err := s.users.GetByExternalID(ctx, claims.UserID)
	if err != nil {
		if errors.Is(err, user.ErrNotFound) {
			return Principal{}, &RejectionError{Reason: ReasonUserNotFound}
		}
		return Principal{}, fmt.Errorf("resolve authenticated user: %w", err)
	}
	if !u.Enabled {
		return Principal{}, &RejectionError{Reason: ReasonUserDisabled}
	}
	return Principal{Claims: claims, Actor: audit.Actor{UserID: u.ID, Email: u.Email}}, nil
}
