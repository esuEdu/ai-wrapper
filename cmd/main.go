package main

import (
	"log"
	"os"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/esuEdu/ai-wrapper/internal/ai"
	"github.com/esuEdu/ai-wrapper/internal/ai/provider/openai"
	"github.com/esuEdu/ai-wrapper/internal/cli"
)

func main() {
	// Logger
	logger := log.New(os.Stdout, "[AI] ", log.LstdFlags)

	// Setup AI service
	openAIKey := os.Getenv("OPENAI_API_KEY")
	if openAIKey == "" {
		log.Fatal("OPENAI_API_KEY is required")
	}

	aiService := ai.NewService(logger)
	openAIConfig := &ai.Config{
		APIKey:     openAIKey,
		Timeout:    30 * time.Second,
		MaxRetries: 3,
	}
	openAIProvider := openai.NewClient(openAIConfig)
	aiService.RegisterProvider("openai", openAIProvider)

	// Start Bubble Tea program
	p := tea.NewProgram(cli.NewChatModel(aiService))
	if err := p.Start(); err != nil {
		log.Fatal(err)
	}
}
