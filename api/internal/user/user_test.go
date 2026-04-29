package user

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/luizhenrique-dev/guild-banker/api/internal/audit"
)

func TestUser_New(t *testing.T) {
	externalID := "user-123"
	name := "Luiz Silva"
	email := "luiz@example.com"
	createdBy := "admin"

	u, err := New(name, email, externalID, createdBy)

	assert.NoError(t, err)
	assert.EqualValues(t, 0, u.ID)
	assert.Equal(t, externalID, u.ExternalID)
	assert.Equal(t, name, u.Name)
	assert.Equal(t, email, u.Email)
	assert.Equal(t, createdBy, u.CreatedBy)
	assert.True(t, u.Enabled)
	assert.False(t, u.CreatedAt.IsZero())
	assert.Nil(t, u.UpdatedAt)
	assert.Nil(t, u.UpdatedBy)
	assert.Nil(t, u.DisabledAt)
	assert.Nil(t, u.DisabledBy)
}

func TestUser_Enable(t *testing.T) {
	u := &User{
		DisableEntry: audit.DisableEntry{
			Enabled:    false,
			DisabledAt: new(time.Now()),
			DisabledBy: new("admin"),
		},
	}

	u.Enable("admin")

	assert.True(t, u.Enabled)
	assert.Nil(t, u.DisabledAt)
	assert.Nil(t, u.DisabledBy)
	assert.NotNil(t, u.UpdatedAt)
	assert.NotNil(t, u.UpdatedBy)
}

func TestUser_Disable(t *testing.T) {
	u := &User{
		DisableEntry: audit.DisableEntry{Enabled: true},
	}
	disabledBy := "admin-2"

	u.Disable(disabledBy)

	assert.False(t, u.Enabled)
	assert.NotNil(t, u.DisabledAt)
	assert.NotNil(t, u.DisabledBy)
	assert.Equal(t, disabledBy, *u.DisabledBy)
	assert.WithinDuration(t, time.Now(), *u.DisabledAt, 1*time.Second)
}

func TestUser_Sync(t *testing.T) {
	u, err := New("Luiz Silva", "luiz@example.com", "user-123", "admin")
	assert.NoError(t, err)
	assert.Nil(t, u.UpdatedAt)

	newName := "Jhon Doe"
	newEmail := "jhon@example.com"
	updatedBy := "test@example.com"

	u.Sync(newName, newEmail, updatedBy)

	assert.EqualValues(t, 0, u.ID)
	assert.Equal(t, "user-123", u.ExternalID)
	assert.Equal(t, newName, u.Name)
	assert.Equal(t, newEmail, u.Email)
	assert.NotNil(t, u.UpdatedAt)
	assert.NotNil(t, u.UpdatedBy)
	assert.WithinDuration(t, time.Now(), *u.UpdatedAt, 1*time.Second)
}
