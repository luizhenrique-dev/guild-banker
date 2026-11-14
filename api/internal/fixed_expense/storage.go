package fixedexpense

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/shopspring/decimal"
)

type fixedExpenseEntity struct {
	ID        int64           `db:"id"`
	GuildID   int64           `db:"guild_id"`
	Name      string          `db:"name"`
	Amount    decimal.Decimal `db:"amount"`
	DueDay    int             `db:"due_day"`
	Category  string          `db:"category"`
	Status    string          `db:"status"`
	CreatedAt time.Time       `db:"created_at"`
	CreatedBy string          `db:"created_by"`
	UpdatedAt *time.Time      `db:"updated_at"`
	UpdatedBy *string         `db:"updated_by"`
}

type postgresStorage struct {
	db *sqlx.DB
}

func NewStorage(db *sqlx.DB) Storage {
	return &postgresStorage{db: db}
}

func (s *postgresStorage) Create(ctx context.Context, fixedExpense *FixedExpense) error {
	const query = `
		INSERT INTO fixed_expense (guild_id, name, amount, due_day, category, status, created_at, created_by, updated_at, updated_by)
		VALUES (:guild_id, :name, :amount, :due_day, :category, :status, :created_at, :created_by, :updated_at, :updated_by)
		RETURNING id
	`

	params := map[string]any{
		"guild_id":   fixedExpense.GuildID,
		"name":       fixedExpense.Name,
		"amount":     fixedExpense.Amount,
		"due_day":    fixedExpense.DueDay,
		"category":   string(fixedExpense.Category),
		"status":     string(fixedExpense.Status),
		"created_at": fixedExpense.CreatedAt,
		"created_by": fixedExpense.CreatedBy,
		"updated_at": fixedExpense.UpdatedAt,
		"updated_by": fixedExpense.UpdatedBy,
	}

	rows, err := s.db.NamedQueryContext(ctx, query, params)
	if err != nil {
		return fmt.Errorf("create fixed expense: %w", err)
	}
	defer func() { _ = rows.Close() }()

	if rows.Next() {
		if err := rows.Scan(&fixedExpense.ID); err != nil {
			return fmt.Errorf("scan fixed expense id: %w", err)
		}
	}

	if err := rows.Err(); err != nil {
		return fmt.Errorf("read fixed expense id: %w", err)
	}
	return nil
}

func (s *postgresStorage) GetByIDAndGuild(ctx context.Context, id, guildID int64) (*FixedExpense, error) {
	const query = `
		SELECT id, guild_id, name, amount, due_day, category, status, created_at, created_by, updated_at, updated_by
		FROM fixed_expense
		WHERE id = $1 AND guild_id = $2
	`

	var entity fixedExpenseEntity
	if err := s.db.GetContext(ctx, &entity, query, id, guildID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrFixedExpenseNotFound
		}

		return nil, fmt.Errorf("get fixed expense by id and guild: %w", err)
	}

	return toDomain(entity), nil
}

func (s *postgresStorage) ListActiveByGuild(ctx context.Context, guildID int64) ([]*FixedExpense, error) {
	return s.listByGuild(ctx, guildID, false)
}

func (s *postgresStorage) ListByGuild(ctx context.Context, guildID int64) ([]*FixedExpense, error) {
	return s.listByGuild(ctx, guildID, true)
}

func (s *postgresStorage) listByGuild(ctx context.Context, guildID int64, includePaused bool) ([]*FixedExpense, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	const query = `
		SELECT id, guild_id, name, amount, due_day, category, status, created_at, created_by, updated_at, updated_by
		FROM fixed_expense
		WHERE guild_id = $1 AND (status = $2 OR ($3 AND status = $4))
		ORDER BY due_day ASC, name ASC
	`

	entities := make([]fixedExpenseEntity, 0)
	if err := s.db.SelectContext(ctx, &entities, query, guildID, StatusActive, includePaused, StatusPaused); err != nil {
		return nil, fmt.Errorf("list fixed expenses by guild: %w", err)
	}

	fixedExpenses := make([]*FixedExpense, 0, len(entities))
	for _, entity := range entities {
		fixedExpenses = append(fixedExpenses, toDomain(entity))
	}

	return fixedExpenses, nil
}

func (s *postgresStorage) Update(ctx context.Context, fixedExpense *FixedExpense) error {
	const query = `
		UPDATE fixed_expense
		SET amount = $1,
			due_day = $2,
			category = $3,
			status = $4,
			updated_at = $5,
			updated_by = $6
		WHERE id = $7 AND guild_id = $8
	`

	result, err := s.db.ExecContext(
		ctx,
		query,
		fixedExpense.Amount,
		fixedExpense.DueDay,
		fixedExpense.Category,
		fixedExpense.Status,
		fixedExpense.UpdatedAt,
		fixedExpense.UpdatedBy,
		fixedExpense.ID,
		fixedExpense.GuildID,
	)
	if err != nil {
		return fmt.Errorf("update fixed expense: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("update fixed expense rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return ErrFixedExpenseNotFound
	}

	return nil
}

func (s *postgresStorage) Delete(ctx context.Context, id, guildID int64) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	result, err := s.db.ExecContext(ctx, "DELETE FROM fixed_expense WHERE id=$1 AND guild_id=$2", id, guildID)
	if err != nil {
		return fmt.Errorf("delete fixed expense: %w", err)
	}
	count, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("count deleted fixed expenses: %w", err)
	}
	if count == 0 {
		return ErrFixedExpenseNotFound
	}
	return nil
}

func toDomain(entity fixedExpenseEntity) *FixedExpense {
	return &FixedExpense{
		ID:        entity.ID,
		GuildID:   entity.GuildID,
		Name:      entity.Name,
		Amount:    entity.Amount,
		DueDay:    entity.DueDay,
		Category:  Category(entity.Category),
		Status:    Status(entity.Status),
		CreatedAt: entity.CreatedAt,
		CreatedBy: entity.CreatedBy,
		UpdatedAt: entity.UpdatedAt,
		UpdatedBy: entity.UpdatedBy,
	}
}

func (s *postgresStorage) IsGuildMember(ctx context.Context, guildID, userID int64) (bool, error) {
	const query = `SELECT EXISTS (SELECT 1 FROM guild_member WHERE guild_id = $1 AND user_id = $2)`
	var isMember bool
	if err := s.db.GetContext(ctx, &isMember, query, guildID, userID); err != nil {
		return false, fmt.Errorf("check guild membership: %w", err)
	}
	return isMember, nil
}
