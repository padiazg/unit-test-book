# Closure-Check

Five chapters constructing a composable assertion system. Starting from typed check functions, the pattern evolves through collection builders and factory closures into a reusable testing framework.

## Chapters
- [chapter-05-typed-check-functions.md](chapter-05-typed-check-functions.md) -- Typed Check Functions — `type checkFn func(t, result, error)` enables composable, reusable assertion blocks
- [chapter-06-check-collection-builder.md](chapter-06-check-collection-builder.md) -- Check Collection Builder — `var check = func(fns...)` collects multiple check functions into a single slice
- [chapter-07-check-factory-closures.md](chapter-07-check-factory-closures.md) -- Check Factory Closures — `checkStatus(want)` returns a closure that captures the expected value
- [chapter-08-error-message-verification.md](chapter-08-error-message-verification.md) -- Error Message Verification — `assert.Contains(err.Error(), "substring")` for error message inspection
- [chapter-09-output-string-inspection.md](chapter-09-output-string-inspection.md) -- Output String Inspection — `assert.Contains/NotContains` on string output for content verification

