package composable_check_navigation

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// ────────────────────────────────────────────────────────────
// Level 0: check function type for the top-level result
// ────────────────────────────────────────────────────────────

type checkAnalyzeFn func(*testing.T, *Report, error)

var checkAnalyze = func(fns ...checkAnalyzeFn) []checkAnalyzeFn { return fns }

// ────────────────────────────────────────────────────────────
// Level 1: navigator — selects Entries[i] and runs sub-checks
//          against that single entry.
// ────────────────────────────────────────────────────────────

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

// ────────────────────────────────────────────────────────────
// Level 2: navigator — selects Details[i] inside an Entry
//          and runs sub-checks against that single detail.
// ────────────────────────────────────────────────────────────

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

// ────────────────────────────────────────────────────────────
// Leaf-level check factories
// ────────────────────────────────────────────────────────────

func checkSchema(want string) checkAnalyzeFn {
	return func(t *testing.T, r *Report, err error) {
		t.Helper()
		assert.Equal(t, want, r.Schema)
	}
}

func checkVersion(want string) checkAnalyzeFn {
	return func(t *testing.T, r *Report, err error) {
		t.Helper()
		assert.Equal(t, want, r.Version)
	}
}

func checkNoError() checkAnalyzeFn {
	return func(t *testing.T, _ *Report, err error) {
		t.Helper()
		assert.NoError(t, err)
	}
}

func checkEntryFile(want string) checkEntryFn {
	return func(t *testing.T, e Entry) {
		t.Helper()
		assert.Equal(t, want, e.File)
	}
}

func checkEntryPackage(want string) checkEntryFn {
	return func(t *testing.T, e Entry) {
		t.Helper()
		assert.Equal(t, want, e.Package)
	}
}

func checkEntryFunction(want string) checkEntryFn {
	return func(t *testing.T, e Entry) {
		t.Helper()
		assert.Equal(t, want, e.Function)
	}
}

func checkEntryScore(want float64) checkEntryFn {
	return func(t *testing.T, e Entry) {
		t.Helper()
		assert.InDelta(t, want, e.Score, 0.01)
	}
}

func checkEntryDetailsLen(want int) checkEntryFn {
	return func(t *testing.T, e Entry) {
		t.Helper()
		assert.Len(t, e.Details, want)
	}
}

func checkDetailType(want string) checkDetailFn {
	return func(t *testing.T, d Detail) {
		t.Helper()
		assert.Equal(t, want, d.Type)
	}
}

func checkDetailLine(want int) checkDetailFn {
	return func(t *testing.T, d Detail) {
		t.Helper()
		assert.Equal(t, want, d.Line)
	}
}

func checkDetailStatus(want string) checkDetailFn {
	return func(t *testing.T, d Detail) {
		t.Helper()
		assert.Equal(t, want, d.Status)
	}
}

// ────────────────────────────────────────────────────────────
// Tests
// ────────────────────────────────────────────────────────────

func TestAnalyze(t *testing.T) {
	tests := []struct {
		name   string
		data   string
		checks []checkAnalyzeFn
	}{
		{
			name: "success_report_metadata",
			data: "sample",
			checks: checkAnalyze(
				checkNoError(),
				checkSchema("https://example.com/report-v1.json"),
				checkVersion("1.0.0"),
			),
		},
		{
			name: "success_entry_details",
			data: "sample",
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
			data: "sample",
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
			data: "sample",
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

// ────────────────────────────────────────────────────────────────
// Before — standalone tests that the table-driven cases replaced.
// Kept as commented examples of the refactoring pattern:
// scattered func TestXxx(t *testing.T) → one test table with
// composable navigator checks.
//
// Each replacing case references the original via a "replaces"
// comment in the test table above.
// ────────────────────────────────────────────────────────────────

// func TestAnalyze_report_metadata(t *testing.T) {
// 	r, err := Analyze("sample")
// 	assert.NoError(t, err)
// 	assert.Equal(t, "https://example.com/report-v1.json", r.Schema)
// 	assert.Equal(t, "1.0.0", r.Version)
// }

// func TestAnalyze_first_entry(t *testing.T) {
// 	r, err := Analyze("sample")
// 	assert.NoError(t, err)
// 	entry := r.Entries[0]
// 	assert.Equal(t, "/main.go", entry.File)
// 	assert.Equal(t, "main", entry.Package)
// 	assert.Equal(t, "Run", entry.Function)
// 	assert.InDelta(t, 12.5, entry.Score, 0.01)
// }

// func TestAnalyze_first_entry_details(t *testing.T) {
// 	r, err := Analyze("sample")
// 	assert.NoError(t, err)
// 	entry := r.Entries[0]
// 	assert.Len(t, entry.Details, 2)
// 	assert.Equal(t, "complexity", entry.Details[0].Type)
// 	assert.Equal(t, 15, entry.Details[0].Line)
// 	assert.Equal(t, "fail", entry.Details[0].Status)
// 	assert.Equal(t, "coverage", entry.Details[1].Type)
// 	assert.Equal(t, 20, entry.Details[1].Line)
// 	assert.Equal(t, "pass", entry.Details[1].Status)
// }

// func TestAnalyze_second_entry_empty_details(t *testing.T) {
// 	r, err := Analyze("sample")
// 	assert.NoError(t, err)
// 	entry := r.Entries[1]
// 	assert.Equal(t, "/util.go", entry.File)
// 	assert.Equal(t, "Helper", entry.Function)
// 	assert.Empty(t, entry.Details)
// }
