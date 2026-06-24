package roundtripper_mock

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type GitHubUser struct {
	Login string `json:"login"`
	ID    int    `json:"id"`
	Name  string `json:"name"`
	URL   string `json:"html_url"`
}

type GitHubClient struct {
	BaseURL string
	client  *http.Client
}

func NewGitHubClient(baseURL string) *GitHubClient {
	return &GitHubClient{
		BaseURL: baseURL,
		client:  &http.Client{},
	}
}

func (c *GitHubClient) GetUser(username string) (*GitHubUser, error) {
	url := fmt.Sprintf("%s/users/%s", c.BaseURL, username)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("executing request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading body: %w", err)
	}

	var user GitHubUser
	if err := json.Unmarshal(body, &user); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	return &user, nil
}
