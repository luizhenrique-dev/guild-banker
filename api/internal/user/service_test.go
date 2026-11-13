package user

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type storageMock struct {
	mock.Mock
}

func (m *storageMock) RecordAccess(ctx context.Context, id int64) error {
	return m.Called(ctx, id).Error(0)
}

func (m *storageMock) Create(ctx context.Context, u *User) error {
	args := m.Called(ctx, u)
	return args.Error(0)
}

func (m *storageMock) Update(ctx context.Context, u *User) error {
	args := m.Called(ctx, u)
	return args.Error(0)
}

func (m *storageMock) GetByID(ctx context.Context, id int64) (*User, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*User), args.Error(1)
}

func (m *storageMock) GetByExternalID(ctx context.Context, externalID string) (*User, error) {
	args := m.Called(ctx, externalID)
	return args.Get(0).(*User), args.Error(1)
}

func (m *storageMock) GetByEmail(ctx context.Context, email string) (*User, error) {
	args := m.Called(ctx, email)
	return args.Get(0).(*User), args.Error(1)
}

func (m *storageMock) Enable(ctx context.Context, id int64) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *storageMock) Disable(ctx context.Context, id int64, disabledBy string) error {
	args := m.Called(ctx, id, disabledBy)
	return args.Error(0)
}

func newService(t *testing.T) (*Service, *storageMock) {
	t.Helper()
	s := &storageMock{}
	return NewService(s), s
}

func stubUser() *User {
	u, _ := New("John Doe", "john@example.com", "ext-123", "admin")
	return u
}

func TestService_Create(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		svc, storage := newService(t)
		storage.On("Create", mock.Anything, mock.AnythingOfType("*user.User")).Return(nil)

		u, err := svc.Create(context.Background(), "John Doe", "john@example.com", "ext-123", "admin")

		assert.NoError(t, err)
		assert.NotNil(t, u)
		assert.Equal(t, "John Doe", u.Name)
		assert.Equal(t, "john@example.com", u.Email)
		assert.Equal(t, "ext-123", u.ExternalID)
		assert.True(t, u.Enabled)
		storage.AssertExpectations(t)
	})

	t.Run("error on invalid params", func(t *testing.T) {
		svc, storage := newService(t)

		u, err := svc.Create(context.Background(), "", "john@example.com", "ext-123", "admin")

		assert.Error(t, err)
		assert.Nil(t, u)
		storage.AssertNotCalled(t, "Create")
	})

	t.Run("error on storage", func(t *testing.T) {
		svc, storage := newService(t)
		storage.On("Create", mock.Anything, mock.AnythingOfType("*user.User")).Return(errors.New("db error"))

		u, err := svc.Create(context.Background(), "John Doe", "john@example.com", "ext-123", "admin")

		assert.Error(t, err)
		assert.Nil(t, u)
		storage.AssertExpectations(t)
	})
}

func TestService_Update(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		svc, storage := newService(t)
		existing := stubUser()
		storage.On("GetByID", mock.Anything, int64(1)).Return(existing, nil)
		storage.On("Update", mock.Anything, existing).Return(nil)

		u, err := svc.Update(context.Background(), 1, "Jane Doe", "jane@example.com", "admin")

		assert.NoError(t, err)
		assert.Equal(t, "Jane Doe", u.Name)
		assert.Equal(t, "jane@example.com", u.Email)
		storage.AssertExpectations(t)
	})

	t.Run("error when id is zero", func(t *testing.T) {
		svc, storage := newService(t)

		u, err := svc.Update(context.Background(), 0, "Jane Doe", "jane@example.com", "admin")

		assert.Error(t, err)
		assert.Nil(t, u)
		storage.AssertNotCalled(t, "GetByID")
		storage.AssertNotCalled(t, "Update")
	})

	t.Run("error on get by id", func(t *testing.T) {
		svc, storage := newService(t)
		storage.On("GetByID", mock.Anything, int64(1)).Return((*User)(nil), errors.New("db error"))

		u, err := svc.Update(context.Background(), 1, "Jane Doe", "jane@example.com", "admin")

		assert.Error(t, err)
		assert.Nil(t, u)
		storage.AssertNotCalled(t, "Update")
	})

	t.Run("error on storage update", func(t *testing.T) {
		svc, storage := newService(t)
		existing := stubUser()
		storage.On("GetByID", mock.Anything, int64(1)).Return(existing, nil)
		storage.On("Update", mock.Anything, existing).Return(errors.New("db error"))

		u, err := svc.Update(context.Background(), 1, "Jane Doe", "jane@example.com", "admin")

		assert.Error(t, err)
		assert.Nil(t, u)
		storage.AssertExpectations(t)
	})
}

