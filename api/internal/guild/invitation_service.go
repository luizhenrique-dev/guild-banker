package guild

import (
	"context"
	"fmt"
	"github.com/luizhenrique-dev/guild-banker/api/internal/audit"
)

func (s *Service) ListInvitations(ctx context.Context, actor audit.Actor) ([]Invitation, error) {
	if err := actor.Validate(); err != nil {
		return nil, fmt.Errorf("list invitations: %w", err)
	}
	email, err := invitationEmail(actor.Email)
	if err != nil {
		return nil, err
	}
	invitations, err := s.storage.ListInvitations(ctx, email)
	if err != nil {
		return nil, fmt.Errorf("list invitations: %w", err)
	}
	return invitations, nil
}

func (s *Service) RespondInvitation(ctx context.Context, id int64, actor audit.Actor, accept bool) (*Guild, error) {
	if err := actor.Validate(); err != nil {
		return nil, fmt.Errorf("respond invitation: %w", err)
	}
	if id <= 0 {
		return nil, ErrInvitationNotFound
	}
	email, err := invitationEmail(actor.Email)
	if err != nil {
		return nil, err
	}
	actor.Email = email
	guild, err := s.storage.RespondInvitation(ctx, id, actor, accept)
	if err != nil {
		return nil, fmt.Errorf("respond invitation: %w", err)
	}
	return guild, nil
}
