// internal/ai/model.go
package ai

import (
	"context"
	"time"
)

// Message represents a single message in a conversation
type Message struct {
	Role    string `json:"role"` // "system", "user", "assistant"
	Content string `json:"content"`
}

// ChatRequest represents a chat completion request
type ChatRequest struct {
	Messages    []Message `json:"messages"`
	Model       string    `json:"model,omitempty"`
	MaxTokens   int       `json:"max_tokens,omitempty"`
	Temperature float64   `json:"temperature,omitempty"`
	Stream      bool      `json:"stream,omitempty"`
}

// ChatResponse represents a chat completion response
type ChatResponse struct {
	ID      string    `json:"id"`
	Model   string    `json:"model"`
	Content string    `json:"content"`
	Usage   Usage     `json:"usage"`
	Created time.Time `json:"created"`
}

// Usage represents token usage information
type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// StreamResponse represents a streaming chat response chunk
type StreamResponse struct {
	ID      string `json:"id"`
	Content string `json:"content"`
	Done    bool   `json:"done"`
	Error   error  `json:"error,omitempty"`
}

// Provider represents an AI provider interface
type Provider interface {
	// Chat sends a chat completion request
	Chat(ctx context.Context, req *ChatRequest) (*ChatResponse, error)

	// ChatStream sends a streaming chat completion request
	ChatStream(ctx context.Context, req *ChatRequest) (<-chan StreamResponse, error)

	// GetName returns the provider name
	GetName() string

	// GetAvailableModels returns available models for this provider
	GetAvailableModels(ctx context.Context) ([]string, error)
}

// Config represents provider configuration
type Config struct {
	APIKey      string            `json:"api_key"`
	BaseURL     string            `json:"base_url,omitempty"`
	Timeout     time.Duration     `json:"timeout,omitempty"`
	MaxRetries  int               `json:"max_retries,omitempty"`
	ExtraParams map[string]string `json:"extra_params,omitempty"`
}

// Error represents an AI provider error
type Error struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Type    string `json:"type"`
}

func (e *Error) Error() string {
	return e.Message
}
