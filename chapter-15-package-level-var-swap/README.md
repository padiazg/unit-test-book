# Chapter 15: Package-Level Var Swap

## Description

Assign a standard library or internal function to a package-level variable (`var randRead = rand.Read`) that production code calls through the variable. Tests save the original value and swap in a stub with `defer` for automatic restoration. This is the simplest possible test seam: no structs, no interfaces, no constructors.

Real-world example: `hexago/internal/core/domain/category.go:33` — `var uuidGenerate = uuid.NewV7` enables deterministic UUID generation in tests.

## Code

```go
var randRead = rand.Read

func GenerateShortCode(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := randRead(bytes); err != nil {
		return "", fmt.Errorf("generating code: %w", err)
	}
	return hex.EncodeToString(bytes)[:length], nil
}
```

## Test

```go
func TestGenerateShortCode(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		code, err := GenerateShortCode(8)
		require.NoError(t, err)
		assert.Len(t, code, 8)
	})

	t.Run("deterministic output", func(t *testing.T) {
		orig := randRead
		randRead = func(b []byte) (int, error) {
			for i := range b {
				b[i] = 0xAB
			}
			return len(b), nil
		}
		defer func() { randRead = orig }()

		code, err := GenerateShortCode(4)
		require.NoError(t, err)
		assert.Equal(t, "abababab", code)
	})

	t.Run("read error", func(t *testing.T) {
		orig := randRead
		randRead = func(b []byte) (int, error) {
			return 0, errors.New("entropy depleted")
		}
		defer func() { randRead = orig }()

		_, err := GenerateShortCode(8)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "entropy depleted")
	})
}
```

## Testing Approach

Package-level var swap:

1. **Minimal seam** — one `var` declaration replaces what would be a field, an interface, or a constructor parameter. The production function still reads like normal code: `randRead(bytes)`.
2. **`defer` restoration** — `defer func() { randRead = orig }()` guarantees the original function is restored even if the test panics. Essential for shared package-level state.
3. **Never parallel** — swapping a package var is global mutable state. Use `t.Parallel()` only when the var is write-once in `TestMain`. Otherwise, tests must run sequentially for the same package.
4. **Outside-in dependency** — the pattern is best for *calls to standard library functions* (rand, os, time). For your own interfaces (repositories, services), prefer constructor injection. Var swap is a quick seam, not an architecture.
