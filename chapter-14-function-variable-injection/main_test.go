package function_variable_injection

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strings"
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
