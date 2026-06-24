# Unit Testing in Go — 29 Patterns from Real Projects

A hands-on guide to Go unit testing through real-world patterns extracted from 7 production Go projects.

Each chapter is a **standalone Go module** with minimal production code, corresponding tests, and an explanation of the testing approach.

## Table of Contents

### Section 1: Foundations
| # | Chapter | Pattern | Source |
|---|---------|---------|--------|
| 01 | [Classic Table-Driven Tests](foundations/chapter-01-classic-table-driven.md) | `wantErr bool` | pantry/domain, hexago/version |
| 02 | [Value Assertions](foundations/chapter-02-value-assertions.md) | `want T` + deep equal | go-testgen/utils, hexago/version |
| 03 | [Fields Struct for Inputs](foundations/chapter-03-fields-struct-inputs.md) | Grouped input structs | hexago/version |
| 04 | [Subtest Naming Strategies](foundations/chapter-04-subtest-naming.md) | Constants as test names | hexago/version |

### Section 2: Closure-Check Pattern
| # | Chapter | Pattern | Source |
|---|---------|---------|--------|
| 05 | [Typed Check Functions](closure-check/chapter-05-typed-check-functions.md) | `type checkXxxFn func(...)` | pantry, jokes |
| 06 | [Check Collection Builder](closure-check/chapter-06-check-collection-builder.md) | `var checkXxx = func(fns...)` | notifier, pantry |
| 07 | [Check Factory Closures](closure-check/chapter-07-check-factory-closures.md) | `checkStatus(want)` closures | go-crap, pantry |
| 08 | [Error Message Verification](closure-check/chapter-08-error-message-verification.md) | `assert.Contains(err.Error())` | notifier, pantry |
| 09 | [Output String Inspection](closure-check/chapter-09-output-string-inspection.md) | `assert.Contains/NotContains` | go-testgen |

### Section 3: Mocking & Dependency Injection
| # | Chapter | Pattern | Source |
|---|---------|---------|--------|
| 10 | [The `before` Hook Pattern](mocking/chapter-10-before-hook-pattern.md) | `before func(*SUT)` | notifier, pantry, jokes |
| 11 | [HTTP Client Interface Mock](mocking/chapter-11-http-client-interface-mock.md) | `HTTPClient` + `mockHTTPClient` | notifier, ollama-tools |
| 12 | [RoundTripper Mock](mocking/chapter-12-roundtripper-mock.md) | `DryRunTransport` | ollama-tools |
| 13 | [testify/mock for Interfaces](mocking/chapter-13-testify-mock-interfaces.md) | `mock.Mock` + `.On().Return()` | pantry, notifier/amqp |
| 14 | [Function Variable Injection](mocking/chapter-14-function-variable-injection.md) | `jsonMarshal`, `httpNewRequest` | notifier |
| 15 | [Package-Level Var Swap](mocking/chapter-15-package-level-var-swap.md) | `randRead` override + defer restore | notifier/utils |

### Section 4: HTTP & I/O Testing
| # | Chapter | Pattern | Source |
|---|---------|---------|--------|
| 16 | [httptest.Server Integration](http-io/chapter-16-httptest-server.md) | Real HTTP test server | notes |
| 17 | [httptest.ResponseRecorder](http-io/chapter-17-httptest-response-recorder.md) | Handler unit tests | pantry/http |
| 18 | [Error Readers](http-io/chapter-18-error-readers.md) | `errorReader` for IO errors | notes |
| 19 | [Temporary Files & Parsing](http-io/chapter-19-temp-files-and-parsing.md) | `t.TempDir()` + file ops | go-testgen, go-crap |

### Section 5: Concurrency
| # | Chapter | Pattern | Source |
|---|---------|---------|--------|
| 20 | [Channel Delivery Tests](concurrency/chapter-20-channel-delivery-tests.md) | Buffered channel + select | notifier |
| 21 | [Panic Recovery in Tests](concurrency/chapter-21-panic-recovery.md) | `defer recover()` + `wantPanic` | notifier |
| 22 | [Goroutine Run Loops](concurrency/chapter-22-goroutine-run-loops.md) | `go n.Run()` + WaitGroup | notifier |
| 23 | [Goroutine Leak Detection](advanced-go/chapter-23-goroutine-leak-detection.md) | `goleak.VerifyTestMain` | go-crap |

### Section 6: Advanced Go Features
| # | Chapter | Pattern | Source |
|---|---------|---------|--------|
| 24 | [AST Parsing for Cyclomatic Complexity](advanced-go/chapter-24-ast-parsing-complexity.md) | `go/parser` + `go/ast` | go-crap/complexity |
| 25 | [Benchmark Tests](advanced-go/chapter-25-benchmark-tests.md) | `b.Benchmark` + `b.Loop()` | go-testgen, go-crap |
| 26 | [Parallel Tests](advanced-go/chapter-26-parallel-tests.md) | `t.Parallel()` | pantry/database |

### Section 7: Integration Patterns
| # | Chapter | Pattern | Source |
|---|---------|---------|--------|
| 27 | [Service Layer with Mocked Ports](integration/chapter-27-service-layer-mocked-ports.md) | `setupService` + `Teardown` | pantry/services |
| 28 | [JSON Format Verification](integration/chapter-28-json-format-verification.md) | Marshal/unmarshal roundtrip | go-crap/report |
| 29 | [Setup/Teardown Fixtures](integration/chapter-29-setup-teardown-fixtures.md) | Fixture struct + teardown | pantry/services |

## How to Use

```bash
# Run tests for a specific chapter
cd chapter-01-classic-table-driven
go test -v ./...

# Run all tests
go test ./chapter-*/...
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
