# Chapter 12: RoundTripper Mock

## Description

Mock `http.RoundTripper` — the interface behind `http.Client` that turns requests into responses. Implement `RoundTrip(*http.Request) (*http.Response, error)` on a stub and inject it via `&http.Client{Transport: mock}`. This gives you full HTTP mocking without changing the production type signature, since the real code uses `*http.Client` directly.

## Code

```go
type GitHubClient struct {
	client  *http.Client
	baseURL string
	token   string
}

func NewGitHubClient(baseURL, token string) *GitHubClient {
	return &GitHubClient{
		client:  &http.Client{Timeout: 10 * time.Second},
		baseURL: baseURL,
		token:   token,
	}
}

func (c *GitHubClient) GetUser(login string) (*GitHubUser, error) {
	url := fmt.Sprintf("%s/users/%s", c.baseURL, url.PathEscape(login))
	resp, err := c.client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("github request: %w", err)
	}
	defer resp.Body.Close()
	// parse response...
}
```

## Test

```go
type mockRoundTripper struct {
	RoundTripFunc func(*http.Request) (*http.Response, error)
}

func (m *mockRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	return m.RoundTripFunc(req)
}

func TestGitHubClient_GetUser(t *testing.T) {
	before := func(t *testing.T) *GitHubClient {
		t.Helper()
		return NewGitHubClient("https://api.github.com", "test-token")
	}

	withTransport := func(t *testing.T, c *GitHubClient, rt *mockRoundTripper) {
		c.client.Transport = rt
	}

	t.Run("success", func(t *testing.T) {
		client := before(t)
		mockRT := &mockRoundTripper{
			RoundTripFunc: func(req *http.Request) (*http.Response, error) {
				assert.Equal(t, "GET", req.Method)
				assert.Contains(t, req.URL.String(), "/users/octocat")
				return &http.Response{
					StatusCode: http.StatusOK,
					Body: io.NopCloser(strings.NewReader(
						`{"login":"octocat","id":1,"name":"Octocat"}`,
					)),
				}, nil
			},
		}
		withTransport(t, client, mockRT)

		user, err := client.GetUser("octocat")
		require.NoError(t, err)
		assert.Equal(t, "octocat", user.Login)
		assert.Equal(t, 1, user.ID)
	})

	t.Run("not found", func(t *testing.T) {
		client := before(t)
		mockRT := &mockRoundTripper{
			RoundTripFunc: func(req *http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: http.StatusNotFound,
					Body:       io.NopCloser(strings.NewReader(`{"message":"Not Found"}`)),
				}, nil
			},
		}
		withTransport(t, client, mockRT)

		_, err := client.GetUser("nonexistent")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "404")
	})

	t.Run("network error", func(t *testing.T) {
		client := before(t)
		mockRT := &mockRoundTripper{
			RoundTripFunc: func(req *http.Request) (*http.Response, error) {
				return nil, errors.New("TLS handshake timeout")
			},
		}
		withTransport(t, client, mockRT)

		_, err := client.GetUser("octocat")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "TLS handshake timeout")
	})
}
```

## Testing Approach

The RoundTripper mock:

1. **Production type unchanged** — `GetUser` receives `*http.Client`. No interface, no abstraction in production code. The mock hooks in at the transport layer.
2. **Request inspection** — inside `RoundTripFunc`, you can assert on `req.Method`, `req.URL`, `req.Header`. The mock verifies *what was sent* before faking *what comes back*.
3. **Grafter pattern** — `withTransport(t, client, mockRT)` is a before-hook variant that mutates the client after construction. Keeps the fixture setup explicit in each test.
4. **Real `*http.Client` behavior preserved** — timeouts, redirects, cookies, and connection pooling all work normally. Only the transport is swapped.
