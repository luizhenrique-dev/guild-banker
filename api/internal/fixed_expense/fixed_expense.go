package fixedexpense

import (
	"errors"
	"time"

	"github.com/shopspring/decimal"

	"github.com/luizhenrique-dev/guild-banker/api/internal/audit"
)

type Category string

const (
	CategoryGrocery        Category = "GROCERY"
	CategoryHousing        Category = "HOUSING"
	CategoryUtilities      Category = "UTILITIES"
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
	CategoryTravel         Category = "TRAVEL"
	CategoryInvestments    Category = "INVESTMENTS"
)

type Status string

const (
	StatusActive    Status = "ACTIVE"
	StatusPaused    Status = "PAUSED"
	StatusCancelled Status = "CANCELLED"
)

type FixedExpense struct {
	ID       int64
	GuildID  int64
	Name     string
	Amount   decimal.Decimal
	DueDay   int
	Category Category
	Status   Status
	audit.Entry
}

func New(guildID int64, name string, amount decimal.Decimal, dueDay int, category Category, createdBy string) (*FixedExpense, error) {
	fe := &FixedExpense{
		GuildID:   guildID,
		Name:      name,
		Amount:    amount,
		DueDay:    dueDay,
		Category:  category,
		Status:    StatusActive,
		CreatedAt: time.Now(),
		CreatedBy: createdBy,
	}

	if err := fe.Validate(); err != nil {
		return nil, err
	}

	return fe, nil
}

func (f *FixedExpense) Update(amount decimal.Decimal, dueDay int, category Category, status Status, changedBy string) error {
	if changedBy == "" {
		return errors.New("changedBy is required")
	}

	f.Amount = amount
	f.DueDay = dueDay
	f.Category = category
	f.Status = status

	if err := f.Validate(); err != nil {
		return err
	}

	f.UpdateEntry(changedBy)

	return nil
}

func (f *FixedExpense) Deactivate(status Status, changedBy string) error {
	if status != StatusPaused && status != StatusCancelled {
		return ErrInvalidFixedExpenseStatus
	}

	f.Status = status
	f.UpdateEntry(changedBy)

	return nil
}

func (f *FixedExpense) Validate() error {
	if f.GuildID <= 0 {
		return errors.New("guildID must be positive")
	}
	if f.Name == "" {
		return errors.New("name is required")
	}
	if f.Amount.LessThanOrEqual(decimal.Zero) {
		return errors.New("amount must be greater than zero")
	}
	const firstDayOfMonth = 1
	const lastDayOfMonth = 31
	if f.DueDay < firstDayOfMonth || f.DueDay > lastDayOfMonth {
		return errors.New("dueDay must be between 1 and 31")
	}
	if !f.Category.IsValid() {
		return errors.New("invalid category")
	}
	if !f.Status.IsValid() {
		return ErrInvalidFixedExpenseStatus
	}
	if f.CreatedBy == "" {
		return errors.New("createdBy is required")
	}

	return nil
}

func (f *FixedExpense) UpdateEntry(changedBy string) {
	f.UpdatedAt = new(time.Now())
	f.UpdatedBy = new(changedBy)
}

func (c Category) IsValid() bool {
	switch c {
	case CategoryGrocery, CategoryHousing, CategoryUtilities, CategorySubscriptions, CategoryInsurance,
		CategoryEducation, CategoryTransportation, CategoryHealth, CategoryPersonal, CategoryTaxes, CategoryOther,
		CategoryFoodAndDining, CategoryEntertainment, CategoryShopping, CategoryPets, CategoryTravel, CategoryInvestments:
		return true
	default:
		return false
	}
}

func (s Status) IsValid() bool {
	switch s {
	case StatusActive, StatusPaused, StatusCancelled:
		return true
	default:
		return false
	}
}
