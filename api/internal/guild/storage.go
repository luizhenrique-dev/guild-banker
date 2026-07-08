package guild

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"

	"github.com/luizhenrique-dev/guild-banker/api/internal/audit"
)

type Storage interface {
	Create(ctx context.Context, g *Guild, creatorUserID int64) error
	GetByID(ctx context.Context, id int64) (*Guild, error)
	NameExists(ctx context.Context, name string, excludeID int64) (bool, error)
	UpdateName(ctx context.Context, g *Guild) error
	Enable(ctx context.Context, id int64, by string, now time.Time) error
	Disable(ctx context.Context, id int64, by string, now time.Time) error
	ListByMember(ctx context.Context, userID int64) ([]*Guild, error)
	IsMember(ctx context.Context, guildID, userID int64) (bool, error)
	InviteByEmail(ctx context.Context, guildID int64, email string, invitedByUserID int64) error
	RemoveMember(ctx context.Context, guildID, userID int64) error
}

type guildEntity struct {
	ID          int64      `db:"id"`
	Name        string     `db:"name"`
	DisplayName string     `db:"display_name"`
	Enabled     bool       `db:"enabled"`
	CreatedAt   time.Time  `db:"created_at"`
	CreatedBy   string     `db:"created_by"`
	UpdatedAt   *time.Time `db:"updated_at"`
	UpdatedBy   *string    `db:"updated_by"`
	DisabledAt  *time.Time `db:"disabled_at"`
	DisabledBy  *string    `db:"disabled_by"`
}

type postgresStorage struct {
	db *sqlx.DB
}

func NewStorage(db *sqlx.DB) Storage {
	return &postgresStorage{db: db}
}

func (s *postgresStorage) Create(ctx context.Context, g *Guild, creatorUserID int64) error {
	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin guild create tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	const createGuildQuery = `
		INSERT INTO guild (name, display_name, created_at, created_by, enabled, updated_at, updated_by, disabled_at, disabled_by)
		VALUES (:name, :display_name, :created_at, :created_by, :enabled, :updated_at, :updated_by, :disabled_at, :disabled_by)
		RETURNING id
	`
	rows, err := tx.NamedQuery(createGuildQuery, toEntity(g))
	if err != nil {
		if isUniqueViolation(err) {
			return ErrGuildNameAlreadyUsed
		}
		return fmt.Errorf("create guild: %w", err)
	}
	defer func() { _ = rows.Close() }()

	if rows.Next() {
		if err := rows.Scan(&g.ID); err != nil {
			return fmt.Errorf("scan guild id: %w", err)
		}
	}

	const addCreatorAsMemberQuery = `
		INSERT INTO guild_member (guild_id, user_id, invited_at, invited_by)
		VALUES ($1, $2, NOW(), $3)
	`
	if _, err := tx.ExecContext(ctx, addCreatorAsMemberQuery, g.ID, creatorUserID, creatorUserID); err != nil {
		return fmt.Errorf("add creator as guild member: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit guild create tx: %w", err)
	}

	return nil
}

func (s *postgresStorage) GetByID(ctx context.Context, id int64) (*Guild, error) {
	const query = `
		SELECT id, name, display_name, enabled, created_at, created_by, updated_at, updated_by, disabled_at, disabled_by
		FROM guild
		WHERE id = $1
	`
	var entity guildEntity
	if err := s.db.GetContext(ctx, &entity, query, id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrGuildNotFound
		}
		return nil, fmt.Errorf("get guild by id: %w", err)
	}
	return toDomain(entity), nil
}

func (s *postgresStorage) NameExists(ctx context.Context, name string, excludeID int64) (bool, error) {
	const query = `
		SELECT EXISTS(
			SELECT 1
			FROM guild
			WHERE LOWER(name) = LOWER($1)
			  AND ($2 = 0 OR id <> $2)
		)
	`
	var exists bool
	if err := s.db.GetContext(ctx, &exists, query, name, excludeID); err != nil {
		return false, fmt.Errorf("check guild name existence: %w", err)
	}
	return exists, nil
}

