# Chapter 30: Composable Check Navigation for Nested Output

## Description

When a production function returns a nested struct (`Report → Entries[] → Details[]`), flat check functions either force repetitive extraction code in each test case or wide assertions with hard-coded indices. *Composable navigator checks* solve this: a check factory selects a sub-element (e.g. `Entries[i]`) and delegates to sub-checks that assert against that level. The same pattern recurs one level deeper (`Details[i]` inside an entry). Each navigator handles bounds checking once; leaf checks stay focused on a single field.

Real-world example: `go-crap/internal/report/json_test.go:24-67` — `checkReportEntries`, `checkEntryMutationDetail`, and their sub-check factories.

## Code

```go
type Report struct {
	Schema  string
	Version string
	Entries []Entry
}

type Entry struct {
	File     string
	Package  string
	Function string
	Score    float64
	Details  []Detail
}

type Detail struct {
	Type   string
	Line   int
	Status string
}

func Analyze(data string) (*Report, error) {
	return &Report{
		Schema:  "https://example.com/report-v1.json",
		Version: "1.0.0",
		Entries: []Entry{
			{File: "/main.go", Package: "main", Function: "Run",
				Score: 12.5, Details: []Detail{
					{Type: "complexity", Line: 15, Status: "fail"},
					{Type: "coverage", Line: 20, Status: "pass"},
				}},
			{File: "/util.go", Package: "main", Function: "Helper",
				Score: 3.0, Details: []Detail{}},
		},
	}, nil
}
```

## Test

```go
// ── Level 0: top-level check type ──────────────────────

type checkAnalyzeFn func(*testing.T, *Report, error)

var checkAnalyze = func(fns ...checkAnalyzeFn) []checkAnalyzeFn { return fns }

// ── Level 1: navigator into Entries[i] ──────────────────

type checkEntryFn func(*testing.T, Entry)

func checkReportEntry(i int, fns ...checkEntryFn) checkAnalyzeFn {
	return func(t *testing.T, r *Report, err error) {
		t.Helper()
		if assert.GreaterOrEqualf(t, len(r.Entries), i+1,
			"Report has enough entries at index %d", i) {
			entry := r.Entries[i]
			for _, fn := range fns {
				fn(t, entry)
			}
		}
	}
}

// ── Level 2: navigator into Details[i] inside an Entry ──

type checkDetailFn func(*testing.T, Detail)

func checkEntryDetail(i int, fns ...checkDetailFn) checkEntryFn {
	return func(t *testing.T, e Entry) {
		t.Helper()
		if assert.GreaterOrEqualf(t, len(e.Details), i+1,
			"Entry has enough details at index %d", i) {
			d := e.Details[i]
			for _, fn := range fns {
				fn(t, d)
			}
		}
	}
}

// ── Leaf factories ──────────────────────────────────────

func checkSchema(want string) checkAnalyzeFn { ... }
func checkEntryFile(want string) checkEntryFn { ... }
func checkDetailType(want string) checkDetailFn { ... }

func TestAnalyze(t *testing.T) {
	tests := []struct {
		name   string
		data   string
		checks []checkAnalyzeFn
	}{
		{
			name: "success_entry_details",
			checks: checkAnalyze(
				checkNoError(),
				checkReportEntry(0,
					checkEntryFile("/main.go"),
					checkEntryPackage("main"),
					checkEntryFunction("Run"),
					checkEntryScore(12.5),
				),
			),
		},
		{
			name: "success_nested_details",
			checks: checkAnalyze(
				checkNoError(),
				checkReportEntry(0,
					checkEntryDetailsLen(2),
					checkEntryDetail(0,
						checkDetailType("complexity"),
						checkDetailLine(15),
						checkDetailStatus("fail"),
					),
					checkEntryDetail(1,
						checkDetailType("coverage"),
						checkDetailLine(20),
						checkDetailStatus("pass"),
					),
				),
			),
		},
		{
			name: "success_entry_without_details",
			checks: checkAnalyze(
				checkNoError(),
				checkReportEntry(1,
					checkEntryFile("/util.go"),
					checkEntryFunction("Helper"),
					checkEntryScore(3.0),
					checkEntryDetailsLen(0),
				),
			),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, err := Analyze(tt.data)
			for _, c := range tt.checks {
				c(t, r, err)
			}
		})
	}
}
```

## Testing Approach

Composable check navigation:

1. **Navigator factories** — `checkReportEntry(i, ...fns)` and `checkEntryDetail(i, ...fns)` return a check at their own level, but internally they descend one layer, select the i-th element, and run the sub-checks against it. Each navigator guards against out-of-bounds with `assert.GreaterOrEqual` before indexing, producing a clear failure message instead of a panic.

2. **Three-level type hierarchy** — `checkAnalyzeFn` → `checkEntryFn` → `checkDetailFn`. Each level has its own function signature matching the struct it asserts against. The navigators cross levels: `checkReportEntry` is a `checkAnalyzeFn` that calls `checkEntryFn` sub-checks.

3. **Composition over flat assertions** — without navigators, a test like `success_nested_details` would need inline code to extract `r.Entries[0].Details[0].Type` and run four separate assertions, repeated for each entry/detail pair. With navigators, the intent is declarative: "at entry 0, at detail 0, check these fields."

4. **Refactoring pattern** — the file keeps commented-out `func TestXxx(t *testing.T)` tests that each covered one slice element inline. The replacing table cases reference the original via `// replaces TestAnalyze_xxx` comments. This preserves the migration path for readers who want to see the "before" state.

