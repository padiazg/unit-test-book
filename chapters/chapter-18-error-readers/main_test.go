package error_readers

import (
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

type errorReader struct{}

func (e errorReader) Read(p []byte) (n int, err error) {
	return 0, errors.New("simulated read failure")
}

func (e errorReader) Close() error {
	return nil
}

func TestProcessAPIResponse(t *testing.T) {
	t.Run("request error", func(t *testing.T) {
		// a URL that will fail to connect
		body, err := ProcessAPIResponse("http://127.0.0.1:1/nonexistent")
		assert.Error(t, err)
		assert.Empty(t, body)
	})
}

type ReadResponseBodyFn func(*testing.T, string, error)

var checkReadResponseBody = func(fns ...ReadResponseBodyFn) []ReadResponseBodyFn { return fns }

func checkError(want string) ReadResponseBodyFn {
	return func(t *testing.T, _ string, err error) {
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

func checkBody(want string) ReadResponseBodyFn {
	return func(t *testing.T, b string, _ error) {
		t.Helper()
		assert.Equal(t, want, b)
	}
}

func TestReadResponseBody(t *testing.T) {
	tests := []struct {
		name   string
		resp   *http.Response
		checks []ReadResponseBodyFn
	}{
		{
			name: "successful read",
			resp: &http.Response{
				Body: io.NopCloser(strings.NewReader(`{"status":"ok"}`)),
			},
			checks: checkReadResponseBody(
				checkError(""),
				checkBody(`{"status":"ok"}`),
			),
		},
		{
			name: "read error",
			resp: &http.Response{Body: errorReader{}},
			checks: checkReadResponseBody(
				checkError("simulated read failure"),
				checkBody(""),
			),
		},
		{
			name: "nil body",
			resp: &http.Response{Body: nil},
			checks: checkReadResponseBody(
				checkError("response body is nil"),
				checkBody(""),
			),
		},
		{
			name: "nil respons",
			resp: nil,
			checks: checkReadResponseBody(
				checkError("response is nil"),
				checkBody(""),
			),
		},
		{
			name: "empty body",
			resp: &http.Response{Body: io.NopCloser(strings.NewReader(""))},
			checks: checkReadResponseBody(
				checkError(""),
				checkBody(""),
			),
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			r, err := ReadResponseBody(tt.resp)
			for _, c := range tt.checks {
				c(t, r, err)
			}
		})
	}
}
