package transaction

import (
	"errors"
	"time"

	"github.com/shopspring/decimal"

	"github.com/luizhenrique-dev/guild-banker/api/internal/audit"
)

type Type string

const (
	TypeExpense Type = "EXPENSE"
	TypeIncome  Type = "INCOME"
)

type Category string

const (
	CategoryHousing        Category = "HOUSING"
	CategorySubscriptions  Category = "SUBSCRIPTIONS"
	CategoryInsurance      Category = "INSURANCE"
	CategoryEducation      Category = "EDUCATION"
	CategoryTransportation Category = "TRANSPORTATION"
	CategoryHealth         Category = "HEALTH"
	CategoryPersonal       Category = "PERSONAL"
	CategoryTaxes          Category = "TAXES"
	CategoryOther          Category = "OTHER"
	CategoryFoodAndDining  Category = "FOOD_AND_DINING"
	CategoryEntertainment  Category = "ENTERTAINMENT"
	CategoryShopping       Category = "SHOPPING"
	CategoryPets           Category = "PETS"
	CategoryGames          Category = "GAMES"
	CategoryTravel         Category = "TRAVEL"
	CategoryInvestments    Category = "INVESTMENTS"
	CategoryDebtRepayment  Category = "DEBT_REPAYMENT"
	CategoryDonations      Category = "DONATIONS"
)

type Status string

const (
	StatusActive    Status = "ACTIVE"
	StatusCancelled Status = "CANCELLED"
)

type Source string

const (
	SourceManual Source = "MANUAL"
	SourceImport Source = "IMPORT"
)

type Visibility string

const (
	VisibilityPrivate Visibility = "PRIVATE"
	VisibilityPublic  Visibility = "PUBLIC"
)

type Transaction struct {
	ID            int64
	Type          Type
	Description   string
	Amount        decimal.Decimal
	Category      Category
	Status        Status
	Source        Source
	Visibility    Visibility
	OccurredAt    time.Time
	UserAccountID int64
	GuildID       int64
	audit.Entry
}

func New(
	typeValue Type,
	description string,
	amount decimal.Decimal,
	category Category,
	visibility Visibility,
	occurredAt time.Time,
	userAccountID int64,
	guildID int64,
	createdBy string,
) (*Transaction, error) {
	t := &Transaction{
		Type:          typeValue,
		Description:   description,
		Amount:        amount,
		Category:      category,
		Status:        StatusActive,
		Source:        SourceManual,
		Visibility:    visibility,
		OccurredAt:    occurredAt,
		UserAccountID: userAccountID,
		GuildID:       guildID,
		Entry: audit.Entry{
			CreatedAt: time.Now(),
			CreatedBy: createdBy,
		},
	}

	if err := t.Validate(); err != nil {
		return nil, err
	}

	return t, nil
}

func (t *Transaction) Validate() error {
	if !t.Type.IsValid() {
		return ErrInvalidTransactionType
	}
	if t.Description == "" {
		return errors.New("description is required")
	}
	if len(t.Description) > 255 {
		return errors.New("description must have at most 255 characters")
	}
	if t.Amount.LessThanOrEqual(decimal.Zero) {
		return errors.New("amount must be greater than zero")
	}
	if !t.Category.IsValid() {
		return errors.New("invalid category")
	}
	if !t.Status.IsValid() {
		return errors.New("invalid status")
	}
	if !t.Source.IsValid() {
		return errors.New("invalid source")
	}
	if !t.Visibility.IsValid() {
		return errors.New("invalid visibility")
	}
	if t.OccurredAt.IsZero() {
		return errors.New("occurredAt is required")
	}
	if t.UserAccountID == 0 {
		return errors.New("userAccountID is required")
	}
	if t.GuildID == 0 {
		return errors.New("guildID is required")
	}
	if t.CreatedBy == "" {
		return errors.New("createdBy is required")
	}

	return nil
}

func (t *Transaction) ApplyUpdate(
	typeValue Type,
	description string,
	amount decimal.Decimal,
	category Category,
	occurredAt time.Time,
	updatedBy string,
) error {
	if updatedBy == "" {
		return errors.New("updatedBy is required")
	}

	t.Type = typeValue
	t.Description = description
	t.Amount = amount
	t.Category = category
	t.OccurredAt = occurredAt

	if err := t.ValidateForUpdate(); err != nil {
		return err
	}

	t.Entry.Update(updatedBy)

	return nil
}

func (t *Transaction) ValidateForUpdate() error {
	if !t.Type.IsValid() {
		return ErrInvalidTransactionType
	}
	if t.Description == "" {
		return errors.New("description is required")
	}
	if len(t.Description) > 255 {
		return errors.New("description must have at most 255 characters")
	}
	if t.Amount.LessThanOrEqual(decimal.Zero) {
		return errors.New("amount must be greater than zero")
	}
	if !t.Category.IsValid() {
		return errors.New("invalid category")
	}
	if t.OccurredAt.IsZero() {
		return errors.New("occurredAt is required")
	}

	return nil
}

func (t *Transaction) Cancel(updatedBy string) error {
	if updatedBy == "" {
		return errors.New("updatedBy is required")
	}

	t.Status = StatusCancelled
	t.Entry.Update(updatedBy)

	return nil
}

func (t *Transaction) SetVisibility(visibility Visibility, updatedBy string) error {
	if updatedBy == "" {
		return errors.New("updatedBy is required")
	}
	if !visibility.IsValid() {
		return errors.New("invalid visibility")
	}

	t.Visibility = visibility
	t.Entry.Update(updatedBy)

	return nil
}

func (t Type) IsValid() bool {
	return t == TypeExpense || t == TypeIncome
}

func (c Category) IsValid() bool {
	switch c {
	case CategoryHousing, CategorySubscriptions, CategoryInsurance, CategoryEducation, CategoryTransportation,
		CategoryHealth, CategoryPersonal, CategoryTaxes, CategoryOther, CategoryFoodAndDining,
		CategoryEntertainment, CategoryShopping, CategoryPets, CategoryGames, CategoryTravel,
		CategoryInvestments, CategoryDebtRepayment, CategoryDonations:
		return true
	default:
		return false
	}
}

func (s Status) IsValid() bool {
	return s == StatusActive || s == StatusCancelled
}

func (s Source) IsValid() bool {
	return s == SourceManual || s == SourceImport
}

func (v Visibility) IsValid() bool {
	return v == VisibilityPrivate || v == VisibilityPublic
}
