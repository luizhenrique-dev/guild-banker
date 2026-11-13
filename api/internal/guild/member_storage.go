package guild

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

func (s *postgresStorage) ListMembers(ctx context.Context, guildID int64) ([]Member, error) {
	ctx, cancel := context.WithTimeout(ctx, invitationStorageTimeout)
	defer cancel()
	var entities []struct {
		ID           int64      `db:"id"`
		Name         string     `db:"name"`
		Email        string     `db:"email"`
		LastAccessAt *time.Time `db:"last_access_at"`
		MemberSince  time.Time  `db:"member_since"`
	}
	const query = `SELECT u.id, u.name, u.email, u.last_access_at, gm.invited_at AS member_since
	 FROM guild_member gm JOIN user_account u ON u.id = gm.user_id
	 WHERE gm.guild_id = $1 ORDER BY gm.invited_at, u.id`
	if err := s.db.SelectContext(ctx, &entities, query, guildID); err != nil {
		return nil, fmt.Errorf("query guild members: %w", err)
	}
	members := make([]Member, 0, len(entities))
	for _, e := range entities {
		members = append(members, Member{ID: e.ID, Name: e.Name, Email: e.Email, LastAccessAt: e.LastAccessAt, MemberSince: e.MemberSince})
	}
	return members, nil
}

func (s *postgresStorage) ListSentInvitations(ctx context.Context, guildID int64) ([]SentInvitation, error) {
	ctx, cancel := context.WithTimeout(ctx, invitationStorageTimeout)
	defer cancel()
	var entities []struct {
		ID             int64     `db:"id"`
		Email          string    `db:"recipient_email"`
		InvitedByName  string    `db:"invited_by_name"`
		InvitedByEmail string    `db:"invited_by_email"`
		CreatedAt      time.Time `db:"created_at"`
	}
	const query = `SELECT i.id, i.recipient_email, i.created_at,
	 u.name AS invited_by_name, u.email AS invited_by_email
	 FROM guild_invitation i JOIN user_account u ON u.id = i.invited_by
	 WHERE i.guild_id = $1 AND i.status = $2 ORDER BY i.created_at, i.id`
	if err := s.db.SelectContext(ctx, &entities, query, guildID, invitationStatusPending); err != nil {
		return nil, fmt.Errorf("query sent invitations: %w", err)
	}
	invitations := make([]SentInvitation, 0, len(entities))
	for _, e := range entities {
		invitations = append(invitations, SentInvitation{ID: e.ID, Email: e.Email, InvitedByName: e.InvitedByName, InvitedByEmail: e.InvitedByEmail, CreatedAt: e.CreatedAt})
	}
	return invitations, nil
}

func (s *postgresStorage) CancelInvitation(ctx context.Context, guildID, invitationID, requesterID int64) error {
	ctx, cancel := context.WithTimeout(ctx, invitationStorageTimeout)
	defer cancel()
	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin invitation cancellation: %w", err)
	}
	defer func() { _ = tx.Rollback() }()
	// Use the same lock order as acceptance: guild first, then invitation.
	var lockedID int64
	err = tx.GetContext(ctx, &lockedID, "SELECT id FROM guild WHERE id=$1 FOR UPDATE", guildID)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrGuildNotFound
	}
	if err != nil {
		return fmt.Errorf("lock cancellation guild: %w", err)
	}
	var isMember bool
	err = tx.GetContext(ctx, &isMember, `SELECT EXISTS(SELECT 1 FROM guild_member WHERE guild_id=$1 AND user_id=$2)`, guildID, requesterID)
	if err != nil {
		return fmt.Errorf("check cancellation membership: %w", err)
	}
	if !isMember {
		return ErrRequesterIsNotMember
	}
	var status string
	err = tx.GetContext(ctx, &status, `SELECT status FROM guild_invitation WHERE guild_id=$1 AND id=$2 FOR UPDATE`, guildID, invitationID)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrInvitationNotFound
	}
	if err != nil {
		return fmt.Errorf("lock cancelled invitation: %w", err)
	}
	if status != invitationStatusPending {
		return ErrInvitationAlreadyAnswered
	}
	_, err = tx.ExecContext(ctx, `UPDATE guild_invitation SET status='CANCELLED', responded_at=clock_timestamp(), responded_by=$2 WHERE id=$1`, invitationID, requesterID)
	if err != nil {
		return fmt.Errorf("persist invitation cancellation: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit invitation cancellation: %w", err)
	}
	return nil
}
