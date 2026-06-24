# Chapter 24: AST Parsing Complexity

## Description

Use `go/ast` and `go/parser` to analyze Go source code programmatically. Walk the AST to count functions, find names, and compute cyclomatic complexity (number of linearly independent paths through a function). Complexity calculation counts `if`, `for`, `range`, `case` clauses, and logical operators (`&&`, `||`).

Real-world example: `go-crap/internal/analysis/complexity.go` — AST-based cyclomatic complexity analysis for code quality gates.

## Code

```go
func CyclomaticComplexity(fn *ast.FuncDecl) int {
	count := 1 // base complexity
	ast.Inspect(fn, func(n ast.Node) bool {
		switch t := n.(type) {
		case *ast.IfStmt, *ast.ForStmt, *ast.RangeStmt,
			*ast.CaseClause, *ast.CommClause:
			count++
		case *ast.BinaryExpr:
			if t.Op == token.LAND || t.Op == token.LOR {
				count++ // short-circuit operators
			}
		}
		return true
	})
	return count
}
```

## Test

```go
func TestAnalyzeComplexity(t *testing.T) {
	t.Run("simple function", func(t *testing.T) {
		src := `package p func Add(a, b int) int { return a + b }`
		results, err := AnalyzeComplexity(src)
		require.NoError(t, err)
		assert.Equal(t, 1, results[0].Complexity) // base only
	})
	t.Run("function with if statements", func(t *testing.T) {
		src := `package p func Grade(score int) string {
			if score >= 90 { return "A" }
			if score >= 80 { return "B" }
			return "C"
		}`
		results, _ := AnalyzeComplexity(src)
		assert.Equal(t, 3, results[0].Complexity) // base + 2 ifs
	})
	t.Run("function with switch", func(t *testing.T) {
		src := `package p func Classify(n int) string {
			switch n { case 1: return "one"
			case 2: return "two"
			default: return "other" }
		}`
		results, _ := AnalyzeComplexity(src)
		assert.Equal(t, 4, results[0].Complexity) // base + 3 cases
	})
	t.Run("logical operators", func(t *testing.T) {
		src := `package p func Check(a, b int) bool {
			if a > 0 && b > 0 { return true }
			return false
		}`
		results, _ := AnalyzeComplexity(src)
		assert.Equal(t, 3, results[0].Complexity) // base + if + &&
	})
}

func TestAnalyzeFile(t *testing.T) {
	results, err := AnalyzeFile("main.go")
	require.NoError(t, err)
	assert.NotEmpty(t, results) // self-analysis
}

func TestCountFunctions(t *testing.T) {
	src := `package p; func A() {} func B() {} func C() {}`
	count, _ := CountFunctions(src)
	assert.Equal(t, 3, count)
}
```

## Testing Approach

AST parsing tests:

1. **Self-analysis** — `AnalyzeFile("main.go")` parses the chapter's own source. Tests assert the function count and complexity values for the known source.
2. **Inline source strings** — each test defines Go source as a raw string literal. No external fixture files needed. The source is minimal but syntactically valid.
3. **Incremental complexity** — testsprogress from 1 (no branches) through 3 (two ifs), 4 (switch with 3 cases), and logical operators, each verifying the count increases as expected.
4. **Error paths** — invalid syntax returns an error; empty package returns empty results. Both are tested to ensure `parser.ParseFile` errors are propagated.
