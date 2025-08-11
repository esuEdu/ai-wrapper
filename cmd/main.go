// cmd/yourapp/main.go
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/esuEdu/ai-wrapper/internal/ai"
	"github.com/esuEdu/ai-wrapper/internal/ai/provider/openai"
)

func main() {
	// Create logger
	logger := log.New(os.Stdout, "[AI] ", log.LstdFlags)

	// Create AI service
	aiService := ai.NewService(logger)

	// Get OpenAI API key from environment
	openAIAPIKey := os.Getenv("OPENAI_API_KEY")
	if openAIAPIKey == "" {
		log.Fatal("OPENAI_API_KEY environment variable is required")
	}

	// Create OpenAI provider
	openAIConfig := &ai.Config{
		APIKey:     openAIAPIKey,
		Timeout:    30 * time.Second,
		MaxRetries: 3,
	}
	openAIProvider := openai.NewClient(openAIConfig)

	// Register provider
	aiService.RegisterProvider("openai", openAIProvider)

	// Context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// Example 1: Simple chat
	fmt.Println("=== Simple Chat Example ===")
	response, err := aiService.SimpleChat(ctx, "openai", "gpt-3.5-turbo", "What is the capital of France?")
	if err != nil {
		log.Printf("Simple chat error: %v", err)
	} else {
		fmt.Printf("Response: %s\n\n", response)
	}

	// Example 2: Chat with history
	fmt.Println("=== Chat with History Example ===")
	messages := []ai.Message{
		{Role: "system", Content: "You are a helpful assistant that speaks like a pirate."},
		{Role: "user", Content: "Tell me about the weather."},
	}

	historyResponse, err := aiService.ChatWithHistory(ctx, "openai", "gpt-3.5-turbo", messages)
	if err != nil {
		log.Printf("Chat with history error: %v", err)
	} else {
		fmt.Printf("Pirate Response: %s\n\n", historyResponse)
	}

	// Example 3: Streaming chat
	fmt.Println("=== Streaming Chat Example ===")
	streamReq := &ai.ChatRequest{
		Messages: []ai.Message{
			{Role: "user", Content: "Write a short poem about artificial intelligence."},
		},
		Model:       "gpt-3.5-turbo",
		MaxTokens:   150,
		Temperature: 0.7,
	}

	stream, err := aiService.ChatStream(ctx, "openai", streamReq)
	if err != nil {
		log.Printf("Stream chat error: %v", err)
	} else {
		fmt.Print("Streamed Response: ")
		for chunk := range stream {
			if chunk.Error != nil {
				log.Printf("Stream error: %v", chunk.Error)
				break
			}

			if chunk.Done {
				fmt.Println("\n[Stream completed]")
				break
			}

			fmt.Print(chunk.Content)
		}
		fmt.Println()
	}

	// Example 4: Provider information
	fmt.Println("\n=== Provider Information ===")
	providers := aiService.ListProviders()
	fmt.Printf("Available providers: %v\n", providers)

	models, err := aiService.GetProviderModels(ctx, "openai")
	if err != nil {
		log.Printf("Error getting models: %v", err)
	} else {
		fmt.Printf("OpenAI models: %v\n", models)
	}

	// Example 5: Health check
	fmt.Println("\n=== Health Check ===")
	if err := aiService.HealthCheck(ctx, "openai"); err != nil {
		log.Printf("Health check failed: %v", err)
	} else {
		fmt.Println("OpenAI provider is healthy")
	}

	// Example 6: Advanced chat with custom parameters
	fmt.Println("\n=== Advanced Chat Example ===")
	advancedReq := &ai.ChatRequest{
		Messages: []ai.Message{
			{Role: "system", Content: "You are a creative writing assistant."},
			{Role: "user", Content: "Write a creative title for a sci-fi story about time travel."},
		},
		Model:       "gpt-4",
		MaxTokens:   50,
		Temperature: 0.9,
	}

	advancedResp, err := aiService.Chat(ctx, "openai", advancedReq)
	if err != nil {
		log.Printf("Advanced chat error: %v", err)
	} else {
		fmt.Printf("Creative Title: %s\n", advancedResp.Content)
		fmt.Printf("Usage: %+v\n", advancedResp.Usage)
		fmt.Printf("Model: %s, ID: %s\n", advancedResp.Model, advancedResp.ID)
	}
}
