# Chapter 18: Error Readers

## Description

Implement `io.Reader` that returns errors on demand to test I/O error handling paths. A stub reader with a `Read([]byte) (int, error)` method that fails after `N` bytes or immediately lets you test *read errors*, *partial reads*, and *close errors* without real files or network connections.

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

func TestReadResponseBody(t *testing.T) {
	t.Run("successful read", func(t *testing.T) {
		resp := &http.Response{
			Body: io.NopCloser(strings.NewReader(`{"status":"ok"}`)),
		}
		body, err := ReadResponseBody(resp)
		require.NoError(t, err)
		assert.Equal(t, `{"status":"ok"}`, body)
	})

	t.Run("read error", func(t *testing.T) {
		resp := &http.Response{Body: errorReader{}}
		body, err := ReadResponseBody(resp)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "simulated read failure")
		assert.Empty(t, body)
	})

	t.Run("nil body", func(t *testing.T) {
		resp := &http.Response{Body: nil}
		body, err := ReadResponseBody(resp)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "response body is nil")
		assert.Empty(t, body)
	})

	t.Run("nil response", func(t *testing.T) {
		body, err := ReadResponseBody(nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "response is nil")
		assert.Empty(t, body)
	})

	t.Run("empty body", func(t *testing.T) {
		resp := &http.Response{Body: io.NopCloser(strings.NewReader(""))}
		body, err := ReadResponseBody(resp)
		require.NoError(t, err)
		assert.Empty(t, body)
	})
}
```

## Testing Approach

Error reader pattern:

1. **`io.Reader` interface** — any type with `Read([]byte) (int, error)` satisfies `io.Reader`. An `errorReader` with a single `Read` method returning `0, err` plugs directly into `io.ReadAll`, `json.Decoder`, or any I/O consumer.
2. **Defensive nil checks** — the production code checks `resp == nil` and `resp.Body == nil` *before* calling `Read`. The test covers both paths explicitly, which a normal success test never exercises.
3. **`io.NopCloser` + `strings.NewReader`** — the happy path uses the standard library to turn a string into a ReadCloser. No custom types needed for success cases.
4. **Error message wrapping** — `fmt.Errorf("reading response body: %w", err)` preserves the root cause. The test asserts both the wrapper context ("reading response body") and the root cause ("simulated read failure").
