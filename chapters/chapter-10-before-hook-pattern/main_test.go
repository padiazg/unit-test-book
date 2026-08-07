package before_hook_pattern

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

type mockGoListRunner struct {
	listFn func(ctx context.Context, patterns []string) ([]byte, error)
}

func (m *mockGoListRunner) List(ctx context.Context, patterns []string) ([]byte, error) {
	return m.listFn(ctx, patterns)
}

type checkResolverResolveFn func(*testing.T, []Target, error)

var checkResolverResolve = func(fns ...checkResolverResolveFn) []checkResolverResolveFn { return fns }

func checkResolveError(want string) checkResolverResolveFn {
	return func(t *testing.T, _ []Target, err error) {
		t.Helper()
		if want == "" {
			assert.NoError(t, err)
			return
		}
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

func checkTargetImportPath(want string) checkResolverResolveFn {
	return func(t *testing.T, got []Target, err error) {
		t.Helper()
		assert.NoError(t, err)
		if assert.Len(t, got, 1) {
			assert.Equal(t, want, got[0].ImportPath)
		}
	}
}

func TestResolver_Resolve(t *testing.T) {
	sampleJSON := `{"ImportPath":"github.com/example/pkg","Dir":"/go/src/github.com/example/pkg","GoFiles":["foo.go"]}`

	tests := []struct {
		name   string
		before func(*Resolver)
		checks []checkResolverResolveFn
	}{
		{
			name: "success",
			before: func(r *Resolver) {
				r.config.Runner = &mockGoListRunner{
					listFn: func(ctx context.Context, patterns []string) ([]byte, error) {
						return []byte(sampleJSON), nil
					},
				}
			},
			checks: checkResolverResolve(
				checkTargetsCount(1),
				checkTargetImportPath("github.com/example/pkg"),
			),
		},
		{
			name: "empty patterns",
			before: func(r *Resolver) {
				r.config.Runner = &mockGoListRunner{
					listFn: func(ctx context.Context, patterns []string) ([]byte, error) {
						return []byte{}, nil
					},
				}
			},
			checks: checkResolverResolve(
				checkTargetsCount(0),
			),
		},
		{
			name: "runner returns error",
			before: func(r *Resolver) {
				r.config.Runner = &mockGoListRunner{
					listFn: func(ctx context.Context, patterns []string) ([]byte, error) {
						return nil, assert.AnError
					},
				}
			},
			checks: checkResolverResolve(
				checkResolveError("assert.AnError"),
			),
		},
		{
			name: "package with error field",
			before: func(r *Resolver) {
				badJSON := `{"ImportPath":"github.com/bad/pkg","Error":{"Err":"build failed"}}`
				r.config.Runner = &mockGoListRunner{
					listFn: func(ctx context.Context, patterns []string) ([]byte, error) {
						return []byte(badJSON), nil
					},
				}
			},
			checks: checkResolverResolve(
				checkResolveError("resolve patterns"),
			),
		},
		{
			name: "include test files",
			before: func(r *Resolver) {
				r.config.Runner = &mockGoListRunner{
					listFn: func(ctx context.Context, patterns []string) ([]byte, error) {
						return []byte(`{"ImportPath":"github.com/example/pkg","Dir":"/go/src/github.com/example/pkg","GoFiles":["foo.go"],"TestGoFiles":["foo_test.go"],"XTestGoFiles":["foo_test_external.go"]}`), nil
					},
				}
			},
			checks: checkResolverResolve(
				checkTargetsCount(1),
			),
		},
		{
			name: "multiple packages",
			before: func(r *Resolver) {
				multiJSON := `{"ImportPath":"github.com/example/a","Dir":"/go/src/github.com/example/a","GoFiles":["a.go"]}` + "\n" +
					`{"ImportPath":"github.com/example/b","Dir":"/go/src/github.com/example/b","GoFiles":["b.go"]}`
				r.config.Runner = &mockGoListRunner{
					listFn: func(ctx context.Context, patterns []string) ([]byte, error) {
						return []byte(multiJSON), nil
					},
				}
			},
			checks: checkResolverResolve(
				checkTargetsCount(2),
				func(t *testing.T, got []Target, err error) {
					t.Helper()
					assert.NoError(t, err)
					if assert.Len(t, got, 2) {
						assert.Equal(t, "github.com/example/a", got[0].ImportPath)
					}
				},
			),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := NewResolver(&ResolverConfig{})
			if tt.before != nil {
				tt.before(r)
			}

			got, err := r.Resolve(context.Background(), []string{"./..."}, false)
			for _, fn := range tt.checks {
				fn(t, got, err)
			}
		})
	}
}

func TestResolver_BeforeHook(t *testing.T) {
	t.Run("resolves single package", func(t *testing.T) {
		f := before(t, `{"ImportPath":"pkg/a","Dir":"/go/pkg/a","GoFiles":["a.go"]}`)
		r := NewResolver(&ResolverConfig{Runner: f.runner, IncludeTests: false})
		targets, err := r.Resolve(context.Background(), []string{"./..."}, false)
		assert.NoError(t, err)
		assert.Len(t, targets, 1)
	})

	t.Run("returns empty for empty json", func(t *testing.T) {
		f := before(t, "")
		r := NewResolver(&ResolverConfig{Runner: f.runner, IncludeTests: false})
		targets, err := r.Resolve(context.Background(), []string{"./..."}, false)
		assert.NoError(t, err)
		assert.Len(t, targets, 0)
	})

	t.Run("isolated runners produce independent results", func(t *testing.T) {
		f1 := before(t, `{"ImportPath":"pkg/a","Dir":"/go/pkg/a","GoFiles":["a.go"]}`)
		f2 := before(t, `{"ImportPath":"pkg/b","Dir":"/go/pkg/b","GoFiles":["b.go"]}`)

		r1 := NewResolver(&ResolverConfig{Runner: f1.runner, IncludeTests: false})
		r2 := NewResolver(&ResolverConfig{Runner: f2.runner, IncludeTests: false})

		t1, err1 := r1.Resolve(context.Background(), []string{"./..."}, false)
		t2, err2 := r2.Resolve(context.Background(), []string{"./..."}, false)
		assert.NoError(t, err1)
		assert.NoError(t, err2)
		assert.Equal(t, "pkg/a", t1[0].ImportPath)
		assert.Equal(t, "pkg/b", t2[0].ImportPath)
	})
}

type beforeReturns struct {
	runner *mockGoListRunner
}

func before(t *testing.T, json string) beforeReturns {
	t.Helper()
	return beforeReturns{
		runner: &mockGoListRunner{
			listFn: func(ctx context.Context, patterns []string) ([]byte, error) {
				return []byte(json), nil
			},
		},
	}
}
