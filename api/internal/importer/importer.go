package importer

import (
	"fmt"
	"strings"
	"time"

	"github.com/shopspring/decimal"

	"github.com/luizhenrique-dev/guild-banker/api/internal/audit"
	"github.com/luizhenrique-dev/guild-banker/api/internal/transaction"
)

type SourceBank string

const (
	SourceBankC6 SourceBank = "C6"
)

type BatchStatus string

const (
	BatchStatusPendingReview BatchStatus = "PENDING_REVIEW"
	BatchStatusCompleted     BatchStatus = "COMPLETED"
	BatchStatusCancelled     BatchStatus = "CANCELLED"
)

type ItemStatus string

const (
	ItemStatusReady     ItemStatus = "READY"
	ItemStatusDuplicate ItemStatus = "DUPLICATE"
	ItemStatusDiscarded ItemStatus = "DISCARDED"
)

type ImportBatch struct {
	ID            int64
	GuildID       int64
	UserAccountID int64
	FileName      string
	SourceBank    SourceBank
	Status        BatchStatus
	audit.Entry
}

type ImportItem struct {
	ID                 int64
	ImportBatchID      int64
	UserAccountID      int64
	OccurredAt         time.Time
	Description        string
	Amount             decimal.Decimal
	Type               transaction.Type
	Category           transaction.Category
	BankCategory       string
	CardLast4          string
	InstallmentCurrent *int
	InstallmentTotal   *int
	Status             ItemStatus
	TransactionID      *int64
	audit.Entry
}

type Summary struct {
	Parsed      int `json:"parsed"`
	Candidates  int `json:"candidates"`
	Duplicates  int `json:"duplicates"`
	SkippedZero int `json:"skippedZero"`
}

func (s SourceBank) IsValid() bool {
	return s == SourceBankC6
}

func (s BatchStatus) IsValid() bool {
	switch s {
	case BatchStatusPendingReview, BatchStatusCompleted, BatchStatusCancelled:
		return true
	default:
		return false
	}
}

func (s ItemStatus) IsValid() bool {
	switch s {
	case ItemStatusReady, ItemStatusDuplicate, ItemStatusDiscarded:
		return true
	default:
		return false
	}
}

func NewBatch(guildID, userAccountID int64, fileName string, sourceBank SourceBank, createdBy string) (*ImportBatch, error) {
	batch := &ImportBatch{
		GuildID:       guildID,
		UserAccountID: userAccountID,
		FileName:      strings.TrimSpace(fileName),
		SourceBank:    sourceBank,
		Status:        BatchStatusPendingReview,
		Entry: audit.Entry{
			CreatedAt: time.Now(),
			CreatedBy: createdBy,
		},
	}

	if err := batch.Validate(); err != nil {
		return nil, err
	}

	return batch, nil
}

func (b *ImportBatch) Validate() error {
	if b.GuildID <= 0 {
		return fmt.Errorf("guild id must be greater than zero")
	}

	if b.UserAccountID <= 0 {
		return fmt.Errorf("user account id must be greater than zero")
	}

	if b.FileName == "" {
		return fmt.Errorf("file name is required")
	}

	if !b.SourceBank.IsValid() {
		return fmt.Errorf("invalid source bank")
	}

	if !b.Status.IsValid() {
		return ErrInvalidImportStatus
	}

	return nil
}

func (i *ImportItem) Validate() error {
	if i.ImportBatchID <= 0 {
		return fmt.Errorf("import batch id must be greater than zero")
	}

	if i.UserAccountID <= 0 {
		return fmt.Errorf("user account id must be greater than zero")
	}

	if i.Description == "" {
		return fmt.Errorf("description is required")
	}

	if i.Amount.LessThanOrEqual(decimal.Zero) {
		return fmt.Errorf("amount must be greater than zero")
	}

	if !i.Type.IsValid() {
		return fmt.Errorf("invalid transaction type")
	}

	if !i.Category.IsValid() {
		return fmt.Errorf("invalid transaction category")
	}

	if !i.Status.IsValid() {
		return fmt.Errorf("invalid item status")
	}

	return nil
}
