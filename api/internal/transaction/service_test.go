package transaction

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"

	"github.com/luizhenrique-dev/guild-banker/api/internal/audit"
)

type storageMock struct {
	isGuildMemberFn   func(ctx context.Context, guildID, userID int64) (bool, error)
	createFn          func(ctx context.Context, transaction *Transaction) error
	listFn            func(ctx context.Context, filter ListFilter) ([]*Transaction, error)
	getByIDForWriteFn func(ctx context.Context, id, guildID, requesterUserID int64) (*Transaction, error)
	getByIDForOwnerFn func(ctx context.Context, id, guildID, requesterUserID int64) (*Transaction, error)
	updateFn          func(ctx context.Context, transaction *Transaction) error
	bulkCategorizeFn  func(ctx context.Context, guildID, requesterUserID int64, transactionIDs []int64, category Category) (int, []BulkFailure, error)
}

func (m *storageMock) IsGuildMember(ctx context.Context, guildID, userID int64) (bool, error) {
	return m.isGuildMemberFn(ctx, guildID, userID)
}
func (m *storageMock) Create(ctx context.Context, transaction *Transaction) error {
	return m.createFn(ctx, transaction)
}
func (m *storageMock) List(ctx context.Context, filter ListFilter) ([]*Transaction, error) {
	return m.listFn(ctx, filter)
}
func (m *storageMock) GetByIDForWrite(ctx context.Context, id, guildID, requesterUserID int64) (*Transaction, error) {
	return m.getByIDForWriteFn(ctx, id, guildID, requesterUserID)
}
func (m *storageMock) GetByIDForOwner(ctx context.Context, id, guildID, requesterUserID int64) (*Transaction, error) {
	return m.getByIDForOwnerFn(ctx, id, guildID, requesterUserID)
}
func (m *storageMock) Update(ctx context.Context, transaction *Transaction) error {
	return m.updateFn(ctx, transaction)
}
func (m *storageMock) BulkCategorize(ctx context.Context, guildID, requesterUserID int64, transactionIDs []int64, category Category) (int, []BulkFailure, error) {
	return m.bulkCategorizeFn(ctx, guildID, requesterUserID, transactionIDs, category)
}

func TestService_Create(t *testing.T) {
	service := NewService(&storageMock{
		isGuildMemberFn: func(ctx context.Context, guildID, userID int64) (bool, error) { return true, nil },
		createFn: func(ctx context.Context, transaction *Transaction) error {
			transaction.ID = 1
			return nil
		},
	})

	transaction, err := service.Create(context.Background(), CreateInput{
		Type:        TypeExpense,
		Description: "Mercado",
		Amount:      decimal.NewFromInt(100),
		Category:    CategoryPersonal,
		OccurredAt:  time.Now(),
		GuildID:     1,
	}, audit.Actor{UserID: 10, Email: "test@example.com"})

	require.NoError(t, err)
	require.Equal(t, int64(1), transaction.ID)
	require.Equal(t, VisibilityPublic, transaction.Visibility)
}

func TestService_List_DefaultStatusActive(t *testing.T) {
	service := NewService(&storageMock{
		isGuildMemberFn: func(ctx context.Context, guildID, userID int64) (bool, error) { return true, nil },
		listFn: func(ctx context.Context, filter ListFilter) ([]*Transaction, error) {
			require.Equal(t, StatusActive, filter.Status)
			return []*Transaction{}, nil
		},
	})

	_, _, err := service.List(context.Background(), ListInput{GuildID: 1}, audit.Actor{UserID: 10, Email: "test@example.com"})
	require.NoError(t, err)
}

func TestService_SetVisibility_Cancelled(t *testing.T) {
	service := NewService(&storageMock{
		isGuildMemberFn: func(ctx context.Context, guildID, userID int64) (bool, error) { return true, nil },
		getByIDForOwnerFn: func(ctx context.Context, id, guildID, requesterUserID int64) (*Transaction, error) {
			return &Transaction{ID: id, Status: StatusCancelled}, nil
		},
	})

	_, err := service.SetVisibility(context.Background(), 1, 2, VisibilityPrivate, audit.Actor{UserID: 10, Email: "test@example.com"})
	require.ErrorIs(t, err, ErrTransactionCancelled)
}

func TestService_BulkCategorize_BatchError(t *testing.T) {
	service := NewService(&storageMock{
		isGuildMemberFn: func(ctx context.Context, guildID, userID int64) (bool, error) { return true, nil },
		bulkCategorizeFn: func(ctx context.Context, guildID, requesterUserID int64, transactionIDs []int64, category Category) (int, []BulkFailure, error) {
			return 0, nil, errors.New("db down")
		},
	})

	result, err := service.BulkCategorize(
		context.Background(),
		1,
		[]int64{1, 2, 3},
		CategoryHealth,
		audit.Actor{UserID: 10, Email: "test@example.com"},
	)

	require.NoError(t, err)
	require.Len(t, result.Failed, 3)
}
