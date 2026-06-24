# Chapter 16: httptest.Server

## Description

Use `httptest.NewServer` to start a real HTTP server on a random port for testing. The production HTTP client connects to the test server URL over real TCP — no transport mocking, no interface stubs. This tests the *full request/response round trip*, including URL construction, header propagation, JSON encoding, and connection pooling.

## Code

```go
type UserAPI struct {
	baseURL    string
	httpClient *http.Client
}

func NewUserAPI(baseURL string) *UserAPI {
	return &UserAPI{
		baseURL:    baseURL,
		httpClient: &http.Client{Timeout: 10 * time.Second},
	}
}

func (a *UserAPI) GetUser(id int) (*User, error) {
	url := fmt.Sprintf("%s/users/%d", a.baseURL, id)
	resp, err := a.httpClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("requesting user: %w", err)
	}
	defer resp.Body.Close()
	// parse JSON response...
}
```

## Test

```go
func TestUserAPI_GetUser(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/users/1", r.URL.Path)

		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"id":1,"name":"One","email":"one@test.com"}`)
	}))
	defer server.Close()

	api := NewUserAPI(server.URL)
	user, err := api.GetUser(1)
	require.NoError(t, err)
	assert.Equal(t, 1, user.ID)
	assert.Equal(t, "One", user.Name)
}

func TestUserAPI_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, `{"error":"not found"}`)
	}))
	defer server.Close()

	api := NewUserAPI(server.URL)
	_, err := api.GetUser(999)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "404")
}

func TestUserAPI_NetworkFailure(t *testing.T) {
	// server closed immediately — client connects to nothing
	server := httptest.NewServer(nil)
	server.Close()

	api := NewUserAPI(server.URL)
	_, err := api.GetUser(1)
	assert.Error(t, err) // connection refused
}
```

## Testing Approach

httptest.Server:

1. **Full HTTP stack** — requests go through the real `http.Client` including redirect handling, timeout, connection pooling, and TLS. Unlike `RoundTripper` mock, this tests the actual request construction path.
2. **Request inspection** — the handler can assert on `r.Method`, `r.URL`, `r.Header`, and `r.Body` *before* sending the response. This validates what the client actually sent, not what we think it sent.
3. **Server per test** — each test creates its own `httptest.NewServer`. Separate servers mean no route collisions or shared state. `defer server.Close()` keeps cleanup automatic.
4. **Network failure simulation** — close the server immediately to test connection-refused paths. No other mocking technique simulates TCP-level failures this easily.
