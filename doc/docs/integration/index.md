# Integration

Three integration patterns: service-layer port mocking with testify/mock, JSON format verification and round-trip testing, and struct-based fixture setup/teardown for isolated test state.

## Chapters
- [chapter-27-service-layer-mocked-ports.md](chapter-27-service-layer-mocked-ports.md) — Service Layer with Mocked Ports — mock repository and email interfaces to test business logic in isolation  
  Source: `chapter-27-service-layer-mocked-ports/`
- [chapter-28-json-format-verification.md](chapter-28-json-format-verification.md) — JSON Format Verification — `assert.JSONEq`, `json.MarshalIndent`, and round-trip serialization tests  
  Source: `chapter-28-json-format-verification/`
- [chapter-29-setup-teardown-fixtures.md](chapter-29-setup-teardown-fixtures.md) — Setup/Teardown Fixtures — struct-based fixtures with Setup/Teardown for isolated, repeatable test state  
  Source: `chapter-29-setup-teardown-fixtures/`

## Running the code

Each chapter is a standalone Go module. To run tests for a chapter:

```bash
cd <source-directory>
go test -v ./...
```

