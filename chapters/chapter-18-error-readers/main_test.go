package error_readers

import (
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type errorReader struct{}

func (e errorReader) Read(p []byte) (n int, err error) {
	return 0, errors.New("simulated read failure")
}

func (e errorReader) Close() error {
	return nil
}

func TestReadResponseBody(t *testing.T) {
	t.Run("successful read", func(t *testing.T) {
		resp := &http.Response{
			Body: io.NopCloser(strings.NewReader(`{"status":"ok"}`)),
		}
		body, err := ReadResponseBody(resp)
		require.NoError(t, err)
		assert.Equal(t, `{"status":"ok"}`, body)
	})

	t.Run("read error", func(t *testing.T) {
		resp := &http.Response{
			Body: errorReader{},
		}
		body, err := ReadResponseBody(resp)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "simulated read failure")
		assert.Empty(t, body)
	})

	t.Run("nil body", func(t *testing.T) {
		resp := &http.Response{Body: nil}
		body, err := ReadResponseBody(resp)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "response body is nil")
		assert.Empty(t, body)
	})

	t.Run("nil response", func(t *testing.T) {
		body, err := ReadResponseBody(nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "response is nil")
		assert.Empty(t, body)
	})

	t.Run("empty body", func(t *testing.T) {
		resp := &http.Response{
			Body: io.NopCloser(strings.NewReader("")),
		}
		body, err := ReadResponseBody(resp)
		require.NoError(t, err)
		assert.Empty(t, body)
	})
}

func TestProcessAPIResponse(t *testing.T) {
	t.Run("request error", func(t *testing.T) {
		// a URL that will fail to connect
		body, err := ProcessAPIResponse("http://127.0.0.1:1/nonexistent")
		assert.Error(t, err)
		assert.Empty(t, body)
	})
}
