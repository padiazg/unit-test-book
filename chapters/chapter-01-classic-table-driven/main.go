package classic_table_driven

import (
	"errors"
	"strings"
)

type Person struct {
	Name  string
	Email string
	Age   int
}

func NewPerson(name, email string, age int) (*Person, error) {
	name = strings.TrimSpace(name)
	email = strings.TrimSpace(email)

	if name == "" {
		return nil, errors.New("name is required")
	}
	if email == "" || !strings.Contains(email, "@") {
		return nil, errors.New("valid email is required")
	}
	if age < 0 || age > 150 {
		return nil, errors.New("age must be between 0 and 150")
	}

	return &Person{Name: name, Email: email, Age: age}, nil
}
