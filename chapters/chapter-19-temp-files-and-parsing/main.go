package temp_files_and_parsing

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"go/types"
	"strings"
)

type Field struct {
	Name string
	Type string
}

type entityCtorInfo struct {
	newArgs      string  // pre-joined argument list for New<entity>(...)
	createFields []Field // fields for Create<entity>Input (the non-auto params)
	needsTime    bool    // generated code references time.*
	needsUUID    bool    // generated code references uuid.*
}

func entityConstructorInfo(entityFile, entityName string) (*entityCtorInfo, error) {
	f, err := parser.ParseFile(token.NewFileSet(), entityFile, nil, 0)
	if err != nil {
		return nil, fmt.Errorf("parsing entity file %s: %w", entityFile, err)
	}

	constructorName := "New" + entityName
	params, found := extractConstructorParams(f, constructorName)
	if !found {
		return nil, fmt.Errorf("constructor %s not found in %s", constructorName, entityFile)
	}

	info := &entityCtorInfo{}
	for _, p := range params {
		switch {
		case isIDParam(p.name) && (p.typ == "string" || p.typ == "uuid.UUID"):
			if p.typ == "uuid.UUID" {
				info.newArgsOrAppend("uuid.New()")
			} else {
				info.newArgsOrAppend("uuid.NewString()")
			}
			info.needsUUID = true
		case p.typ == "time.Time":
			info.newArgsOrAppend("time.Now()")
			info.needsTime = true
		default:
			fieldName := toTitleCase(p.name)
			info.createFields = append(info.createFields, Field{Name: fieldName, Type: p.typ})
			info.newArgsOrAppend("input." + fieldName)
			if strings.Contains(p.typ, "time.") {
				info.needsTime = true
			}
			if strings.Contains(p.typ, "uuid.") {
				info.needsUUID = true
			}
		}
	}
	return info, nil
}

type namedParam struct {
	name string
	typ  string
}

func extractConstructorParams(f *ast.File, funcName string) ([]namedParam, bool) {
	for _, decl := range f.Decls {
		fd, ok := decl.(*ast.FuncDecl)
		if !ok || fd.Name.Name != funcName || fd.Type.Params == nil {
			continue
		}
		var params []namedParam
		for _, pf := range fd.Type.Params.List {
			typ := types.ExprString(pf.Type)
			names := pf.Names
			if len(names) == 0 {
				names = []*ast.Ident{{Name: "_"}}
			}
			for _, n := range names {
				params = append(params, namedParam{name: n.Name, typ: typ})
			}
		}
		return params, true
	}
	return nil, false
}

func isIDParam(name string) bool {
	return strings.EqualFold(name, "id") || strings.EqualFold(name, "uuid")
}

func (c *entityCtorInfo) newArgsOrAppend(arg string) {
	if c.newArgs == "" {
		c.newArgs = arg
	} else {
		c.newArgs += ", " + arg
	}
}

func toTitleCase(s string) string {
	if len(s) == 0 {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}
