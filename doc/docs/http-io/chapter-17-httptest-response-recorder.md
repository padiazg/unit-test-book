# Chapter 17: httptest.ResponseRecorder

## Description

Use `httptest.NewRecorder` to capture HTTP handler output without starting a server. The recorder implements `http.ResponseWriter` and stores the status code, headers, and body. Combined with `httptest.NewRequest`, you can test handlers in isolation — no network, no server lifecycle, just handler logic.

## Code

```go
type TaskHandler struct{}

func (h *TaskHandler) CreateTask(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		respondError(w, http.StatusMethodNotAllowed, "only POST allowed")
		return
	}
	if ct := r.Header.Get("Content-Type"); ct != "application/json" {
		respondError(w, http.StatusUnsupportedMediaType, "JSON required")
		return
	}
	var task Task
	if err := json.NewDecoder(r.Body).Decode(&task); err != nil {
		respondError(w, http.StatusBadRequest, fmt.Sprintf("invalid JSON: %v", err))
		return
	}
	task.Title = strings.TrimSpace(task.Title)
	if task.Title == "" {
		respondError(w, http.StatusBadRequest, "title is required")
		return
	}
	// create task...
}
```

## Test

```go
func TestTaskHandler_CreateTask(t *testing.T) {
	tests := []struct {
		name   string
		method string
		ctype  string
		body   string
		checks []checkHandlerFn
	}{
		{
			name:   "success",
			method: http.MethodPost,
			ctype:  "application/json",
			body:   `{"title":"Buy milk"}`,
			checks: checkHandler(checkStatus(http.StatusCreated), checkTaskTitle("Buy milk")),
		},
		{
			name:   "wrong method",
			method: http.MethodGet,
			ctype:  "application/json",
			body:   `{"title":"test"}`,
			checks: checkHandler(checkStatus(http.StatusMethodNotAllowed), checkError("only POST")),
		},
		{
			name:   "wrong content type",
			method: http.MethodPost,
			ctype:  "text/plain",
			body:   `{"title":"test"}`,
			checks: checkHandler(checkStatus(http.StatusUnsupportedMediaType), checkError("JSON required")),
		},
		{
			name:   "invalid JSON",
			method: http.MethodPost,
			ctype:  "application/json",
			body:   `{bad}`,
			checks: checkHandler(checkStatus(http.StatusBadRequest), checkError("invalid JSON")),
		},
		{
			name:   "empty title",
			method: http.MethodPost,
			ctype:  "application/json",
			body:   `{"title":"  "}`,
			checks: checkHandler(checkStatus(http.StatusBadRequest), checkError("title is required")),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r := httptest.NewRequest(tt.method, "/tasks", strings.NewReader(tt.body))
			r.Header.Set("Content-Type", tt.ctype)

			NewTaskHandler().CreateTask(w, r)

			for _, fn := range tt.checks {
				fn(t, w) // each check receives the recorder
			}
		})
	}
}
```

## Testing Approach

httptest.ResponseRecorder:

1. **Handler in isolation** — the recorder captures exactly what `WriteHeader` and `Write` produce. No routing, no middleware, no server process. Tests focus on handler logic alone.
2. **Closure-check integration** — `checkHandlerFn` is a typed check function that receives `*httptest.ResponseRecorder`. Helper factories like `checkStatus(201)` and `checkError("required")` compose assertions as a slice.
3. **No server lifecycle** — no `defer server.Close()`, no port allocation, no goroutines. Tests run as fast as any non-HTTP table-driven test. The recorder is created and inspected in the same function.
4. **Request construction** — `httptest.NewRequest(method, url, body)` creates a valid `*http.Request` with a `GET` default body or configurable reader. Set headers explicitly for content-type, auth, etc.
