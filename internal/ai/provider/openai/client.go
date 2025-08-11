// internal/ai/provider/openai/client.go
package openai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/esuEdu/ai-wrapper/internal/ai"
)

const (
	DefaultBaseURL = "https://api.openai.com/v1"
	DefaultTimeout = 30 * time.Second
)

// Client represents an OpenAI API client
type Client struct {
	apiKey     string
	baseURL    string
	httpClient *http.Client
	maxRetries int
}

// NewClient creates a new OpenAI client
func NewClient(config *ai.Config) *Client {
	baseURL := config.BaseURL
	if baseURL == "" {
		baseURL = DefaultBaseURL
	}

	timeout := config.Timeout
	if timeout == 0 {
		timeout = DefaultTimeout
	}

	maxRetries := config.MaxRetries
	if maxRetries == 0 {
		maxRetries = 3
	}

	return &Client{
		apiKey:  config.APIKey,
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: timeout,
		},
		maxRetries: maxRetries,
	}
}

// GetName returns the provider name
func (c *Client) GetName() string {
	return "openai"
}

// makeRequest makes an HTTP request to the OpenAI API
func (c *Client) makeRequest(ctx context.Context, method, endpoint string, body interface{}) (*http.Response, error) {
	var reqBody io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		reqBody = bytes.NewBuffer(jsonBody)
	}

	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+endpoint, reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "ai-wrapper/1.0")

	var resp *http.Response
	for i := 0; i <= c.maxRetries; i++ {
		resp, err = c.httpClient.Do(req)
		if err == nil && resp.StatusCode < 500 {
			break
		}

		if resp != nil {
			resp.Body.Close()
		}

		if i == c.maxRetries {
			break
		}

		// Exponential backoff
		time.Sleep(time.Duration(i+1) * time.Second)
	}

	if err != nil {
		return nil, fmt.Errorf("request failed after %d retries: %w", c.maxRetries, err)
	}

	if resp.StatusCode >= 400 {
		defer resp.Body.Close()
		body, _ := io.ReadAll(resp.Body)

		var apiError struct {
			Error struct {
				Message string `json:"message"`
				Type    string `json:"type"`
				Code    string `json:"code"`
			} `json:"error"`
		}

		if err := json.Unmarshal(body, &apiError); err == nil {
			return nil, &ai.Error{
				Code:    apiError.Error.Code,
				Message: apiError.Error.Message,
				Type:    apiError.Error.Type,
			}
		}

		return nil, fmt.Errorf("API error: %d %s", resp.StatusCode, string(body))
	}

	return resp, nil
}
