package transaction

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"github.com/shopspring/decimal"

	"github.com/luizhenrique-dev/guild-banker/api/internal/audit"
)

type transactionEntity struct {
	ID            int64           `db:"id"`
	Type          string          `db:"type"`
	Description   string          `db:"description"`
	Amount        decimal.Decimal `db:"amount"`
	Category      string          `db:"category"`
	Status        string          `db:"status"`
	Source        string          `db:"source"`
	Visibility    string          `db:"visibility"`
	OccurredAt    time.Time       `db:"occurred_at"`
	UserAccountID int64           `db:"user_account_id"`
	GuildID       int64           `db:"guild_id"`
	CreatedAt     time.Time       `db:"created_at"`
	CreatedBy     string          `db:"created_by"`
	UpdatedAt     *time.Time      `db:"updated_at"`
	UpdatedBy     *string         `db:"updated_by"`
}

type postgresStorage struct {
	db *sqlx.DB
}

func NewStorage(db *sqlx.DB) Storage {
	return &postgresStorage{db: db}
}

func (s *postgresStorage) IsGuildMember(ctx context.Context, guildID, userID int64) (bool, error) {
	const query = `SELECT EXISTS(SELECT 1 FROM guild_member WHERE guild_id = $1 AND user_id = $2)`

	var isMember bool
	if err := s.db.GetContext(ctx, &isMember, query, guildID, userID); err != nil {
		return false, fmt.Errorf("check guild member: %w", err)
	}

	return isMember, nil
}

func (s *postgresStorage) Create(ctx context.Context, transaction *Transaction) error {
	const query = `
		INSERT INTO transaction (
			type, description, amount, category, status, source, visibility,
			occurred_at, user_account_id, guild_id, created_at, created_by, updated_at, updated_by
		) VALUES (
			:type, :description, :amount, :category, :status, :source, :visibility,
			:occurred_at, :user_account_id, :guild_id, :created_at, :created_by, :updated_at, :updated_by
		)
		RETURNING id
	`

	params := map[string]any{
		"type":            string(transaction.Type),
		"description":     transaction.Description,
		"amount":          transaction.Amount,
		"category":        string(transaction.Category),
		"status":          string(transaction.Status),
		"source":          string(transaction.Source),
		"visibility":      string(transaction.Visibility),
		"occurred_at":     transaction.OccurredAt,
		"user_account_id": transaction.UserAccountID,
		"guild_id":        transaction.GuildID,
		"created_at":      transaction.CreatedAt,
		"created_by":      transaction.CreatedBy,
		"updated_at":      transaction.UpdatedAt,
		"updated_by":      transaction.UpdatedBy,
	}

	rows, err := s.db.NamedQueryContext(ctx, query, params)
	if err != nil {
		return fmt.Errorf("create transaction: %w", err)
	}
	defer func() { _ = rows.Close() }()

	if rows.Next() {
		if err := rows.Scan(&transaction.ID); err != nil {
			return fmt.Errorf("scan created transaction id: %w", err)
		}
	}

	return nil
}

func (s *postgresStorage) List(ctx context.Context, filter ListFilter) ([]*Transaction, error) {
	query := strings.Builder{}
	query.WriteString(`
		SELECT id, type, description, amount, category, status, source, visibility,
			occurred_at, user_account_id, guild_id, created_at, created_by, updated_at, updated_by
		FROM transaction
		WHERE guild_id = $1
		  AND (visibility = 'PUBLIC' OR user_account_id = $2)
		  AND status = $3
	`)

	args := []any{filter.GuildID, filter.RequesterUserID, filter.Status}
	argPos := 4

	if filter.Visibility != nil {
		fmt.Fprintf(&query, " AND visibility = $%d", argPos)
		args = append(args, *filter.Visibility)
		argPos++
	}
	if filter.DateFrom != nil {
		fmt.Fprintf(&query, " AND occurred_at >= $%d", argPos)
		args = append(args, *filter.DateFrom)
		argPos++
	}
	if filter.DateTo != nil {
		fmt.Fprintf(&query, " AND occurred_at <= $%d", argPos)
		args = append(args, *filter.DateTo)
		argPos++
	}
	if filter.Category != nil {
		fmt.Fprintf(&query, " AND category = $%d", argPos)
		args = append(args, *filter.Category)
		argPos++
	}
	if filter.Type != nil {
		fmt.Fprintf(&query, " AND type = $%d", argPos)
		args = append(args, *filter.Type)
		argPos++
	}
	if filter.Source != nil {
		fmt.Fprintf(&query, " AND source = $%d", argPos)
		args = append(args, *filter.Source)
		argPos++
	}
	if filter.CursorOccurredAt != nil && filter.CursorID != nil {
		fmt.Fprintf(&query, " AND (occurred_at, id) < ($%d, $%d)", argPos, argPos+1)
		args = append(args, *filter.CursorOccurredAt, *filter.CursorID)
		argPos += 2
	}

	fmt.Fprintf(&query, " ORDER BY occurred_at DESC, id DESC LIMIT $%d", argPos)
	args = append(args, filter.Limit)

	entities := make([]transactionEntity, 0)
	if err := s.db.SelectContext(ctx, &entities, query.String(), args...); err != nil {
		return nil, fmt.Errorf("list transactions: %w", err)
	}

	transactions := make([]*Transaction, 0, len(entities))
	for _, entity := range entities {
		transactions = append(transactions, toDomain(entity))
	}

	return transactions, nil
}

