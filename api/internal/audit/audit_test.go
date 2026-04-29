package audit

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestEntry_Update(t *testing.T) {
	t.Run("sets updated fields", func(t *testing.T) {
		e := &Entry{
			CreatedAt: time.Now(),
			CreatedBy: "admin",
		}
		assert.Nil(t, e.UpdatedAt)
		assert.Nil(t, e.UpdatedBy)

		e.Update("editor")

		assert.NotNil(t, e.UpdatedAt)
		assert.NotNil(t, e.UpdatedBy)
		assert.Equal(t, "editor", *e.UpdatedBy)
		assert.WithinDuration(t, time.Now(), *e.UpdatedAt, 1*time.Second)
	})

	t.Run("overwrites previous update", func(t *testing.T) {
		e := &Entry{
			CreatedAt: time.Now(),
			CreatedBy: "admin",
		}
		e.Update("first-editor")
		firstUpdatedAt := *e.UpdatedAt

		e.Update("second-editor")

		assert.Equal(t, "second-editor", *e.UpdatedBy)
		assert.False(t, e.UpdatedAt.Before(firstUpdatedAt))
	})
}

func TestDisableEntry_Enable(t *testing.T) {
	t.Run("enables and clears disable fields", func(t *testing.T) {
		d := &DisableEntry{
			Enabled:    false,
			DisabledAt: new(time.Now()),
			DisabledBy: new("admin"),
		}

		d.Enable("admin")

		assert.True(t, d.Enabled)
		assert.Nil(t, d.DisabledAt)
		assert.Nil(t, d.DisabledBy)
		assert.NotNil(t, d.UpdatedAt)
		assert.NotNil(t, d.UpdatedBy)
		assert.Equal(t, "admin", *d.UpdatedBy)
		assert.WithinDuration(t, time.Now(), *d.UpdatedAt, 1*time.Second)
	})

	t.Run("can enable an already enabled entry", func(t *testing.T) {
		d := &DisableEntry{Enabled: true}

		d.Enable("admin")

		assert.True(t, d.Enabled)
		assert.Nil(t, d.DisabledAt)
		assert.Nil(t, d.DisabledBy)
	})
}

func TestDisableEntry_Disable(t *testing.T) {
	t.Run("disables and sets disable fields", func(t *testing.T) {
		d := &DisableEntry{Enabled: true}
		disabledBy := "admin-2"

		d.Disable(disabledBy)

		assert.False(t, d.Enabled)
		assert.NotNil(t, d.DisabledAt)
		assert.NotNil(t, d.DisabledBy)
		assert.Equal(t, disabledBy, *d.DisabledBy)
		assert.WithinDuration(t, time.Now(), *d.DisabledAt, 1*time.Second)
		assert.NotNil(t, d.UpdatedAt)
		assert.NotNil(t, d.UpdatedBy)
		assert.Equal(t, disabledBy, *d.UpdatedBy)
		assert.WithinDuration(t, time.Now(), *d.UpdatedAt, 1*time.Second)
	})

	t.Run("can disable an already disabled entry", func(t *testing.T) {
		d := &DisableEntry{
			Enabled:    false,
			DisabledAt: new(time.Now()),
			DisabledBy: new("prev-admin"),
		}

		d.Disable("new-admin")

		assert.False(t, d.Enabled)
		assert.Equal(t, "new-admin", *d.DisabledBy)
		assert.Equal(t, "new-admin", *d.UpdatedBy)
	})
}
