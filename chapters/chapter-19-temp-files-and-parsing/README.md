---
section: http-io
---

# Chapter 19: Temp Files and Parsing

## Description

Use `t.TempDir()` + `os.WriteFile` to create on-disk fixture files for functions that read and parse them. The example in this chapter is a Go source parser: tests write an entity constructor file into a temp directory, and production code parses it with `go/parser` / `go/ast` to extract the constructor's parameters. Each case writes its own fixture, and `t.TempDir()` cleans it up automatically — no shared fixtures, no collisions, no leftovers.

Real-world example: `hexago/internal/generator/service_test.go` — `t.TempDir()` + `os.WriteFile` for entity constructor files parsed with `go/parser`.

## Code

```go
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
```

## Test

```go
type entityConstructorInfoFn func(*testing.T, *entityCtorInfo, error)

var checkentityConstructorInfo = func(fns ...entityConstructorInfoFn) []entityConstructorInfoFn { return fns }

func checkentityConstructorInfoError(want string) entityConstructorInfoFn {
    return func(t *testing.T, _ *entityCtorInfo, err error) {
        t.Helper()
        if want == "" {
            assert.NoErrorf(t, err, "checkentityConstructorInfoError: expected no error, got %v", err)
            return
        }
        if assert.Errorf(t, err, "checkentityConstructorInfoError: expected error %q", want) {
            assert.Containsf(t, err.Error(), want, "checkentityConstructorInfoError mismatch")
        }
    }
}

func checkNewArgs(want string) entityConstructorInfoFn {
    return func(t *testing.T, info *entityCtorInfo, _ error) {
        t.Helper()
        assert.Equal(t, want, info.newArgs)
    }
}

func checkCreateFields(want []Field) entityConstructorInfoFn {
    return func(t *testing.T, info *entityCtorInfo, _ error) {
        t.Helper()
        assert.Equal(t, want, info.createFields)
    }
}

func checkNeedsUUID(want bool) entityConstructorInfoFn {
    return func(t *testing.T, info *entityCtorInfo, _ error) {
        t.Helper()
        assert.Equal(t, want, info.needsUUID)
    }
}

func checkNeedsTime(want bool) entityConstructorInfoFn {
    return func(t *testing.T, info *entityCtorInfo, _ error) {
        t.Helper()
        assert.Equal(t, want, info.needsTime)
    }
}

func writeEntityFile(t *testing.T, content string) string {
    t.Helper()
    dir := t.TempDir()
    p := filepath.Join(dir, "entity.go")
    if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
        t.Fatal(err)
    }
    return p
}

func Test_entityConstructorInfo(t *testing.T) {
    tests := []struct {
        name       string
        entityName string
        content    string
        checks     []entityConstructorInfoFn
    }{
        {
            name:       "string id plus extra field",
            entityName: "Category",
            content:    "package categories\nfunc NewCategory(id string, description string) (*Category, error) { return nil, nil }\n",
            checks: checkentityConstructorInfo(
                checkentityConstructorInfoError(""),
                checkNewArgs(`uuid.NewString(), input.Description`),
                checkCreateFields([]Field{{Name: "Description", Type: "string"}}),
                checkNeedsUUID(true),
            ),
        },
        {
            name:       "uuid id and timestamp",
            entityName: "User",
            content:    "package users\nfunc NewUser(id uuid.UUID, createdAt time.Time) (*User, error) { return nil, nil }\n",
            checks: checkentityConstructorInfo(
                checkentityConstructorInfoError(""),
                checkNewArgs("uuid.New(), time.Now()"),
                checkNeedsUUID(true),
                checkNeedsTime(true),
            ),
        },
        {
            name:       "id only",
            entityName: "Note",
            content:    "package notes\nfunc NewNote(id string) (*Note, error) { return nil, nil }\n",
            checks: checkentityConstructorInfo(
                checkentityConstructorInfoError(""),
                checkNewArgs("uuid.NewString()"),
                checkNeedsUUID(true),
            ),
        },
        {
            name:       "no id field",
            entityName: "Greeting",
            content:    "package greetings\nfunc NewGreeting(name string, email string) (*Greeting, error) { return nil, nil }\n",
            checks: checkentityConstructorInfo(
                checkentityConstructorInfoError(""),
                checkNewArgs("input.Name, input.Email"),
                checkCreateFields([]Field{{Name: "Name", Type: "string"}, {Name: "Email", Type: "string"}}),
            ),
        },
        {
            name:       "grouped param names",
            entityName: "Pair",
            content:    "package pairs\nfunc NewPair(a, b string) (*Pair, error) { return nil, nil }\n",
            checks: checkentityConstructorInfo(
                checkentityConstructorInfoError(""),
                checkNewArgs("input.A, input.B"),
                checkCreateFields([]Field{{Name: "A", Type: "string"}, {Name: "B", Type: "string"}}),
            ),
        },
        {
            name:       "pointer and slice fields become input fields",
            entityName: "Order",
            content:    "package orders\nfunc NewOrder(id string, total float64, items []string, at *time.Time) (*Order, error) { return nil, nil }\n",
            checks: checkentityConstructorInfo(
                checkentityConstructorInfoError(""),
                checkNewArgs("uuid.NewString(), input.Total, input.Items, input.At"),
                checkCreateFields([]Field{
                    {Name: "Total", Type: "float64"},
                    {Name: "Items", Type: "[]string"},
                    {Name: "At", Type: "*time.Time"},
                }),
                checkNeedsUUID(true),
                checkNeedsTime(true),
            ),
        },
        {
            name:       "constructor missing",
            entityName: "Category",
            content:    "package categories\n",
            checks: checkentityConstructorInfo(
                checkentityConstructorInfoError("constructor NewCategory not found"),
            ),
        },
    }
    for _, tt := range tests {
        tt := tt
        t.Run(tt.name, func(t *testing.T) {
            path := writeEntityFile(t, tt.content)
            r, err := entityConstructorInfo(path, tt.entityName)
            for _, c := range tt.checks {
                c(t, r, err)
            }
        })
    }
}

func Test_entityConstructorInfo_missingFile(t *testing.T) {
    _, err := entityConstructorInfo(filepath.Join(t.TempDir(), "nope.go"), "Category")
    assert.Error(t, err)
}
```

## Scaffold

Generate test scaffolding with go-testgen:

```shell
go-testgen report . --format table
go-testgen gen . entityConstructorInfo
```

## Testing Approach

Temp files and parsing:

1. **Per-case fixtures with `t.TempDir()`** — the `writeEntityFile` helper writes each case's `entity.go` into a fresh temp directory that Go removes automatically. No shared fixtures, no `os.RemoveAll`, no leftover files between runs.
2. **Write, then parse** — the fixture content is raw source text (a constructor signature), written to a real file on disk. `entityConstructorInfo` exercises the full `go/parser` + `go/ast` path against an actual file, including the file-not-found error case.
3. **Signature-driven table cases** — the seven rows cover every decision branch of the parser: string id → `uuid.NewString()`, `uuid.UUID` id → `uuid.New()`, `time.Time` → `time.Now()`, grouped parameter names (`a, b string`), pointer and slice types, and a file with no constructor at all.
4. **Distinct error paths** — a missing file (parse failure) and a missing constructor (AST lookup failure) are separate cases with separate checks, so a regression in either branch is reported at the exact line.
