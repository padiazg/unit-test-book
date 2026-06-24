# Chapter 14: Function Variable Injection

## Description

Store function references (`json.Marshal`, `http.NewRequest`, `time.Now`) as struct fields so tests can replace them without interfaces or mocking frameworks. Each function becomes a test seam: the production version is the real standard library call; the test version returns controlled values or records what was called.

## Code

```go
type WebhookSender struct {
	Endpoint       string
	client         HTTPClient
	jsonMarshal    func(v any) ([]byte, error)
	httpNewRequest func(method, url string, body io.Reader) (*http.Request, error)
}

func NewWebhookSender(endpoint string) *WebhookSender {
	return &WebhookSender{
		Endpoint:       endpoint,
		client:         &http.Client{},
		jsonMarshal:    json.Marshal,
		httpNewRequest: http.NewRequest,
	}
}

func (s *WebhookSender) Send(event, data string) error {
	payload := WebhookPayload{
		Event:     event,
		Data:      data,
		Timestamp: time.Now(),
	}
	body, err := s.jsonMarshal(payload)
	// ... builds request, sends via s.client
}
```

## Test

```go
func TestWebhookSender_Send(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		var capturedBody []byte
		var capturedURL string

		sender := &WebhookSender{
			Endpoint: "https://hooks.example.com/evt",
			jsonMarshal: func(v any) ([]byte, error) {
				return json.Marshal(v)
			},
			httpNewRequest: func(method, url string, body io.Reader) (*http.Request, error) {
				capturedURL = url
				capturedBody, _ = io.ReadAll(body)
				return http.NewRequest(method, url, body)
			},
			client: &mockHTTPClient{
				DoFunc: func(req *http.Request) (*http.Response, error) {
					return &http.Response{
						StatusCode: http.StatusOK,
						Body:       io.NopCloser(strings.NewReader(`{}`)),
					}, nil
				},
			},
		}

		err := sender.Send("user.created", `{"id":42}`)
		require.NoError(t, err)
		assert.Contains(t, string(capturedBody), "user.created")
		assert.Contains(t, string(capturedBody), `"data":"{\"id\":42}"`)
	})

	t.Run("marshal error", func(t *testing.T) {
		sender := &WebhookSender{
			Endpoint: "https://hooks.example.com/evt",
			jsonMarshal: func(v any) ([]byte, error) {
				return nil, errors.New("marshaling failed")
			},
		}
		err := sender.Send("test", "data")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "marshaling failed")
	})
}
```

## Testing Approach

Function variable injection:

1. **No interfaces needed** — standard library functions (`json.Marshal`, `http.NewRequest`) become struct fields with the same signature. Tests construct the struct directly with injected stubs.
2. **Capture and verify** — `httpNewRequest` is replaced with a closure that captures `capturedBody` and `capturedURL`. Tests both inject behavior and verify *what was sent* — all without touching the network.
3. **Error path injection** — `jsonMarshal` can be set to return an error on demand. Testing the "marshal failed" path would require a malformed struct with real `json.Marshal`; with injection, it's one line.
4. **Zero dependency** — the pattern uses only closures and struct fields. No testify/mock, no code generation. Works with any function you want to control in tests.
