package openai

import (
	"context"
	"encoding/json"
	"fmt"
)

// OpenAI API request/response structure
type openAIModel struct {
	Id      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Owned   string `json:"owned_by"`
}

type openAIModelResponse struct {
	Object string        `json:"object"`
	Models []openAIModel `json:"data"`
}

func (c *Client) GetAvailableModels(ctx context.Context) ([]string, error) {
	// Make request
	resp, err := c.makeRequest(ctx, "GET", "/models", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	fmt.Print(resp)

	var openAIResp openAIModelResponse
	if err := json.NewDecoder(resp.Body).Decode(&openAIResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Extract names
	ids := make([]string, len(openAIResp.Models))
	for i, m := range openAIResp.Models {
		ids[i] = m.Id
	}

	return ids, nil
}
