package importer

import (
	"io"
	"time"

	"github.com/shopspring/decimal"

	"github.com/luizhenrique-dev/guild-banker/api/internal/transaction"
)

type UploadInput struct {
	GuildID    int64
	FileName   string
	File       io.Reader
	SourceBank SourceBank
}

type UpdateItemInput struct {
	GuildID     int64
	ImportID    int64
	ItemID      int64
	Description *string
	OccurredAt  *time.Time
	Amount      *decimal.Decimal
	Type        *transaction.Type
	Category    *transaction.Category
}

type ConfirmInput struct {
	GuildID  int64
	ImportID int64
}

type ExistingKey struct {
	OccurredAt  time.Time
	Description string
	Amount      decimal.Decimal
	CardLast4   string
}
