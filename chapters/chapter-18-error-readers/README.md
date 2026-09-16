---
section: http-io
---

# Chapter 18: Error Readers

## Description

Implement `io.Reader` that returns errors on demand to test I/O error handling paths. A stub reader with a `Read([]byte) (int, error)` method that fails after `N` bytes or immediately lets you test *read errors*, *partial reads*, and *close errors* without real files or network connections.

Real-world example: `jokes/internal/adapters/secondary/external/joke_client_test.go:68` — `errReader` simulates body read failures in HTTP response tests.

## Code

```go
func ReadResponseBody(resp *http.Response) (string, error) {
    if resp == nil {
        return "", fmt.Errorf("response is nil")
    }
    if resp.Body == nil {
        return "", fmt.Errorf("response body is nil")
    }
    defer resp.Body.Close()

    body, err := io.ReadAll(resp.Body)
    if err != nil {
        return "", fmt.Errorf("reading response body: %w", err)
    }

    return string(body), nil
}

func ProcessAPIResponse(url string) (string, error) {
    resp, err := http.Get(url)
    if err != nil {
        return "", fmt.Errorf("request failed: %w", err)
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        return "", fmt.Errorf("unexpected status: %d", resp.StatusCode)
    }

    return ReadResponseBody(resp)
}
```

## Test

```go
type errorReader struct{}

func (e errorReader) Read(p []byte) (n int, err error) {
    return 0, errors.New("simulated read failure")
}

func (e errorReader) Close() error {
    return nil
}

func TestProcessAPIResponse(t *testing.T) {
    t.Run("request error", func(t *testing.T) {
        // a URL that will fail to connect
        body, err := ProcessAPIResponse("http://127.0.0.1:1/nonexistent")
        assert.Error(t, err)
        assert.Empty(t, body)
    })
}

type ReadResponseBodyFn func(*testing.T, string, error)

var checkReadResponseBody = func(fns ...ReadResponseBodyFn) []ReadResponseBodyFn { return fns }

func checkError(want string) ReadResponseBodyFn {
    return func(t *testing.T, _ string, err error) {
        t.Helper()
        if want == "" {
            assert.NoErrorf(t, err, "checkError: expected no error, got %v", err)
            return
        }
        if assert.Errorf(t, err, "checkError: expected error %q", want) {
            assert.Containsf(t, err.Error(), want, "checkError mismatch")
        }
    }
}

func checkBody(want string) ReadResponseBodyFn {
    return func(t *testing.T, b string, _ error) {
        t.Helper()
        assert.Equal(t, want, b)
    }
}

func TestReadResponseBody(t *testing.T) {
    tests := []struct {
        name   string
        resp   *http.Response
        checks []ReadResponseBodyFn
    }{
        {
            name: "successful read",
            resp: &http.Response{
                Body: io.NopCloser(strings.NewReader(`{"status":"ok"}`)),
            },
            checks: checkReadResponseBody(
                checkError(""),
                checkBody(`{"status":"ok"}`),
            ),
        },
        {
            name: "read error",
            resp: &http.Response{Body: errorReader{}},
            checks: checkReadResponseBody(
                checkError("simulated read failure"),
                checkBody(""),
            ),
        },
        {
            name: "nil body",
            resp: &http.Response{Body: nil},
            checks: checkReadResponseBody(
                checkError("response body is nil"),
                checkBody(""),
            ),
        },
        {
            name: "nil respons",
            resp: nil,
            checks: checkReadResponseBody(
                checkError("response is nil"),
                checkBody(""),
            ),
        },
        {
            name: "empty body",
            resp: &http.Response{Body: io.NopCloser(strings.NewReader(""))},
            checks: checkReadResponseBody(
                checkError(""),
                checkBody(""),
            ),
        },
    }
    for _, tt := range tests {
        tt := tt
        t.Run(tt.name, func(t *testing.T) {
            r, err := ReadResponseBody(tt.resp)
            for _, c := range tt.checks {
                c(t, r, err)
            }
        })
    }
}
```

## Scaffold

Generate test scaffolding with go-testgen:

```shell
go-testgen report . --format table
go-testgen gen . ReadResponseBody
go-testgen gen . ProcessAPIResponse --style simple
```

## Testing Approach

Error reader pattern:

1. **`io.Reader` interface** — any type with `Read([]byte) (int, error)` satisfies `io.Reader`. An `errorReader` with a single `Read` method returning `0, err` plugs directly into `io.ReadAll`, `json.Decoder`, or any I/O consumer.
2. **Defensive nil checks** — the production code checks `resp == nil` and `resp.Body == nil` *before* calling `Read`. The test covers both paths explicitly, which a normal success test never exercises.
3. **`io.NopCloser` + `strings.NewReader`** — the happy path uses the standard library to turn a string into a ReadCloser. No custom types needed for success cases.
4. **Error message wrapping** — `fmt.Errorf("reading response body: %w", err)` preserves the root cause. The test asserts both the wrapper context ("reading response body") and the root cause ("simulated read failure").
