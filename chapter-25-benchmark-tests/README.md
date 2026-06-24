# Chapter 25: Benchmark Tests

## Description

Use Go's built-in `testing.B` to write benchmarks that measure performance. Compare implementations (string concatenation, sorting, parsing) to make data-driven decisions. Benchmarks use `b.N` iterations, `b.ResetTimer()` to exclude setup, and sub-benchmarks with `b.Run()` for comparison.

Real-world example: `go-testgen/internal/generator/template_test.go` — benchmarks comparing text/template vs handwritten builders.

## Code

```go
func ConcatJoin(items []string) string {
	return strings.Join(items, ",")
}

func ConcatPlus(items []string) string {
	out := ""
	for _, s := range items { out += s + "," }
	if len(out) > 0 { return out[:len(out)-1] }
	return out
}

func ConcatBuilder(items []string) string {
	var b strings.Builder
	for i, s := range items {
		if i > 0 { b.WriteString(",") }
		b.WriteString(s)
	}
	return b.String()
}

var sink interface{}
```

## Test

```go
func BenchmarkConcatJoin(b *testing.B) {
	items := []string{"alpha", "beta", "gamma", "delta", "epsilon"}
	for i := 0; i < b.N; i++ { ConcatJoin(items) }
}

func BenchmarkConcatPlus(b *testing.B) {
	items := []string{"alpha", "beta", "gamma", "delta", "epsilon"}
	for i := 0; i < b.N; i++ { ConcatPlus(items) }
}

func BenchmarkConcatBuilder(b *testing.B) {
	items := []string{"alpha", "beta", "gamma", "delta", "epsilon"}
	for i := 0; i < b.N; i++ { ConcatBuilder(items) }
}

func BenchmarkSortStdLib(b *testing.B) {
	nums := []int{3, 1, 4, 1, 5, 9, 2, 6, 5, 3}
	b.Run("sort.Ints", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			tmp := make([]int, len(nums)); copy(tmp, nums)
			sort.Ints(tmp)
		}
	})
	b.Run("sort.Slice", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			tmp := make([]int, len(nums)); copy(tmp, nums)
			sort.Slice(tmp, func(i, j int) bool { return tmp[i] < tmp[j] })
		}
	})
}

func BenchmarkLargeConcat(b *testing.B) {
	items := make([]string, 1000)
	for i := range items { items[i] = "value" }
	b.Run("Join", func(b *testing.B) {
		for i := 0; i < b.N; i++ { sink = strings.Join(items, ",") }
	})
	b.Run("Builder", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			var bld strings.Builder
			for j, s := range items {
				if j > 0 { bld.WriteString(",") }
				bld.WriteString(s)
			}
			sink = bld.String()
		}
	})
}

func BenchmarkFilterAdults(b *testing.B) {
	users := make([]User, 1000)
	for i := range users { users[i] = User{Name: "User", Age: i % 30} }
	b.ResetTimer()
	for i := 0; i < b.N; i++ { FilterAdults(users) }
}
```

## Testing Approach

Benchmark tests:

1. **`b.N` iterations** — the framework adjusts `b.N` until the benchmark runs for a stable duration (~1s). Tests should be instrumented with `b.ResetTimer()` after expensive setup.
2. **`var sink interface{}`** — prevents the compiler from optimizing away the result. Assign benchmark output to a package-level `sink` variable so the call isn't eliminated as dead code.
3. **Sub-benchmarks for comparison** — `b.Run("Join", ...)` / `b.Run("Builder", ...)` in the same parent benchmark produce comparable results under `go test -bench=.`. Use `benchstat` for statistical comparison across runs.
4. **Setup outside timer** — create test data before `b.ResetTimer()`. In `BenchmarkFilterAdults`, the 1000-element slice is created and populated before the timer resets, so only the filtering is measured.
