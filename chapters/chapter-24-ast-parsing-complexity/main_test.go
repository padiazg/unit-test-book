package ast_parsing_complexity

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAnalyzeComplexity(t *testing.T) {
	t.Run("simple function", func(t *testing.T) {
		src := `package p
func Add(a, b int) int { return a + b }`
		results, err := AnalyzeComplexity(src)
		require.NoError(t, err)
		require.Len(t, results, 1)
		assert.Equal(t, "Add", results[0].FuncName)
		assert.Equal(t, 1, results[0].Complexity)
	})

	t.Run("function with if statements", func(t *testing.T) {
		src := `package p
func Grade(score int) string {
	if score >= 90 { return "A" }
	if score >= 80 { return "B" }
	return "C"
}`
		results, err := AnalyzeComplexity(src)
		require.NoError(t, err)
		require.Len(t, results, 1)
		assert.Equal(t, 3, results[0].Complexity)
	})

	t.Run("function with switch", func(t *testing.T) {
		src := `package p
func Classify(n int) string {
	switch n {
	case 1: return "one"
	case 2: return "two"
	default: return "other"
	}
}`
		results, err := AnalyzeComplexity(src)
		require.NoError(t, err)
		require.Len(t, results, 1)
		// base + 3 case clauses (1, 2, default)
		assert.Equal(t, 4, results[0].Complexity)
	})

	t.Run("function with for loop", func(t *testing.T) {
		src := `package p
func Sum(nums []int) int {
	s := 0
	for _, n := range nums { s += n }
	return s
}`
		results, err := AnalyzeComplexity(src)
		require.NoError(t, err)
		require.Len(t, results, 1)
		assert.Equal(t, 2, results[0].Complexity) // base + range
	})

	t.Run("complex function", func(t *testing.T) {
		src := `package p
func Validate(u struct{ Name, Email string }) []string {
	var errs []string
	if u.Name == "" { errs = append(errs, "name required") }
	if u.Email == "" { errs = append(errs, "email required") }
	for _, e := range errs { _ = e }
	return errs
}`
		results, err := AnalyzeComplexity(src)
		require.NoError(t, err)
		require.Len(t, results, 1)
		assert.Equal(t, 4, results[0].Complexity) // base + 2 if + 1 range
	})
}

func TestAnalyzeFile(t *testing.T) {
	t.Run("self-analysis", func(t *testing.T) {
		results, err := AnalyzeFile("main.go")
		require.NoError(t, err)
		assert.NotEmpty(t, results)
		for _, r := range results {
			assert.GreaterOrEqual(t, r.Complexity, 1)
		}
	})

	t.Run("non-existent file", func(t *testing.T) {
		_, err := AnalyzeFile("/nonexistent/file.go")
		assert.Error(t, err)
	})
}

func TestAnalyzeFile_WithTempFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.go")
	err := os.WriteFile(path, []byte(`package p
func Foo() int {
	if true { return 1 }
	return 2
}`), 0644)
	require.NoError(t, err)

	results, err := AnalyzeFile(path)
	require.NoError(t, err)
	require.Len(t, results, 1)
	assert.Equal(t, "Foo", results[0].FuncName)
	assert.Equal(t, 2, results[0].Complexity)
}

func TestCountFunctions(t *testing.T) {
	src := `package p
func A() {}
func B() {}
func C() {}`
	count, err := CountFunctions(src)
	require.NoError(t, err)
	assert.Equal(t, 3, count)
}

func TestFindFunctionNames(t *testing.T) {
	src := `package p
func Alpha() {}
func Beta() {}`
	names, err := FindFunctionNames(src)
	require.NoError(t, err)
	assert.Equal(t, []string{"Alpha", "Beta"}, names)
}

func TestAnalyzeComplexity_InvalidSyntax(t *testing.T) {
	_, err := AnalyzeComplexity(`package p func {`)
	assert.Error(t, err)
}

func TestAnalyzeComplexity_Empty(t *testing.T) {
	results, err := AnalyzeComplexity("package p")
	require.NoError(t, err)
	assert.Empty(t, results)
}

func TestCyclomaticComplexity_IfWithLogicalOps(t *testing.T) {
	src := `package p
func Check(a, b int) bool {
	if a > 0 && b > 0 { return true }
	return false
}`
	results, err := AnalyzeComplexity(src)
	require.NoError(t, err)
	require.Len(t, results, 1)
	// base + 1 if + 1 logical AND
	assert.Equal(t, 3, results[0].Complexity)
}
