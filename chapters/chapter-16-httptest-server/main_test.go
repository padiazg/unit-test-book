package httptest_server

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

type checkUserFn func(*testing.T, *User, error)

var checkUser = func(fns ...checkUserFn) []checkUserFn { return fns }

func setupServer(payload string, statusCode int) *httptest.Server {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if statusCode == 0 {
			statusCode = http.StatusOK
		}
		w.WriteHeader(statusCode)
		fmt.Fprint(w, payload)
	}))

	return server
}

func checkError(want string) checkUserFn {
	return func(t *testing.T, _ *User, err error) {
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

func checkId(want int) checkUserFn {
	return func(t *testing.T, u *User, _ error) {
		t.Helper()
		assert.Equal(t, want, u.ID)
	}
}

func checkName(want string) checkUserFn {
	return func(t *testing.T, u *User, _ error) {
		t.Helper()
		assert.Equal(t, want, u.Name)
	}
}

func checkEmail(want string) checkUserFn {
	return func(t *testing.T, u *User, _ error) {
		t.Helper()
		assert.Equal(t, want, u.Email)
	}
}

func TestUserAPI_GetUser(t *testing.T) {
	tests := []struct {
		name       string
		id         int
		statusCode int
		payload    string
		checks     []checkUserFn
		before     func(*UserAPI)
	}{
		{
			name:       "success",
			id:         1,
			payload:    `{"id":1,"name":"Alice","email":"alice@example.com"}`,
			statusCode: http.StatusOK,
			checks: checkUser(
				checkError(""),
				checkId(1),
				checkName("Alice"),
				checkEmail("alice@example.com"),
			),
		},
		{
			name:       "not-found",
			statusCode: http.StatusNotFound,
			checks: checkUser(
				checkError("404:"),
			),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := setupServer(tt.payload, tt.statusCode)
			defer server.Close()

			s := NewUserAPI(server.URL)
			if tt.before != nil {
				tt.before(s)
			}

			r, err := s.GetUser(tt.id)
			for _, c := range tt.checks {
				c(t, r, err)
			}
		})
	}
}

func TestUserAPI_CreateUser(t *testing.T) {
	tests := []struct {
		name       string
		userName   string
		email      string
		statusCode int
		checks     []checkUserFn
		before     func(*UserAPI)
	}{
		{
			name:       "success",
			userName:   "Jhon",
			email:      "jhon@example.com",
			statusCode: http.StatusCreated,
			checks: checkUser(
				checkError(""),
				checkId(3),
				checkName("Jhon"),
				checkEmail("jhon@example.com"),
			),
		},
		{
			name:       "server-error",
			statusCode: http.StatusInternalServerError,
			checks: checkUser(
				checkError("create failed with status 500"),
			),
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {

			server := setupServer(fmt.Sprintf(`{"id":3,"name":"%s","email":"%s"}`, tt.userName, tt.email), tt.statusCode)
			defer server.Close()

			s := NewUserAPI(server.URL)
			if tt.before != nil {
				tt.before(s)
			}
			r, err := s.CreateUser(tt.userName, tt.email)
			for _, c := range tt.checks {
				c(t, r, err)
			}
		})
	}
}
