# Contributing to unit-test-book

Thanks for your interest in contributing.

## Reporting bugs

Open an issue using the bug report template. Include:
- Go version (`go version`)
- OS and architecture
- Steps to reproduce
- Expected vs actual behavior

## Proposing features

Open an issue using the feature request template.
Describe the use case before proposing a solution.

## Development setup

```bash
git clone https://github.com/padiazg/unit-test-book.git
cd unit-test-book
go test ./chapter-*/...
```

Requirements: Go 1.26+, golangci-lint.

## Pull request process

1. Fork the repo and create a feature branch from `master`.
2. Make your changes. Keep commits atomic.
3. Run `go test ./chapter-*/...` and `golangci-lint run ./...` — both must pass.
4. Open a PR against `master`. Fill in the PR template.
5. One approval required before merge.

## Code style

- Tests: table-driven with `testify/assert`.
- Commits: conventional commits (`feat:`, `fix:`, `docs:`, `refactor:`, `test:`).
- Each chapter is a standalone Go module under `chapter-XX-name/`.

## License

By contributing you agree your contributions are licensed under the MIT License.
