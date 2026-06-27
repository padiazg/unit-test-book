package httptest_response_recorder

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type checkHandlerFn func(*testing.T, *httptest.ResponseRecorder)

var checkHandler = func(fns ...checkHandlerFn) []checkHandlerFn { return fns }

func TestTaskHandler_CreateTask(t *testing.T) {
	checkStatus := func(want int) checkHandlerFn {
		return func(t *testing.T, w *httptest.ResponseRecorder) {
			t.Helper()
			assert.Equal(t, want, w.Code)
		}
	}

	checkError := func(want string) checkHandlerFn {
		return func(t *testing.T, w *httptest.ResponseRecorder) {
			t.Helper()
			var body map[string]string
			json.NewDecoder(w.Body).Decode(&body)
			assert.Contains(t, body["error"], want)
		}
	}

	checkTaskTitle := func(want string) checkHandlerFn {
		return func(t *testing.T, w *httptest.ResponseRecorder) {
			t.Helper()
			var task Task
			json.NewDecoder(w.Body).Decode(&task)
			assert.Equal(t, want, task.Title)
		}
	}

	tests := []struct {
		name   string
		method string
		ctype  string
		body   string
		checks []checkHandlerFn
	}{
		{
			name:   "successful creation",
			method: http.MethodPost,
			ctype:  "application/json",
			body:   `{"title":"Buy milk"}`,
			checks: checkHandler(
				checkStatus(http.StatusCreated),
				checkTaskTitle("Buy milk"),
			),
		},
		{
			name:   "wrong method",
			method: http.MethodGet,
			ctype:  "application/json",
			body:   `{"title":"test"}`,
			checks: checkHandler(
				checkStatus(http.StatusMethodNotAllowed),
				checkError("only POST"),
			),
		},
		{
			name:   "wrong content type",
			method: http.MethodPost,
			ctype:  "text/plain",
			body:   `{"title":"test"}`,
			checks: checkHandler(
				checkStatus(http.StatusUnsupportedMediaType),
				checkError("JSON required"),
			),
		},
		{
			name:   "invalid JSON",
			method: http.MethodPost,
			ctype:  "application/json",
			body:   `{invalid}`,
			checks: checkHandler(
				checkStatus(http.StatusBadRequest),
				checkError("invalid JSON"),
			),
		},
		{
			name:   "empty title",
			method: http.MethodPost,
			ctype:  "application/json",
			body:   `{"title":"  "}`,
			checks: checkHandler(
				checkStatus(http.StatusBadRequest),
				checkError("title is required"),
			),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r := httptest.NewRequest(tt.method, "/tasks", strings.NewReader(tt.body))
			r.Header.Set("Content-Type", tt.ctype)

			h := NewTaskHandler()
			h.CreateTask(w, r)

			for _, fn := range tt.checks {
				fn(t, w)
			}
		})
	}
}

func TestTaskHandler_GetTask(t *testing.T) {
	t.Run("existing task", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/tasks/42", nil)

		NewTaskHandler().GetTask(w, r)

		assert.Equal(t, http.StatusOK, w.Code)

		var task Task
		require.NoError(t, json.NewDecoder(w.Body).Decode(&task))
		assert.Equal(t, "42", task.ID)
	})

	t.Run("missing id", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/tasks/", nil)

		NewTaskHandler().GetTask(w, r)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("wrong method", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodPost, "/tasks/1", nil)

		NewTaskHandler().GetTask(w, r)

		assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
	})
}
