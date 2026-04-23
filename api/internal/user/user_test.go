package user

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestUser_New(t *testing.T) {
	externalID := "user-123"
	name := "Luiz Silva"
	email := "luiz@example.com"
	createdBy := "admin"

	u := New(name, email, externalID, createdBy)

	assert.EqualValues(t, 0, u.ID)
	assert.Equal(t, externalID, u.ExternalID)
	assert.Equal(t, name, u.Name)
	assert.Equal(t, email, u.Email)
	assert.Equal(t, createdBy, u.CreatedBy)
	assert.True(t, u.Enabled)
	assert.False(t, u.CreatedAt.IsZero())
	assert.Nil(t, u.UpdatedAt)
	assert.Nil(t, u.DisabledAt)
	assert.Nil(t, u.DisabledBy)
}

func TestUser_Enable(t *testing.T) {
	u := &User{
		Enabled:    false,
		DisabledAt: new(time.Now()),
		DisabledBy: new("admin"),
	}

	u.Enable()

	assert.True(t, u.Enabled)
	assert.Nil(t, u.DisabledAt)
	assert.Nil(t, u.DisabledBy)
}

func TestUser_Disable(t *testing.T) {
	u := &User{
		Enabled: true,
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
	externalID := "user-123"
	name := "Luiz Silva"
	email := "luiz@example.com"
	createdBy := "admin"

	u := New(name, email, externalID, createdBy)
	assert.Nil(t, u.UpdatedAt)

	newName := "Jhon Doe"
	newEmail := "jhon@example.com"

	u.Sync(newName, newEmail)

	assert.EqualValues(t, 0, u.ID)
	assert.Equal(t, externalID, u.ExternalID)
	assert.Equal(t, newName, u.Name)
	assert.Equal(t, newEmail, u.Email)
	assert.NotNil(t, u.UpdatedAt)
	assert.WithinDuration(t, time.Now(), *u.UpdatedAt, 1*time.Second)
}
