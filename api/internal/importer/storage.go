package importer

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/shopspring/decimal"

	"github.com/luizhenrique-dev/guild-banker/api/internal/audit"
	"github.com/luizhenrique-dev/guild-banker/api/internal/transaction"
)

type postgresStorage struct {
	db *sqlx.DB
}

func NewStorage(db *sqlx.DB) Storage {
	return &postgresStorage{db: db}
}

type importBatchEntity struct {
	ID            int64      `db:"id"`
	GuildID       int64      `db:"guild_id"`
	UserAccountID int64      `db:"user_account_id"`
	FileName      string     `db:"file_name"`
	SourceBank    string     `db:"source_bank"`
	Status        string     `db:"status"`
	CreatedAt     time.Time  `db:"created_at"`
	CreatedBy     string     `db:"created_by"`
	UpdatedAt     *time.Time `db:"updated_at"`
	UpdatedBy     *string    `db:"updated_by"`
}

type importItemEntity struct {
	ID                 int64           `db:"id"`
	ImportBatchID      int64           `db:"import_batch_id"`
	UserAccountID      int64           `db:"user_account_id"`
	OccurredAt         time.Time       `db:"occurred_at"`
	Description        string          `db:"description"`
	Amount             decimal.Decimal `db:"amount"`
	Type               string          `db:"type"`
	Category           string          `db:"category"`
	BankCategory       string          `db:"bank_category"`
	CardLast4          string          `db:"card_last4"`
	InstallmentCurrent *int            `db:"installment_current"`
	InstallmentTotal   *int            `db:"installment_total"`
	Status             string          `db:"status"`
	TransactionID      *int64          `db:"transaction_id"`
	CreatedAt          time.Time       `db:"created_at"`
	CreatedBy          string          `db:"created_by"`
	UpdatedAt          *time.Time      `db:"updated_at"`
	UpdatedBy          *string         `db:"updated_by"`
}

func (s *postgresStorage) IsGuildMember(ctx context.Context, guildID, userID int64) (bool, error) {
	const query = `SELECT EXISTS(SELECT 1 FROM guild_member WHERE guild_id = $1 AND user_id = $2)`

	var isMember bool
	if err := s.db.GetContext(ctx, &isMember, query, guildID, userID); err != nil {
		return false, fmt.Errorf("check guild member: %w", err)
	}

	return isMember, nil
}

