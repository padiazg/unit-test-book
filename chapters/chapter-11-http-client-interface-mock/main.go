package http_client_interface_mock

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type Notification struct {
	Event string
	Data  interface{}
}

type Result struct {
	Error   error
	Success bool
}

type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type Config struct {
	Headers  map[string]string
	Name     string
	Endpoint string
	Insecure bool
}

type WebhookNotifier struct {
	*Config
	client         HTTPClient
	jsonMarshal    func(v any) ([]byte, error)
	httpNewRequest func(method, url string, body io.Reader) (*http.Request, error)
}

func New(config *Config) *WebhookNotifier {
	return (&WebhookNotifier{}).New(config)
}

func (n *WebhookNotifier) New(config *Config) *WebhookNotifier {
	if config == nil {
		config = &Config{}
	}

	n.Config = config
	n.jsonMarshal = json.Marshal
	n.httpNewRequest = http.NewRequest

	return n
}

func (n *WebhookNotifier) Type() string {
	return "webhook"
}

func (n *WebhookNotifier) Name() string {
	return n.Config.Name
}

func (n *WebhookNotifier) Connect() error {
	return nil
}

func (n *WebhookNotifier) Close() error {
	return nil
}

// Deliver sends a notification to the webhook endpoint
func (n *WebhookNotifier) Deliver(message *Notification) *Result {
	// Serialize the notification data to JSON
	payload, err := n.jsonMarshal(message)
	if err != nil {
		return &Result{Success: false, Error: err}
	}

	// Send the POST request to the webhook endpoint
	r, err := n.httpNewRequest(http.MethodPost, n.Endpoint, bytes.NewBuffer(payload))
	if err != nil {
		return &Result{Success: false, Error: err}
	}

	// Set headers
	r.Header.Set("Content-Type", "application/json")
	for k, v := range n.Headers {
		r.Header.Set(k, v)
	}

	resp, err := n.getClient().Do(r)
	if err != nil {
		return &Result{Success: false, Error: err}
	}
	defer resp.Body.Close()

	// Check the response status code
	if resp.StatusCode != http.StatusOK {
		return &Result{Success: false, Error: fmt.Errorf("webhook returned non-OK status: %d", resp.StatusCode)}
	}

	return &Result{Success: true}
}

func (n *WebhookNotifier) getClient() HTTPClient {
	if n.client != nil {
		return n.client
	}

	n.client = &http.Client{}
	if n.Insecure {
		n.client.(*http.Client).Transport = &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
	}

	return n.client
}
