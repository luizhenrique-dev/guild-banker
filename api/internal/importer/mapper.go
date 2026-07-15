package importer

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/shopspring/decimal"

	"github.com/luizhenrique-dev/guild-banker/api/internal/audit"
	"github.com/luizhenrique-dev/guild-banker/api/internal/transaction"
)

func mapC6Category(value string) transaction.Category {
	category := strings.ToLower(strings.TrimSpace(value))

	switch {
	case strings.Contains(category, "supermercados"):
		return transaction.CategoryGrocery
	case strings.Contains(category, "restaurante"):
		return transaction.CategoryFoodAndDining
	case strings.Contains(category, "companhia aérea"), strings.Contains(category, "t&e"):
		return transaction.CategoryTravel
	case strings.Contains(category, "educacional"):
		return transaction.CategoryEducation
	case strings.Contains(category, "automotivo"):
		return transaction.CategoryTransportation
	case strings.Contains(category, "médica"), strings.Contains(category, "odontológica"):
		return transaction.CategoryHealth
	case strings.Contains(category, "materiais de construção"):
		return transaction.CategoryHousing
	case strings.Contains(category, "seguro"):
		return transaction.CategoryInsurance
	case strings.Contains(category, "serviços pessoais"):
		return transaction.CategoryPersonalCare
	case strings.Contains(category, "transporte"):
		return transaction.CategoryTransportation
	case strings.Contains(category, "elétrico"), strings.Contains(category, "varejo"):
		return transaction.CategoryShopping
	default:
		return transaction.CategoryOther
	}
}

func mapAmountAndType(value decimal.Decimal) (decimal.Decimal, transaction.Type, bool) {
	if value.Equal(decimal.Zero) {
		return decimal.Zero, "", true
	}

	if value.IsNegative() {
		return value.Abs(), transaction.TypeIncome, false
	}

	return value.Abs(), transaction.TypeExpense, false
}

func parseInstallment(installment string) (*int, *int, error) {
	value := strings.TrimSpace(installment)
	if value == "" || strings.EqualFold(value, "única") {
		return nil, nil, nil
	}

	currentStr, totalStr, found := strings.Cut(value, "/")
	if !found {
		return nil, nil, fmt.Errorf("invalid installment format")
	}

	current, err := strconv.Atoi(strings.TrimSpace(currentStr))
	if err != nil {
		return nil, nil, fmt.Errorf("parse installment current: %w", err)
	}

	total, err := strconv.Atoi(strings.TrimSpace(totalStr))
	if err != nil {
		return nil, nil, fmt.Errorf("parse installment total: %w", err)
	}

	if current <= 0 || total <= 0 || current > total {
		return nil, nil, fmt.Errorf("invalid installment values")
	}

	return &current, &total, nil
}

func mapParsedRowToItem(importBatchID int64, userAccountID int64, row ParsedRow, createdBy string, now time.Time) (*ImportItem, bool, error) {
	amount, typeValue, skippedZero := mapAmountAndType(row.Amount)
	if skippedZero {
		return nil, true, nil
	}

	installmentCurrent, installmentTotal, err := parseInstallment(row.Installment)
	if err != nil {
		return nil, false, err
	}

	item := &ImportItem{
		ImportBatchID:      importBatchID,
		UserAccountID:      userAccountID,
		OccurredAt:         row.OccurredAt,
		Description:        strings.TrimSpace(row.Description),
		Amount:             amount,
		Type:               typeValue,
		Category:           mapC6Category(row.BankCategory),
		BankCategory:       strings.TrimSpace(row.BankCategory),
		CardLast4:          strings.TrimSpace(row.CardLast4),
		InstallmentCurrent: installmentCurrent,
		InstallmentTotal:   installmentTotal,
		Status:             ItemStatusReady,
		Entry: audit.Entry{
			CreatedAt: now,
			CreatedBy: createdBy,
		},
	}

	if err := item.Validate(); err != nil {
		return nil, false, err
	}

	return item, false, nil
}
