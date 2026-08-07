# Chapter 01: Classic Table-Driven Tests

## Description

A method that iterates a slice and checks each element against a threshold is tested by enumerating empty lists, all-below, at-boundary, one-exceeding, all-above, and single-entry cases. Each case uses a struct with `name`, `entries`, `threshold`, and `want bool` to drive a `t.Run` subtest.

Real-world example:

- `go-crap/internal/scan/entries_test.go:438` — `TestEntries_ThresholdExceeded`  

## Code

```go
package classic_table_driven

// CRAPEntry represents a function with its computed CRAP score.
type CRAPEntry struct {
    FuncName      string
    EffectiveCRAP float64
}

// Entries wraps a list of CRAP entries.
type Entries struct {
    List []CRAPEntry
}

// ThresholdExceeded returns true if any entry exceeds the given threshold.
func (e *Entries) ThresholdExceeded(threshold float64) bool {
    for _, entry := range e.List {
        if entry.EffectiveCRAP > threshold {
            return true
        }
    }
    return false
}
```

## Test

```go
func TestEntries_ThresholdExceeded(t *testing.T) {
    const fn = "ThresholdExceeded"
    tests := []struct {
        name      string
        entries   *Entries
        threshold float64
        want      bool
    }{
        {
            name:      "empty_list_returns_false",
            entries:   &Entries{List: []CRAPEntry{}},
            threshold: 100.0,
            want:      false,
        },
        {
            name: "all_below_threshold",
            entries: &Entries{List: []CRAPEntry{
                {FuncName: "a", EffectiveCRAP: 50},
                {FuncName: "b", EffectiveCRAP: 30},
            }},
            threshold: 100.0,
            want:      false,
        },
        {
            name: "all_at_threshold_not_exceeded",
            entries: &Entries{List: []CRAPEntry{
                {FuncName: "exact", EffectiveCRAP: 100},
            }},
            threshold: 100.0,
            want:      false,
        },
        {
            name: "one_entry_exceeding_threshold",
            entries: &Entries{List: []CRAPEntry{
                {FuncName: "low", EffectiveCRAP: 30},
                {FuncName: "high", EffectiveCRAP: 200},
            }},
            threshold: 100.0,
            want:      true,
        },
        {
            name: "all_above_threshold",
            entries: &Entries{List: []CRAPEntry{
                {FuncName: "a", EffectiveCRAP: 200},
                {FuncName: "b", EffectiveCRAP: 150},
            }},
            threshold: 100.0,
            want:      true,
        },
        {
            name: "single_entry_below_threshold",
            entries: &Entries{List: []CRAPEntry{
                {FuncName: "single", EffectiveCRAP: 50},
            }},
            threshold: 100.0,
            want:      false,
        },
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got := tt.entries.ThresholdExceeded(tt.threshold)
            assert.Equal(t, tt.want, got, "%s: %s", fn, tt.name)
        })
    }
}
```

## Scaffold

Generate test scaffolding with go-testgen:

```shell
go-testgen report . --format table
go-testgen gen . Entries.ThresholdExceeded
```

## Testing Approach

The table-driven `ThresholdExceeded` test covers:

1. **Empty list** — `&Entries{List: []CRAPEntry{}}` should return `false`; no entries means nothing can exceed the threshold.
2. **All below** — when every entry's `EffectiveCRAP` is strictly less than the threshold, the result is `false`.
3. **At boundary** — `EffectiveCRAP == threshold` is NOT exceeded (the check uses `>`, not `>=`). This is a critical boundary case.
4. **One exceeding** — only one entry above the threshold among several below; the method should short-circuit and return `true`.
5. **All above** — every entry exceeds the threshold; confirms the method works when all values pass.
6. **Single entry** — a list with one item below threshold; verifies the loop runs at least once and returns `false` correctly.

The `const fn = "ThresholdExceeded"` plus `%s: %s` format in `assert.Equal` makes test failure output self-documenting: `ThresholdExceeded: all_at_threshold_not_exceeded`.
