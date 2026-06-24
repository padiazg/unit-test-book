package before_hook_pattern

import (
	"fmt"
	"time"
)

type RateLimiter interface {
	Allow(key string) (bool, time.Duration)
}

type TokenBucket struct {
	Rate  int
	Burst int
}

func (tb *TokenBucket) Allow(key string) (bool, time.Duration) {
	return true, 0
}

type APIStrategy string

const (
	StrategyFixedWindow APIStrategy = "fixed-window"
	StrategyTokenBucket APIStrategy = "token-bucket"
)

type APIClient struct {
	Strategy    APIStrategy
	RateLimiter RateLimiter
	APIKey      string
	BaseURL     string
	now         func() time.Time
}

func NewAPIClient(apiKey, baseURL string) *APIClient {
	return &APIClient{
		APIKey:  apiKey,
		BaseURL: baseURL,
		now:     time.Now,
	}
}

func (c *APIClient) Call(endpoint string) (string, error) {
	if c.RateLimiter != nil {
		ok, retryAfter := c.RateLimiter.Allow(c.APIKey)
		if !ok {
			return "", fmt.Errorf("rate limit exceeded, retry after %s", retryAfter)
		}
	}

	return fmt.Sprintf("%s/%s response ok", c.BaseURL, endpoint), nil
}
