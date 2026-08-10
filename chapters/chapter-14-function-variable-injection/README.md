---
section: mocking
---

# Chapter 14: Function Variable Injection

## Description

Store function references (`json.Marshal`, `http.NewRequest`, `time.Now`) as struct fields so tests can replace them without interfaces or mocking frameworks. Each function becomes a test seam: the production version is the real standard library call; the test version returns controlled values or records what was called.

Real-world example: `go-crap/internal/core/services/task.go:15` — `GetEnv` and `LookupEnv` are function fields injected in tests.

## Code

```go
type WebhookPayload struct {
    Timestamp time.Time `json:"timestamp"`
    Data      string    `json:"data"`
    Event     string    `json:"event"`
}

type WebhookSender struct {
    httpNewRequest func(method, url string, body io.Reader) (*http.Request, error)
    jsonMarshal    func(v any) ([]byte, error)
    client         HTTPClient
    Endpoint       string
}

func NewWebhookSender(endpoint string) *WebhookSender {
    return &WebhookSender{
        Endpoint:       endpoint,
        client:         &http.Client{},
        jsonMarshal:    json.Marshal,
        httpNewRequest: http.NewRequest,
    }
}

type HTTPClient interface {
    Do(req *http.Request) (*http.Response, error)
}

func (s *WebhookSender) Send(event, data string) error {
    payload := WebhookPayload{
        Event:     event,
        Data:      data,
        Timestamp: time.Now(),
    }

    body, err := s.jsonMarshal(payload)
    if err != nil {
        return fmt.Errorf("serializing payload: %w", err)
    }

    req, err := s.httpNewRequest(http.MethodPost, s.Endpoint, strings.NewReader(string(body)))
    if err != nil {
        return fmt.Errorf("creating request: %w", err)
    }
    req.Header.Set("Content-Type", "application/json")

    resp, err := s.client.Do(req)
    if err != nil {
        return fmt.Errorf("sending request: %w", err)
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        return fmt.Errorf("unexpected status: %d", resp.StatusCode)
    }

    return nil
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

type checkSendFn func(*testing.T, error)

var checkSend = func(fns ...checkSendFn) []checkSendFn { return fns }

func TestWebhookSender_Send(t *testing.T) {
    checkError := func(want string) checkSendFn {
        return func(t *testing.T, err error) {
            t.Helper()
            require.Error(t, err)
            assert.Contains(t, err.Error(), want)
        }
    }

    checkSuccess := func() checkSendFn {
        return func(t *testing.T, err error) {
            t.Helper()
            assert.NoError(t, err)
        }
    }

    tests := []struct {
        name   string
        before func(*WebhookSender)
        checks []checkSendFn
    }{
        {
            name: "successful send",
            before: func(s *WebhookSender) {
                s.client = &mockHTTPClient{
                    DoFunc: func(req *http.Request) (*http.Response, error) {
                        return &http.Response{
                            StatusCode: http.StatusOK,
                            Body:       io.NopCloser(strings.NewReader(`ok`)),
                        }, nil
                    },
                }
            },
            checks: checkSend(checkSuccess()),
        },
        {
            name: "json marshal error",
            before: func(s *WebhookSender) {
                s.jsonMarshal = func(v any) ([]byte, error) {
                    return nil, fmt.Errorf("json: unexpected error")
                }
            },
            checks: checkSend(checkError("serializing payload")),
        },
        {
            name: "http new request error",
            before: func(s *WebhookSender) {
                s.httpNewRequest = func(method, url string, body io.Reader) (*http.Request, error) {
                    return nil, fmt.Errorf("invalid method")
                }
            },
            checks: checkSend(checkError("creating request")),
        },
        {
            name: "http client error",
            before: func(s *WebhookSender) {
                s.client = &mockHTTPClient{
                    DoFunc: func(req *http.Request) (*http.Response, error) {
                        return nil, fmt.Errorf("connection refused")
                    },
                }
            },
            checks: checkSend(checkError("sending request")),
        },
        {
            name: "non-ok status",
            before: func(s *WebhookSender) {
                s.client = &mockHTTPClient{
                    DoFunc: func(req *http.Request) (*http.Response, error) {
                        return &http.Response{
                            StatusCode: http.StatusForbidden,
                            Body:       io.NopCloser(strings.NewReader(`forbidden`)),
                        }, nil
                    },
                }
            },
            checks: checkSend(checkError("unexpected status: 403")),
        },
        {
            name:   "nil hooks work with defaults",
            before: nil,
            checks: checkSend(checkError("connection refused")), // will actually error since no real server
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            s := NewWebhookSender("http://localhost:9999/webhook")
            if tt.before != nil {
                tt.before(s)
            }

            err := s.Send("user.created", `{"id":42}`)
            for _, fn := range tt.checks {
                fn(t, err)
            }
        })
    }
}

func TestNewWebhookSender_Defaults(t *testing.T) {
    s := NewWebhookSender("http://example.com")
    assert.NotNil(t, s.client)
    assert.NotNil(t, s.jsonMarshal)
    assert.NotNil(t, s.httpNewRequest)

    b, err := s.jsonMarshal(map[string]string{"a": "b"})
    require.NoError(t, err)
    assert.True(t, bytes.Contains(b, []byte(`"a":"b"`)))
}
```

## Scaffold

Generate test scaffolding with go-testgen:

```shell
go-testgen report . --format table
go-testgen gen . Sender.Send
go-testgen gen . Sender.Defaults --style simple
```

## Testing Approach

Function variable injection:

1. **No interfaces needed for functions** — standard library functions (`json.Marshal`, `http.NewRequest`) become struct fields with the same signature. Tests construct the struct with the constructor and override fields via before hooks.
2. **Table-driven before hooks** — the `before` function mutates the injected function fields (`jsonMarshal`, `httpNewRequest`, `client`) per case. One test function covers success plus every failure path (marshal, request creation, client send, non-OK status, defaults) without touching the network.
3. **Error path injection** — `jsonMarshal`, `httpNewRequest`, or `client` can be swapped to return an error on demand. Testing the "marshal failed" path would require a malformed struct with real `json.Marshal`; with injection, it's a one-line stub.
4. **Standard assertions only** — testify is used for plain `assert`/`require` inside typed check closures; the pattern itself is only closures and struct fields. No mock framework, no code generation, no capture-and-verify scaffolding. Works with any function you want to control in tests.
