package error_message_verification

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type checkParseEmailFn func(*testing.T, *EmailAddress, error)

var checkParseEmail = func(fns ...checkParseEmailFn) []checkParseEmailFn { return fns }

func TestParseEmail(t *testing.T) {
	checkSuccess := func() checkParseEmailFn {
		return func(t *testing.T, e *EmailAddress, err error) {
			t.Helper()
			require.NoError(t, err)
			assert.NotNil(t, e)
		}
	}

	checkError := func(want string) checkParseEmailFn {
		return func(t *testing.T, e *EmailAddress, err error) {
			t.Helper()
			require.Error(t, err)
			assert.Contains(t, err.Error(), want)
			assert.Nil(t, e)
		}
	}

	checkLocal := func(want string) checkParseEmailFn {
		return func(t *testing.T, e *EmailAddress, _ error) {
			t.Helper()
			assert.Equal(t, want, e.Local)
		}
	}

	checkDomain := func(want string) checkParseEmailFn {
		return func(t *testing.T, e *EmailAddress, _ error) {
			t.Helper()
			assert.Equal(t, want, e.Domain)
		}
	}

	tests := []struct {
		name   string
		input  string
		checks []checkParseEmailFn
	}{
		{
			name:  "valid email",
			input: "user@example.com",
			checks: checkParseEmail(
				checkSuccess(),
				checkLocal("user"),
				checkDomain("example.com"),
			),
		},
		{
			name:  "empty input",
			input: "",
			checks: checkParseEmail(
				checkError("email address is empty"),
			),
		},
		{
			name:  "whitespace only",
			input: "  ",
			checks: checkParseEmail(
				checkError("email address is empty"),
			),
		},
		{
			name:  "missing @ symbol",
			input: "notanemail",
			checks: checkParseEmail(
				checkError("must contain exactly one @"),
			),
		},
		{
			name:  "multiple @ symbols",
			input: "a@b@c.com",
			checks: checkParseEmail(
				checkError("must contain exactly one @"),
			),
		},
		{
			name:  "empty local part",
			input: "@example.com",
			checks: checkParseEmail(
				checkError("empty local part"),
			),
		},
		{
			name:  "empty domain",
			input: "user@",
			checks: checkParseEmail(
				checkError("empty domain"),
			),
		},
		{
			name:  "domain without dot",
			input: "user@example",
			checks: checkParseEmail(
				checkError("must contain a dot"),
			),
		},
		{
			name:  "local part too long",
			input: "aaaaaaaaaabbbbbbbbbbccccccccccddddddddddeeeeeeeeeeffffffffffggggggggg@example.com",
			checks: checkParseEmail(
				checkError("exceeds 64 characters"),
			),
		},
		{
			name:  "invalid characters",
			input: "user name@example.com",
			checks: checkParseEmail(
				checkError("invalid characters"),
			),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e, err := ParseEmail(tt.input)
			for _, c := range tt.checks {
				c(t, e, err)
			}
		})
	}
}
