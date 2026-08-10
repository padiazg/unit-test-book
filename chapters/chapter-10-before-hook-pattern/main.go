package before_hook_pattern

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
)

// GoListRunner runs `go list -json -e` and returns stdout bytes.
type GoListRunner interface {
	List(ctx context.Context, patterns []string) ([]byte, error)
}

// goListRunner is the production GoListRunner: it invokes the go command.
type goListRunner struct{}

func (goListRunner) List(ctx context.Context, patterns []string) ([]byte, error) {
	args := append([]string{"list", "-json", "-e"}, patterns...)
	return exec.CommandContext(ctx, "go", args...).Output()
}

// Target represents a single Go package resolved from a pattern.
type Target struct {
	Dir        string
	ImportPath string
	ModuleDir  string
	ModulePath string
	Files      []string
	TestFiles  []string
}

type packageError struct {
	ImportPath string
	Pos        string
	Err        string
}

// listPackage is a minimal representation of go list -json output for a package.
type listPackage struct {
	Error        packageError
	Dir          string
	ImportPath   string
	ModuleDir    string
	ModulePath   string
	GoFiles      []string
	TestGoFiles  []string
	XTestGoFiles []string
}

// ResolverConfig holds configuration for the Resolver.
type ResolverConfig struct {
	Runner       GoListRunner
	IncludeTests bool
}

// Resolver resolves package patterns via `go list -json -e`.
type Resolver struct {
	config *ResolverConfig
}

// NewResolver creates a Resolver with the given options.
func NewResolver(config *ResolverConfig) *Resolver {
	if config == nil {
		config = &ResolverConfig{}
	}
	if config.Runner == nil {
		config.Runner = goListRunner{}
	}
	return &Resolver{config: config}
}

// Resolve resolves package patterns via `go list -json -e`.
func (r *Resolver) Resolve(ctx context.Context, patterns []string, includeTests bool) ([]Target, error) {
	stdout, err := r.config.Runner.List(ctx, patterns)
	if err != nil {
		return nil, err
	}

	var b bytes.Buffer
	b.Write(stdout)

	targets, errors, err := getTargets(&b, includeTests)
	if err != nil {
		return nil, err
	}

	if len(targets) == 0 && len(errors) > 0 {
		return nil, fmt.Errorf("resolve patterns: %s", strings.Join(errors, "; "))
	}

	return targets, nil
}

func getTargets(stdout *bytes.Buffer, includeTests bool) ([]Target, []string, error) {
	dec := json.NewDecoder(stdout)
	var targets []Target
	var errors []string

	for dec.More() {
		var pkg listPackage
		if err := dec.Decode(&pkg); err != nil {
			return nil, nil, fmt.Errorf("resolve patterns: decode JSON: %w", err)
		}

		if pkg.Error.Err != "" {
			errors = append(errors, fmt.Sprintf("%s: %s", pkg.ImportPath, pkg.Error.Err))
			continue
		}

		t := Target{
			ImportPath: pkg.ImportPath,
			Dir:        pkg.Dir,
			ModulePath: pkg.ModulePath,
			ModuleDir:  pkg.ModuleDir,
		}

		for _, f := range pkg.GoFiles {
			abs := filepath.Join(pkg.Dir, f)
			t.Files = append(t.Files, abs)
		}

		if includeTests {
			for _, f := range append(pkg.TestGoFiles, pkg.XTestGoFiles...) {
				abs := filepath.Join(pkg.Dir, f)
				t.TestFiles = append(t.TestFiles, abs)
			}
		}

		targets = append(targets, t)
	}

	return targets, errors, nil
}
