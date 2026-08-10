package http_client_interface_mock

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"reflect"
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

type checkResultFn func(*testing.T, *Result)

func checkResult(fns ...checkResultFn) []checkResultFn {
	return fns
}

func checkResultError(want string) checkResultFn {
	return func(t *testing.T, r *Result) {
		t.Helper()

		if want == "" {
			require.NoError(t, r.Error)
			assert.True(t, r.Success)
			return
		}

		require.Error(t, r.Error)
		assert.Contains(t, r.Error.Error(), want)
		assert.False(t, r.Success)
	}
}

func TestWebhookNotifier_TypeNameConnectClose(t *testing.T) {
	n := New(&Config{Name: "webhook-01"})

	assert.Equal(t, "webhook", n.Type())
	assert.Equal(t, "webhook-01", n.Name())
	assert.NoError(t, n.Connect())
	assert.NoError(t, n.Close())
}

func TestWebhookNotifier_Deliver(t *testing.T) {
	tests := []struct {
		name    string
		config  *Config
		message *Notification
		before  func(*WebhookNotifier)
		checks  []checkResultFn
	}{
		{
			name: "json-marshal-error",
			before: func(n *WebhookNotifier) {
				n.jsonMarshal = func(_ any) ([]byte, error) { return nil, fmt.Errorf("error from json.Marshal") }
			},
			checks: checkResult(
				checkResultError("error from json.Marshal"),
			),
		},
		{
			name: "http-newrequest-error",
			before: func(n *WebhookNotifier) {
				n.httpNewRequest = func(_, _ string, _ io.Reader) (*http.Request, error) {
					return nil, fmt.Errorf("test error on http.NewRequest")
				}
			},
			checks: checkResult(
				checkResultError("test error on http.NewRequest"),
			),
		},
		{
			name: "client-do-error",
			before: func(n *WebhookNotifier) {
				n.client = &mockHTTPClient{
					DoFunc: func(req *http.Request) (*http.Response, error) {
						return nil, errors.New("test http new request error")
					},
				}
			},
			checks: checkResult(
				checkResultError("test http new request error"),
			),
		},
		{
			name:    "http-status-code-not-ok",
			config:  &Config{Endpoint: "http://localhost:8080/webhook"},
			message: &Notification{Event: "test-event", Data: "test-data"},
			before: func(n *WebhookNotifier) {
				n.client = &mockHTTPClient{
					DoFunc: func(req *http.Request) (*http.Response, error) {
						return &http.Response{
							StatusCode: http.StatusForbidden,
							Body:       io.NopCloser(bytes.NewBufferString(`Forbidden`)),
						}, nil
					},
				}
			},
			checks: checkResult(
				checkResultError("webhook returned non-OK status: 403"),
			),
		},
		{
			name: "http-status-code-ok",
			config: &Config{
				Endpoint: "http://localhost:8080/webhook",
				Headers: map[string]string{
					"Header-XYZ": "xyz",
				},
			},
			message: &Notification{Event: "test-event", Data: "test-data"},
			before: func(n *WebhookNotifier) {
				n.client = &mockHTTPClient{
					DoFunc: func(req *http.Request) (*http.Response, error) {
						return &http.Response{
							StatusCode: http.StatusOK,
							Body:       io.NopCloser(bytes.NewBufferString(`Ok`)),
						}, nil
					},
				}
			},
			checks: checkResult(
				checkResultError(""),
			),
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			n := New(tt.config)

			if tt.before != nil {
				tt.before(n)
			}

			r := n.Deliver(tt.message)

			for _, c := range tt.checks {
				c(t, r)
			}
		})
	}
}

func TestWebhookNotifier_getClient(t *testing.T) {
	type testCase struct {
		name   string
		config *Config
		before func(*WebhookNotifier)
		checks []func(*testing.T, *WebhookNotifier)
	}

	var (
		checkClientTypeFn = func(clientType interface{}) func(*testing.T, *WebhookNotifier) {
			return func(t *testing.T, n *WebhookNotifier) {
				t.Helper()
				assert.Equalf(t, reflect.TypeOf(clientType), reflect.TypeOf(n.client), "getClient type = %v, expected %v", reflect.TypeOf(n.client), reflect.TypeOf(clientType))
			}
		}

		checkClientInsecure = func(want bool) func(*testing.T, *WebhookNotifier) {
			return func(t *testing.T, n *WebhookNotifier) {
				t.Helper()
				insecureSkipVerify := false
				if transport, ok := n.client.(*http.Client).Transport.(*http.Transport); ok {
					insecureSkipVerify = transport.TLSClientConfig.InsecureSkipVerify
				}
				assert.Equalf(t, want, insecureSkipVerify, "getClient insecure = %t, expected %t", insecureSkipVerify, want)
			}
		}
	)

	tests := []testCase{
		{
			name: "existing-client",
			before: func(n *WebhookNotifier) {
				n.client = &mockHTTPClient{}
			},
			checks: []func(*testing.T, *WebhookNotifier){
				checkClientTypeFn(&mockHTTPClient{}),
			},
		},
		{
			name: "empty-client",
			before: func(n *WebhookNotifier) {
				n.client = nil
			},
			checks: []func(*testing.T, *WebhookNotifier){
				checkClientTypeFn(&http.Client{}),
				checkClientInsecure(false),
			},
		},
		{
			name:   "insecure-client",
			config: &Config{Insecure: true},
			before: func(n *WebhookNotifier) {
				n.client = nil
			},
			checks: []func(*testing.T, *WebhookNotifier){
				checkClientTypeFn(&http.Client{}),
				checkClientInsecure(true),
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			n := New(tt.config)

			if tt.before != nil {
				tt.before(n)
			}

			_ = n.getClient()

			for _, c := range tt.checks {
				c(t, n)
			}
		})
	}
}
