// Package sendkit provides a Go client for the SendKit email API.
package sendkit

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

const defaultBaseURL = "https://api.sendkit.dev"

// Client is the SendKit API client.
type Client struct {
	apiKey  string
	baseURL string
	http    *http.Client

	// Emails provides access to the emails API.
	Emails *EmailsService
}

// Option configures the Client.
type Option func(*Client)

// WithBaseURL sets a custom API base URL.
func WithBaseURL(url string) Option {
	return func(c *Client) {
		c.baseURL = url
	}
}

// WithHTTPClient sets a custom HTTP client.
func WithHTTPClient(client *http.Client) Option {
	return func(c *Client) {
		c.http = client
	}
}

// NewClient creates a new SendKit client.
// If apiKey is empty, it reads from the SENDKIT_API_KEY environment variable.
func NewClient(apiKey string, opts ...Option) (*Client, error) {
	if apiKey == "" {
		apiKey = os.Getenv("SENDKIT_API_KEY")
	}

	if apiKey == "" {
		return nil, fmt.Errorf("sendkit: missing API key, pass it to NewClient or set SENDKIT_API_KEY")
	}

	c := &Client{
		apiKey:  apiKey,
		baseURL: defaultBaseURL,
		http:    http.DefaultClient,
	}

	for _, opt := range opts {
		opt(c)
	}

	c.Emails = &EmailsService{client: c}

	return c, nil
}

func (c *Client) doRequest(ctx context.Context, method, path string, body any) ([]byte, error) {
	var reqBody io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("sendkit: failed to marshal request body: %w", err)
		}
		reqBody = bytes.NewReader(b)
	}

	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, reqBody)
	if err != nil {
		return nil, fmt.Errorf("sendkit: failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("sendkit: request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("sendkit: failed to read response body: %w", err)
	}

	if resp.StatusCode >= 400 {
		var apiErr APIError
		if err := json.Unmarshal(respBody, &apiErr); err != nil {
			return nil, &APIError{
				Name:       "application_error",
				Message:    resp.Status,
				StatusCode: resp.StatusCode,
			}
		}
		apiErr.StatusCode = resp.StatusCode
		return nil, &apiErr
	}

	return respBody, nil
}
