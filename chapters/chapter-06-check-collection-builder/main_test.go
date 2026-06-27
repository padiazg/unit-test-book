package check_collection_builder

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

type checkValidateFn func(*testing.T, *ValidationReport)

var checkValidate = func(fns ...checkValidateFn) []checkValidateFn { return fns }

func TestValidateDocument(t *testing.T) {
	checkValid := func(want bool) checkValidateFn {
		return func(t *testing.T, r *ValidationReport) {
			t.Helper()
			assert.Equal(t, want, r.Valid)
		}
	}

	checkErrorCount := func(want int) checkValidateFn {
		return func(t *testing.T, r *ValidationReport) {
			t.Helper()
			assert.Len(t, r.Errors, want)
		}
	}

	checkWarningCount := func(want int) checkValidateFn {
		return func(t *testing.T, r *ValidationReport) {
			t.Helper()
			assert.Len(t, r.Warnings, want)
		}
	}

	checkErrorContains := func(want string) checkValidateFn {
		return func(t *testing.T, r *ValidationReport) {
			t.Helper()
			assert.Contains(t, strings.Join(r.Errors, " "), want)
		}
	}

	checkWarningContains := func(want string) checkValidateFn {
		return func(t *testing.T, r *ValidationReport) {
			t.Helper()
			assert.Contains(t, strings.Join(r.Warnings, " "), want)
		}
	}

	tests := []struct {
		name   string
		doc    *Document
		checks []checkValidateFn
	}{
		{
			name: "valid document",
			doc:  &Document{Title: "Go Tips", Content: "Use interfaces to decouple code.", Tags: []string{"go", "testing"}},
			checks: checkValidate(
				checkValid(true),
				checkErrorCount(0),
				checkWarningCount(0),
			),
		},
		{
			name: "missing title",
			doc:  &Document{Title: "", Content: "Some content here for validation.", Tags: nil},
			checks: checkValidate(
				checkValid(false),
				checkErrorCount(1),
				checkErrorContains("title"),
			),
		},
		{
			name: "missing content",
			doc:  &Document{Title: "My Doc", Content: ""},
			checks: checkValidate(
				checkValid(false),
				checkErrorCount(1),
				checkErrorContains("content"),
			),
		},
		{
			name: "title and content both missing",
			doc:  &Document{Title: "  ", Content: "  "},
			checks: checkValidate(
				checkValid(false),
				checkErrorCount(2),
				checkErrorContains("title"),
				checkErrorContains("content"),
			),
		},
		{
			name: "nil document",
			doc:  nil,
			checks: checkValidate(
				checkValid(false),
				checkErrorCount(1),
				checkErrorContains("nil"),
			),
		},
		{
			name: "short content warning",
			doc:  &Document{Title: "Hi", Content: "Short"},
			checks: checkValidate(
				checkValid(true),
				checkWarningCount(1),
				checkWarningContains("too short"),
			),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := ValidateDocument(tt.doc)
			for _, c := range tt.checks {
				c(t, r)
			}
		})
	}
}
