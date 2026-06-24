package function_variable_injection

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type WebhookPayload struct {
	Event     string    `json:"event"`
	Data      string    `json:"data"`
	Timestamp time.Time `json:"timestamp"`
}

type WebhookSender struct {
	Endpoint       string
	client         HTTPClient
	jsonMarshal    func(v any) ([]byte, error)
	httpNewRequest func(method, url string, body io.Reader) (*http.Request, error)
}

func NewWebhookSender(endpoint string) *WebhookSender {
	return &WebhookSender{
		Endpoint:       endpoint,
		client:         &http.Client{},
		jsonMarshal:    json.Marshal,
		httpNewRequest: http.NewRequest,
	}
}

type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

func (s *WebhookSender) Send(event, data string) error {
	payload := WebhookPayload{
		Event:     event,
		Data:      data,
		Timestamp: time.Now(),
	}

	body, err := s.jsonMarshal(payload)
	if err != nil {
		return fmt.Errorf("serializing payload: %w", err)
	}

	req, err := s.httpNewRequest(http.MethodPost, s.Endpoint, strings.NewReader(string(body)))
	if err != nil {
		return fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("sending request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	return nil
}