func TestService_GetByID(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		svc, storage := newService(t)
		existing := stubUser()
		storage.On("GetByID", mock.Anything, int64(1)).Return(existing, nil)

		u, err := svc.GetByID(context.Background(), 1)

		assert.NoError(t, err)
		assert.Equal(t, existing, u)
		storage.AssertExpectations(t)
	})

	t.Run("error when id is zero", func(t *testing.T) {
		svc, storage := newService(t)

		u, err := svc.GetByID(context.Background(), 0)

		assert.Error(t, err)
		assert.Nil(t, u)
		storage.AssertNotCalled(t, "GetByID")
	})

	t.Run("error on storage", func(t *testing.T) {
		svc, storage := newService(t)
		storage.On("GetByID", mock.Anything, int64(1)).Return((*User)(nil), errors.New("db error"))

		u, err := svc.GetByID(context.Background(), 1)

		assert.Error(t, err)
		assert.Nil(t, u)
		storage.AssertExpectations(t)
	})
}

func TestService_GetByExternalID(t *testing.T) {
	t.Run("not found is preserved", func(t *testing.T) {
		svc, storage := newService(t)
		storage.On("GetByExternalID", mock.Anything, "missing").Return((*User)(nil), ErrNotFound)
		u, err := svc.GetByExternalID(context.Background(), "missing")
		assert.Nil(t, u)
		assert.ErrorIs(t, err, ErrNotFound)
		storage.AssertExpectations(t)
	})
	t.Run("success", func(t *testing.T) {
		svc, storage := newService(t)
		existing := stubUser()
		storage.On("GetByExternalID", mock.Anything, "ext-123").Return(existing, nil)

		u, err := svc.GetByExternalID(context.Background(), "ext-123")

		assert.NoError(t, err)
		assert.Equal(t, existing, u)
		storage.AssertExpectations(t)
	})

	t.Run("error when externalID is empty", func(t *testing.T) {
		svc, storage := newService(t)

		u, err := svc.GetByExternalID(context.Background(), "")

		assert.Error(t, err)
		assert.Nil(t, u)
		storage.AssertNotCalled(t, "GetByExternalID")
	})

	t.Run("error on storage", func(t *testing.T) {
		svc, storage := newService(t)
		storage.On("GetByExternalID", mock.Anything, "ext-123").Return((*User)(nil), errors.New("db error"))

		u, err := svc.GetByExternalID(context.Background(), "ext-123")

		assert.Error(t, err)
		assert.Nil(t, u)
		storage.AssertExpectations(t)
	})
}

func TestService_GetByEmail(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		svc, storage := newService(t)
		existing := stubUser()
		storage.On("GetByEmail", mock.Anything, "john@example.com").Return(existing, nil)

		u, err := svc.GetByEmail(context.Background(), "john@example.com")

		assert.NoError(t, err)
		assert.Equal(t, existing, u)
		storage.AssertExpectations(t)
	})

	t.Run("error when email is empty", func(t *testing.T) {
		svc, storage := newService(t)

		u, err := svc.GetByEmail(context.Background(), "")

		assert.Error(t, err)
		assert.Nil(t, u)
		storage.AssertNotCalled(t, "GetByEmail")
	})

	t.Run("error on storage", func(t *testing.T) {
		svc, storage := newService(t)
		storage.On("GetByEmail", mock.Anything, "john@example.com").Return((*User)(nil), errors.New("db error"))

		u, err := svc.GetByEmail(context.Background(), "john@example.com")

		assert.Error(t, err)
		assert.Nil(t, u)
		storage.AssertExpectations(t)
	})
}

func TestService_Enable(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		svc, storage := newService(t)
		storage.On("Enable", mock.Anything, int64(1)).Return(nil)

		err := svc.Enable(context.Background(), 1)

		assert.NoError(t, err)
		storage.AssertExpectations(t)
	})

	t.Run("error when id is zero", func(t *testing.T) {
		svc, storage := newService(t)

		err := svc.Enable(context.Background(), 0)

		assert.Error(t, err)
		storage.AssertNotCalled(t, "Enable")
	})

	t.Run("error on storage", func(t *testing.T) {
		svc, storage := newService(t)
		storage.On("Enable", mock.Anything, int64(1)).Return(errors.New("db error"))

		err := svc.Enable(context.Background(), 1)

		assert.Error(t, err)
		storage.AssertExpectations(t)
	})
}

