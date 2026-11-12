package guild

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/lib/pq"
	"github.com/luizhenrique-dev/guild-banker/api/internal/audit"
)

func (s *postgresStorage) InviteByEmail(ctx context.Context, guildID int64, email string, invitedByUserID int64) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin invitation: %w", err)
	}
	defer func() { _ = tx.Rollback() }()
	var lockedID int64
	err = tx.GetContext(ctx, &lockedID, "SELECT id FROM guild WHERE id=$1 FOR UPDATE", guildID)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrGuildNotFound
	}
	if err != nil {
		return fmt.Errorf("lock invitation guild: %w", err)
	}
	var isMember bool
	err = tx.GetContext(ctx, &isMember, `SELECT EXISTS(SELECT 1 FROM guild_member WHERE guild_id=$1 AND user_id=$2)`, guildID, invitedByUserID)
	if err != nil {
		return fmt.Errorf("check inviter membership: %w", err)
	}
	if !isMember {
		return ErrRequesterIsNotMember
	}
	memberEmails := []string{}
	err = tx.SelectContext(ctx, &memberEmails, `SELECT u.email FROM guild_member gm JOIN user_account u ON u.id=gm.user_id WHERE gm.guild_id=$1`, guildID)
	if err != nil {
		return fmt.Errorf("check invitation recipient membership: %w", err)
	}
	for _, memberEmail := range memberEmails {
		canonical, err := invitationEmail(memberEmail)
		if err == nil && canonical == email {
			return ErrGuildMemberAlreadySet
		}
	}
	_, err = tx.ExecContext(ctx, `INSERT INTO guild_invitation(guild_id,recipient_email,invited_by) VALUES ($1,$2,$3)`, guildID, email, invitedByUserID)
	if err != nil {
		if constraint, ok := errors.AsType[*pq.Error](err); ok && constraint.Code == "23505" && constraint.Constraint == "idx_guild_invitation_pending" {
			return ErrInvitationAlreadyPending
		}
		return fmt.Errorf("insert invitation: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit invitation: %w", err)
	}
	return nil
}

type invitationEntity struct {
	ID             int64  `db:"id"`
	GuildID        int64  `db:"guild_id"`
	Name           string `db:"name"`
	DisplayName    string `db:"display_name"`
	Description    string `db:"description"`
	Enabled        bool   `db:"enabled"`
	InvitedByName  string `db:"invited_by_name"`
	InvitedByEmail string `db:"invited_by_email"`
}

func (s *postgresStorage) ListInvitations(ctx context.Context, email string) ([]Invitation, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	entities := []invitationEntity{}
	err := s.db.SelectContext(ctx, &entities, `SELECT i.id, g.id AS guild_id, g.name,g.display_name,g.description,g.enabled,
 u.name AS invited_by_name,u.email AS invited_by_email
 FROM guild_invitation i JOIN guild g ON g.id=i.guild_id JOIN user_account u ON u.id=i.invited_by
 WHERE i.recipient_email=$1 AND i.status='PENDING' ORDER BY i.created_at,i.id`, email)
	if err != nil {
		return nil, fmt.Errorf("query pending invitations: %w", err)
	}
	invitations := make([]Invitation, 0, len(entities))
	for _, e := range entities {
		invitations = append(invitations, Invitation{ID: e.ID, Guild: Guild{ID: e.GuildID, Name: e.Name, DisplayName: e.DisplayName, Description: e.Description}, InvitedByName: e.InvitedByName, InvitedByEmail: e.InvitedByEmail})
		invitations[len(invitations)-1].Guild.Enabled = e.Enabled
	}
	return invitations, nil
}

func (s *postgresStorage) RespondInvitation(ctx context.Context, id int64, actor audit.Actor, accept bool) (*Guild, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin invitation response: %w", err)
	}
	defer func() { _ = tx.Rollback() }()
	var guildID int64
	err = tx.GetContext(ctx, &guildID, "SELECT guild_id FROM guild_invitation WHERE id=$1 AND recipient_email=$2", id, actor.Email)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrInvitationNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("find invitation: %w", err)
	}
	var guild guildEntity
	err = tx.GetContext(ctx, &guild, `SELECT id,name,display_name,description,enabled,created_at,created_by,updated_at,updated_by,disabled_at,disabled_by FROM guild WHERE id=$1 FOR UPDATE`, guildID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrInvitationNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("lock response guild: %w", err)
	}
	var invitation struct {
		Status    string `db:"status"`
		InvitedBy int64  `db:"invited_by"`
	}
	err = tx.GetContext(ctx, &invitation, `SELECT status,invited_by FROM guild_invitation WHERE id=$1 AND recipient_email=$2 FOR UPDATE`, id, actor.Email)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrInvitationNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("lock invitation: %w", err)
	}
	if invitation.Status != "PENDING" {
		return nil, ErrInvitationAlreadyAnswered
	}
	status := "DECLINED"
	if accept {
		_, err = tx.ExecContext(ctx, `INSERT INTO guild_member(guild_id,user_id,invited_by) VALUES($1,$2,$3) ON CONFLICT(guild_id,user_id) DO NOTHING`, guildID, actor.UserID, invitation.InvitedBy)
		if err != nil {
			return nil, fmt.Errorf("associate invitation recipient: %w", err)
		}
		status = "ACCEPTED"
	}
	_, err = tx.ExecContext(ctx, `UPDATE guild_invitation SET status=$2,responded_at=NOW(),responded_by=$3 WHERE id=$1`, id, status, actor.UserID)
	if err != nil {
		return nil, fmt.Errorf("record invitation response: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit invitation response: %w", err)
	}
	return toDomain(guild), nil
}
