# Unit Testing in Go — Patterns from real projects

Full documentation site: https://padiazg.github.io/unit-test-book

A hands-on guide to Go unit testing through real-world patterns extracted from 7 production Go projects.

Each chapter is a **standalone Go module** with production code, tests, and an explanation of the testing approach.

## Table of Contents

### Section 1: Foundations
| # | Chapter | Pattern | Source |
|---|---------|---------|--------|
| 01 | [Classic Table-Driven Tests](chapters/chapter-01-classic-table-driven/README.md) | `wantErr bool` | pantry/domain, hexago/version |
| 02 | [Value Assertions](chapters/chapter-02-value-assertions/README.md) | `want T` + deep equal | go-testgen/utils, hexago/version |
| 03 | [Fields Struct for Inputs](chapters/chapter-03-fields-struct-inputs/README.md) | Grouped input structs | hexago/version |
| 04 | [Subtest Naming Strategies](chapters/chapter-04-subtest-naming/README.md) | Constants as test names | hexago/version |

### Section 2: Closure-Check Pattern
| # | Chapter | Pattern | Source |
|---|---------|---------|--------|
| 05 | [Typed Check Functions](chapters/chapter-05-typed-check-functions/README.md) | `type checkXxxFn func(...)` | pantry, jokes |
| 06 | [Check Collection Builder](chapters/chapter-06-check-collection-builder/README.md) | `var checkXxx = func(fns...)` | notifier, pantry |
| 07 | [Check Factory Closures](chapters/chapter-07-check-factory-closures/README.md) | `checkStatus(want)` closures | go-crap, pantry |
| 08 | [Error Message Verification](chapters/chapter-08-error-message-verification/README.md) | `assert.Contains(err.Error())` | notifier, pantry |
| 09 | [Output String Inspection](chapters/chapter-09-output-string-inspection/README.md) | `assert.Contains/NotContains` | go-testgen |
| 30 | [Composable Check Navigation](chapters/chapter-30-composable-check-navigation/README.md) | Navigator factories for nested output | go-crap/report |
| 31 | [Inline Check Closures](chapters/chapter-31-inline-check-closures/README.md) | Inline closures when checks are single-use | go-crap/coverage |

### Section 3: Mocking & Dependency Injection
| # | Chapter | Pattern | Source |
|---|---------|---------|--------|
| 10 | [The `before` Hook Pattern](chapters/chapter-10-before-hook-pattern/README.md) | `before func(*SUT)` | notifier, pantry, jokes |
| 11 | [HTTP Client Interface Mock](chapters/chapter-11-http-client-interface-mock/README.md) | `HTTPClient` + `mockHTTPClient` | notifier, ollama-tools |
| 12 | [RoundTripper Mock](chapters/chapter-12-roundtripper-mock/README.md) | `DryRunTransport` | ollama-tools |
| 13 | [testify/mock for Interfaces](chapters/chapter-13-testify-mock-interfaces/README.md) | `mock.Mock` + `.On().Return()` | pantry, notifier/amqp |
| 14 | [Function Variable Injection](chapters/chapter-14-function-variable-injection/README.md) | `jsonMarshal`, `httpNewRequest` | notifier |
| 15 | [Package-Level Var Swap](chapters/chapter-15-package-level-var-swap/README.md) | `randRead` override + defer restore | notifier/utils |
| 32 | [Extracting Interfaces from Third-Party Deps](chapters/chapter-32-interface-extraction-3rd-party/README.md) | Interface extraction + `testify/mock` | go-aqi/sps30 |

### Section 4: HTTP & I/O Testing
| # | Chapter | Pattern | Source |
|---|---------|---------|--------|
| 16 | [httptest.Server Integration](chapters/chapter-16-httptest-server/README.md) | Real HTTP test server | notes |
| 17 | [httptest.ResponseRecorder](chapters/chapter-17-httptest-response-recorder/README.md) | Handler unit tests | pantry/http |
| 18 | [Error Readers](chapters/chapter-18-error-readers/README.md) | `errorReader` for IO errors | notes |
| 19 | [Temporary Files & Parsing](chapters/chapter-19-temp-files-and-parsing/README.md) | `t.TempDir()` + file ops | go-testgen, go-crap |

### Section 5: Concurrency
| # | Chapter | Pattern | Source |
|---|---------|---------|--------|
| 20 | [Channel Delivery Tests](chapters/chapter-20-channel-delivery-tests/README.md) | Buffered channel + select | notifier |
| 21 | [Panic Recovery in Tests](chapters/chapter-21-panic-recovery/README.md) | `defer recover()` + `wantPanic` | notifier |
| 22 | [Goroutine Run Loops](chapters/chapter-22-goroutine-run-loops/README.md) | `go n.Run()` + WaitGroup | notifier |

### Section 6: Advanced Go Features
| # | Chapter | Pattern | Source |
|---|---------|---------|--------|
| 23 | [Goroutine Leak Detection](chapters/chapter-23-goroutine-leak-detection/README.md) | `goleak.VerifyTestMain` | go-crap |
| 24 | [AST Parsing for Cyclomatic Complexity](chapters/chapter-24-ast-parsing-complexity/README.md) | `go/parser` + `go/ast` | go-crap/complexity |
| 25 | [Benchmark Tests](chapters/chapter-25-benchmark-tests/README.md) | `b.Benchmark` + `b.Loop()` | go-testgen, go-crap |
| 26 | [Parallel Tests](chapters/chapter-26-parallel-tests/README.md) | `t.Parallel()` | pantry/database |

### Section 7: Integration Patterns
| # | Chapter | Pattern | Source |
|---|---------|---------|--------|
| 27 | [Service Layer with Mocked Ports](chapters/chapter-27-service-layer-mocked-ports/README.md) | `setupService` + `Teardown` | pantry/services |
| 28 | [JSON Format Verification](chapters/chapter-28-json-format-verification/README.md) | Marshal/unmarshal roundtrip | go-crap/report |
| 29 | [Setup/Teardown Fixtures](chapters/chapter-29-setup-teardown-fixtures/README.md) | Fixture struct + teardown | pantry/services |

## How to Use

```bash
# Run tests for a specific chapter
cd chapters/chapter-01-classic-table-driven
go test -v ./...

# Run all tests
go test ./chapters/chapter-*/...
```

## Repository structure

| Path | Description |
|------|-------------|
| `chapters/chapter-XX-*/` | 31 standalone Go modules with code + tests + README |
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
