package ast_parsing_complexity

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
)

type ComplexityResult struct {
	Position   token.Position
	FuncName   string
	Complexity int
}

func AnalyzeComplexity(src string) ([]ComplexityResult, error) {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "", src, parser.ParseComments)
	if err != nil {
		return nil, fmt.Errorf("parsing source: %w", err)
	}

	var results []ComplexityResult
	for _, decl := range f.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok {
			continue
		}
		c := CyclomaticComplexity(fn)
		results = append(results, ComplexityResult{
			FuncName:   fn.Name.Name,
			Complexity: c,
			Position:   fset.Position(fn.Pos()),
		})
	}
	return results, nil
}

func CyclomaticComplexity(fn *ast.FuncDecl) int {
	count := 1 // base complexity
	ast.Inspect(fn, func(n ast.Node) bool {
		switch t := n.(type) {
		case *ast.IfStmt, *ast.ForStmt, *ast.RangeStmt,
			*ast.CaseClause, *ast.CommClause:
			count++
		case *ast.BinaryExpr:
			if t.Op == token.LAND || t.Op == token.LOR {
				count++
			}
		}
		return true
	})
	return count
}

func AnalyzeFile(path string) ([]ComplexityResult, error) {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
	if err != nil {
		return nil, fmt.Errorf("parsing file: %w", err)
	}

	var results []ComplexityResult
	for _, decl := range f.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok {
			continue
		}
		c := CyclomaticComplexity(fn)
		results = append(results, ComplexityResult{
			FuncName:   fn.Name.Name,
			Complexity: c,
			Position:   fset.Position(fn.Pos()),
		})
	}
	return results, nil
}

func CountFunctions(src string) (int, error) {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "", src, parser.ParseComments)
	if err != nil {
		return 0, fmt.Errorf("parsing source: %w", err)
	}
	count := 0
	for _, decl := range f.Decls {
		if _, ok := decl.(*ast.FuncDecl); ok {
			count++
		}
	}
	return count, nil
}

func FindFunctionNames(src string) ([]string, error) {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "", src, parser.ParseComments)
	if err != nil {
		return nil, fmt.Errorf("parsing source: %w", err)
	}
	var names []string
	for _, decl := range f.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if ok {
			names = append(names, fn.Name.Name)
		}
	}
	return names, nil
}
