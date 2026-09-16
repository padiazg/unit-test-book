package temp_files_and_parsing

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

type entityConstructorInfoFn func(*testing.T, *entityCtorInfo, error)

var checkentityConstructorInfo = func(fns ...entityConstructorInfoFn) []entityConstructorInfoFn { return fns }

func checkentityConstructorInfoError(want string) entityConstructorInfoFn {
	return func(t *testing.T, _ *entityCtorInfo, err error) {
		t.Helper()
		if want == "" {
			assert.NoErrorf(t, err, "checkentityConstructorInfoError: expected no error, got %v", err)
			return
		}
		if assert.Errorf(t, err, "checkentityConstructorInfoError: expected error %q", want) {
			assert.Containsf(t, err.Error(), want, "checkentityConstructorInfoError mismatch")
		}
	}
}

func checkNewArgs(want string) entityConstructorInfoFn {
	return func(t *testing.T, info *entityCtorInfo, _ error) {
		t.Helper()
		assert.Equal(t, want, info.newArgs)
	}
}

func checkCreateFields(want []Field) entityConstructorInfoFn {
	return func(t *testing.T, info *entityCtorInfo, _ error) {
		t.Helper()
		assert.Equal(t, want, info.createFields)
	}
}

func checkNeedsUUID(want bool) entityConstructorInfoFn {
	return func(t *testing.T, info *entityCtorInfo, _ error) {
		t.Helper()
		assert.Equal(t, want, info.needsUUID)
	}
}

func checkNeedsTime(want bool) entityConstructorInfoFn {
	return func(t *testing.T, info *entityCtorInfo, _ error) {
		t.Helper()
		assert.Equal(t, want, info.needsTime)
	}
}

func writeEntityFile(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	p := filepath.Join(dir, "entity.go")
	if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	return p
}

func Test_entityConstructorInfo(t *testing.T) {
	tests := []struct {
		name       string
		entityName string
		content    string
		checks     []entityConstructorInfoFn
	}{
		{
			name:       "string id plus extra field",
			entityName: "Category",
			content:    "package categories\nfunc NewCategory(id string, description string) (*Category, error) { return nil, nil }\n",
			checks: checkentityConstructorInfo(
				checkentityConstructorInfoError(""),
				checkNewArgs(`uuid.NewString(), input.Description`),
				checkCreateFields([]Field{{Name: "Description", Type: "string"}}),
				checkNeedsUUID(true),
			),
		},
		{
			name:       "uuid id and timestamp",
			entityName: "User",
			content:    "package users\nfunc NewUser(id uuid.UUID, createdAt time.Time) (*User, error) { return nil, nil }\n",
			checks: checkentityConstructorInfo(
				checkentityConstructorInfoError(""),
				checkNewArgs("uuid.New(), time.Now()"),
				checkNeedsUUID(true),
				checkNeedsTime(true),
			),
		},
		{
			name:       "id only",
			entityName: "Note",
			content:    "package notes\nfunc NewNote(id string) (*Note, error) { return nil, nil }\n",
			checks: checkentityConstructorInfo(
				checkentityConstructorInfoError(""),
				checkNewArgs("uuid.NewString()"),
				checkNeedsUUID(true),
			),
		},
		{
			name:       "no id field",
			entityName: "Greeting",
			content:    "package greetings\nfunc NewGreeting(name string, email string) (*Greeting, error) { return nil, nil }\n",
			checks: checkentityConstructorInfo(
				checkentityConstructorInfoError(""),
				checkNewArgs("input.Name, input.Email"),
				checkCreateFields([]Field{{Name: "Name", Type: "string"}, {Name: "Email", Type: "string"}}),
			),
		},
		{
			name:       "grouped param names",
			entityName: "Pair",
			content:    "package pairs\nfunc NewPair(a, b string) (*Pair, error) { return nil, nil }\n",
			checks: checkentityConstructorInfo(
				checkentityConstructorInfoError(""),
				checkNewArgs("input.A, input.B"),
				checkCreateFields([]Field{{Name: "A", Type: "string"}, {Name: "B", Type: "string"}}),
			),
		},
		{
			name:       "pointer and slice fields become input fields",
			entityName: "Order",
			content:    "package orders\nfunc NewOrder(id string, total float64, items []string, at *time.Time) (*Order, error) { return nil, nil }\n",
			checks: checkentityConstructorInfo(
				checkentityConstructorInfoError(""),
				checkNewArgs("uuid.NewString(), input.Total, input.Items, input.At"),
				checkCreateFields([]Field{
					{Name: "Total", Type: "float64"},
					{Name: "Items", Type: "[]string"},
					{Name: "At", Type: "*time.Time"},
				}),
				checkNeedsUUID(true),
				checkNeedsTime(true),
			),
		},
		{
			name:       "constructor missing",
			entityName: "Category",
			content:    "package categories\n",
			checks: checkentityConstructorInfo(
				checkentityConstructorInfoError("constructor NewCategory not found"),
			),
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			path := writeEntityFile(t, tt.content)
			r, err := entityConstructorInfo(path, tt.entityName)
			for _, c := range tt.checks {
				c(t, r, err)
			}
		})
	}
}

func Test_entityConstructorInfo_missingFile(t *testing.T) {
	_, err := entityConstructorInfo(filepath.Join(t.TempDir(), "nope.go"), "Category")
	assert.Error(t, err)
}
