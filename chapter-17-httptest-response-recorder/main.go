package httptest_response_recorder

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

type Task struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Completed   bool   `json:"completed"`
}

type TaskHandler struct{}

func NewTaskHandler() *TaskHandler {
	return &TaskHandler{}
}

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

	task.ID = fmt.Sprintf("task-%d", len(task.Title))
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(task)
}

func (h *TaskHandler) GetTask(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		respondError(w, http.StatusMethodNotAllowed, "only GET allowed")
		return
	}

	id := strings.TrimPrefix(r.URL.Path, "/tasks/")
	if id == "" || id == r.URL.Path {
		respondError(w, http.StatusBadRequest, "task ID required")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(Task{ID: id, Title: "Sample Task"})
}

func respondError(w http.ResponseWriter, status int, msg string) {
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{"error": msg})
}
