package guild

import (
	"context"
	"fmt"
)

func (s *Service) ListMembers(ctx context.Context, requesterID, guildID int64) ([]Member, error) {
	if err := s.ensureRequesterMember(ctx, requesterID, guildID, "list members"); err != nil {
		return nil, err
	}
	members, err := s.storage.ListMembers(ctx, guildID)
	if err != nil {
		return nil, fmt.Errorf("list members: %w", err)
	}
	return members, nil
}

func (s *Service) ListSentInvitations(ctx context.Context, requesterID, guildID int64) ([]SentInvitation, error) {
	if err := s.ensureRequesterMember(ctx, requesterID, guildID, "list sent invitations"); err != nil {
		return nil, err
	}
	invitations, err := s.storage.ListSentInvitations(ctx, guildID)
	if err != nil {
		return nil, fmt.Errorf("list sent invitations: %w", err)
	}
	return invitations, nil
}

func (s *Service) CancelInvitation(ctx context.Context, requesterID, guildID, invitationID int64) error {
	if err := s.ensureRequesterMember(ctx, requesterID, guildID, "cancel invitation"); err != nil {
		return err
	}
	if invitationID <= 0 {
		return ErrInvitationNotFound
	}
	if err := s.storage.CancelInvitation(ctx, guildID, invitationID, requesterID); err != nil {
		return fmt.Errorf("cancel invitation: %w", err)
	}
	return nil
}
