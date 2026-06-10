package transaction

import (
	"time"

	"github.com/shopspring/decimal"
)

type createTransactionRequest struct {
	Type        Type            `json:"type"`
	Description string          `json:"description"`
	Amount      decimal.Decimal `json:"amount"`
	Category    Category        `json:"category"`
	Visibility  Visibility      `json:"visibility"`
	OccurredAt  time.Time       `json:"occurredAt"`
}

type updateTransactionRequest struct {
	Type        *Type            `json:"type"`
	Description *string          `json:"description"`
	Amount      *decimal.Decimal `json:"amount"`
	Category    *Category        `json:"category"`
	OccurredAt  *time.Time       `json:"occurredAt"`
}

type setVisibilityRequest struct {
	Visibility Visibility `json:"visibility"`
}

type bulkCategorizeRequest struct {
	TransactionIDs []int64  `json:"transactionIds"`
	Category       Category `json:"category"`
}

type listTransactionsResponse struct {
	Items      []transactionResponse `json:"items"`
	NextCursor string                `json:"nextCursor,omitempty"`
}

type transactionResponse struct {
	ID            int64           `json:"id"`
	Type          Type            `json:"type"`
	Description   string          `json:"description"`
	Amount        decimal.Decimal `json:"amount"`
	Category      Category        `json:"category"`
	Status        Status          `json:"status"`
	Source        Source          `json:"source"`
	Visibility    Visibility      `json:"visibility"`
	OccurredAt    time.Time       `json:"occurredAt"`
	GuildID       int64           `json:"guildId"`
	UserAccountID int64           `json:"userAccountId"`
	CreatedAt     time.Time       `json:"createdAt"`
	UpdatedAt     *time.Time      `json:"updatedAt,omitempty"`
}

type setVisibilityResponse struct {
	ID         int64      `json:"id"`
	Visibility Visibility `json:"visibility"`
	UpdatedAt  *time.Time `json:"updatedAt,omitempty"`
}

func toResponse(transaction *Transaction) transactionResponse {
	return transactionResponse{
		ID:            transaction.ID,
		Type:          transaction.Type,
		Description:   transaction.Description,
		Amount:        transaction.Amount,
		Category:      transaction.Category,
		Status:        transaction.Status,
		Source:        transaction.Source,
		Visibility:    transaction.Visibility,
		OccurredAt:    transaction.OccurredAt,
		GuildID:       transaction.GuildID,
		UserAccountID: transaction.UserAccountID,
		CreatedAt:     transaction.CreatedAt,
		UpdatedAt:     transaction.UpdatedAt,
	}
}
