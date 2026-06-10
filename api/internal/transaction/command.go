package transaction

import (
	"time"

	"github.com/shopspring/decimal"
)

type CreateInput struct {
	Type        Type
	Description string
	Amount      decimal.Decimal
	Category    Category
	Visibility  Visibility
	OccurredAt  time.Time
	GuildID     int64
}

type UpdateInput struct {
	Type        *Type
	Description *string
	Amount      *decimal.Decimal
	Category    *Category
	OccurredAt  *time.Time
}

type ListInput struct {
	GuildID    int64
	Cursor     string
	Limit      int
	DateFrom   *time.Time
	DateTo     *time.Time
	Category   *Category
	Type       *Type
	Source     *Source
	Status     *Status
	Visibility *Visibility
}

type ListFilter struct {
	GuildID          int64
	RequesterUserID  int64
	CursorOccurredAt *time.Time
	CursorID         *int64
	Limit            int
	DateFrom         *time.Time
	DateTo           *time.Time
	Category         *Category
	Type             *Type
	Source           *Source
	Status           Status
	Visibility       *Visibility
}

type BulkCategorizeResult struct {
	Updated int           `json:"updated"`
	Failed  []BulkFailure `json:"failed"`
}

type BulkFailure struct {
	TransactionID int64  `json:"transactionId"`
	Reason        string `json:"reason"`
}