func (s *postgresStorage) CreateBatch(ctx context.Context, batch *ImportBatch, items []*ImportItem) error {
	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx create batch: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	const createBatchQuery = `
		INSERT INTO import_batch (guild_id, user_account_id, file_name, source_bank, status, created_at, created_by, updated_at, updated_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id
	`

	if err := tx.QueryRowxContext(ctx, createBatchQuery,
		batch.GuildID, batch.UserAccountID, batch.FileName, string(batch.SourceBank), string(batch.Status),
		batch.CreatedAt, batch.CreatedBy, batch.UpdatedAt, batch.UpdatedBy,
	).Scan(&batch.ID); err != nil {
		return fmt.Errorf("insert import batch: %w", err)
	}

	const createItemQuery = `
		INSERT INTO import_item (
			import_batch_id, user_account_id, occurred_at, description, amount, type, category,
			bank_category, card_last4, installment_current, installment_total, status, transaction_id,
			created_at, created_by, updated_at, updated_by
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7,
			$8, $9, $10, $11, $12, $13,
			$14, $15, $16, $17
		)
		RETURNING id
	`

	for _, item := range items {
		item.ImportBatchID = batch.ID
		if err := tx.QueryRowxContext(ctx, createItemQuery,
			item.ImportBatchID,
			item.UserAccountID,
			item.OccurredAt,
			item.Description,
			item.Amount,
			string(item.Type),
			string(item.Category),
			item.BankCategory,
			item.CardLast4,
			item.InstallmentCurrent,
			item.InstallmentTotal,
			string(item.Status),
			item.TransactionID,
			item.CreatedAt,
			item.CreatedBy,
			item.UpdatedAt,
			item.UpdatedBy,
		).Scan(&item.ID); err != nil {
			return fmt.Errorf("insert import item: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit create batch: %w", err)
	}

	return nil
}

func (s *postgresStorage) GetBatch(ctx context.Context, guildID, importID, requesterUserID int64) (*ImportBatch, []*ImportItem, error) {
	const batchQuery = `
		SELECT id, guild_id, user_account_id, file_name, source_bank, status, created_at, created_by, updated_at, updated_by
		FROM import_batch
		WHERE id = $1 AND guild_id = $2 AND user_account_id = $3
	`

	var batchEntity importBatchEntity
	if err := s.db.GetContext(ctx, &batchEntity, batchQuery, importID, guildID, requesterUserID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil, ErrImportBatchNotFound
		}
		return nil, nil, fmt.Errorf("get import batch: %w", err)
	}

	const itemsQuery = `
		SELECT id, import_batch_id, user_account_id, occurred_at, description, amount, type, category, bank_category,
			card_last4, installment_current, installment_total, status, transaction_id, created_at, created_by, updated_at, updated_by
		FROM import_item
		WHERE import_batch_id = $1
		ORDER BY id ASC
	`

	entities := make([]importItemEntity, 0)
	if err := s.db.SelectContext(ctx, &entities, itemsQuery, importID); err != nil {
		return nil, nil, fmt.Errorf("list import items: %w", err)
	}

	items := make([]*ImportItem, 0, len(entities))
	for _, entity := range entities {
		items = append(items, toDomainItem(entity))
	}

	return toDomainBatch(batchEntity), items, nil
}

func (s *postgresStorage) GetItemForUpdate(ctx context.Context, guildID, importID, itemID, requesterUserID int64) (*ImportItem, error) {
	const query = `
		SELECT i.id, i.import_batch_id, i.user_account_id, i.occurred_at, i.description, i.amount, i.type, i.category,
			i.bank_category, i.card_last4, i.installment_current, i.installment_total, i.status, i.transaction_id,
			i.created_at, i.created_by, i.updated_at, i.updated_by
		FROM import_item i
		JOIN import_batch b ON b.id = i.import_batch_id
		WHERE i.id = $1 AND i.import_batch_id = $2 AND b.guild_id = $3 AND b.user_account_id = $4
	`

	var entity importItemEntity
	if err := s.db.GetContext(ctx, &entity, query, itemID, importID, guildID, requesterUserID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrImportItemNotFound
		}
		return nil, fmt.Errorf("get import item for update: %w", err)
	}

	return toDomainItem(entity), nil
}

