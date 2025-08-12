package cli

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/esuEdu/ai-wrapper/internal/ai"
)

type ChatModel struct {
	aiService *ai.Service
	textarea  textarea.Model
	history   []string
	messages  []ai.Message
	err       error
}

type responseMsg string
type errMsg error

func NewChatModel(aiService *ai.Service) ChatModel {
	ta := textarea.New()
	ta.Placeholder = "Ask me something..."
	ta.Focus()

	initialMessages := []ai.Message{
		{Role: "system", Content: "You are a helpful assistant."},
	}

	return ChatModel{
		aiService: aiService,
		textarea:  ta,
		history:   []string{"Welcome to AI Chat!"},
		messages:  initialMessages,
	}
}

func (m ChatModel) Init() tea.Cmd {
	return nil
}

func (m ChatModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			return m, tea.Quit
		case "enter":
			question := strings.TrimSpace(m.textarea.Value())
			if question != "" {
				m.history = append(m.history, fmt.Sprintf("You: %s", question))

				m.messages = append(m.messages, ai.Message{
					Role:    "user",
					Content: question,
				})

				m.textarea.Reset()
				return m, m.askAI()
			}
		}

	case responseMsg:
		m.history = append(m.history, fmt.Sprintf("AI: %s", string(msg)))

		m.messages = append(m.messages, ai.Message{
			Role:    "assistant",
			Content: string(msg),
		})

		return m, nil

	case errMsg:
		m.err = msg
		m.history = append(m.history, fmt.Sprintf("Error: %v", msg))
		return m, nil
	}

	var cmd tea.Cmd
	m.textarea, cmd = m.textarea.Update(msg)
	return m, cmd
}

func (m ChatModel) View() string {
	var b strings.Builder
	for _, line := range m.history {
		b.WriteString(line + "\n")
	}
	b.WriteString("\n" + m.textarea.View())
	b.WriteString("\n(Press Esc or Ctrl+C to quit)")
	return b.String()
}

func (m ChatModel) askAI() tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		defer cancel()

		// Start the stream with the full conversation so far
		stream, err := m.aiService.ChatStream(ctx, "openai", &ai.ChatRequest{
			Model:    "gpt-5",
			Messages: m.messages,
		})
		if err != nil {
			return errMsg(err)
		}

		var aiReply strings.Builder

		// Process the streaming chunks
		for chunk := range stream {
			if chunk.Error != nil {
				return errMsg(chunk.Error)
			}
			if chunk.Done {
				break
			}

			// Append the partial content
			aiReply.WriteString(chunk.Content)

			// Update terminal view as chunks arrive
			m.history = append(m.history, fmt.Sprintf("AI (partial): %s", chunk.Content))
		}

		fullReply := aiReply.String()

		// Store in conversation context
		m.messages = append(m.messages, ai.Message{
			Role:    "assistant",
			Content: fullReply,
		})

		// Show final reply
		return responseMsg(fullReply)
	}
}
