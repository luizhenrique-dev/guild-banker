package importer

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/shopspring/decimal"

	"github.com/luizhenrique-dev/guild-banker/api/internal/audit"
)

type Storage interface {
	IsGuildMember(ctx context.Context, guildID, userID int64) (bool, error)
	CreateBatch(ctx context.Context, batch *ImportBatch, items []*ImportItem) error
	GetBatch(ctx context.Context, guildID, importID, requesterUserID int64) (*ImportBatch, []*ImportItem, error)
	GetItemForUpdate(ctx context.Context, guildID, importID, itemID, requesterUserID int64) (*ImportItem, error)
	UpdateItem(ctx context.Context, item *ImportItem) error
	DiscardItem(ctx context.Context, guildID, importID, itemID, requesterUserID int64, updatedBy string) error
	FindDuplicateKeys(ctx context.Context, guildID, requesterUserID int64, keys []ExistingKey) (map[string]struct{}, error)
	ConfirmBatch(ctx context.Context, batch *ImportBatch, itemIDs []int64, actor audit.Actor) error
}

type Service struct {
	storage Storage
	parser  Parser
}

func NewService(storage Storage, parser Parser) *Service {
	return &Service{storage: storage, parser: parser}
}

func (s *Service) Upload(ctx context.Context, input UploadInput, actor audit.Actor) (*ImportBatch, []*ImportItem, Summary, error) {
	if err := s.authorizeGuildMember(ctx, input.GuildID, actor); err != nil {
		return nil, nil, Summary{}, err
	}

	rows, err := s.parser.Parse(input.File)
	if err != nil {
		return nil, nil, Summary{}, fmt.Errorf("parse csv: %w", err)
	}

	batch, err := NewBatch(input.GuildID, actor.UserID, input.FileName, input.SourceBank, actor.Email)
	if err != nil {
		return nil, nil, Summary{}, fmt.Errorf("create batch: %w", err)
	}

	now := time.Now()
	items := make([]*ImportItem, 0, len(rows))
	summary := Summary{Parsed: len(rows)}
	for _, row := range rows {
		item, skippedZero, mapErr := mapParsedRowToItem(1, actor.UserID, row, actor.Email, now)
		if mapErr != nil {
			return nil, nil, Summary{}, fmt.Errorf("map row: %w", mapErr)
		}

		if skippedZero {
			summary.SkippedZero++
			continue
		}

		items = append(items, item)
	}

	summary.Candidates = len(items)
	s.markDuplicates(ctx, input.GuildID, actor.UserID, items, &summary)

	if err := s.storage.CreateBatch(ctx, batch, items); err != nil {
		return nil, nil, Summary{}, fmt.Errorf("create batch with items: %w", err)
	}

	return batch, items, summary, nil
}

func (s *Service) GetByID(ctx context.Context, guildID, importID int64, actor audit.Actor) (*ImportBatch, []*ImportItem, error) {
	if err := s.authorizeGuildMember(ctx, guildID, actor); err != nil {
		return nil, nil, err
	}

	batch, items, err := s.storage.GetBatch(ctx, guildID, importID, actor.UserID)
	if err != nil {
		return nil, nil, fmt.Errorf("get import batch: %w", err)
	}

	return batch, items, nil
}

func (s *Service) UpdateItem(ctx context.Context, input UpdateItemInput, actor audit.Actor) (*ImportItem, error) {
	if err := s.authorizeGuildMember(ctx, input.GuildID, actor); err != nil {
		return nil, err
	}

	item, err := s.storage.GetItemForUpdate(ctx, input.GuildID, input.ImportID, input.ItemID, actor.UserID)
	if err != nil {
		return nil, fmt.Errorf("get import item for update: %w", err)
	}

	if input.Description != nil {
		item.Description = strings.TrimSpace(*input.Description)
	}
	if input.OccurredAt != nil {
		item.OccurredAt = *input.OccurredAt
	}
	if input.Amount != nil {
		item.Amount = input.Amount.Abs()
	}
	if input.Type != nil {
		item.Type = *input.Type
	}
	if input.Category != nil {
		item.Category = *input.Category
	}

	item.Status = ItemStatusReady
	item.Update(actor.Email)

	if err := item.Validate(); err != nil {
		return nil, fmt.Errorf("validate item update: %w", err)
	}

	if err := s.storage.UpdateItem(ctx, item); err != nil {
		return nil, fmt.Errorf("update import item: %w", err)
	}

	return item, nil
}