func TestService_Disable(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		svc, storage := newService(t)
		storage.On("Disable", mock.Anything, int64(1), "admin").Return(nil)

		err := svc.Disable(context.Background(), 1, "admin")

		assert.NoError(t, err)
		storage.AssertExpectations(t)
	})

	t.Run("error when id is zero", func(t *testing.T) {
		svc, storage := newService(t)

		err := svc.Disable(context.Background(), 0, "admin")

		assert.Error(t, err)
		storage.AssertNotCalled(t, "Disable")
	})

	t.Run("error when disabledBy is empty", func(t *testing.T) {
		svc, storage := newService(t)

		err := svc.Disable(context.Background(), 1, "")

		assert.Error(t, err)
		storage.AssertNotCalled(t, "Disable")
	})

	t.Run("error on storage", func(t *testing.T) {
		svc, storage := newService(t)
		storage.On("Disable", mock.Anything, int64(1), "admin").Return(errors.New("db error"))

		err := svc.Disable(context.Background(), 1, "admin")

		assert.Error(t, err)
		storage.AssertExpectations(t)
	})
}

func (m *storageMock) CreateIfAbsent(ctx context.Context, u *User) (*User, error) {
	args := m.Called(ctx, u)
	return args.Get(0).(*User), args.Error(1)
}

func (m *storageMock) UpdateProfile(ctx context.Context, u *User) (*User, error) {
	args := m.Called(ctx, u)
	return args.Get(0).(*User), args.Error(1)
}

func TestServiceEnsureAccount(t *testing.T) {
	for _, tt := range []struct {
		name, profileName, email        string
		missing, disabled               bool
		lookupErr, createErr, updateErr error
	}{
		{name: "create", missing: true},
		{name: "equal"},
		{name: "name", profileName: "New Name"},
		{name: "email", email: "NEW@example.test"},
		{name: "both", profileName: " New Name ", email: "new@example.test"},
		{name: "case diff", email: "JOHN@example.com"},
		{name: "disabled", disabled: true, profileName: "New Name"},
		{name: "lookup canceled", lookupErr: context.Canceled},
		{name: "create conflict", missing: true, createErr: ErrEmailConflict},
		{name: "update conflict", profileName: "New Name", updateErr: ErrEmailConflict},
		{name: "update canceled", profileName: "New Name", updateErr: context.Canceled},
	} {
		t.Run(tt.name, func(t *testing.T) {
			svc, storage := newService(t)
			t.Cleanup(func() { storage.AssertExpectations(t) })
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			existing := stubUser()
			existing.ID = 42
			existing.Enabled = !tt.disabled
			name, email := existing.Name, existing.Email
			if tt.profileName != "" {
				name = tt.profileName
			}
			if tt.email != "" {
				email = tt.email
			}
			before := *existing
			lookupErr := tt.lookupErr
			if tt.missing {
				lookupErr = ErrNotFound
			}
			storage.On("GetByExternalID", ctx, "ext-123").Return(existing, lookupErr).Once()
			if tt.missing {
				storage.On("CreateIfAbsent", ctx, mock.MatchedBy(func(u *User) bool {
					return u.ExternalID == "ext-123" && u.Name == name && u.Email == strings.ToLower(email) && u.CreatedBy == strings.ToLower(email) && u.Enabled && u.UpdatedAt == nil
				})).Return(existing, tt.createErr).Once()
			}
			canonical := strings.ToLower(email)
			changed := name != existing.Name || canonical != existing.Email
			persisted := *existing
			persisted.Name = name
			persisted.Email = canonical
			if lookupErr == nil && !tt.disabled && changed {
				storage.On("UpdateProfile", ctx, mock.MatchedBy(func(u *User) bool {
					return u.ID == 42 && u.ExternalID == before.ExternalID && u.CreatedAt.Equal(before.CreatedAt) && u.CreatedBy == before.CreatedBy && u.Name == name && u.Email == canonical && u.UpdatedBy != nil && *u.UpdatedBy == canonical
				})).Return(&persisted, tt.updateErr).Once()
			}
			got, err := svc.EnsureAccount(ctx, "ext-123", name, email)
			expected := tt.lookupErr
			if tt.createErr != nil {
				expected = tt.createErr
			}
			if tt.updateErr != nil {
				expected = tt.updateErr
			}
			if tt.disabled {
				expected = ErrDisabled
			}
			if expected != nil {
				require.ErrorIs(t, err, expected)
				assert.Nil(t, got)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, int64(42), got.ID)
			if !changed {
				assert.Equal(t, before, *got)
				storage.AssertNotCalled(t, "UpdateProfile", mock.Anything, mock.Anything)
			}
			if changed {
				assert.Same(t, &persisted, got)
			}
		})
	}
	for _, values := range [][3]string{{"", "Name", "a@b.test"}, {" ", "Name", "a@b.test"}, {"sub", " ", "a@b.test"}, {"sub", "Name", "\t"}} {
		svc, storage := newService(t)
		got, err := svc.EnsureAccount(context.Background(), values[0], values[1], values[2])
		require.Error(t, err)
		assert.Nil(t, got)
		storage.AssertExpectations(t)
	}
}
