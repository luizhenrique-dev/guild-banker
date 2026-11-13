package auth

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/luizhenrique-dev/guild-banker/api/internal/audit"
	"github.com/luizhenrique-dev/guild-banker/api/internal/emailaddress"
	"github.com/luizhenrique-dev/guild-banker/api/internal/user"
)

type tokenVerifier interface {
	Verify(ctx context.Context, raw string) (Claims, error)
}

type accountResolver interface {
	EnsureAccount(ctx context.Context, externalID, name, email string) (*user.User, error)
	RecordAccess(ctx context.Context, id int64) error
}

type Service struct {
	verifier tokenVerifier
	users    accountResolver
}

func NewService(verifier tokenVerifier, users accountResolver) *Service {
	return &Service{verifier: verifier, users: users}
}

func (s *Service) Authenticate(ctx context.Context, raw string) (Principal, error) {
	claims, err := s.verifier.Verify(ctx, raw)
	if err != nil {
		return Principal{}, fmt.Errorf("verify access token: %w", err)
	}
	if strings.TrimSpace(claims.Name) == "" || strings.TrimSpace(claims.Email) == "" {
		return Principal{}, &RejectionError{Reason: ReasonIncompleteProfile}
	}
	if !claims.EmailVerified {
		return Principal{}, &RejectionError{Reason: ReasonEmailUnverified}
	}
	u, err := s.users.EnsureAccount(ctx, claims.UserID, claims.Name, claims.Email)
	if err != nil {
		if errors.Is(err, emailaddress.ErrInvalid) {
			return Principal{}, &RejectionError{Reason: ReasonIncompleteProfile}
		}
		if errors.Is(err, user.ErrDisabled) {
			return Principal{}, &RejectionError{Reason: ReasonUserDisabled}
		}
		if errors.Is(err, user.ErrEmailConflict) {
			return Principal{}, &RejectionError{Reason: ReasonEmailConflict}
		}
		return Principal{}, fmt.Errorf("resolve authenticated user: %w", err)
	}
	if !u.Enabled {
		return Principal{}, &RejectionError{Reason: ReasonUserDisabled}
	}
	if err := s.users.RecordAccess(ctx, u.ID); err != nil {
		return Principal{}, fmt.Errorf("record authenticated access: %w", err)
	}
	return Principal{Claims: claims, Actor: audit.Actor{UserID: u.ID, Email: u.Email}}, nil
}
