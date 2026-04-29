package user

import (
	"errors"
	"time"

	"github.com/luizhenrique-dev/guild-banker/api/internal/audit"
)

type User struct {
	ID         int64
	ExternalID string
	Name       string
	Email      string
	audit.DisableEntry
}

func New(name, email, externalID, createdBy string) (*User, error) {
	u := &User{
		ExternalID: externalID,
		Name:       name,
		Email:      email,
		DisableEntry: audit.DisableEntry{
			Enabled: true,
			Entry: audit.Entry{
				CreatedAt: time.Now(),
				CreatedBy: createdBy,
			},
		},
	}

	if err := u.validate(); err != nil {
		return nil, err
	}

	return u, nil
}

func (u *User) Sync(name, email, by string) {
	u.Name = name
	u.Email = email
	u.Update(by)
}

func (u *User) validate() error {
	if u.Name == "" {
		return errors.New("name is required")
	}
	if u.Email == "" {
		return errors.New("email is required")
	}
	if u.ExternalID == "" {
		return errors.New("externalID is required")
	}
	if u.CreatedBy == "" {
		return errors.New("createdBy is required")
	}
	return nil
}
