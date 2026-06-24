# HTTP / I/O

Four patterns for testing HTTP handlers and I/O operations. Uses httptest.Server for integration tests, ResponseRecorder for handler unit tests, error readers for failure paths, and temp files for filesystem tests.

## Chapters
- [chapter-16-httptest-server.md](chapter-16-httptest-server.md) — httptest.Server Integration — start a real HTTP server on a random port for full-stack HTTP tests  
  Source: `chapter-16-httptest-server/`
- [chapter-17-httptest-response-recorder.md](chapter-17-httptest-response-recorder.md) — httptest.ResponseRecorder — capture handler output without starting a server for pure handler unit tests  
  Source: `chapter-17-httptest-response-recorder/`
- [chapter-18-error-readers.md](chapter-18-error-readers.md) — Error Readers — implement `io.Reader` that returns errors on demand to test I/O failure paths  
  Source: `chapter-18-error-readers/`
- [chapter-19-temp-files-and-parsing.md](chapter-19-temp-files-and-parsing.md) — Temporary Files & Parsing — use `t.TempDir()` for isolated file I/O with automatic cleanup  
  Source: `chapter-19-temp-files-and-parsing/`

## Running the code

Each chapter is a standalone Go module. To run tests for a chapter:

```bash
cd <source-directory>
go test -v ./...
```

