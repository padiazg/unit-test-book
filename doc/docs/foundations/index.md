# Foundations

Four classic table-driven testing patterns. Each chapter builds on the previous, from basic `wantErr` assertions through value inspection and subtest naming conventions.

## Chapters
- [chapter-01-classic-table-driven.md](chapter-01-classic-table-driven.md) — Classic Table-Driven Tests — struct fields define test cases with `wantErr bool` for expected pass/fail outcomes  
  Source: `chapter-01-classic-table-driven/`
- [chapter-02-value-assertions.md](chapter-02-value-assertions.md) — Value Assertions — compare production output against a `want T` value using `assert.Equal` or `assert.InDelta`  
  Source: `chapter-02-value-assertions/`
- [chapter-03-fields-struct-inputs.md](chapter-03-fields-struct-inputs.md) — Fields Struct for Inputs — group related test inputs into a dedicated struct for reusable test vectors  
  Source: `chapter-03-fields-struct-inputs/`
- [chapter-04-subtest-naming.md](chapter-04-subtest-naming.md) — Subtest Naming Strategies — use descriptive constants or natural language as subtest names  
  Source: `chapter-04-subtest-naming/`

## Running the code

Each chapter is a standalone Go module. To run tests for a chapter:

```bash
cd <source-directory>
go test -v ./...
```

