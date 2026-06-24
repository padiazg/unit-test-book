package service_layer_mocked_ports

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type checkUserServiceFn func(*testing.T, *User, error)

var checkUserService = func(fns ...checkUserServiceFn) []checkUserServiceFn { return fns }

func checkUser(name string) checkUserServiceFn {
	return func(t *testing.T, u *User, err error) {
		t.Helper()
		require.NoError(t, err)
		assert.Equal(t, name, u.Name)
	}
}

func checkError(want string) checkUserServiceFn {
	return func(t *testing.T, u *User, err error) {
		t.Helper()
		require.Error(t, err)
		assert.Contains(t, err.Error(), want)
	}
}

type mockUserRepository struct {
	mock.Mock
}

func (m *mockUserRepository) FindByID(id string) (*User, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*User), args.Error(1)
}

func (m *mockUserRepository) FindByEmail(email string) (*User, error) {
	args := m.Called(email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*User), args.Error(1)
}

func (m *mockUserRepository) Save(user *User) error {
	args := m.Called(user)
	return args.Error(0)
}

type mockEmailSender struct {
	mock.Mock
}

func (m *mockEmailSender) SendWelcome(user *User) error {
	args := m.Called(user)
	return args.Error(0)
}

func TestUserService_Register(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		repo := &mockUserRepository{}
		email := &mockEmailSender{}

		repo.On("FindByEmail", "alice@test.com").Return(nil, nil)
		repo.On("Save", mock.MatchedBy(func(u *User) bool {
			return u.Email == "alice@test.com"
		})).Return(nil)
		email.On("SendWelcome", mock.MatchedBy(func(u *User) bool {
			return u.Email == "alice@test.com"
		})).Return(nil)

		svc := NewUserService(repo, email)
		user, err := svc.Register("Alice", "alice@test.com", 30)
		require.NoError(t, err)
		assert.Equal(t, "Alice", user.Name)
		assert.Equal(t, "alice@test.com", user.Email)
		repo.AssertExpectations(t)
		email.AssertExpectations(t)
	})

	t.Run("empty email", func(t *testing.T) {
		svc := NewUserService(&mockUserRepository{}, &mockEmailSender{})
		_, err := svc.Register("Alice", "", 30)
		assert.ErrorIs(t, err, ErrEmailRequired)
	})

	t.Run("duplicate email", func(t *testing.T) {
		repo := &mockUserRepository{}
		email := &mockEmailSender{}

		existing := &User{ID: "usr_old", Email: "alice@test.com"}
		repo.On("FindByEmail", "alice@test.com").Return(existing, nil)

		svc := NewUserService(repo, email)
		_, err := svc.Register("Alice", "alice@test.com", 30)
		assert.ErrorIs(t, err, ErrDuplicateEmail)
		repo.AssertExpectations(t)
	})

	t.Run("notification failure", func(t *testing.T) {
		repo := &mockUserRepository{}
		email := &mockEmailSender{}

		repo.On("FindByEmail", "alice@test.com").Return(nil, nil)
		repo.On("Save", mock.Anything).Return(nil)
		email.On("SendWelcome", mock.Anything).Return(ErrNotification)

		svc := NewUserService(repo, email)
		_, err := svc.Register("Alice", "alice@test.com", 30)
		assert.ErrorIs(t, err, ErrNotification)
		repo.AssertExpectations(t)
		email.AssertExpectations(t)
	})
}

func TestUserService_GetUser(t *testing.T) {
	t.Run("existing user", func(t *testing.T) {
		repo := &mockUserRepository{}
		repo.On("FindByID", "usr_1").Return(&User{ID: "usr_1", Name: "Alice"}, nil)

		svc := NewUserService(repo, &mockEmailSender{})
		user, err := svc.GetUser("usr_1")
		require.NoError(t, err)
		assert.Equal(t, "Alice", user.Name)
	})

	t.Run("user not found", func(t *testing.T) {
		repo := &mockUserRepository{}
		repo.On("FindByID", "missing").Return(nil, nil)

		svc := NewUserService(repo, &mockEmailSender{})
		_, err := svc.GetUser("missing")
		assert.ErrorIs(t, err, ErrUserNotFound)
	})

	t.Run("empty id", func(t *testing.T) {
		svc := NewUserService(&mockUserRepository{}, &mockEmailSender{})
		_, err := svc.GetUser("")
		assert.Error(t, err)
	})
}

func TestUserService_UpdateEmail(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		repo := &mockUserRepository{}
		repo.On("FindByID", "usr_1").Return(&User{ID: "usr_1", Email: "old@test.com"}, nil)
		repo.On("FindByEmail", "new@test.com").Return(nil, nil)
		repo.On("Save", mock.MatchedBy(func(u *User) bool {
			return u.Email == "new@test.com"
		})).Return(nil)

		svc := NewUserService(repo, &mockEmailSender{})
		err := svc.UpdateEmail("usr_1", "new@test.com")
		require.NoError(t, err)
	})

	t.Run("empty email", func(t *testing.T) {
		svc := NewUserService(&mockUserRepository{}, &mockEmailSender{})
		err := svc.UpdateEmail("usr_1", "")
		assert.ErrorIs(t, err, ErrEmailRequired)
	})

	t.Run("same email different user", func(t *testing.T) {
		repo := &mockUserRepository{}
		repo.On("FindByID", "usr_1").Return(&User{ID: "usr_1", Email: "old@test.com"}, nil)
		repo.On("FindByEmail", "taken@test.com").Return(&User{ID: "usr_2"}, nil)

		svc := NewUserService(repo, &mockEmailSender{})
		err := svc.UpdateEmail("usr_1", "taken@test.com")
		assert.ErrorIs(t, err, ErrDuplicateEmail)
	})

	t.Run("same email same user", func(t *testing.T) {
		repo := &mockUserRepository{}
		repo.On("FindByID", "usr_1").Return(&User{ID: "usr_1", Email: "old@test.com"}, nil)
		repo.On("FindByEmail", "same@test.com").Return(&User{ID: "usr_1"}, nil)
		repo.On("Save", mock.Anything).Return(nil)

		svc := NewUserService(repo, &mockEmailSender{})
		err := svc.UpdateEmail("usr_1", "same@test.com")
		require.NoError(t, err)
	})
}
