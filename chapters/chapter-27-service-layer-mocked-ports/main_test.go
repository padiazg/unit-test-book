package service_layer_mocked_ports

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type checkUserServiceFn func(*testing.T, *User, error)

var checkUserService = func(fns ...checkUserServiceFn) []checkUserServiceFn { return fns }

func checkUser(name, email string) checkUserServiceFn {
	return func(t *testing.T, u *User, err error) {
		t.Helper()
		require.NoError(t, err)
		require.NotNil(t, u)
		assert.Equal(t, name, u.Name)
		assert.Equal(t, email, u.Email)
	}
}

func checkNoError() checkUserServiceFn {
	return func(t *testing.T, u *User, err error) {
		t.Helper()
		require.NoError(t, err)
	}
}

func checkError(want error) checkUserServiceFn {
	return func(t *testing.T, u *User, err error) {
		t.Helper()
		require.Error(t, err)
		if want != nil {
			assert.ErrorIs(t, err, want)
		}
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

type testFixtures struct {
	repo  *mockUserRepository
	email *mockEmailSender
}

func (f *testFixtures) Teardown(t *testing.T) {
	t.Helper()
	f.repo.AssertExpectations(t)
	f.email.AssertExpectations(t)
}

func setupService(t *testing.T) (*UserService, *testFixtures) {
	t.Helper()
	repo := &mockUserRepository{}
	email := &mockEmailSender{}
	svc := NewUserService(repo, email)
	return svc, &testFixtures{repo: repo, email: email}
}

func TestUserService_Register(t *testing.T) {
	tests := []struct {
		name    string
		nameArg string
		email   string
		age     int
		before  func(*testFixtures)
		checks  []checkUserServiceFn
	}{
		{
			name:    "success",
			nameArg: "Alice",
			email:   "alice@test.com",
			age:     30,
			before: func(f *testFixtures) {
				f.repo.On("FindByEmail", "alice@test.com").Return(nil, nil)
				f.repo.On("Save", mock.MatchedBy(func(u *User) bool {
					return u.Email == "alice@test.com"
				})).Return(nil)
				f.email.On("SendWelcome", mock.MatchedBy(func(u *User) bool {
					return u.Email == "alice@test.com"
				})).Return(nil)
			},
			checks: checkUserService(
				checkUser("Alice", "alice@test.com"),
			),
		},
		{
			name:    "empty email",
			nameArg: "Alice",
			email:   "",
			age:     30,
			before:  nil,
			checks: checkUserService(
				checkError(ErrEmailRequired),
			),
		},
		{
			name:    "duplicate email",
			nameArg: "Alice",
			email:   "alice@test.com",
			age:     30,
			before: func(f *testFixtures) {
				f.repo.On("FindByEmail", "alice@test.com").Return(&User{ID: "usr_old", Email: "alice@test.com"}, nil)
			},
			checks: checkUserService(
				checkError(ErrDuplicateEmail),
			),
		},
		{
			name:    "notification failure",
			nameArg: "Alice",
			email:   "alice@test.com",
			age:     30,
			before: func(f *testFixtures) {
				f.repo.On("FindByEmail", "alice@test.com").Return(nil, nil)
				f.repo.On("Save", mock.Anything).Return(nil)
				f.email.On("SendWelcome", mock.Anything).Return(ErrNotification)
			},
			checks: checkUserService(
				checkError(ErrNotification),
			),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc, fixtures := setupService(t)
			defer fixtures.Teardown(t)
			if tt.before != nil {
				tt.before(fixtures)
			}
			user, err := svc.Register(tt.nameArg, tt.email, tt.age)
			for _, c := range tt.checks {
				c(t, user, err)
			}
		})
	}
}

func TestUserService_GetUser(t *testing.T) {
	tests := []struct {
		name   string
		id     string
		before func(*testFixtures)
		checks []checkUserServiceFn
	}{
		{
			name: "existing user",
			id:   "usr_1",
			before: func(f *testFixtures) {
				f.repo.On("FindByID", "usr_1").Return(&User{ID: "usr_1", Name: "Alice"}, nil)
			},
			checks: checkUserService(
				checkUser("Alice", ""),
			),
		},
		{
			name: "user not found",
			id:   "missing",
			before: func(f *testFixtures) {
				f.repo.On("FindByID", "missing").Return(nil, nil)
			},
			checks: checkUserService(
				checkError(ErrUserNotFound),
			),
		},
		{
			name:   "empty id",
			id:     "",
			before: nil,
			checks: checkUserService(
				checkError(nil),
			),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc, fixtures := setupService(t)
			defer fixtures.Teardown(t)
			if tt.before != nil {
				tt.before(fixtures)
			}
			user, err := svc.GetUser(tt.id)
			for _, c := range tt.checks {
				c(t, user, err)
			}
		})
	}
}

func TestUserService_UpdateEmail(t *testing.T) {
	tests := []struct {
		name   string
		id     string
		email  string
		before func(*testFixtures)
		checks []checkUserServiceFn
	}{
		{
			name:  "success",
			id:    "usr_1",
			email: "new@test.com",
			before: func(f *testFixtures) {
				f.repo.On("FindByID", "usr_1").Return(&User{ID: "usr_1", Email: "old@test.com"}, nil)
				f.repo.On("FindByEmail", "new@test.com").Return(nil, nil)
				f.repo.On("Save", mock.MatchedBy(func(u *User) bool {
					return u.Email == "new@test.com"
				})).Return(nil)
			},
			checks: checkUserService(
				checkNoError(),
			),
		},
		{
			name:   "empty email",
			id:     "usr_1",
			email:  "",
			before: nil,
			checks: checkUserService(
				checkError(ErrEmailRequired),
			),
		},
		{
			name:  "same email different user",
			id:    "usr_1",
			email: "taken@test.com",
			before: func(f *testFixtures) {
				f.repo.On("FindByID", "usr_1").Return(&User{ID: "usr_1", Email: "old@test.com"}, nil)
				f.repo.On("FindByEmail", "taken@test.com").Return(&User{ID: "usr_2"}, nil)
			},
			checks: checkUserService(
				checkError(ErrDuplicateEmail),
			),
		},
		{
			name:  "same email same user",
			id:    "usr_1",
			email: "same@test.com",
			before: func(f *testFixtures) {
				f.repo.On("FindByID", "usr_1").Return(&User{ID: "usr_1", Email: "old@test.com"}, nil)
				f.repo.On("FindByEmail", "same@test.com").Return(&User{ID: "usr_1"}, nil)
				f.repo.On("Save", mock.Anything).Return(nil)
			},
			checks: checkUserService(
				checkNoError(),
			),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc, fixtures := setupService(t)
			defer fixtures.Teardown(t)
			if tt.before != nil {
				tt.before(fixtures)
			}
			err := svc.UpdateEmail(tt.id, tt.email)
			for _, c := range tt.checks {
				c(t, nil, err)
			}
		})
	}
}
