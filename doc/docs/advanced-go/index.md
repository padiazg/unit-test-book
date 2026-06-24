# Advanced Go

Four advanced testing topics: goroutine leak detection with goleak, AST-based cyclomatic complexity analysis, benchmark-driven performance comparison, and parallel test safety with -race.

## Chapters
- [chapter-23-goroutine-leak-detection.md](chapter-23-goroutine-leak-detection.md) — Goroutine Leak Detection — `goleak.VerifyNone(t)` catches leaked goroutines after each test completes  
  Source: `chapter-23-goroutine-leak-detection/`
- [chapter-24-ast-parsing-complexity.md](chapter-24-ast-parsing-complexity.md) — AST Parsing for Cyclomatic Complexity — walk the Go AST with `ast.Inspect` to compute complexity metrics  
  Source: `chapter-24-ast-parsing-complexity/`
- [chapter-25-benchmark-tests.md](chapter-25-benchmark-tests.md) — Benchmark Tests — compare implementations with `b.N` loops, sub-benchmarks, and `b.ResetTimer()`  
  Source: `chapter-25-benchmark-tests/`
- [chapter-26-parallel-tests.md](chapter-26-parallel-tests.md) — Parallel Tests — `t.Parallel()` with safe (mutex/atomic) vs unsafe concurrent access patterns  
  Source: `chapter-26-parallel-tests/`

## Running the code

Each chapter is a standalone Go module. To run tests for a chapter:

```bash
cd <source-directory>
go test -v ./...
```

