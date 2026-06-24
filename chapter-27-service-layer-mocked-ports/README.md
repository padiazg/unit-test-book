# Chapter 27: Service Layer Mocked Ports

## Description

Test the service layer by mocking its port interfaces (repository, email sender) with `testify/mock`. The service contains business logic without I/O; tests verify validation, error handling, and correct delegation to ports. Each port is mocked independently, so a repository failure or email failure can be tested in isolation.

Real-world example: `pantry/internal/adapters/primary/http/product_handler_test.go` — `mockProductService` and `mockProductRepository` using testify/mock.

## Code

```go
type UserRepository interface {
	FindByID(id string) (*User, error)
	FindByEmail(email string) (*User, error)
	Save(user *User) error
}

type EmailSender interface {
	SendWelcome(user *User) error
}

type UserService struct {
	repo  UserRepository
	email EmailSender
}

func (s *UserService) Register(name, email string, age int) (*User, error) {
	if email == "" { return nil, ErrEmailRequired }
	existing, err := s.repo.FindByEmail(email)
	if existing != nil { return nil, ErrDuplicateEmail }
	user := &User{ID: fmt.Sprintf("usr_%s", email), Name: name, Email: email, Age: age}
	if err := s.repo.Save(user); err != nil { return nil, err }
	if err := s.email.SendWelcome(user); err != nil { return nil, err }
	return user, nil
}
```

## Test

```go
type mockUserRepository struct { mock.Mock }
func (m *mockUserRepository) FindByID(id string) (*User, error) {
	args := m.Called(id)
	if args.Get(0) == nil { return nil, args.Error(1) }
	return args.Get(0).(*User), args.Error(1)
}
func (m *mockUserRepository) Save(user *User) error {
	args := m.Called(user)
	return args.Error(0)
}

func TestUserService_Register(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		repo := &mockUserRepository{}
		email := &mockEmailSender{}
		repo.On("FindByEmail", "alice@test.com").Return(nil, nil)
		repo.On("Save", mock.Anything).Return(nil)
		email.On("SendWelcome", mock.Anything).Return(nil)

		svc := NewUserService(repo, email)
		user, err := svc.Register("Alice", "alice@test.com", 30)
		require.NoError(t, err)
		assert.Equal(t, "Alice", user.Name)
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
		repo.On("FindByEmail", "alice@test.com").
			Return(&User{Email: "alice@test.com"}, nil)
		svc := NewUserService(repo, &mockEmailSender{})
		_, err := svc.Register("Alice", "alice@test.com", 30)
		assert.ErrorIs(t, err, ErrDuplicateEmail)
	})
}

func TestUserService_GetUser(t *testing.T) {
	t.Run("not found", func(t *testing.T) {
		repo := &mockUserRepository{}
		repo.On("FindByID", "missing").Return(nil, nil)
		svc := NewUserService(repo, &mockEmailSender{})
		_, err := svc.GetUser("missing")
		assert.ErrorIs(t, err, ErrUserNotFound)
	})
}
```

## Testing Approach

Service layer mocked ports:

1. **Interface segregation** — `UserRepository` (data) and `EmailSender` (notification) are separate interfaces. Tests mock only the port they need. A repository test doesn't need email, and vice versa.
2. **`testify/mock` expectations** — `On("FindByEmail", "alice@test.com").Return(nil, nil)` sets up both input matching and output values. `AssertExpectations(t)` verifies every expected call happened exactly once.
3. **Error path coverage** — each port method has a failure variant tested independently: email validation before repository calls, duplicate detection, and notification failure after save.
4. **`mock.Anything` vs specific matchers** — `mock.Anything` accepts any argument for flexible matching. Use specific values (`"alice@test.com"`) when the exact argument matters, `mock.MatchedBy(func)` for custom validation.
