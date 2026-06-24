# Unit Testing in Go — 29 Patterns from Real Projects

Full documentation site: https://padiazg.github.io/unit-test-book

A hands-on guide to Go unit testing through real-world patterns extracted from 7 production Go projects.

Each chapter is a **standalone Go module** (`chapter-XX-*/`) with production code, tests, and an explanation of the testing approach.

## Run the code

```bash
# Run tests for a specific chapter
cd chapter-01-classic-table-driven
go test -v ./...

# Run all chapter tests
go test ./chapter-*/...
```

## Repository structure

| Path | Description |
|------|-------------|
| `chapter-XX-*/` | 29 standalone Go modules with code + tests + README |
| `doc/` | MkDocs site (generates the documentation site) |
| `doc/docs/` | Generated markdown pages (37 files) |
| `doc/generate.py` | Script that builds mkdocs pages from chapter READMEs |
| `doc/mkdocs.yml` | MkDocs site configuration |

## Build docs locally

```bash
cd doc
pip install mkdocs-material pymdown-extensions
mkdocs serve
```
