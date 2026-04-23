package user

import (
	"time"
)

type User struct {
	ID         int64
	ExternalID string
	Name       string
	Email      string
	CreatedAt  time.Time
	CreatedBy  string
	Enabled    bool
	UpdatedAt  *time.Time
	DisabledAt *time.Time
	DisabledBy *string
}

func New(name, email, externalID, createdBy string) *User {
	return &User{
		ExternalID: externalID,
		Name:       name,
		Email:      email,
		CreatedAt:  time.Now(),
		CreatedBy:  createdBy,
		Enabled:    true,
	}
}

func (u *User) Enable() {
	u.Enabled = true
	u.DisabledAt = nil
	u.DisabledBy = nil
	u.UpdatedAt = new(time.Now())
}

func (u *User) Disable(disabledBy string) {
	u.Enabled = false
	u.DisabledAt = new(time.Now())
	u.DisabledBy = new(disabledBy)
}

func (u *User) Sync(name, email string) {
	u.Name = name
	u.Email = email
	u.UpdatedAt = new(time.Now())
}
