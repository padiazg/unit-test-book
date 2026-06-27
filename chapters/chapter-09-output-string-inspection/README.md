# Chapter 09: Output String Inspection

## Description

When testing functions that produce formatted string output (tables, reports, code generation, CSV), use `assert.Contains` / `assert.NotContains` to verify the output contains critical elements without matching the entire string. This avoids brittle exact-match assertions while still guaranteeing the output has the required structure.

String inspection is the primary testing strategy for code generators. In `go-testgen`, generator tests verify that the generated source code contains expected function signatures, struct definitions, and assertion calls.

Real-world examples:

- `go-testgen/internal/generator/generator_test.go:59-65` — verifying test source code has `type FooFn func(`, `var checkFoo = func(`, `func TestFoo(t *testing.T)`  
- `go-testgen/internal/generator/gen_table_test.go:17-23` — table generator output must contain `want string`, `wantErr string`, `assert.Equal`  

## Code

```go
type ReportRow struct {
	Label string
	Unit  string
	Value int
}

func FormatTable(rows []ReportRow) string {
	// produces formatted table:
	// LABEL       | VALUE | UNIT
	// ------------+-------+------
	// CPU         |    45 | %
}

func FormatCSV(rows []ReportRow) string {
	// produces CSV: CPU,45,%
}
```

## Test

```go
type checkFormatTableFn func(*testing.T, string)

var checkFormatTable = func(fns ...checkFormatTableFn) []checkFormatTableFn { return fns }

func TestFormatTable(t *testing.T) {
	checkContains := func(want string) checkFormatTableFn {
		return func(t *testing.T, got string) {
			t.Helper()
			assert.Contains(t, got, want)
		}
	}

	checkNotContains := func(want string) checkFormatTableFn {
		return func(t *testing.T, got string) {
			t.Helper()
			assert.NotContains(t, got, want)
		}
	}

	tests := []struct {
		name   string
		rows   []ReportRow
		checks []checkFormatTableFn
	}{
		{
			name:   "empty rows",
			rows:   nil,
			checks: checkFormatTable(
				checkContains("(empty)"),
				checkNotContains("LABEL"),   // header not printed for empty
			),
		},
		{
			name: "single row",
			rows: []ReportRow{{Label: "CPU", Value: 45, Unit: "%"}},
			checks: checkFormatTable(
				checkContains("CPU"),
				checkContains("45"),
				checkContains("%"),
				checkContains("LABEL"),
				checkNotContains("MEM"),
			),
		},
		{
			name: "multiple rows",
			rows: []ReportRow{
				{Label: "CPU", Value: 45, Unit: "%"},
				{Label: "MEM", Value: 2048, Unit: "MB"},
			},
			checks: checkFormatTable(
				checkContains("CPU"),
				checkContains("MEM"),
				checkContains("45"),
				checkContains("2048"),
			),
		},
	}
}

func TestFormatCSV(t *testing.T) {
	// CSV uses exact matching (want string)
	tests := []struct {
		name string
		rows []ReportRow
		want string
	}{
		{name: "empty", rows: nil, want: ""},
		{name: "single", rows: []ReportRow{{Label: "CPU", Value: 45, Unit: "%"}}, want: "CPU,45,%\n"},
		{name: "multiple", rows: []ReportRow{
			{Label: "CPU", Value: 45, Unit: "%"},
			{Label: "MEM", Value: 2048, Unit: "MB"},
		}, want: "CPU,45,%\nMEM,2048,MB\n"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FormatCSV(tt.rows)
			assert.Equal(t, tt.want, got)
		})
	}
}
```

## Testing Approach

String output inspection:

1. **`Contains` for tables, `Equal` for structured formats** — table output includes formatting (borders, spacing, alignment) that shouldn't be exact-matched; `Contains` verifies data presence. CSV has a defined structure where exact matching makes sense.
2. **`NotContains` for negative verification** — "empty input should not produce a header row" is tested via `checkNotContains("LABEL")`. This catches logic errors where the function produces output when it shouldn't.
3. **Resilient to formatting changes** — if the table column width changes from 10 to 12 characters, `Contains` assertions still pass. Only the `Equal`-based CSV test needs updating. Use the right assertion for the right format.
4. **Code generation testing** — this pattern is essential for testing code generators (like `go-testgen`). Instead of generating and compiling full files, verify the generated source contains the expected declarations and patterns.
