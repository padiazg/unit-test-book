# Chapter 11: HTTP Client Interface Mock

## Description

Define an `HTTPClient` interface with a `Do(*http.Request) (*http.Response, error)` method matching the `http.Client` signature. Production code uses this interface; tests provide a stub that returns canned responses. This is the simplest and most testable HTTP mocking strategy: no test server, no transport hacking, just an interface.

## Code

```go
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type WeatherClient struct {
	client HTTPClient
	apiKey string
}

func NewWeatherClient(client HTTPClient, apiKey string) *WeatherClient {
	return &WeatherClient{client: client, apiKey: apiKey}
}

func (c *WeatherClient) GetWeather(city string) (*WeatherResponse, error) {
	url := fmt.Sprintf("https://api.weather.com/v1/%s?key=%s", url.PathEscape(city), c.apiKey)
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("weather request: %w", err)
	}
	// parse response...
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

func TestWeatherClient_GetWeather(t *testing.T) {
	type fields struct {
		client *mockHTTPClient
	}

	before := func(t *testing.T) fields {
		t.Helper()
		return fields{client: &mockHTTPClient{}}
	}

	t.Run("success", func(t *testing.T) {
		f := before(t)
		f.client.DoFunc = func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusOK,
				Body: io.NopCloser(strings.NewReader(
					`{"city":"London","temp_c":15.5,"condition":"Cloudy"}`,
				)),
			}, nil
		}

		c := NewWeatherClient(f.client, "test-key")
		w, err := c.GetWeather("London")
		require.NoError(t, err)
		assert.Equal(t, "London", w.City)
		assert.Equal(t, 15.5, w.TempC)
	})

	t.Run("API error", func(t *testing.T) {
		f := before(t)
		f.client.DoFunc = func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusInternalServerError,
				Body:       io.NopCloser(strings.NewReader(`{}`)),
			}, nil
		}

		_, err := NewWeatherClient(f.client, "test-key").GetWeather("London")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "500")
	})

	t.Run("network failure", func(t *testing.T) {
		f := before(t)
		f.client.DoFunc = func(req *http.Request) (*http.Response, error) {
			return nil, errors.New("connection refused")
		}

		_, err := NewWeatherClient(f.client, "test-key").GetWeather("London")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "connection refused")
	})

	t.Run("invalid JSON", func(t *testing.T) {
		f := before(t)
		f.client.DoFunc = func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader(`not json`)),
			}, nil
		}

		_, err := NewWeatherClient(f.client, "test-key").GetWeather("London")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "parse")
	})
}
```

## Testing Approach

The HTTP client interface mock:

1. **Interface segregation** — `HTTPClient{ Do(*http.Request) (*http.Response, error) }` is a single-method interface. Production `http.Client` satisfies it. The stub implements it with a function field.
2. **Per-test behavior** — each subtest sets `DoFunc` to return exactly what it needs (success, errors, status codes). No global mock state leaks between tests.
3. **Error paths visible** — network failures, HTTP errors, and malformed responses are all trivially testable by changing what `DoFunc` returns. No need to start/stop test servers.
4. **Zero dependencies** — the mock is a 6-line struct. No testify/mock, no httptest. The pattern scales: add a `DoFunc` field and each test configures it inline.
