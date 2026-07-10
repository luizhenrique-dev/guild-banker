package importer

import (
	"time"

	"github.com/shopspring/decimal"

	"github.com/luizhenrique-dev/guild-banker/api/internal/transaction"
)

type uploadResponse struct {
	ImportID int64          `json:"importId"`
	Status   BatchStatus    `json:"status"`
	FileName string         `json:"fileName"`
	Summary  Summary        `json:"summary"`
	Items    []itemResponse `json:"items"`
}

type itemResponse struct {
	ItemID       int64                `json:"itemId"`
	OccurredAt   time.Time            `json:"occurredAt"`
	Description  string               `json:"description"`
	Amount       decimal.Decimal      `json:"amount"`
	Type         transaction.Type     `json:"type"`
	Category     transaction.Category `json:"category"`
	BankCategory string               `json:"bankCategory"`
	CardLast4    string               `json:"cardLast4"`
	Installment  string               `json:"installment"`
	Status       ItemStatus           `json:"status"`
}

type updateItemRequest struct {
	Description *string               `json:"description"`
	OccurredAt  *time.Time            `json:"occurredAt"`
	Amount      *decimal.Decimal      `json:"amount"`
	Type        *transaction.Type     `json:"type"`
	Category    *transaction.Category `json:"category"`
}

type confirmResponse struct {
	ImportID int64       `json:"importId"`
	Created  int         `json:"created"`
	Skipped  int         `json:"skipped"`
	Status   BatchStatus `json:"status"`
}

type getBatchResponse struct {
	ImportID int64          `json:"importId"`
	Status   BatchStatus    `json:"status"`
	FileName string         `json:"fileName"`
	Items    []itemResponse `json:"items"`
}

func toItemResponse(item *ImportItem) itemResponse {
	installment := "Única"
	if item.InstallmentCurrent != nil && item.InstallmentTotal != nil {
		installment = decimal.NewFromInt(int64(*item.InstallmentCurrent)).String() + "/" + decimal.NewFromInt(int64(*item.InstallmentTotal)).String()
	}

	return itemResponse{
		ItemID:       item.ID,
		OccurredAt:   item.OccurredAt,
		Description:  item.Description,
		Amount:       item.Amount,
		Type:         item.Type,
		Category:     item.Category,
		BankCategory: item.BankCategory,
		CardLast4:    item.CardLast4,
		Installment:  installment,
		Status:       item.Status,
	}
}