func (s *postgresStorage) UpdateName(ctx context.Context, g *Guild) error {
	const query = `
		UPDATE guild
		SET display_name = :display_name,
		    updated_at   = :updated_at,
		    updated_by   = :updated_by
		WHERE id = :id
	`
	result, err := s.db.NamedExecContext(ctx, query, toEntity(g))
	if err != nil {
		if isUniqueViolation(err) {
			return ErrGuildNameAlreadyUsed
		}
		return fmt.Errorf("update guild name: %w", err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("get affected rows in update guild name: %w", err)
	}
	if rowsAffected == 0 {
		return ErrGuildNotFound
	}
	return nil
}

func (s *postgresStorage) Enable(ctx context.Context, id int64, by string, now time.Time) error {
	const query = `
		UPDATE guild
		SET enabled = TRUE, disabled_at = NULL, disabled_by = NULL, updated_at = $3, updated_by = $2
		WHERE id = $1
	`
	result, err := s.db.ExecContext(ctx, query, id, by, now)
	if err != nil {
		return fmt.Errorf("enable guild: %w", err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("get affected rows in enable guild: %w", err)
	}
	if rowsAffected == 0 {
		return ErrGuildNotFound
	}
	return nil
}

func (s *postgresStorage) Disable(ctx context.Context, id int64, by string, now time.Time) error {
	const query = `
		UPDATE guild
        SET enabled = FALSE, disabled_at = $3, disabled_by = $2, updated_at = $3, updated_by = $2
        WHERE id = $1
	`
	result, err := s.db.ExecContext(ctx, query, id, by, now)
	if err != nil {
		return fmt.Errorf("disable guild: %w", err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("get affected rows in disable guild: %w", err)
	}
	if rowsAffected == 0 {
		return ErrGuildNotFound
	}
	return nil
}

func (s *postgresStorage) ListByMember(ctx context.Context, userID int64) ([]*Guild, error) {
	const query = `
		SELECT g.id, g.name, g.display_name, g.enabled, g.created_at, g.created_by, g.updated_at, g.updated_by, g.disabled_at, g.disabled_by
		FROM guild g
		INNER JOIN guild_member gm ON gm.guild_id = g.id
		WHERE gm.user_id = $1
		ORDER BY g.id ASC
	`
	entities := make([]guildEntity, 0)
	if err := s.db.SelectContext(ctx, &entities, query, userID); err != nil {
		return nil, fmt.Errorf("list guilds by member: %w", err)
	}

	guilds := make([]*Guild, 0, len(entities))
	for _, entity := range entities {
		guilds = append(guilds, toDomain(entity))
	}
	return guilds, nil
}

func (s *postgresStorage) IsMember(ctx context.Context, guildID, userID int64) (bool, error) {
	const query = `
		SELECT EXISTS(
			SELECT 1
			FROM guild_member
			WHERE guild_id = $1
			  AND user_id = $2
		)
	`
	var exists bool
	if err := s.db.GetContext(ctx, &exists, query, guildID, userID); err != nil {
		return false, fmt.Errorf("check guild member: %w", err)
	}
	return exists, nil
}

func (s *postgresStorage) InviteByEmail(ctx context.Context, guildID int64, email string, invitedByUserID int64) error {
	const query = `
		INSERT INTO guild_member (guild_id, user_id, invited_at, invited_by)
		SELECT $1, u.id, NOW(), $3
		FROM users u
		WHERE LOWER(u.email) = LOWER($2)
	`
	result, err := s.db.ExecContext(ctx, query, guildID, email, invitedByUserID)
	if err != nil {
		if isUniqueViolation(err) {
			return ErrGuildMemberAlreadySet
		}
		err = mapFKViolation(err)
		return fmt.Errorf("invite user by email: %w", err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("get affected rows in invite user: %w", err)
	}
	if rowsAffected == 0 {
		return ErrUserNotFound
	}
	return nil
}

func (s *postgresStorage) RemoveMember(ctx context.Context, guildID, userID int64) error {
	const query = `
		DELETE FROM guild_member
		WHERE guild_id = $1
		  AND user_id = $2
	`
	result, err := s.db.ExecContext(ctx, query, guildID, userID)
	if err != nil {
		return fmt.Errorf("remove guild member: %w", err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("get affected rows in remove guild member: %w", err)
	}
	if rowsAffected == 0 {
		return ErrGuildMemberNotFound
	}
	return nil
}

func toEntity(g *Guild) guildEntity {
	return guildEntity{
		ID:          g.ID,
		Name:        g.Name,
		DisplayName: g.DisplayName,
		Enabled:     g.Enabled,
		CreatedAt:   g.CreatedAt,
		CreatedBy:   g.CreatedBy,
		UpdatedAt:   g.UpdatedAt,
		UpdatedBy:   g.UpdatedBy,
		DisabledAt:  g.DisabledAt,
		DisabledBy:  g.DisabledBy,
	}
}

func toDomain(e guildEntity) *Guild {
	return &Guild{
		ID:          e.ID,
		Name:        e.Name,
		DisplayName: e.DisplayName,
		DisableEntry: audit.DisableEntry{
			Enabled: e.Enabled,
			Entry: audit.Entry{
				CreatedAt: e.CreatedAt,
				CreatedBy: e.CreatedBy,
				UpdatedAt: e.UpdatedAt,
				UpdatedBy: e.UpdatedBy,
			},
			DisabledAt: e.DisabledAt,
			DisabledBy: e.DisabledBy,
		},
	}
}

func isUniqueViolation(err error) bool {
	var pqErr *pq.Error
	return errors.As(err, &pqErr) && pqErr.Code == "23505"
}

func mapFKViolation(err error) error {
	if pqErr, ok := errors.AsType[*pq.Error](err); ok && pqErr.Code == "23503" {
		if strings.Contains(pqErr.Constraint, "guild") {
			return ErrGuildNotFound
		}
		if strings.Contains(pqErr.Constraint, "user") {
			return ErrUserNotFound
		}
	}
	return err
}
