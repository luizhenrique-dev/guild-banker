package fixedexpense

import (
	"encoding/json"
	"fmt"

	"github.com/shopspring/decimal"
)

type createFixedExpenseRequest struct {
	ResponsibleUserID *int64          `json:"responsible_user_id"`
	Name              string          `json:"name"`
	Amount            decimal.Decimal `json:"amount"`
	DueDay            int             `json:"due_day"`
	Category          Category        `json:"category"`
}

type updateFixedExpenseRequest struct {
	ResponsibleUserID optionalUserID   `json:"responsible_user_id"`
	Amount            *decimal.Decimal `json:"amount"`
	DueDay            *int             `json:"due_day"`
	Category          *Category        `json:"category"`
	Status            *Status          `json:"status"`
}

type deactivateFixedExpenseRequest struct {
	Status Status `json:"status"`
}

type fixedExpenseResponse struct {
	ResponsibleUserID *int64          `json:"responsible_user_id"`
	GuildID           int64           `json:"guild_id"`
	ID                int64           `json:"id"`
	Name              string          `json:"name"`
	Amount            decimal.Decimal `json:"amount"`
	DueDay            int             `json:"due_day"`
	Category          Category        `json:"category"`
	Status            Status          `json:"status"`
}

func toResponse(fixedExpense *FixedExpense) fixedExpenseResponse {
	return fixedExpenseResponse{
		ResponsibleUserID: fixedExpense.ResponsibleUserID,
		ID:                fixedExpense.ID,
		GuildID:           fixedExpense.GuildID,
		Name:              fixedExpense.Name,
		Amount:            fixedExpense.Amount,
		DueDay:            fixedExpense.DueDay,
		Category:          fixedExpense.Category,
		Status:            fixedExpense.Status,
	}
}

// optionalUserID distinguishes an omitted PATCH field from an explicit null.
type optionalUserID struct {
	value   *int64
	present bool
}

func (id *optionalUserID) UnmarshalJSON(data []byte) error {
	if err := json.Unmarshal(data, &id.value); err != nil {
		return fmt.Errorf("decode responsible user id: %w", err)
	}
	id.present = true
	return nil
}
