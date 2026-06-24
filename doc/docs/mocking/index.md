# Mocking

Six techniques for isolating external dependencies. Covers HTTP client interfaces, RoundTripper transport mocking, testify/mock, function variable injection, and package-level var swap.

## Chapters
- [chapter-10-before-hook-pattern.md](chapter-10-before-hook-pattern.md) — The `before` Hook Pattern — a typed fixture function returns fresh test state for each case  
  Source: `chapter-10-before-hook-pattern/`
- [chapter-11-http-client-interface-mock.md](chapter-11-http-client-interface-mock.md) — HTTP Client Interface Mock — define an `HTTPClient{ Do() }` interface and stub its single method  
  Source: `chapter-11-http-client-interface-mock/`
- [chapter-12-roundtripper-mock.md](chapter-12-roundtripper-mock.md) — RoundTripper Mock — implement `http.RoundTripper` to mock at the transport layer without changing production types  
  Source: `chapter-12-roundtripper-mock/`
- [chapter-13-testify-mock-interfaces.md](chapter-13-testify-mock-interfaces.md) — testify/mock for Interfaces — embed `mock.Mock`, use `On().Return()` for interface mock expectations  
  Source: `chapter-13-testify-mock-interfaces/`
- [chapter-14-function-variable-injection.md](chapter-14-function-variable-injection.md) — Function Variable Injection — store `json.Marshal`, `http.NewRequest` as struct fields for test seams  
  Source: `chapter-14-function-variable-injection/`
- [chapter-15-package-level-var-swap.md](chapter-15-package-level-var-swap.md) — Package-Level Var Swap — override a package variable and restore with `defer` for minimal seam injection  
  Source: `chapter-15-package-level-var-swap/`

## Running the code

Each chapter is a standalone Go module. To run tests for a chapter:

```bash
cd <source-directory>
go test -v ./...
```

