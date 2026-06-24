# Closure-Check

Five chapters constructing a composable assertion system. Starting from typed check functions, the pattern evolves through collection builders and factory closures into a reusable testing framework.

## Chapters
- [chapter-05-typed-check-functions.md](chapter-05-typed-check-functions.md) — Typed Check Functions — `type checkFn func(t, result, error)` enables composable, reusable assertion blocks  
  Source: `chapter-05-typed-check-functions/`
- [chapter-06-check-collection-builder.md](chapter-06-check-collection-builder.md) — Check Collection Builder — `var check = func(fns...)` collects multiple check functions into a single slice  
  Source: `chapter-06-check-collection-builder/`
- [chapter-07-check-factory-closures.md](chapter-07-check-factory-closures.md) — Check Factory Closures — `checkStatus(want)` returns a closure that captures the expected value  
  Source: `chapter-07-check-factory-closures/`
- [chapter-08-error-message-verification.md](chapter-08-error-message-verification.md) — Error Message Verification — `assert.Contains(err.Error(), "substring")` for error message inspection  
  Source: `chapter-08-error-message-verification/`
- [chapter-09-output-string-inspection.md](chapter-09-output-string-inspection.md) — Output String Inspection — `assert.Contains/NotContains` on string output for content verification  
  Source: `chapter-09-output-string-inspection/`

## Running the code

Each chapter is a standalone Go module. To run tests for a chapter:

```bash
cd <source-directory>
go test -v ./...
```

