package client

import (
	"time"

	"github.com/go-resty/resty/v2"
)

const (
	defaultTimeout = 30 * time.Second
	maxRetries     = 3
	retryWaitTime  = 1 * time.Second
	maxRetryWait   = 10 * time.Second
)

func New(authClient *resty.Client) *resty.Client {
	client := resty.New()
	
	// Copy settings from auth client
	client.SetBaseURL(authClient.BaseURL)
	client.SetCookies(authClient.Cookies)
	client.SetHeader("User-Agent", authClient.Header.Get("User-Agent"))
	
	// Configure retry and timeout settings
	client.
		SetTimeout(defaultTimeout).
		SetRetryCount(maxRetries).
		SetRetryWaitTime(retryWaitTime).
		SetRetryMaxWaitTime(maxRetryWait).
		AddRetryCondition(func(r *resty.Response, err error) bool {
			// Retry on connection errors
			if err != nil {
				return true
			}
			// Retry on 429 (rate limit) and 5xx errors
			return r.StatusCode() == 429 || r.StatusCode() >= 500
		})

	// Add response middleware for better error handling
	client.OnAfterResponse(func(c *resty.Client, resp *resty.Response) error {
		// Log rate limit headers if present
		if remaining := resp.Header().Get("X-RateLimit-Remaining"); remaining != "" {
			// Could log this for debugging
		}
		return nil
	})

	return client
}