---
section: mocking
---

# Chapter 12: RoundTripper Mock

## Description

Mock `http.RoundTripper` — the interface behind `http.Client` that turns requests into responses. Implement `RoundTrip(*http.Request) (*http.Response, error)` on a stub and inject it via `&http.Client{Transport: mock}`. This gives you full HTTP mocking without changing the production type signature, since the real code uses `*http.Client` directly. In the test the mock is a `DryRunTransport` — a zero-work stand-in that just calls a `RoundTripFn` you set per test case.

Real-world example: `jokes/internal/adapters/secondary/external/joke_client.go:27` — `jokeAPI` uses `*http.Client` directly; tests inject a mock `http.RoundTripper`.

## Code

```go
type GitHubUser struct {
    Login string `json:"login"`
    Name  string `json:"name"`
    URL   string `json:"html_url"`
    ID    int    `json:"id"`
}

type GitHubClient struct {
    client  *http.Client
    BaseURL string
}

func NewGitHubClient(baseURL string) *GitHubClient {
    return &GitHubClient{
        BaseURL: baseURL,
        client:  &http.Client{},
    }
}

func (c *GitHubClient) GetUser(username string) (*GitHubUser, error) {
    url := fmt.Sprintf("%s/users/%s", c.BaseURL, username)
    req, err := http.NewRequest(http.MethodGet, url, nil)
    if err != nil {
        return nil, fmt.Errorf("creating request: %w", err)
    }

    resp, err := c.client.Do(req)
    if err != nil {
        return nil, fmt.Errorf("executing request: %w", err)
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        return nil, fmt.Errorf("unexpected status: %d", resp.StatusCode)
    }

    body, err := io.ReadAll(resp.Body)
    if err != nil {
        return nil, fmt.Errorf("reading body: %w", err)
    }

    var user GitHubUser
    if err := json.Unmarshal(body, &user); err != nil {
        return nil, fmt.Errorf("decoding response: %w", err)
    }

    return &user, nil
}
```

## Test

```go
type DryRunTransport struct {
    http.RoundTripper
    RoundTripFn func(r *http.Request) (*http.Response, error)
}

func (dr *DryRunTransport) RoundTrip(r *http.Request) (*http.Response, error) {
    return dr.RoundTripFn(r)
}

type checkGitHubUserFn func(*testing.T, *GitHubUser, error)

var checkGitHubUser = func(fns ...checkGitHubUserFn) []checkGitHubUserFn { return fns }

func TestGitHubClient_GetUser(t *testing.T) {
    checkError := func(want string) checkGitHubUserFn {
        return func(t *testing.T, _ *GitHubUser, err error) {
            t.Helper()
            require.Error(t, err)
            assert.Contains(t, err.Error(), want)
        }
    }

    checkLogin := func(want string) checkGitHubUserFn {
        return func(t *testing.T, u *GitHubUser, err error) {
            t.Helper()
            require.NoError(t, err)
            assert.Equal(t, want, u.Login)
        }
    }

    tests := []struct {
        name   string
        before func(*GitHubClient)
        checks []checkGitHubUserFn
    }{
        {
            name: "successful response",
            before: func(c *GitHubClient) {
                c.client.Transport = &DryRunTransport{
                    RoundTripFn: func(r *http.Request) (*http.Response, error) {
                        user := GitHubUser{Login: "padiazg", ID: 123, Name: "Pato Diaz"}
                        data, _ := json.Marshal(user)
                        return &http.Response{
                            StatusCode: http.StatusOK,
                            Body:       io.NopCloser(bytes.NewReader(data)),
                            Header:     make(http.Header),
                        }, nil
                    },
                }
            },
            checks: checkGitHubUser(
                checkLogin("padiazg"),
            ),
        },
        {
            name: "not found",
            before: func(c *GitHubClient) {
                c.client.Transport = &DryRunTransport{
                    RoundTripFn: func(r *http.Request) (*http.Response, error) {
                        return &http.Response{
                            StatusCode: http.StatusNotFound,
                            Body:       io.NopCloser(strings.NewReader(`{"message":"Not Found"}`)),
                            Header:     make(http.Header),
                        }, nil
                    },
                }
            },
            checks: checkGitHubUser(
                checkError("unexpected status: 404"),
            ),
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            c := NewGitHubClient("https://api.github.com")
            if tt.before != nil {
                tt.before(c)
            }

            user, err := c.GetUser("padiazg")
            for _, fn := range tt.checks {
                fn(t, user, err)
            }
        })
    }
}
```

## Scaffold

Generate test scaffolding with go-testgen:

```shell
go-testgen report . --format table
go-testgen gen . GitHubClient.GetUser
```

## Testing Approach

The RoundTripper mock:

1. **Production type unchanged** — `GetUser` receives `*http.Client`. No interface, no abstraction in production code. The mock hooks in at the transport layer.
2. **`DryRunTransport`** — embeds `http.RoundTripper` so the struct satisfies the interface, and carries a `RoundTripFn` func field set per test case. There is no pre-built behavior to stub; each case decides what `RoundTrip` returns.
3. **Named check functions with collection builder** — `checkGitHubUserFn` types the assertion block; `checkLogin` and `checkError` are reusable closures. `checkGitHubUser` is a variadic collection builder that composes the checks applied to each case's result.
4. **Grafter pattern** — `before func(*GitHubClient)` is a before-hook variant that swaps in the `DryRunTransport` after construction. Keeps the fixture setup explicit in each test.
5. **Real `*http.Client` behavior preserved** — timeouts, redirects, cookies, and connection pooling all work normally. Only the transport is swapped.
