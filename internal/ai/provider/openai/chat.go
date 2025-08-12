package openai

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/esuEdu/ai-wrapper/internal/ai"
)

// OpenAI API request/response structures
type openAIMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type openAIChatRequest struct {
	Model       string          `json:"model"`
	Messages    []openAIMessage `json:"messages"`
	MaxTokens   int             `json:"max_tokens,omitempty"`
	Temperature float64         `json:"temperature,omitempty"`
	Stream      bool            `json:"stream,omitempty"`
}

type openAIChatResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Model   string `json:"model"`
	Choices []struct {
		Index   int `json:"index"`
		Message struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"message"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
}

type openAIStreamResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Model   string `json:"model"`
	Choices []struct {
		Index int `json:"index"`
		Delta struct {
			Role    string `json:"role,omitempty"`
			Content string `json:"content,omitempty"`
		} `json:"delta"`
		FinishReason *string `json:"finish_reason"`
	} `json:"choices"`
}

// ChatStream sends a streaming chat completion request
func (c *Client) ChatStream(ctx context.Context, req *ai.ChatRequest) (<-chan ai.StreamResponse, error) {
	// Convert to OpenAI format
	openAIReq := &openAIChatRequest{
		Model:       req.Model,
		MaxTokens:   req.MaxTokens,
		Temperature: req.Temperature,
		Stream:      true,
	}

	// Set default model if not specified
	if openAIReq.Model == "" {
		openAIReq.Model = "gpt-3.5-turbo"
	}

	// Convert messages
	for _, msg := range req.Messages {
		openAIReq.Messages = append(openAIReq.Messages, openAIMessage{
			Role:    msg.Role,
			Content: msg.Content,
		})
	}

	// Make request
	resp, err := c.makeRequest(ctx, "POST", "/chat/completions", openAIReq)
	if err != nil {
		return nil, err
	}

	// Create response channel
	responseChan := make(chan ai.StreamResponse, 100)

	// Start goroutine to handle streaming
	go func() {
		defer resp.Body.Close()
		defer close(responseChan)

		scanner := bufio.NewScanner(resp.Body)
		var responseID string

		for scanner.Scan() {
			line := scanner.Text()

			// Skip empty lines and comments
			if line == "" || strings.HasPrefix(line, ":") {
				continue
			}

			// Parse Server-Sent Events format
			if strings.HasPrefix(line, "data: ") {
				data := strings.TrimPrefix(line, "data: ")

				// Check for end of stream
				if data == "[DONE]" {
					responseChan <- ai.StreamResponse{
						ID:   responseID,
						Done: true,
					}
					return
				}

				// Parse JSON
				var streamResp openAIStreamResponse
				if err := json.Unmarshal([]byte(data), &streamResp); err != nil {
					responseChan <- ai.StreamResponse{
						Error: fmt.Errorf("failed to parse stream response: %w", err),
					}
					return
				}

				responseID = streamResp.ID

				// Extract content from delta
				if len(streamResp.Choices) > 0 {
					content := streamResp.Choices[0].Delta.Content
					done := streamResp.Choices[0].FinishReason != nil

					responseChan <- ai.StreamResponse{
						ID:      streamResp.ID,
						Content: content,
						Done:    done,
					}

					if done {
						return
					}
				}
			}
		}

		if err := scanner.Err(); err != nil {
			responseChan <- ai.StreamResponse{
				Error: fmt.Errorf("stream scanning error: %w", err),
			}
		}
	}()

	return responseChan, nil
}
