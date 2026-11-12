package guild

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/lib/pq"
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
	err = tx.GetContext(ctx, &isMember, `SELECT EXISTS(
 SELECT 1 FROM guild_member gm JOIN user_account u ON u.id=gm.user_id
 WHERE gm.guild_id=$1 AND LOWER(BTRIM(u.email))=$2)`, guildID, email)
	if err != nil {
		return fmt.Errorf("check invitation recipient membership: %w", err)
	}
	if isMember {
		return ErrGuildMemberAlreadySet
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
