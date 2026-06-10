package transaction

import (
	"context"
	"errors"
	"fmt"

	"github.com/luizhenrique-dev/guild-banker/api/internal/audit"
	"github.com/luizhenrique-dev/guild-banker/api/pkg/cursor"
)

const (
	defaultListLimit = 100
	maxBatchSize     = 250
)

type Storage interface {
	IsGuildMember(ctx context.Context, guildID, userID int64) (bool, error)
	Create(ctx context.Context, transaction *Transaction) error
	List(ctx context.Context, filter ListFilter) ([]*Transaction, error)
	GetByIDForWrite(ctx context.Context, id, guildID, requesterUserID int64) (*Transaction, error)
	GetByIDForOwner(ctx context.Context, id, guildID, requesterUserID int64) (*Transaction, error)
	Update(ctx context.Context, transaction *Transaction) error
	BulkCategorize(ctx context.Context, guildID, requesterUserID int64, transactionIDs []int64, category Category) (int, []BulkFailure, error)
}

type Service struct {
	storage Storage
}

func NewService(storage Storage) *Service {
	return &Service{storage: storage}
}

func (s *Service) Create(ctx context.Context, input CreateInput, actor audit.Actor) (*Transaction, error) {
	if err := s.authorizeGuildMember(ctx, "create transaction", input.GuildID, actor); err != nil {
		return nil, err
	}

	visibility := input.Visibility
	if visibility == "" {
		visibility = VisibilityPublic
	}

	transaction, err := New(
		input.Type,
		input.Description,
		input.Amount,
		input.Category,
		visibility,
		input.OccurredAt,
		actor.UserID,
		input.GuildID,
		actor.Email,
	)
	if err != nil {
		return nil, fmt.Errorf("create transaction: %w", err)
	}

	if err := s.storage.Create(ctx, transaction); err != nil {
		return nil, fmt.Errorf("create transaction: %w", err)
	}

	return transaction, nil
}

func (s *Service) List(ctx context.Context, input ListInput, actor audit.Actor) ([]*Transaction, string, error) {
	if err := s.authorizeGuildMember(ctx, "list transactions", input.GuildID, actor); err != nil {
		return nil, "", err
	}

	limit := input.Limit
	if limit <= 0 {
		limit = defaultListLimit
	}

	filter := ListFilter{
		GuildID:         input.GuildID,
		RequesterUserID: actor.UserID,
		Limit:           limit,
		DateFrom:        input.DateFrom,
		DateTo:          input.DateTo,
		Category:        input.Category,
		Type:            input.Type,
		Source:          input.Source,
		Visibility:      input.Visibility,
	}
	if input.Status != nil {
		filter.Status = *input.Status
	} else {
		filter.Status = StatusActive
	}

	if input.Cursor != "" {
		occurredAt, id, err := cursor.Decode(input.Cursor)
		if err != nil {
			return nil, "", fmt.Errorf("list transactions: %w", err)
		}
		filter.CursorOccurredAt = &occurredAt
		filter.CursorID = &id
	}

	transactions, err := s.storage.List(ctx, filter)
	if err != nil {
		return nil, "", fmt.Errorf("list transactions: %w", err)
	}

	nextCursor := ""
	if len(transactions) == limit {
		last := transactions[len(transactions)-1]
		nextCursor, err = cursor.Encode(last.OccurredAt, last.ID)
		if err != nil {
			return nil, "", fmt.Errorf("list transactions: encode cursor: %w", err)
		}
	}

	return transactions, nextCursor, nil
}

func (s *Service) Update(ctx context.Context, guildID, transactionID int64, input UpdateInput, actor audit.Actor) (*Transaction, error) {
	if err := s.authorizeGuildMember(ctx, "update transaction", guildID, actor); err != nil {
		return nil, err
	}
	if transactionID == 0 {
		return nil, errors.New("update transaction: transactionID is required")
	}

	transaction, err := s.storage.GetByIDForWrite(ctx, transactionID, guildID, actor.UserID)
	if err != nil {
		return nil, fmt.Errorf("update transaction: get transaction: %w", err)
	}

	typeValue := transaction.Type
	if input.Type != nil {
		typeValue = *input.Type
	}
	description := transaction.Description
	if input.Description != nil {
		description = *input.Description
	}
	amount := transaction.Amount
	if input.Amount != nil {
		amount = *input.Amount
	}
	category := transaction.Category
	if input.Category != nil {
		category = *input.Category
	}
	occurredAt := transaction.OccurredAt
	if input.OccurredAt != nil {
		occurredAt = *input.OccurredAt
	}

	if err := transaction.ApplyUpdate(typeValue, description, amount, category, occurredAt, actor.Email); err != nil {
		return nil, fmt.Errorf("update transaction: %w", err)
	}

	if err := s.storage.Update(ctx, transaction); err != nil {
		return nil, fmt.Errorf("update transaction: %w", err)
	}

	return transaction, nil
}

