package fixedexpense

import "github.com/shopspring/decimal"

type createFixedExpenseRequest struct {
	Name     string          `json:"name"`
	Amount   decimal.Decimal `json:"amount"`
	DueDay   int             `json:"due_day"`
	Category Category        `json:"category"`
}

type updateFixedExpenseRequest struct {
	Amount   *decimal.Decimal `json:"amount"`
	DueDay   *int             `json:"due_day"`
	Category *Category        `json:"category"`
	Status   *Status          `json:"status"`
}

type deactivateFixedExpenseRequest struct {
	Status Status `json:"status"`
}

type fixedExpenseResponse struct {
	GuildID  int64           `json:"guild_id"`
	ID       int64           `json:"id"`
	Name     string          `json:"name"`
	Amount   decimal.Decimal `json:"amount"`
	DueDay   int             `json:"due_day"`
	Category Category        `json:"category"`
	Status   Status          `json:"status"`
}

func toResponse(fixedExpense *FixedExpense) fixedExpenseResponse {
	return fixedExpenseResponse{
		ID:       fixedExpense.ID,
		GuildID:  fixedExpense.GuildID,
		Name:     fixedExpense.Name,
		Amount:   fixedExpense.Amount,
		DueDay:   fixedExpense.DueDay,
		Category: fixedExpense.Category,
		Status:   fixedExpense.Status,
	}
}
