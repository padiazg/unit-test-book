# Unit Testing in Go — 29 Patterns from Real Projects

Full documentation site: https://padiazg.github.io/unit-test-book

A hands-on guide to Go unit testing through real-world patterns extracted from 7 production Go projects.

Each chapter is a **standalone Go module** with production code, tests, and an explanation of the testing approach.

## Table of Contents

### Section 1: Foundations
| # | Chapter | Pattern | Source |
|---|---------|---------|--------|
| 01 | [Classic Table-Driven Tests](chapter-01-classic-table-driven/README.md) | `wantErr bool` | pantry/domain, hexago/version |
| 02 | [Value Assertions](chapter-02-value-assertions/README.md) | `want T` + deep equal | go-testgen/utils, hexago/version |
| 03 | [Fields Struct for Inputs](chapter-03-fields-struct-inputs/README.md) | Grouped input structs | hexago/version |
| 04 | [Subtest Naming Strategies](chapter-04-subtest-naming/README.md) | Constants as test names | hexago/version |

### Section 2: Closure-Check Pattern
| # | Chapter | Pattern | Source |
|---|---------|---------|--------|
| 05 | [Typed Check Functions](chapter-05-typed-check-functions/README.md) | `type checkXxxFn func(...)` | pantry, jokes |
| 06 | [Check Collection Builder](chapter-06-check-collection-builder/README.md) | `var checkXxx = func(fns...)` | notifier, pantry |
| 07 | [Check Factory Closures](chapter-07-check-factory-closures/README.md) | `checkStatus(want)` closures | go-crap, pantry |
| 08 | [Error Message Verification](chapter-08-error-message-verification/README.md) | `assert.Contains(err.Error())` | notifier, pantry |
| 09 | [Output String Inspection](chapter-09-output-string-inspection/README.md) | `assert.Contains/NotContains` | go-testgen |

### Section 3: Mocking & Dependency Injection
| # | Chapter | Pattern | Source |
|---|---------|---------|--------|
| 10 | [The `before` Hook Pattern](chapter-10-before-hook-pattern/README.md) | `before func(*SUT)` | notifier, pantry, jokes |
| 11 | [HTTP Client Interface Mock](chapter-11-http-client-interface-mock/README.md) | `HTTPClient` + `mockHTTPClient` | notifier, ollama-tools |
| 12 | [RoundTripper Mock](chapter-12-roundtripper-mock/README.md) | `DryRunTransport` | ollama-tools |
| 13 | [testify/mock for Interfaces](chapter-13-testify-mock-interfaces/README.md) | `mock.Mock` + `.On().Return()` | pantry, notifier/amqp |
| 14 | [Function Variable Injection](chapter-14-function-variable-injection/README.md) | `jsonMarshal`, `httpNewRequest` | notifier |
| 15 | [Package-Level Var Swap](chapter-15-package-level-var-swap/README.md) | `randRead` override + defer restore | notifier/utils |

### Section 4: HTTP & I/O Testing
| # | Chapter | Pattern | Source |
|---|---------|---------|--------|
| 16 | [httptest.Server Integration](chapter-16-httptest-server/README.md) | Real HTTP test server | notes |
| 17 | [httptest.ResponseRecorder](chapter-17-httptest-response-recorder/README.md) | Handler unit tests | pantry/http |
| 18 | [Error Readers](chapter-18-error-readers/README.md) | `errorReader` for IO errors | notes |
| 19 | [Temporary Files & Parsing](chapter-19-temp-files-and-parsing/README.md) | `t.TempDir()` + file ops | go-testgen, go-crap |

### Section 5: Concurrency
| # | Chapter | Pattern | Source |
|---|---------|---------|--------|
| 20 | [Channel Delivery Tests](chapter-20-channel-delivery-tests/README.md) | Buffered channel + select | notifier |
| 21 | [Panic Recovery in Tests](chapter-21-panic-recovery/README.md) | `defer recover()` + `wantPanic` | notifier |
| 22 | [Goroutine Run Loops](chapter-22-goroutine-run-loops/README.md) | `go n.Run()` + WaitGroup | notifier |

### Section 6: Advanced Go Features
| # | Chapter | Pattern | Source |
|---|---------|---------|--------|
| 23 | [Goroutine Leak Detection](chapter-23-goroutine-leak-detection/README.md) | `goleak.VerifyTestMain` | go-crap |
| 24 | [AST Parsing for Cyclomatic Complexity](chapter-24-ast-parsing-complexity/README.md) | `go/parser` + `go/ast` | go-crap/complexity |
| 25 | [Benchmark Tests](chapter-25-benchmark-tests/README.md) | `b.Benchmark` + `b.Loop()` | go-testgen, go-crap |
| 26 | [Parallel Tests](chapter-26-parallel-tests/README.md) | `t.Parallel()` | pantry/database |

### Section 7: Integration Patterns
| # | Chapter | Pattern | Source |
|---|---------|---------|--------|
| 27 | [Service Layer with Mocked Ports](chapter-27-service-layer-mocked-ports/README.md) | `setupService` + `Teardown` | pantry/services |
| 28 | [JSON Format Verification](chapter-28-json-format-verification/README.md) | Marshal/unmarshal roundtrip | go-crap/report |
| 29 | [Setup/Teardown Fixtures](chapter-29-setup-teardown-fixtures/README.md) | Fixture struct + teardown | pantry/services |

## How to Use

```bash
# Run tests for a specific chapter
cd chapter-01-classic-table-driven
go test -v ./...

# Run all tests
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

## Real-World Sources

The patterns in this book are extracted from these production Go projects:

| Project | Description |
|---------|-------------|
| [hexago](https://github.com/padiazg/hexago) | Go project scaffolding with hexagonal architecture |
| [jokes](https://github.com/padiazg/jokes) | Joke API client |
| [go-testgen](https://github.com/padiazg/go-testgen) | Go test generator |
| [go-crap](https://github.com/padiazg/go-crap) | CRAP score analysis tool |
| [ollama-tools](https://github.com/padiazg/ollama-tools) | Ollama model tools |
| [pantry](https://github.com/padiazg/pantry) | Inventory management with hexagonal architecture |
| [notifier](https://github.com/padiazg/notifier) | Multi-channel notification system |
