package http_client_interface_mock

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockHTTPClient struct {
	DoFunc func(req *http.Request) (*http.Response, error)
}

func (m *mockHTTPClient) Do(req *http.Request) (*http.Response, error) {
	return m.DoFunc(req)
}

type checkWeatherFn func(*testing.T, *WeatherResponse, error)

var checkWeather = func(fns ...checkWeatherFn) []checkWeatherFn { return fns }

func TestWeatherClient_GetWeather(t *testing.T) {
	checkError := func(want string) checkWeatherFn {
		return func(t *testing.T, _ *WeatherResponse, err error) {
			t.Helper()
			require.Error(t, err)
			assert.Contains(t, err.Error(), want)
		}
	}

	checkCity := func(want string) checkWeatherFn {
		return func(t *testing.T, r *WeatherResponse, err error) {
			t.Helper()
			require.NoError(t, err)
			assert.Equal(t, want, r.City)
		}
	}

	checkTemperature := func(want float64) checkWeatherFn {
		return func(t *testing.T, r *WeatherResponse, err error) {
			t.Helper()
			require.NoError(t, err)
			assert.Equal(t, want, r.Temperature)
		}
	}

	tests := []struct {
		name   string
		before func(*WeatherClient)
		checks []checkWeatherFn
	}{
		{
			name: "successful response",
			before: func(c *WeatherClient) {
				c.client = &mockHTTPClient{
					DoFunc: func(req *http.Request) (*http.Response, error) {
						resp := WeatherResponse{City: "London", Temperature: 15.5, Unit: "C"}
						data, _ := json.Marshal(resp)
						return &http.Response{
							StatusCode: http.StatusOK,
							Body:       io.NopCloser(bytes.NewReader(data)),
						}, nil
					},
				}
			},
			checks: checkWeather(
				checkCity("London"),
				checkTemperature(15.5),
			),
		},
		{
			name: "not found",
			before: func(c *WeatherClient) {
				c.client = &mockHTTPClient{
					DoFunc: func(req *http.Request) (*http.Response, error) {
						return &http.Response{
							StatusCode: http.StatusNotFound,
							Body:       io.NopCloser(bytes.NewReader([]byte(`{}`))),
						}, nil
					},
				}
			},
			checks: checkWeather(
				checkError("unexpected status: 404"),
			),
		},
		{
			name: "server error",
			before: func(c *WeatherClient) {
				c.client = &mockHTTPClient{
					DoFunc: func(req *http.Request) (*http.Response, error) {
						return &http.Response{
							StatusCode: http.StatusInternalServerError,
							Body:       io.NopCloser(bytes.NewReader([]byte(`{}`))),
						}, nil
					},
				}
			},
			checks: checkWeather(
				checkError("unexpected status: 500"),
			),
		},
		{
			name: "network error",
			before: func(c *WeatherClient) {
				c.client = &mockHTTPClient{
					DoFunc: func(req *http.Request) (*http.Response, error) {
						return nil, assert.AnError
					},
				}
			},
			checks: checkWeather(
				checkError("executing request"),
			),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewWeatherClient("https://api.weather.com", "test-key")
			if tt.before != nil {
				tt.before(c)
			}

			weather, err := c.GetWeather("London")
			for _, fn := range tt.checks {
				fn(t, weather, err)
			}
		})
	}
}
