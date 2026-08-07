---
section: mocking
---

# Chapter 10: Before Hook Pattern

## Description

Extract test setup to a `before` hook function that returns a fresh fixture state. Each test case calls `before()` explicitly instead of a shared `setup()` mutated by every test. This prevents test pollution from shared mutable state (rate limiters, timestamps, sequence generators).

Real-world example: `go-crap/internal/packages/packages_test.go:TestResolver_Resolve` — `mockGoListRunner` returns a fresh mock runner for each test case, isolating JSON responses per subtest.

## Code

```go
type Resolver struct {
    runner GoListRunner
}

type GoListRunner interface {
    List(ctx context.Context, patterns []string) ([]byte, error)
}

func (r *Resolver) Resolve(ctx context.Context, patterns []string, includeTests bool) ([]Target, error) {
    stdout, err := r.runner.List(ctx, patterns)
    if err != nil {
        return nil, err
    }
    // ... parse JSON, build Targets
    return targets, nil
}
```

## Test

```go
func checkResolveError(want string) checkResolverResolveFn {
    return func(t *testing.T, _ []Target, err error) {
        t.Helper()
        if assert.Error(t, err) {
            assert.Contains(t, err.Error(), want)
        }
    }
}

func checkTargetsCount(want int) checkResolverResolveFn {
    return func(t *testing.T, got []Target, err error) {
        t.Helper()
        assert.NoError(t, err)
        assert.Len(t, got, want)
    }
}

func TestResolver_Resolve(t *testing.T) {
    tests := []struct {
        name   string
        before func(*Resolver)
        checks []checkResolverResolveFn
    }{
        {
            name: "success",
            before: func(r *Resolver) {
                r.runner = &mockGoListRunner{
                    listFn: func(ctx context.Context, patterns []string) ([]byte, error) {
                        return []byte(`{"ImportPath":"pkg/a","Dir":"/go/pkg/a","GoFiles":["a.go"]}`), nil
                    },
                }
            },
            checks: checkResolverResolve(
                checkTargetsCount(1),
                checkTargetImportPath("pkg/a"),
            ),
        },
        {
            name: "runner returns error",
            before: func(r *Resolver) {
                r.runner = &mockGoListRunner{
                    listFn: func(context.Context, []string) ([]byte, error) {
                        return nil, assert.AnError
                    },
                }
            },
            checks: checkResolverResolve(
                checkResolveError("assert.AnError"),
            ),
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            r := NewResolver(&ResolverConfig{})
            if tt.before != nil {
                tt.before(r)
            }
            got, err := r.Resolve(ctx, []string{"./..."}, false)
            for _, fn := range tt.checks {
                fn(t, got, err)
            }
        })
    }
}
```

## Scaffold

Generate test scaffolding with go-testgen:

```shell
go-testgen report . --format table
go-testgen gen . Resolver.Resolve
```

## Testing Approach

The before hook pattern:

1. **Fresh state per case** — `before()` returns a new `mockGoListRunner` with the configured JSON response. Each subtest gets its own mock instance. No risk of test A's mock state leaking into test B.
2. **Typed return struct** — `beforeReturns` documents exactly what the test fixture provides. Adding a new dependency (e.g. a clock, a config flag) doesn't require changing every test's setup, just the struct and the `before` function.
3. **Explicit over shared** — no `init()` or `TestMain` setup. Every subtest calls `before()` and sees the fixture it depends on. Makes the test self-documenting.
4. **`t.Helper()`** — the `before` hook calls `t.Helper()`, so failure line numbers point to the test assertion, not inside the setup function.

### Beyond Setup: Before/Checks/After Phases

When tests involve goroutines, mocked dependencies, or explicit teardown, the `before` pattern extends to a three-phase lifecycle:

- **Before** — configure mocks and set up the system under test
- **Exercise** — call the method being tested
- **Checks** — assert on captured output
- **After** — trigger shutdown, cancel context, verify channel cleanup

The tests in Chapter 22 (Sensor) and Chapter 33 (Event-Driven Run Loop) demonstrate this: mock transport expectations are wired in `before()`, the `Run()` channel is observed with `select` + timeout, received values are checked by typed closures, and `after()` calls `s.Stop()` or `cancel()` to tear down the goroutine and verify channel closure.

This structured lifecycle makes it easy to add new test cases: each entry in the table specifies only what differs — mock setup, assertions, and teardown — without repeating the boilerplate of creating contexts, starting goroutines, or draining channels.
