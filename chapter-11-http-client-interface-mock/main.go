package http_client_interface_mock

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type WeatherClient struct {
	BaseURL string
	APIKey  string
	client  HTTPClient
}

func NewWeatherClient(baseURL, apiKey string) *WeatherClient {
	return &WeatherClient{
		BaseURL: baseURL,
		APIKey:  apiKey,
		client:  &http.Client{},
	}
}

type WeatherResponse struct {
	City        string  `json:"city"`
	Temperature float64 `json:"temperature"`
	Unit        string  `json:"unit"`
}

func (c *WeatherClient) GetWeather(city string) (*WeatherResponse, error) {
	url := fmt.Sprintf("%s/weather?city=%s&apikey=%s", c.BaseURL, city, c.APIKey)
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

	var weather WeatherResponse
	if err := json.Unmarshal(body, &weather); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	return &weather, nil
}
