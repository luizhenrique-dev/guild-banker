package audit

import "time"

type Entry struct {
	CreatedAt time.Time
	CreatedBy string
	UpdatedAt *time.Time
	UpdatedBy *string
}

func (e *Entry) Update(by string) {
	e.UpdatedAt = new(time.Now())
	e.UpdatedBy = new(by)
}

type DisableEntry struct {
	Entry
	Enabled    bool
	DisabledAt *time.Time
	DisabledBy *string
}

func (d *DisableEntry) Enable(by string) {
	d.Enabled = true
	d.DisabledAt = nil
	d.DisabledBy = nil
	d.Update(by)
}

func (d *DisableEntry) Disable(by string) {
	d.Enabled = false
	d.DisabledAt = new(time.Now())
	d.DisabledBy = new(by)
	d.Update(by)
}
