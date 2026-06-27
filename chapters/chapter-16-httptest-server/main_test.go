package httptest_server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUserAPI_GetUser(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/users/42", r.URL.Path)

		user := User{ID: 42, Name: "Alice", Email: "alice@example.com"}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(user)
	}))
	defer server.Close()

	api := NewUserAPI(server.URL)
	user, err := api.GetUser(42)

	require.NoError(t, err)
	assert.Equal(t, 42, user.ID)
	assert.Equal(t, "Alice", user.Name)
	assert.Equal(t, "alice@example.com", user.Email)
}

func TestUserAPI_GetUser_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`user not found`))
	}))
	defer server.Close()

	api := NewUserAPI(server.URL)
	user, err := api.GetUser(999)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unexpected status 404")
	assert.Nil(t, user)
}

func TestUserAPI_CreateUser(t *testing.T) {
	createdUser := User{ID: 1, Name: "Bob", Email: "bob@test.com"}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/users", r.URL.Path)
		assert.Equal(t, http.MethodPost, r.Method)

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(createdUser)
	}))
	defer server.Close()

	api := NewUserAPI(server.URL)
	user, err := api.CreateUser("Bob", "bob@test.com")

	require.NoError(t, err)
	assert.Equal(t, 1, user.ID)
	assert.Equal(t, "Bob", user.Name)
}

func TestUserAPI_RequestFailure(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	api := NewUserAPI(server.URL)
	user, err := api.GetUser(1)

	assert.Error(t, err)
	assert.Nil(t, user)
}

func TestUserAPI_ServerDown(t *testing.T) {
	// point to a port with no server
	api := NewUserAPI("http://localhost:1")
	user, err := api.GetUser(1)

	assert.Error(t, err)
	assert.Nil(t, user)
}

func TestUserAPI_MultipleCalls(t *testing.T) {
	callCount := 0

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++

		switch r.URL.Path {
		case "/users/1":
			json.NewEncoder(w).Encode(User{ID: 1, Name: "One"})
		case "/users/2":
			json.NewEncoder(w).Encode(User{ID: 2, Name: "Two"})
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	api := NewUserAPI(server.URL)

	u1, _ := api.GetUser(1)
	assert.Equal(t, "One", u1.Name)

	u2, _ := api.GetUser(2)
	assert.Equal(t, "Two", u2.Name)

	assert.Equal(t, 2, callCount)
}

func TestUserAPI_ConcurrentRequests(t *testing.T) {
	mu := sync.Mutex{}
	calls := make(map[int]bool)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.URL.Path[len("/users/"):]
		mu.Lock()
		calls[len(calls)] = true
		mu.Unlock()
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"id":%s,"name":"User"}`, id)
	}))
	defer server.Close()

	api := NewUserAPI(server.URL)

	var wg sync.WaitGroup
	for i := 1; i <= 5; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			user, err := api.GetUser(id)
			assert.NoError(t, err)
			if err == nil {
				assert.Equal(t, id, user.ID)
			}
		}(i)
	}
	wg.Wait()
}
