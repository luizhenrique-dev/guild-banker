package guild

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestGuild_New(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		g, err := New("devs", "Developers", "admin@example.com")

		assert.NoError(t, err)
		assert.EqualValues(t, 0, g.ID)
		assert.Equal(t, "devs", g.Name)
		assert.Equal(t, "Developers", g.DisplayName)
		assert.Equal(t, "admin@example.com", g.CreatedBy)
		assert.True(t, g.Enabled)
		assert.False(t, g.CreatedAt.IsZero())
		assert.Nil(t, g.UpdatedAt)
		assert.Nil(t, g.UpdatedBy)
		assert.Nil(t, g.DisabledAt)
		assert.Nil(t, g.DisabledBy)
	})

	t.Run("error when name is empty", func(t *testing.T) {
		g, err := New("", "Developers", "admin@example.com")

		assert.Error(t, err)
		assert.Nil(t, g)
		assert.EqualError(t, err, "name is required")
	})

	t.Run("error when display name is empty", func(t *testing.T) {
		g, err := New("devs", "", "admin@example.com")

		assert.Error(t, err)
		assert.Nil(t, g)
		assert.EqualError(t, err, "displayName is required")
	})

	t.Run("error when created by is empty", func(t *testing.T) {
		g, err := New("devs", "Developers", "")

		assert.Error(t, err)
		assert.Nil(t, g)
		assert.EqualError(t, err, "createdBy is required")
	})
}

func TestGuild_Rename(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		g, err := New("devs", "Developers", "admin@example.com")
		assert.NoError(t, err)
		assert.Nil(t, g.UpdatedAt)

		err = g.Rename("Platform", "owner@example.com")

		assert.NoError(t, err)
		assert.Equal(t, "Platform", g.DisplayName)
		assert.NotNil(t, g.UpdatedAt)
		assert.NotNil(t, g.UpdatedBy)
		assert.Equal(t, "owner@example.com", *g.UpdatedBy)
		assert.WithinDuration(t, time.Now(), *g.UpdatedAt, time.Second)
	})

	t.Run("error when name is empty", func(t *testing.T) {
		g, err := New("devs", "Developers", "admin@example.com")
		assert.NoError(t, err)

		err = g.Rename("", "owner@example.com")

		assert.EqualError(t, err, "name is required")
	})

	t.Run("error when by is empty", func(t *testing.T) {
		g, err := New("devs", "Developers", "admin@example.com")
		assert.NoError(t, err)

		err = g.Rename("platform", "")

		assert.EqualError(t, err, "by is required")
	})
}