func (s *postgresStorage) UpdateItem(ctx context.Context, item *ImportItem) error {
	const query = `
		UPDATE import_item
		SET occurred_at = $1,
			description = $2,
			amount = $3,
			type = $4,
			category = $5,
			status = $6,
			updated_at = $7,
			updated_by = $8
		WHERE id = $9
	`

	result, err := s.db.ExecContext(ctx, query,
		item.OccurredAt,
		item.Description,
		item.Amount,
		string(item.Type),
		string(item.Category),
		string(item.Status),
		item.UpdatedAt,
		item.UpdatedBy,
		item.ID,
	)
	if err != nil {
		return fmt.Errorf("update import item: %w", err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("update import item rows affected: %w", err)
	}

	if affected == 0 {
		return ErrImportItemNotFound
	}

	return nil
}

func (s *postgresStorage) DiscardItem(ctx context.Context, guildID, importID, itemID, requesterUserID int64, updatedBy string) error {
	now := time.Now()
	const query = `
		UPDATE import_item i
		SET status = 'DISCARDED', updated_at = $1, updated_by = $2
		FROM import_batch b
		WHERE i.id = $3 AND i.import_batch_id = $4 AND b.id = i.import_batch_id AND b.guild_id = $5 AND b.user_account_id = $6
	`

	result, err := s.db.ExecContext(ctx, query, now, updatedBy, itemID, importID, guildID, requesterUserID)
	if err != nil {
		return fmt.Errorf("discard import item: %w", err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("discard import item rows affected: %w", err)
	}

	if affected == 0 {
		return ErrImportItemNotFound
	}

	return nil
}

func (s *postgresStorage) FindDuplicateKeys(ctx context.Context, guildID, requesterUserID int64, keys []ExistingKey) (map[string]struct{}, error) {
	if len(keys) == 0 {
		return map[string]struct{}{}, nil
	}

	const query = `
		SELECT occurred_at, description, amount
		FROM transaction
		WHERE guild_id = $1
		  AND user_account_id = $2
		  AND status = 'ACTIVE'
	`

	rows, err := s.db.QueryxContext(ctx, query, guildID, requesterUserID)
	if err != nil {
		return nil, fmt.Errorf("query duplicate candidates: %w", err)
	}
	defer func() { _ = rows.Close() }()

	type candidate struct {
		OccurredAt  time.Time       `db:"occurred_at"`
		Description string          `db:"description"`
		Amount      decimal.Decimal `db:"amount"`
	}

	keyMap := make(map[string]struct{})
	for rows.Next() {
		var c candidate
		if err := rows.StructScan(&c); err != nil {
			return nil, fmt.Errorf("scan duplicate candidate: %w", err)
		}

		for _, current := range keys {
			if current.OccurredAt.UTC().Format(time.DateOnly) != c.OccurredAt.UTC().Format(time.DateOnly) {
				continue
			}
			if !strings.EqualFold(strings.TrimSpace(current.Description), strings.TrimSpace(c.Description)) {
				continue
			}
			if !current.Amount.Equal(c.Amount) {
				continue
			}

			keyMap[duplicateKey(current.OccurredAt, current.Description, current.Amount, current.CardLast4)] = struct{}{}
		}
	}

	return keyMap, nil
}

func (s *postgresStorage) ConfirmBatch(ctx context.Context, batch *ImportBatch, itemIDs []int64, actor audit.Actor) error {
	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx confirm batch: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	if len(itemIDs) > 0 {
		items := make([]importItemEntity, 0)
		query, args, err := sqlx.In(`
			SELECT id, import_batch_id, user_account_id, occurred_at, description, amount, type, category, bank_category,
				card_last4, installment_current, installment_total, status, transaction_id, created_at, created_by, updated_at, updated_by
			FROM import_item
			WHERE import_batch_id = ? AND id IN (?)
		`, batch.ID, itemIDs)
		if err != nil {
			return fmt.Errorf("build list confirm items query: %w", err)
		}

		query = tx.Rebind(query)
		if err := tx.SelectContext(ctx, &items, query, args...); err != nil {
			return fmt.Errorf("list confirm items: %w", err)
		}

		for _, item := range items {
			txEntity, newErr := transaction.New(
				transaction.Type(item.Type),
				item.Description,
				item.Amount,
				transaction.Category(item.Category),
				transaction.VisibilityPublic,
				item.OccurredAt,
				actor.UserID,
				batch.GuildID,
				actor.Email,
			)
			if newErr != nil {
				return fmt.Errorf("create transaction from import item: %w", newErr)
			}
			txEntity.Source = transaction.SourceImport

			const createTransactionQuery = `
				INSERT INTO transaction (
					type, description, amount, category, status, source, visibility,
					occurred_at, user_account_id, guild_id, created_at, created_by, updated_at, updated_by
				) VALUES (
					$1, $2, $3, $4, $5, $6, $7,
					$8, $9, $10, $11, $12, $13, $14
				)
				RETURNING id
			`

			if err := tx.QueryRowxContext(ctx, createTransactionQuery,
				string(txEntity.Type),
				txEntity.Description,
				txEntity.Amount,
				string(txEntity.Category),
				string(txEntity.Status),
				string(txEntity.Source),
				string(txEntity.Visibility),
				txEntity.OccurredAt,
				txEntity.UserAccountID,
				txEntity.GuildID,
				txEntity.CreatedAt,
				txEntity.CreatedBy,
				txEntity.UpdatedAt,
				txEntity.UpdatedBy,
			).Scan(&txEntity.ID); err != nil {
				return fmt.Errorf("insert transaction from import item: %w", err)
			}

			const updateItemQuery = `
				UPDATE import_item
				SET transaction_id = $1, updated_at = $2, updated_by = $3
				WHERE id = $4
			`

			now := time.Now()
			if _, err := tx.ExecContext(ctx, updateItemQuery, txEntity.ID, now, actor.Email, item.ID); err != nil {
				return fmt.Errorf("link import item to transaction: %w", err)
			}
		}
	}

	now := time.Now()
	batch.Status = BatchStatusCompleted
	batch.UpdatedAt = &now
	batch.UpdatedBy = &actor.Email

	const updateBatchQuery = `
		UPDATE import_batch
		SET status = $1, updated_at = $2, updated_by = $3
		WHERE id = $4
	`

	if _, err := tx.ExecContext(ctx, updateBatchQuery, string(batch.Status), batch.UpdatedAt, batch.UpdatedBy, batch.ID); err != nil {
		return fmt.Errorf("complete import batch: %w", err)
	}

	if len(itemIDs) > 0 {
		query, args, err := sqlx.In(`
			UPDATE import_item
			SET status = 'DISCARDED', updated_at = ?, updated_by = ?
			WHERE import_batch_id = ?
			  AND id NOT IN (?)
			  AND status = 'READY'
		`, now, actor.Email, batch.ID, itemIDs)
		if err != nil {
			return fmt.Errorf("build discard remaining items query: %w", err)
		}

		query = tx.Rebind(query)
		if _, err := tx.ExecContext(ctx, query, args...); err != nil {
			return fmt.Errorf("discard remaining items: %w", err)
		}
	} else {
		const discardAllReadyQuery = `
			UPDATE import_item
			SET status = 'DISCARDED', updated_at = $1, updated_by = $2
			WHERE import_batch_id = $3 AND status = 'READY'
		`
		if _, err := tx.ExecContext(ctx, discardAllReadyQuery, now, actor.Email, batch.ID); err != nil {
			return fmt.Errorf("discard all ready items: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit confirm batch: %w", err)
	}

	return nil
}

func toDomainBatch(entity importBatchEntity) *ImportBatch {
	return &ImportBatch{
		ID:            entity.ID,
		GuildID:       entity.GuildID,
		UserAccountID: entity.UserAccountID,
		FileName:      entity.FileName,
		SourceBank:    SourceBank(entity.SourceBank),
		Status:        BatchStatus(entity.Status),
		Entry: audit.Entry{
			CreatedAt: entity.CreatedAt,
			CreatedBy: entity.CreatedBy,
			UpdatedAt: entity.UpdatedAt,
			UpdatedBy: entity.UpdatedBy,
		},
	}
}

func toDomainItem(entity importItemEntity) *ImportItem {
	return &ImportItem{
		ID:                 entity.ID,
		ImportBatchID:      entity.ImportBatchID,
		UserAccountID:      entity.UserAccountID,
		OccurredAt:         entity.OccurredAt,
		Description:        entity.Description,
		Amount:             entity.Amount,
		Type:               transaction.Type(entity.Type),
		Category:           transaction.Category(entity.Category),
		BankCategory:       entity.BankCategory,
		CardLast4:          entity.CardLast4,
		InstallmentCurrent: entity.InstallmentCurrent,
		InstallmentTotal:   entity.InstallmentTotal,
		Status:             ItemStatus(entity.Status),
		TransactionID:      entity.TransactionID,
		Entry: audit.Entry{
			CreatedAt: entity.CreatedAt,
			CreatedBy: entity.CreatedBy,
			UpdatedAt: entity.UpdatedAt,
			UpdatedBy: entity.UpdatedBy,
		},
	}
}