func (s *Service) Delete(ctx context.Context, guildID, transactionID int64, actor audit.Actor) error {
	if err := s.authorizeGuildMember(ctx, "delete transaction", guildID, actor); err != nil {
		return err
	}
	if transactionID == 0 {
		return errors.New("delete transaction: transactionID is required")
	}

	transaction, err := s.storage.GetByIDForWrite(ctx, transactionID, guildID, actor.UserID)
	if err != nil {
		return fmt.Errorf("delete transaction: get transaction: %w", err)
	}

	if err := transaction.Cancel(actor.Email); err != nil {
		return fmt.Errorf("delete transaction: %w", err)
	}

	if err := s.storage.Update(ctx, transaction); err != nil {
		return fmt.Errorf("delete transaction: %w", err)
	}

	return nil
}

func (s *Service) BulkCategorize(
	ctx context.Context,
	guildID int64,
	transactionIDs []int64,
	category Category,
	actor audit.Actor,
) (*BulkCategorizeResult, error) {
	if err := s.authorizeGuildMember(ctx, "bulk categorize transactions", guildID, actor); err != nil {
		return nil, err
	}
	if !category.IsValid() {
		return nil, errors.New("bulk categorize transactions: invalid category")
	}

	result := &BulkCategorizeResult{Failed: make([]BulkFailure, 0)}
	if len(transactionIDs) == 0 {
		return result, nil
	}

	for i := 0; i < len(transactionIDs); i += maxBatchSize {
		end := min(i+maxBatchSize, len(transactionIDs))
		updated, failed, batchErr := s.storage.BulkCategorize(ctx, guildID, actor.UserID, transactionIDs[i:end], category)
		result.Updated += updated
		result.Failed = append(result.Failed, failed...)
		if batchErr != nil {
			for _, id := range transactionIDs[i:end] {
				result.Failed = append(result.Failed, BulkFailure{TransactionID: id, Reason: "storage_error"})
			}
		}
	}

	return result, nil
}

func (s *Service) SetVisibility(
	ctx context.Context,
	guildID, transactionID int64,
	visibility Visibility,
	actor audit.Actor,
) (*Transaction, error) {
	if err := s.authorizeGuildMember(ctx, "set transaction visibility", guildID, actor); err != nil {
		return nil, err
	}
	if transactionID == 0 {
		return nil, errors.New("set transaction visibility: transactionID is required")
	}

	transaction, err := s.storage.GetByIDForOwner(ctx, transactionID, guildID, actor.UserID)
	if err != nil {
		return nil, fmt.Errorf("set transaction visibility: get transaction: %w", err)
	}

	if transaction.Status == StatusCancelled {
		return nil, ErrTransactionCancelled
	}

	if err := transaction.SetVisibility(visibility, actor.Email); err != nil {
		return nil, fmt.Errorf("set transaction visibility: %w", err)
	}

	if err := s.storage.Update(ctx, transaction); err != nil {
		return nil, fmt.Errorf("set transaction visibility: %w", err)
	}

	return transaction, nil
}

func (s *Service) authorizeGuildMember(ctx context.Context, op string, guildID int64, actor audit.Actor) error {
	if err := actor.Validate(); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	if guildID == 0 {
		return fmt.Errorf("%s: guildID is required", op)
	}

	isMember, err := s.storage.IsGuildMember(ctx, guildID, actor.UserID)
	if err != nil {
		return fmt.Errorf("%s: check requester membership: %w", op, err)
	}
	if !isMember {
		return ErrRequesterIsNotMember
	}
	return nil
}
