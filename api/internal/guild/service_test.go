package guild

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/luizhenrique-dev/guild-banker/api/internal/audit"
)

type storageMock struct {
	mock.Mock
}

func (m *storageMock) Create(ctx context.Context, g *Guild, creatorUserID int64) error {
	args := m.Called(ctx, g, creatorUserID)
	return args.Error(0)
}

func (m *storageMock) GetByID(ctx context.Context, id int64) (*Guild, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*Guild), args.Error(1)
}

func (m *storageMock) NameExists(ctx context.Context, name string, excludeID int64) (bool, error) {
	args := m.Called(ctx, name, excludeID)
	return args.Bool(0), args.Error(1)
}

func (m *storageMock) UpdateName(ctx context.Context, g *Guild) error {
	args := m.Called(ctx, g)
	return args.Error(0)
}

func (m *storageMock) Enable(ctx context.Context, id int64, by string, now time.Time) error {
	args := m.Called(ctx, id, by, now)
	return args.Error(0)
}

func (m *storageMock) Disable(ctx context.Context, id int64, by string, now time.Time) error {
	args := m.Called(ctx, id, by, now)
	return args.Error(0)
}

func (m *storageMock) ListByMember(ctx context.Context, userID int64) ([]*Guild, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]*Guild), args.Error(1)
}

func (m *storageMock) IsMember(ctx context.Context, guildID, userID int64) (bool, error) {
	args := m.Called(ctx, guildID, userID)
	return args.Bool(0), args.Error(1)
}

func (m *storageMock) InviteByEmail(ctx context.Context, guildID int64, email string, invitedByUserID int64) error {
	args := m.Called(ctx, guildID, email, invitedByUserID)
	return args.Error(0)
}

func (m *storageMock) RemoveMember(ctx context.Context, guildID, userID int64) error {
	args := m.Called(ctx, guildID, userID)
	return args.Error(0)
}

func newService(t *testing.T) (*Service, *storageMock) {
	t.Helper()
	s := &storageMock{}
	return NewService(s), s
}

func stubGuild() *Guild {
	g, _ := New("devs", "Developers", "admin@example.com")
	g.ID = 1
	return g
}

func validActor() audit.Actor {
	return audit.Actor{UserID: 1, Email: "admin@example.com"}
}

func TestService_Create(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		svc, storage := newService(t)
		actor := validActor()
		storage.On("NameExists", mock.Anything, "devs", int64(0)).Return(false, nil)
		storage.On("Create", mock.Anything, mock.AnythingOfType("*guild.Guild"), actor.UserID).Return(nil)

		g, err := svc.Create(context.Background(), "devs", "Developers", actor)

		assert.NoError(t, err)
		assert.NotNil(t, g)
		assert.Equal(t, "devs", g.Name)
		assert.Equal(t, "Developers", g.DisplayName)
		storage.AssertExpectations(t)
	})

	t.Run("error when actor is invalid", func(t *testing.T) {
		svc, storage := newService(t)

		g, err := svc.Create(context.Background(), "devs", "Developers", audit.Actor{})

		assert.Error(t, err)
		assert.Nil(t, g)
		storage.AssertNotCalled(t, "NameExists")
		storage.AssertNotCalled(t, "Create")
	})

	t.Run("error on name exists check", func(t *testing.T) {
		svc, storage := newService(t)
		actor := validActor()
		storage.On("NameExists", mock.Anything, "devs", int64(0)).Return(false, errors.New("db error"))

		g, err := svc.Create(context.Background(), "devs", "Developers", actor)

		assert.Error(t, err)
		assert.Nil(t, g)
		storage.AssertNotCalled(t, "Create")
	})

	t.Run("error when name already used", func(t *testing.T) {
		svc, storage := newService(t)
		actor := validActor()
		storage.On("NameExists", mock.Anything, "devs", int64(0)).Return(true, nil)

		g, err := svc.Create(context.Background(), "devs", "Developers", actor)

		assert.ErrorIs(t, err, ErrGuildNameAlreadyUsed)
		assert.Nil(t, g)
		storage.AssertNotCalled(t, "Create")
	})
}

func TestService_UpdateName(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		svc, storage := newService(t)
		actor := validActor()
		existing := stubGuild()
		storage.On("IsMember", mock.Anything, int64(1), actor.UserID).Return(true, nil)
		storage.On("NameExists", mock.Anything, "Platform", int64(1)).Return(false, nil)
		storage.On("GetByID", mock.Anything, int64(1)).Return(existing, nil)
		storage.On("UpdateName", mock.Anything, existing).Return(nil)

		g, err := svc.UpdateName(context.Background(), 1, "Platform", actor)

		assert.NoError(t, err)
		assert.Equal(t, "Platform", g.DisplayName)
		storage.AssertExpectations(t)
	})

	t.Run("error when requester is not member", func(t *testing.T) {
		svc, storage := newService(t)
		actor := validActor()
		storage.On("IsMember", mock.Anything, int64(1), actor.UserID).Return(false, nil)

		g, err := svc.UpdateName(context.Background(), 1, "platform", actor)

		assert.ErrorIs(t, err, ErrRequesterIsNotMember)
		assert.Nil(t, g)
		storage.AssertNotCalled(t, "NameExists")
		storage.AssertNotCalled(t, "GetByID")
		storage.AssertNotCalled(t, "UpdateName")
	})
}

