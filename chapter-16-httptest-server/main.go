package httptest_server

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type UserAPI struct {
	client  *http.Client
	BaseURL string
}

func NewUserAPI(baseURL string) *UserAPI {
	return &UserAPI{BaseURL: baseURL, client: &http.Client{}}
}

type User struct {
	Email string `json:"email"`
	Name  string `json:"name"`
	ID    int    `json:"id"`
}

func (api *UserAPI) GetUser(id int) (*User, error) {
	url := fmt.Sprintf("%s/users/%d", api.BaseURL, id)
	resp, err := api.client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(body))
	}

	var user User
	if err := json.Unmarshal(body, &user); err != nil {
		return nil, fmt.Errorf("decoding user: %w", err)
	}

	return &user, nil
}

func (api *UserAPI) CreateUser(name, email string) (*User, error) {
	url := fmt.Sprintf("%s/users", api.BaseURL)
	resp, err := api.client.Post(url, "application/json", nil)
	if err != nil {
		return nil, fmt.Errorf("create request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("create failed with status %d", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	var user User
	_ = json.Unmarshal(body, &user)
	return &user, nil
}
