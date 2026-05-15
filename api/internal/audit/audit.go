package audit

import (
	"errors"
	"time"
)

type Entry struct {
	CreatedAt time.Time
	CreatedBy string
	UpdatedAt *time.Time
	UpdatedBy *string
}

func (e *Entry) Update(changedBy string) {
	e.UpdatedAt = new(time.Now())
	e.UpdatedBy = new(changedBy)
}

type DisableEntry struct {
	Entry
	Enabled    bool
	DisabledAt *time.Time
	DisabledBy *string
}

func (d *DisableEntry) Enable(changedBy string) {
	d.Enabled = true
	d.DisabledAt = nil
	d.DisabledBy = nil
	d.Update(changedBy)
}

func (d *DisableEntry) Disable(changedBy string) {
	d.Enabled = false
	d.DisabledAt = new(time.Now())
	d.DisabledBy = new(changedBy)
	d.Update(changedBy)
}

type Actor struct {
	UserID int64
	Email  string
}

func (a Actor) Validate() error {
	if a.UserID == 0 {
		return errors.New("actor: userID is required")
	}
	if a.Email == "" {
		return errors.New("actor: e-mail is required")
	}
	return nil
}
