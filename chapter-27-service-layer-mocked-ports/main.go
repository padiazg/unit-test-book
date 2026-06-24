package service_layer_mocked_ports

import (
	"errors"
	"fmt"
)

var (
	ErrUserNotFound   = errors.New("user not found")
	ErrEmailRequired  = errors.New("email is required")
	ErrNotification   = errors.New("notification failed")
	ErrDuplicateEmail = errors.New("email already exists")
)

type User struct {
	ID    string
	Name  string
	Email string
	Age   int
}

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

func NewUserService(repo UserRepository, email EmailSender) *UserService {
	return &UserService{repo: repo, email: email}
}

func (s *UserService) Register(name, email string, age int) (*User, error) {
	if email == "" {
		return nil, ErrEmailRequired
	}

	existing, err := s.repo.FindByEmail(email)
	if err != nil {
		return nil, fmt.Errorf("checking email: %w", err)
	}
	if existing != nil {
		return nil, ErrDuplicateEmail
	}

	user := &User{
		ID:    fmt.Sprintf("usr_%s", email),
		Name:  name,
		Email: email,
		Age:   age,
	}

	if err := s.repo.Save(user); err != nil {
		return nil, fmt.Errorf("saving user: %w", err)
	}

	if err := s.email.SendWelcome(user); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrNotification, err)
	}

	return user, nil
}

func (s *UserService) GetUser(id string) (*User, error) {
	if id == "" {
		return nil, fmt.Errorf("user ID is required")
	}

	user, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, ErrUserNotFound
	}
	return user, nil
}

func (s *UserService) UpdateEmail(id, email string) error {
	if email == "" {
		return ErrEmailRequired
	}

	user, err := s.repo.FindByID(id)
	if err != nil {
		return err
	}
	if user == nil {
		return ErrUserNotFound
	}

	existing, err := s.repo.FindByEmail(email)
	if err != nil {
		return fmt.Errorf("checking email: %w", err)
	}
	if existing != nil && existing.ID != id {
		return ErrDuplicateEmail
	}

	user.Email = email
	return s.repo.Save(user)
}
