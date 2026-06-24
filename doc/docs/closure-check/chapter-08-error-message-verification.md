# Chapter 08: Error Message Verification

## Description

A test that only checks `err != nil` misses half the story. Error message verification asserts that the error message *contains* expected text using `assert.Contains(t, err.Error(), want)`. This ensures the right error context is propagated — critical for debugging production failures where the difference between "not found" and "invalid input" determines the recovery path.

Real-world examples:
- `pantry/internal/core/domain/product_test.go:101` — `assert.Containsf(t, err.Error(), want, ...)`
- `notifier/model/test_utils.go:51-52` — `CheckResultError` uses `assert.Contains`
- `go-crap/internal/scan/scan_test.go:29-39` — `checkScanError(want string)`

## Code

```go
func ParseEmail(raw string) (*EmailAddress, error) {
	// validates email format, returns specific error messages:
	// - "email address is empty"
	// - "must contain exactly one @ symbol"
	// - "empty local part"
	// - "empty domain"
	// - "exceeds 64 characters"
	// - "must contain a dot"
	// - "contains invalid characters"
}
```

## Test

```go
type checkParseEmailFn func(*testing.T, *EmailAddress, error)

var checkParseEmail = func(fns ...checkParseEmailFn) []checkParseEmailFn { return fns }

func TestParseEmail(t *testing.T) {
	checkError := func(want string) checkParseEmailFn {
		return func(t *testing.T, e *EmailAddress, err error) {
			t.Helper()
			require.Error(t, err)
			assert.Contains(t, err.Error(), want)
			assert.Nil(t, e)
		}
	}

	tests := []struct {
		name   string
		input  string
		checks []checkParseEmailFn
	}{
		{
			name:  "empty input",
			input: "",
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
			name:  "domain without dot",
			input: "user@example",
			checks: checkParseEmail(
				checkError("must contain a dot"),
			),
		},
		// ... 7+ error cases each with distinct message
	}
}
```

## Testing Approach

Error message verification:

1. **`assert.Contains` over `assert.Equal`** — error messages often include dynamic content (the input value appears in quotes). `assert.Contains` checks for the key phrase while ignoring dynamic parts, making tests resilient to formatting changes.
2. **`require.Error` first** — always use `require.Error` (not `assert.Error`) before checking the message. If `err` is nil, `err.Error()` panics. `require` stops the test immediately, preventing a panic.
3. **Distinct messages per error type** — each validation failure produces a different message. The test verifies not just "an error occurred" but "the RIGHT error occurred." This catches bugs where validation short-circuits to a generic error instead of the specific one.
4. **Test as documentation** — the test table reads as a specification: "empty input → `email address is empty`", "missing @ → `must contain exactly one @`". A developer can understand the validation rules just by reading the test cases.