func (s *postgresStorage) GetByIDForWrite(ctx context.Context, id, guildID, requesterUserID int64) (*Transaction, error) {
	const query = `
		SELECT id, type, description, amount, category, status, source, visibility,
			occurred_at, user_account_id, guild_id, created_at, created_by, updated_at, updated_by
		FROM transaction
		WHERE id = $1
		  AND guild_id = $2
		  AND (visibility = 'PUBLIC' OR user_account_id = $3)
	`

	var entity transactionEntity
	if err := s.db.GetContext(ctx, &entity, query, id, guildID, requesterUserID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrTransactionNotFound
		}
		return nil, fmt.Errorf("get transaction by id for write: %w", err)
	}

	return toDomain(entity), nil
}

func (s *postgresStorage) GetByIDForOwner(ctx context.Context, id, guildID, requesterUserID int64) (*Transaction, error) {
	const query = `
		SELECT id, type, description, amount, category, status, source, visibility,
			occurred_at, user_account_id, guild_id, created_at, created_by, updated_at, updated_by
		FROM transaction
		WHERE id = $1
		  AND guild_id = $2
		  AND user_account_id = $3
	`

	var entity transactionEntity
	if err := s.db.GetContext(ctx, &entity, query, id, guildID, requesterUserID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrTransactionNotFound
		}
		return nil, fmt.Errorf("get transaction by id for owner: %w", err)
	}

	return toDomain(entity), nil
}

func (s *postgresStorage) Update(ctx context.Context, transaction *Transaction) error {
	const query = `
		UPDATE transaction
		SET type = $1,
			description = $2,
			amount = $3,
			category = $4,
			status = $5,
			visibility = $6,
			occurred_at = $7,
			updated_at = $8,
			updated_by = $9
		WHERE id = $10
	`

	result, err := s.db.ExecContext(
		ctx,
		query,
		transaction.Type,
		transaction.Description,
		transaction.Amount,
		transaction.Category,
		transaction.Status,
		transaction.Visibility,
		transaction.OccurredAt,
		transaction.UpdatedAt,
		transaction.UpdatedBy,
		transaction.ID,
	)
	if err != nil {
		return fmt.Errorf("update transaction: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("update transaction rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return ErrTransactionNotFound
	}

	return nil
}

func (s *postgresStorage) BulkCategorize(
	ctx context.Context,
	guildID, requesterUserID int64,
	transactionIDs []int64,
	category Category,
) (int, []BulkFailure, error) {
	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return 0, nil, fmt.Errorf("bulk categorize begin transaction: %w", err)
	}

	defer func() {
		_ = tx.Rollback()
	}()

	const selectQuery = `
		SELECT id, visibility, user_account_id
		FROM transaction
		WHERE guild_id = $1 AND id = ANY($2)
	`

	type checkEntity struct {
		ID            int64  `db:"id"`
		Visibility    string `db:"visibility"`
		UserAccountID int64  `db:"user_account_id"`
	}

	rows := make([]checkEntity, 0)
	if err := tx.SelectContext(ctx, &rows, selectQuery, guildID, pq.Array(transactionIDs)); err != nil {
		return 0, nil, fmt.Errorf("bulk categorize select: %w", err)
	}

	allowedIDs := make([]int64, 0, len(rows))
	found := make(map[int64]bool, len(rows))
	failed := make([]BulkFailure, 0)

	for _, row := range rows {
		found[row.ID] = true
		if row.Visibility == string(VisibilityPrivate) && row.UserAccountID != requesterUserID {
			failed = append(failed, BulkFailure{TransactionID: row.ID, Reason: "forbidden"})
			continue
		}
		allowedIDs = append(allowedIDs, row.ID)
	}

	for _, id := range transactionIDs {
		if !found[id] {
			failed = append(failed, BulkFailure{TransactionID: id, Reason: "not_found"})
		}
	}

	updated := 0
	if len(allowedIDs) > 0 {
		const updateQuery = `
			UPDATE transaction
			SET category = $1
			WHERE guild_id = $2
			  AND id = ANY($3)
		`

		result, err := tx.ExecContext(ctx, updateQuery, category, guildID, pq.Array(allowedIDs))
		if err != nil {
			return 0, nil, fmt.Errorf("bulk categorize update: %w", err)
		}

		rowsAffected, err := result.RowsAffected()
		if err != nil {
			return 0, nil, fmt.Errorf("bulk categorize rows affected: %w", err)
		}
		updated = int(rowsAffected)
	}

	if err := tx.Commit(); err != nil {
		return 0, nil, fmt.Errorf("bulk categorize commit: %w", err)
	}

	return updated, failed, nil
}

func toDomain(entity transactionEntity) *Transaction {
	return &Transaction{
		ID:            entity.ID,
		Type:          Type(entity.Type),
		Description:   entity.Description,
		Amount:        entity.Amount,
		Category:      Category(entity.Category),
		Status:        Status(entity.Status),
		Source:        Source(entity.Source),
		Visibility:    Visibility(entity.Visibility),
		OccurredAt:    entity.OccurredAt,
		UserAccountID: entity.UserAccountID,
		GuildID:       entity.GuildID,
		Entry: audit.Entry{
			CreatedAt: entity.CreatedAt,
			CreatedBy: entity.CreatedBy,
			UpdatedAt: entity.UpdatedAt,
			UpdatedBy: entity.UpdatedBy,
		},
	}
}
