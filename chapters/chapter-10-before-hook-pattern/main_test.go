package before_hook_pattern

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type mockRateLimiter struct {
	allowFn func(key string) (bool, time.Duration)
}

func (m *mockRateLimiter) Allow(key string) (bool, time.Duration) {
	return m.allowFn(key)
}

type checkAPICallFn func(*testing.T, string, error)

var checkAPICall = func(fns ...checkAPICallFn) []checkAPICallFn { return fns }

func TestAPIClient_Call(t *testing.T) {
	checkSuccess := func(want string) checkAPICallFn {
		return func(t *testing.T, got string, err error) {
			t.Helper()
			assert.NoError(t, err)
			assert.Contains(t, got, want)
		}
	}

	checkError := func(want string) checkAPICallFn {
		return func(t *testing.T, _ string, err error) {
			t.Helper()
			assert.Error(t, err)
			assert.Contains(t, err.Error(), want)
		}
	}

	tests := []struct {
		name   string
		before func(*APIClient)
		checks []checkAPICallFn
	}{
		{
			name: "no rate limiter",
			before: func(c *APIClient) {
				c.RateLimiter = nil
			},
			checks: checkAPICall(
				checkSuccess("response ok"),
			),
		},
		{
			name: "rate limiter allows",
			before: func(c *APIClient) {
				c.RateLimiter = &mockRateLimiter{
					allowFn: func(key string) (bool, time.Duration) { return true, 0 },
				}
			},
			checks: checkAPICall(
				checkSuccess("response ok"),
			),
		},
		{
			name: "rate limiter blocks",
			before: func(c *APIClient) {
				c.RateLimiter = &mockRateLimiter{
					allowFn: func(key string) (bool, time.Duration) { return false, 30 * time.Second },
				}
			},
			checks: checkAPICall(
				checkError("rate limit exceeded"),
			),
		},
		{
			name: "custom base URL",
			before: func(c *APIClient) {
				c.BaseURL = "https://custom.api.com/v2"
			},
			checks: checkAPICall(
				checkSuccess("https://custom.api.com/v2/data"),
			),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewAPIClient("key-123", "https://api.example.com/v1")
			if tt.before != nil {
				tt.before(c)
			}

			got, err := c.Call("data")
			for _, fn := range tt.checks {
				fn(t, got, err)
			}
		})
	}
}
