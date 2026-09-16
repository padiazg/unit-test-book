---
section: http-io
---

# Chapter 16: httptest.Server

## Description

Use `httptest.NewServer` to start a real HTTP server on a random port for testing. The production HTTP client connects to the test server URL over real TCP — no transport mocking, no interface stubs. This tests the *full request/response round trip*, including URL construction, header propagation, JSON encoding, and connection pooling.

Real-world example: `go-testgen/internal/llm/client_test.go:25` — `httptest.NewServer` for LLM API integration tests.

## Code

```go
type UserAPI struct {
    client  *http.Client
    BaseURL string
}

func NewUserAPI(baseURL string) *UserAPI {
    return &UserAPI{BaseURL: baseURL, client: &http.Client{}}
}

type User struct {
    Email string `json:"email"`
    Name  string `json:"name"`
    ID    int    `json:"id"`
}

func (api *UserAPI) GetUser(id int) (*User, error) {
    url := fmt.Sprintf("%s/users/%d", api.BaseURL, id)
    resp, err := api.client.Get(url)
    if err != nil {
        return nil, fmt.Errorf("request failed: %w", err)
    }
    defer resp.Body.Close()

    body, err := io.ReadAll(resp.Body)
    if err != nil {
        return nil, fmt.Errorf("reading response: %w", err)
    }

    if resp.StatusCode != http.StatusOK {
        return nil, fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(body))
    }

    var user User
    if err := json.Unmarshal(body, &user); err != nil {
        return nil, fmt.Errorf("decoding user: %w", err)
    }

    return &user, nil
}

func (api *UserAPI) CreateUser(name, email string) (*User, error) {
    url := fmt.Sprintf("%s/users", api.BaseURL)

    var buf bytes.Buffer
    fmt.Fprintf(&buf, `{"name":"%s","email":"%s"}`, name, email)

    resp, err := api.client.Post(url, "application/json", &buf)
    if err != nil {
        return nil, fmt.Errorf("create request failed: %w", err)
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusCreated {
        return nil, fmt.Errorf("create failed with status %d", resp.StatusCode)
    }

    body, _ := io.ReadAll(resp.Body)
    var user User
    _ = json.Unmarshal(body, &user)
    return &user, nil
}
```

## Test

```go
type checkUserFn func(*testing.T, *User, error)

var checkUser = func(fns ...checkUserFn) []checkUserFn { return fns }

func setupServer(payload string, statusCode int) *httptest.Server {
    server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Content-Type", "application/json")
        if statusCode == 0 {
            statusCode = http.StatusOK
        }
        w.WriteHeader(statusCode)
        fmt.Fprint(w, payload)
    }))

    return server
}

func checkError(want string) checkUserFn {
    return func(t *testing.T, _ *User, err error) {
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

func checkId(want int) checkUserFn {
    return func(t *testing.T, u *User, _ error) {
        t.Helper()
        assert.Equal(t, want, u.ID)
    }
}

func checkName(want string) checkUserFn {
    return func(t *testing.T, u *User, _ error) {
        t.Helper()
        assert.Equal(t, want, u.Name)
    }
}

func checkEmail(want string) checkUserFn {
    return func(t *testing.T, u *User, _ error) {
        t.Helper()
        assert.Equal(t, want, u.Email)
    }
}

func TestUserAPI_GetUser(t *testing.T) {
    tests := []struct {
        name       string
        id         int
        statusCode int
        payload    string
        checks     []checkUserFn
        before     func(*UserAPI)
    }{
        {
            name:       "success",
            id:         1,
            payload:    `{"id":1,"name":"Alice","email":"alice@example.com"}`,
            statusCode: http.StatusOK,
            checks: checkUser(
                checkError(""),
                checkId(1),
                checkName("Alice"),
                checkEmail("alice@example.com"),
            ),
        },
        {
            name:       "not-found",
            statusCode: http.StatusNotFound,
            checks: checkUser(
                checkError("404:"),
            ),
        },
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            server := setupServer(tt.payload, tt.statusCode)
            defer server.Close()

            s := NewUserAPI(server.URL)
            if tt.before != nil {
                tt.before(s)
            }

            r, err := s.GetUser(tt.id)
            for _, c := range tt.checks {
                c(t, r, err)
            }
        })
    }
}

func TestUserAPI_CreateUser(t *testing.T) {
    tests := []struct {
        name       string
        userName   string
        email      string
        statusCode int
        checks     []checkUserFn
        before     func(*UserAPI)
    }{
        {
            name:       "success",
            userName:   "Jhon",
            email:      "jhon@example.com",
            statusCode: http.StatusCreated,
            checks: checkUser(
                checkError(""),
                checkId(3),
                checkName("Jhon"),
                checkEmail("jhon@example.com"),
            ),
        },
        {
            name:       "server-error",
            statusCode: http.StatusInternalServerError,
            checks: checkUser(
                checkError("create failed with status 500"),
            ),
        },
    }
    for _, tt := range tests {
        tt := tt
        t.Run(tt.name, func(t *testing.T) {

            server := setupServer(fmt.Sprintf(`{"id":3,"name":"%s","email":"%s"}`, tt.userName, tt.email), tt.statusCode)
            defer server.Close()

            s := NewUserAPI(server.URL)
            if tt.before != nil {
                tt.before(s)
            }
            r, err := s.CreateUser(tt.userName, tt.email)
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
go-testgen gen . UserAPI.CreateUser
go-testgen gen . UserAPI.GetUser
```

## Testing Approach

httptest.Server:

1. **Full HTTP stack** — requests go through the real `http.Client` including redirect handling, timeout, connection pooling, and TLS. Unlike `RoundTripper` mock, this tests the actual request construction path.
2. **Request inspection** — the handler can assert on `r.Method`, `r.URL`, `r.Header`, and `r.Body` *before* sending the response. This validates what the client actually sent, not what we think it sent.
3. **Server per test** — each test creates its own `httptest.NewServer`. Separate servers mean no route collisions or shared state. `defer server.Close()` keeps cleanup automatic.
4. **Network failure simulation** — close the server immediately to test connection-refused paths. No other mocking technique simulates TCP-level failures this easily.
