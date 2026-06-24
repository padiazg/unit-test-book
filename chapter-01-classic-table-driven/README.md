# Chapter 01: Classic Table-Driven Tests (`wantErr bool`)

## Description

The most common Go testing pattern: a constructor that returns `(T, error)` is tested by enumerating valid and invalid inputs. Each test case has a `wantErr bool` that drives a branch: on expected errors, assert error presence and nil result; on success, assert no error and non-nil result with correct field values.

Real-world examples:
- `pantry/internal/core/domain/product_test.go:13` — `TestNewProduct`
- `pantry/internal/core/domain/category_test.go:11` — `TestNewCategory`
- `hexago/pkg/version/version_test.go:15` — `TestVersionParseVersion`

## Code

```go
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
```

## Test

```go
func TestNewPerson(t *testing.T) {
	tests := []struct {
		name    string
		person  string
		email   string
		age     int
		wantErr bool
	}{
		{name: "valid person", person: "Alice", email: "alice@example.com", age: 30, wantErr: false},
		{name: "empty name", person: "", email: "bob@example.com", age: 25, wantErr: true},
		{name: "whitespace name", person: "  ", email: "bob@example.com", age: 25, wantErr: true},
		{name: "missing email", person: "Bob", email: "", age: 25, wantErr: true},
		{name: "invalid email format", person: "Bob", email: "not-an-email", age: 25, wantErr: true},
		{name: "negative age", person: "Charlie", email: "charlie@example.com", age: -1, wantErr: true},
		{name: "age too high", person: "Diana", email: "diana@example.com", age: 200, wantErr: true},
		{name: "age zero", person: "Eve", email: "eve@example.com", age: 0, wantErr: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, err := NewPerson(tt.person, tt.email, tt.age)

			if tt.wantErr {
				require.Error(t, err)
				assert.Nil(t, p)
				return
			}

			require.NoError(t, err)
			assert.NotNil(t, p)
			assert.Equal(t, tt.person, p.Name)
			assert.Equal(t, tt.email, p.Email)
			assert.Equal(t, tt.age, p.Age)
		})
	}
}
```

## Testing Approach

The `wantErr bool` pattern works as follows:

1. **Happy path branch** — when `wantErr` is `false`, assert no error (`require.NoError`) then validate the returned value with `assert.Equal` for each field. Using `require` for the error check ensures the test stops immediately if the value is nil.
2. **Error path branch** — when `wantErr` is `true`, assert error presence (`require.Error`) and nil value (`assert.Nil`). The `return` after the check prevents further nil-pointer panics.
3. **Boundary cases** — include edge values (zero, empty string, boundary max) to verify validation logic. The `age: 0` case confirms zero is valid; `age: -1` and `age: 200` confirm the boundary rejection.
4. **Whitespace handling** — testing `"  "` as input verifies that trimming happens before validation.

This pattern is preferred over separate test functions per case because it makes adding new cases trivial (one struct entry) and visually groups all scenarios in one place.
