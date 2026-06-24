package roundtripper_mock

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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
