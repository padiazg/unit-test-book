---
section: mocking
---

# Chapter 11: HTTP Client Interface Mock

## Description

Define an `HTTPClient` interface with a `Do(*http.Request) (*http.Response, error)` method matching the `http.Client` signature. Production code uses this interface; tests provide a stub that returns canned responses. This is the simplest and most testable HTTP mocking strategy: no test server, no transport hacking, just an interface.

`WebhookNotifier.Deliver` marshals a notification to JSON and POSTs it to a configured endpoint through `getClient()`, which returns the injected `HTTPClient` or lazily creates a real `*http.Client`. Because delivery only depends on the interface, tests replace the client with `mockHTTPClient` and control every response — success, non-OK status, network failure — without touching the network.

Real-world example: `notifier/connector/webhook/webhook.go:17` — `type HTTPClient interface { Do(*http.Request) (*http.Response, error) }`; mocked at `webhook_test.go:20` (`mockHTTPClient`) and exercised in `webhook_test.go:305` (`TestWebhookNotifier_Deliver`).

## Code

```go
type HTTPClient interface {
    Do(req *http.Request) (*http.Response, error)
}

type WebhookNotifier struct {
    *Config
    client         HTTPClient
    jsonMarshal    func(v any) ([]byte, error)
    httpNewRequest func(method, url string, body io.Reader) (*http.Request, error)
}

// Deliver sends a notification to the webhook endpoint
func (n *WebhookNotifier) Deliver(message *Notification) *Result {
    payload, err := n.jsonMarshal(message)
    if err != nil {
        return &Result{Success: false, Error: err}
    }

    r, err := n.httpNewRequest(http.MethodPost, n.Endpoint, bytes.NewBuffer(payload))
    if err != nil {
        return &Result{Success: false, Error: err}
    }

    r.Header.Set("Content-Type", "application/json")
    for k, v := range n.Headers {
        r.Header.Set(k, v)
    }

    resp, err := n.getClient().Do(r)
    if err != nil {
        return &Result{Success: false, Error: err}
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        return &Result{Success: false, Error: fmt.Errorf("webhook returned non-OK status: %d", resp.StatusCode)}
    }

    return &Result{Success: true}
}
```

## Test

```go
type mockHTTPClient struct {
    DoFunc func(req *http.Request) (*http.Response, error)
}

func (m *mockHTTPClient) Do(req *http.Request) (*http.Response, error) {
    return m.DoFunc(req)
}

func TestWebhookNotifier_Deliver(t *testing.T) {
    tests := []struct {
        name    string
        config  *Config
        message *Notification
        before  func(*WebhookNotifier)
        checks  []checkResultFn
    }{
        {
            name: "client-do-error",
            before: func(n *WebhookNotifier) {
                n.client = &mockHTTPClient{
                    DoFunc: func(req *http.Request) (*http.Response, error) {
                        return nil, errors.New("test http new request error")
                    },
                }
            },
            checks: checkResult(checkResultError("test http new request error")),
        },
        {
            name:    "http-status-code-not-ok",
            config:  &Config{Endpoint: "http://localhost:8080/webhook"},
            message: &Notification{Event: "test-event", Data: "test-data"},
            before: func(n *WebhookNotifier) {
                n.client = &mockHTTPClient{
                    DoFunc: func(req *http.Request) (*http.Response, error) {
                        return &http.Response{
                            StatusCode: http.StatusForbidden,
                            Body:       io.NopCloser(bytes.NewBufferString(`Forbidden`)),
                        }, nil
                    },
                }
            },
            checks: checkResult(checkResultError("webhook returned non-OK status: 403")),
        },
        {
            name: "http-status-code-ok",
            config: &Config{
                Endpoint: "http://localhost:8080/webhook",
                Headers:  map[string]string{"Header-XYZ": "xyz"},
            },
            message: &Notification{Event: "test-event", Data: "test-data"},
            before: func(n *WebhookNotifier) {
                n.client = &mockHTTPClient{
                    DoFunc: func(req *http.Request) (*http.Response, error) {
                        return &http.Response{
                            StatusCode: http.StatusOK,
                            Body:       io.NopCloser(bytes.NewBufferString(`Ok`)),
                        }, nil
                    },
                }
            },
            checks: checkResult(checkResultError("")),
        },
    }
    // ... table-driven runner applies before hooks, calls Deliver, runs checks
}
```

## Scaffold

Generate test scaffolding with go-testgen:

```shell
go-testgen report . --format table
go-testgen gen . Notifier.Deliver
```

## Testing Approach

The HTTP client interface mock:

1. **Interface segregation** — `HTTPClient{ Do(*http.Request) (*http.Response, error) }` is a single-method interface. Production `*http.Client` satisfies it; `mockHTTPClient` implements it with a `DoFunc` field. `getClient()` returns the injected mock when present, otherwise a real client.
2. **Per-case behavior** — each table row sets `before` to wire the mock: `DoFunc` returns exactly what the case needs (network error, 403, 200). No global mock state leaks between tests.
3. **Error paths visible** — network failures, HTTP errors, and marshaling/request-construction failures are all trivially testable by swapping `DoFunc` (or `jsonMarshal`/`httpNewRequest`). No need to start/stop test servers.
4. **Zero dependencies** — the mock is a 6-line struct. No testify/mock, no httptest. The pattern scales: add a `DoFunc` field and each test configures it inline.
