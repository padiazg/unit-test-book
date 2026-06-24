# Chapter 02: Value Assertions (`want T`)

## Description

Extends the classic table-driven test with a `want T` field for expected return values. Used for pure functions that map an input deterministically to an output — no side effects, no errors, just input → output. The assertion becomes a simple `assert.Equal(t, tt.want, got)`.

Real-world examples:

- `go-testgen/internal/utils/parse_test.go:9` — `TestParseInt`  
- `go-testgen/internal/generator/generator_test.go:105` — `TestQualifiedTypeName_Array`  
- `hexago/pkg/version/version_test.go:141` — `TestVersionString`  

## Code

```go
package value_assertions

func ParseInt(input string) int {
	if len(input) == 0 {
		return 0
	}

	var (
		result int
		sign   = 1
		start  = 0
	)

	if input[0] == '-' {
		sign = -1
		start = 1
	} else if input[0] == '+' {
		start = 1
	}

	for i := start; i < len(input); i++ {
		ch := input[i]
		if ch < '0' || ch > '9' {
			return 0
		}
		result = result*10 + int(ch-'0')
	}

	return result * sign
}
```

## Test

```go
func TestParseInt(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  int
	}{
		{name: "valid positive", input: "42", want: 42},
		{name: "valid zero", input: "0", want: 0},
		{name: "valid negative", input: "-5", want: -5},
		{name: "leading plus", input: "+7", want: 7},
		{name: "empty string", input: "", want: 0},
		{name: "non-numeric", input: "abc", want: 0},
		{name: "with spaces", input: " 3", want: 0},
		{name: "trailing chars", input: "12ab", want: 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ParseInt(tt.input)
			assert.Equal(t, tt.want, got)
		})
	}
}
```

## Testing Approach

The `want T` pattern:

1. **Direct assertion** — `assert.Equal(t, tt.want, got)` is the only assertion needed. No error branching, no nil checks. This makes the test table the single source of truth for expected behavior.
2. **Exhaustive edge cases** — the 8 test cases cover: positive, zero, negative, explicit plus sign, empty string, non-numeric, leading whitespace, and trailing garbage. Each edge case is a single line.
3. **Zero-value default** — the function returns `0` for invalid input. The test documents this contract explicitly through the `want` field.
4. **Pure function advantage** — because `ParseInt` has no side effects, every call is independent. Tests can run in parallel without shared state. The assertion is always the same shape: `assert.Equal(t, tt.want, got)`.

This pattern is ideal for parsing, formatting, calculation, and transformation functions where the output is completely determined by the input.