func (s *Service) DiscardItem(ctx context.Context, guildID, importID, itemID int64, actor audit.Actor) error {
	if err := s.authorizeGuildMember(ctx, guildID, actor); err != nil {
		return err
	}

	if err := s.storage.DiscardItem(ctx, guildID, importID, itemID, actor.UserID, actor.Email); err != nil {
		return fmt.Errorf("discard import item: %w", err)
	}

	return nil
}

func (s *Service) Confirm(ctx context.Context, input ConfirmInput, actor audit.Actor) (int, int, error) {
	if err := s.authorizeGuildMember(ctx, input.GuildID, actor); err != nil {
		return 0, 0, err
	}

	batch, items, err := s.storage.GetBatch(ctx, input.GuildID, input.ImportID, actor.UserID)
	if err != nil {
		return 0, 0, fmt.Errorf("get import batch for confirmation: %w", err)
	}

	if batch.Status != BatchStatusPendingReview {
		return 0, 0, ErrInvalidImportStatus
	}

	itemIDs := make([]int64, 0)
	created := 0
	skipped := 0
	for _, item := range items {
		if item.Status != ItemStatusReady {
			skipped++
			continue
		}

		itemIDs = append(itemIDs, item.ID)
		created++
	}

	if err := s.storage.ConfirmBatch(ctx, batch, itemIDs, actor); err != nil {
		return 0, 0, fmt.Errorf("confirm import batch: %w", err)
	}

	return created, skipped, nil
}

func (s *Service) markDuplicates(ctx context.Context, guildID, requesterUserID int64, items []*ImportItem, summary *Summary) {
	seen := make(map[string]struct{}, len(items))
	keys := make([]ExistingKey, 0, len(items))
	for _, item := range items {
		key := duplicateKey(item.OccurredAt, item.Description, item.Amount, item.CardLast4)
		if _, exists := seen[key]; exists {
			item.Status = ItemStatusDuplicate
			summary.Duplicates++
			continue
		}

		seen[key] = struct{}{}
		keys = append(keys, ExistingKey{OccurredAt: item.OccurredAt, Description: item.Description, Amount: item.Amount, CardLast4: item.CardLast4})
	}

	existing, err := s.storage.FindDuplicateKeys(ctx, guildID, requesterUserID, keys)
	if err != nil {
		return
	}

	for _, item := range items {
		if item.Status == ItemStatusDuplicate {
			continue
		}

		key := duplicateKey(item.OccurredAt, item.Description, item.Amount, item.CardLast4)
		if _, exists := existing[key]; exists {
			item.Status = ItemStatusDuplicate
			summary.Duplicates++
		}
	}
}

func (s *Service) authorizeGuildMember(ctx context.Context, guildID int64, actor audit.Actor) error {
	isMember, err := s.storage.IsGuildMember(ctx, guildID, actor.UserID)
	if err != nil {
		return fmt.Errorf("check guild member: %w", err)
	}

	if !isMember {
		return ErrRequesterIsNotMember
	}

	return nil
}

func duplicateKey(occurredAt time.Time, description string, amount decimal.Decimal, cardLast4 string) string {
	return strings.ToLower(strings.TrimSpace(description)) + "|" + occurredAt.UTC().Format(time.DateOnly) + "|" + amount.StringFixed(2) + "|" + strings.TrimSpace(cardLast4)
}
