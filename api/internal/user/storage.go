package user

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"

	"github.com/luizhenrique-dev/guild-banker/api/internal/audit"
)

type Storage interface {
	Create(ctx context.Context, u *User) error
	Update(ctx context.Context, u *User) error
	GetByID(ctx context.Context, id int64) (*User, error)
	GetByExternalID(ctx context.Context, externalID string) (*User, error)
	GetByEmail(ctx context.Context, email string) (*User, error)
	Enable(ctx context.Context, id int64) error
	Disable(ctx context.Context, id int64, disabledBy string) error
}

type userEntity struct {
	ID         int64      `db:"id"`
	ExternalID string     `db:"external_id"`
	Name       string     `db:"name"`
	Email      string     `db:"email"`
	CreatedAt  time.Time  `db:"created_at"`
	CreatedBy  string     `db:"created_by"`
	Enabled    bool       `db:"enabled"`
	UpdatedAt  *time.Time `db:"updated_at"`
	DisabledAt *time.Time `db:"disabled_at"`
	DisabledBy *string    `db:"disabled_by"`
}

type postgresStorage struct {
	db *sqlx.DB
}

func NewStorage(db *sqlx.DB) Storage {
	return &postgresStorage{db: db}
}

func (s *postgresStorage) Create(ctx context.Context, u *User) error {
	const query = `
		INSERT INTO user_account (external_id, name, email, created_at, created_by, enabled, updated_at, disabled_at, disabled_by)
		VALUES (:external_id, :name, :email, :created_at, :created_by, :enabled, :updated_at, :disabled_at, :disabled_by)
		RETURNING id
	`
	entity := toEntity(u)
	rows, err := s.db.NamedQueryContext(ctx, query, entity)
	if err != nil {
		return fmt.Errorf("create user: %w", err)
	}
	defer func() { _ = rows.Close() }()

	if rows.Next() {
		if err := rows.Scan(&u.ID); err != nil {
			return fmt.Errorf("scan created user id: %w", err)
		}
	}
	return nil
}

func (s *postgresStorage) Update(ctx context.Context, u *User) error {
	const query = `
		UPDATE user_account
		SET name        = :name,
		    email       = :email,
		    enabled     = :enabled,
		    updated_at  = :updated_at,
		    disabled_at = :disabled_at,
		    disabled_by = :disabled_by
		WHERE id = :id
	`
	if _, err := s.db.NamedExecContext(ctx, query, toEntity(u)); err != nil {
		return fmt.Errorf("update user: %w", err)
	}
	return nil
}

func (s *postgresStorage) GetByID(ctx context.Context, id int64) (*User, error) {
	const query = `
		SELECT id, external_id, name, email, created_at, created_by, enabled, updated_at, disabled_at, disabled_by
		FROM user_account
		WHERE id = $1
	`
	return s.get(ctx, query, id)
}

func (s *postgresStorage) GetByExternalID(ctx context.Context, externalID string) (*User, error) {
	const query = `
		SELECT id, external_id, name, email, created_at, created_by, enabled, updated_at, disabled_at, disabled_by
		FROM user_account
		WHERE external_id = $1
	`
	return s.get(ctx, query, externalID)
}

func (s *postgresStorage) GetByEmail(ctx context.Context, email string) (*User, error) {
	const query = `
		SELECT id, external_id, name, email, created_at, created_by, enabled, updated_at, disabled_at, disabled_by
		FROM user_account
		WHERE email = $1
	`
	return s.get(ctx, query, email)
}

func (s *postgresStorage) Enable(ctx context.Context, id int64) error {
	const query = `
		UPDATE user_account
		SET enabled = TRUE, disabled_at = NULL, disabled_by = NULL, updated_at = NOW()
		WHERE id = $1
	`
	if _, err := s.db.ExecContext(ctx, query, id); err != nil {
		return fmt.Errorf("enable user: %w", err)
	}
	return nil
}

func (s *postgresStorage) Disable(ctx context.Context, id int64, disabledBy string) error {
	const query = `
		UPDATE user_account
		SET enabled = FALSE, disabled_at = NOW(), disabled_by = $1, updated_at = NOW()
		WHERE id = $2
	`
	if _, err := s.db.ExecContext(ctx, query, disabledBy, id); err != nil {
		return fmt.Errorf("disable user: %w", err)
	}
	return nil
}

func (s *postgresStorage) get(ctx context.Context, query string, arg any) (*User, error) {
	var entity userEntity
	if err := s.db.GetContext(ctx, &entity, query, arg); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("user not found: %w", err)
		}
		return nil, fmt.Errorf("query user: %w", err)
	}
	return toDomain(entity), nil
}

func toEntity(u *User) userEntity {
	return userEntity{
		ID:         u.ID,
		ExternalID: u.ExternalID,
		Name:       u.Name,
		Email:      u.Email,
		CreatedAt:  u.CreatedAt,
		CreatedBy:  u.CreatedBy,
		Enabled:    u.Enabled,
		UpdatedAt:  u.UpdatedAt,
		DisabledAt: u.DisabledAt,
		DisabledBy: u.DisabledBy,
	}
}

func toDomain(e userEntity) *User {
	return &User{
		ID:         e.ID,
		ExternalID: e.ExternalID,
		Name:       e.Name,
		Email:      e.Email,
		DisableEntry: audit.DisableEntry{
			Enabled: e.Enabled,
			Entry: audit.Entry{
				CreatedAt: e.CreatedAt,
				CreatedBy: e.CreatedBy,
				UpdatedAt: e.UpdatedAt,
				UpdatedBy: e.DisabledBy,
			},
			DisabledAt: e.DisabledAt,
			DisabledBy: e.DisabledBy,
		},
	}
}
