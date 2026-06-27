# Chapter 06: Check Collection Builder

## Description

A var-level helper function wraps check functions into a slice: `var checkXxx = func(fns ...checkXxxFn) []checkXxxFn { return fns }`. This gives callers a clean, expressive API: `checks: checkValidate(checkValid(true), checkErrorCount(1))` instead of the noisier `checks: []checkValidateFn{checkValid(true), checkErrorCount(1)}`.

Real-world examples:

- `pantry/internal/core/domain/product_test.go:91` — `var checkProductApplyMovement`  
- `notifier/model/test_utils.go:13-19` — `CheckNotifier`, `CheckResult`  
- `jokes/internal/adapters/secondary/external/joke_client_test.go:13` — `checkNewJokeClient`  

## Code

```go
type ValidationReport struct {
	Valid    bool
	Errors   []string
	Warnings []string
}

func ValidateDocument(doc *Document) *ValidationReport {
	// validates document, returns report with errors/warnings
}
```

## Test

```go
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

	tests := []struct {
		name   string
		doc    *Document
		checks []checkValidateFn
	}{
		{
			name: "valid document",
			doc:  &Document{Title: "Go Tips", Content: "Use interfaces to decouple code."},
			checks: checkValidate(        // <--- builder call, no [] type wrapper
				checkValid(true),
				checkErrorCount(0),
			),
		},
		{
			name: "missing title",
			doc:  &Document{Title: "", Content: "Some content here."},
			checks: checkValidate(
				checkValid(false),
				checkErrorCount(1),
			),
		},
	}
	// ...
}
```

## Testing Approach

The check collection builder:

1. **Cleaner syntax** — `checkValidate(checkValid(true))` reads better than `[]checkValidateFn{checkValid(true)}`. The varargs builder eliminates the slice literal wrapper.
2. **Zero-value consistency** — `checkValidate()` returns an empty slice, meaning "no assertions to run." This is useful for placeholder/TODO test cases that need to compile but aren't filled in yet.
3. **Builder naming convention** — the variable is named after the check type without `Fn` suffix: `checkValidate` builds `checkValidateFn`. This creates a natural naming: `checkXxx` returns `[]checkXxxFn`.
4. **Composable with `t.Helper()`** — all inner check factories call `t.Helper()`, so the stack trace points to the test case, not the helper. The builder itself (`var checkValidate`) is trivially simple and doesn't need `t.Helper()`.
