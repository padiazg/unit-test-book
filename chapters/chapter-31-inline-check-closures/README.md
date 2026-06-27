# Chapter 31: Inline Check Closures

## Description

Not every check needs a factory function. When a closure is used exactly once, define it inline in the test case: `checks: checkXxx(func(t *testing.T, got Result, err error) { ... })`. This avoids the ceremony of extracting a named function that has no reuse. The decision is purely about reuse frequency — there is no architectural benefit to extracting single-use checks into named factories.

Real-world example: `go-crap/internal/coverage/scanner_test.go:137-161` — inline closures used directly in test table cases where each assertion pattern appears once.

## Code

```go
type Config struct {
	Host string
	Port int
	Key  string
}

func ValidateConfig(cfg Config) error {
	if cfg.Host == "" {
		return errors.New("host is required")
	}
	if cfg.Port <= 0 || cfg.Port > 65535 {
		return fmt.Errorf("port must be between 1 and 65535, got %d", cfg.Port)
	}
	if cfg.Key == "" {
		return errors.New("api key is required")
	}
	return nil
}
```

## Test

```go
type checkValidateFn func(*testing.T, error)

var checkValidate = func(fns ...checkValidateFn) []checkValidateFn { return fns }

func TestValidateConfig(t *testing.T) {
	tests := []struct {
		name   string
		cfg    Config
		checks []checkValidateFn
	}{
		{
			name: "valid config",
			cfg:  Config{Host: "example.com", Port: 443, Key: "sk-abc"},
			checks: checkValidate(
				func(t *testing.T, err error) {
					t.Helper()
					assert.NoError(t, err)
				},
			),
		},
		{
			name: "missing host",
			cfg:  Config{Port: 443, Key: "sk-abc"},
			checks: checkValidate(
				func(t *testing.T, err error) {
					t.Helper()
					if assert.Error(t, err) {
						assert.Contains(t, err.Error(), "host is required")
					}
				},
			),
		},
		{
			name: "zero port",
			cfg:  Config{Host: "example.com", Port: 0, Key: "sk-abc"},
			checks: checkValidate(
				func(t *testing.T, err error) {
					t.Helper()
					if assert.Error(t, err) {
						assert.Contains(t, err.Error(), "port must be between 1 and 65535")
					}
				},
			),
		},
		// ... more cases follow the same inline pattern
	}
}
```

## Testing Approach

Inline check closures:

1. **No factory overhead** — each closure is written directly in the test case where it appears. This is the simplest form of the closure-check pattern: define the assertion where you use it, call `t.Helper()`, and move on.

2. **Reuse frequency drives extraction** — the moment the same assertion pattern appears in a second test case, extract it into a factory function (Chapter 07). Until then, the inline form is less code, less indirection, and equally readable.

3. **Contrast with over-factoring** — the test file includes a commented-out section showing the same tests with unnecessary factory extraction. Compare the two: the factories (`checkSuccess`, `checkError`, `checkErrorNotContains`) are each used once, yet add 15 lines of indirection. The inline version is shorter and the assertion logic sits next to the test data it validates.

4. **Zero-cost flexibility** — inline closures can reference local variables, `t.TempDir()`, or loop variables without plumbing them through factory parameters. For unique assertions this is simpler than capturing values into a factory argument.