func TestService_Enable(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		svc, storage := newService(t)
		actor := validActor()
		g := stubGuild()
		g.Enabled = false
		storage.On("IsMember", mock.Anything, int64(1), actor.UserID).Return(true, nil)
		storage.On("GetByID", mock.Anything, int64(1)).Return(g, nil)
		storage.On("Enable", mock.Anything, int64(1), actor.Email, mock.AnythingOfType("time.Time")).Return(nil)

		err := svc.Enable(context.Background(), 1, actor)

		assert.NoError(t, err)
		storage.AssertExpectations(t)
	})

	t.Run("error when guild already enabled", func(t *testing.T) {
		svc, storage := newService(t)
		actor := validActor()
		g := stubGuild()
		g.Enabled = true
		storage.On("IsMember", mock.Anything, int64(1), actor.UserID).Return(true, nil)
		storage.On("GetByID", mock.Anything, int64(1)).Return(g, nil)

		err := svc.Enable(context.Background(), 1, actor)

		assert.ErrorIs(t, err, ErrGuildAlreadyEnabled)
		storage.AssertNotCalled(t, "Enable")
	})
}

func TestService_Disable(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		svc, storage := newService(t)
		actor := validActor()
		g := stubGuild()
		g.Enabled = true
		storage.On("IsMember", mock.Anything, int64(1), actor.UserID).Return(true, nil)
		storage.On("GetByID", mock.Anything, int64(1)).Return(g, nil)
		storage.On("Disable", mock.Anything, int64(1), actor.Email, mock.AnythingOfType("time.Time")).Return(nil)

		err := svc.Disable(context.Background(), 1, actor)

		assert.NoError(t, err)
		storage.AssertExpectations(t)
	})

	t.Run("error when guild already disabled", func(t *testing.T) {
		svc, storage := newService(t)
		actor := validActor()
		g := stubGuild()
		g.Enabled = false
		storage.On("IsMember", mock.Anything, int64(1), actor.UserID).Return(true, nil)
		storage.On("GetByID", mock.Anything, int64(1)).Return(g, nil)

		err := svc.Disable(context.Background(), 1, actor)

		assert.ErrorIs(t, err, ErrGuildAlreadyDisabled)
		storage.AssertNotCalled(t, "Disable")
	})
}

func TestService_ListByMember(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		svc, storage := newService(t)
		guilds := []*Guild{stubGuild()}
		storage.On("ListByMember", mock.Anything, int64(1)).Return(guilds, nil)

		result, err := svc.ListByMember(context.Background(), 1)

		assert.NoError(t, err)
		assert.Equal(t, guilds, result)
		storage.AssertExpectations(t)
	})

	t.Run("error when user id is zero", func(t *testing.T) {
		svc, storage := newService(t)

		result, err := svc.ListByMember(context.Background(), 0)

		assert.Error(t, err)
		assert.Nil(t, result)
		storage.AssertNotCalled(t, "ListByMember")
	})
}

func TestService_InviteUser(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		svc, storage := newService(t)
		storage.On("IsMember", mock.Anything, int64(1), int64(1)).Return(true, nil)
		storage.On("InviteByEmail", mock.Anything, int64(1), "member@example.com", int64(1)).Return(nil)

		err := svc.InviteUser(context.Background(), 1, 1, "member@example.com")

		assert.NoError(t, err)
		storage.AssertExpectations(t)
	})

	t.Run("error when email is invalid", func(t *testing.T) {
		svc, storage := newService(t)
		storage.On("IsMember", mock.Anything, int64(1), int64(1)).Return(true, nil)

		err := svc.InviteUser(context.Background(), 1, 1, "invalid")

		assert.Error(t, err)
		storage.AssertNotCalled(t, "InviteByEmail")
	})
}

func TestService_RemoveUser(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		svc, storage := newService(t)
		storage.On("IsMember", mock.Anything, int64(1), int64(1)).Return(true, nil)
		storage.On("RemoveMember", mock.Anything, int64(1), int64(2)).Return(nil)

		err := svc.RemoveUser(context.Background(), 1, 1, 2)

		assert.NoError(t, err)
		storage.AssertExpectations(t)
	})

	t.Run("error when trying to remove yourself", func(t *testing.T) {
		svc, storage := newService(t)
		storage.On("IsMember", mock.Anything, int64(1), int64(1)).Return(true, nil)

		err := svc.RemoveUser(context.Background(), 1, 1, 1)

		assert.EqualError(t, err, "remove user: cannot remove yourself")
		storage.AssertNotCalled(t, "RemoveMember")
	})
}
