package ai

import (
	"context"
	"fmt"
	"log"
	"time"
)

// Service represents the high-level AI service
type Service struct {
	providers map[string]Provider
	logger    *log.Logger
}

// NewService creates a new AI service
func NewService(logger *log.Logger) *Service {
	if logger == nil {
		logger = log.Default()
	}

	return &Service{
		providers: make(map[string]Provider),
		logger:    logger,
	}
}

// RegisterProvider registers a new AI provider
func (s *Service) RegisterProvider(name string, provider Provider) {
	s.providers[name] = provider
	s.logger.Printf("Registered AI provider: %s", name)
}

// GetProvider returns a provider by name
func (s *Service) GetProvider(name string) (Provider, error) {
	provider, exists := s.providers[name]
	if !exists {
		return nil, fmt.Errorf("provider '%s' not found", name)
	}
	return provider, nil
}

// ListProviders returns all registered provider names
func (s *Service) ListProviders() []string {
	var names []string
	for name := range s.providers {
		names = append(names, name)
	}
	return names
}

// ChatStream sends a streaming chat request to a specific provider
func (s *Service) ChatStream(ctx context.Context, providerName string, req *ChatRequest) (<-chan StreamResponse, error) {
	provider, err := s.GetProvider(providerName)
	if err != nil {
		return nil, err
	}

	s.logger.Printf("Starting stream chat request to %s with model %s", providerName, req.Model)

	stream, err := provider.ChatStream(ctx, req)
	if err != nil {
		s.logger.Printf("Stream chat request to %s failed: %v", providerName, err)
		return nil, err
	}

	// Wrap the stream to add logging
	loggedStream := make(chan StreamResponse, 100)

	go func() {
		defer close(loggedStream)
		start := time.Now()
		totalChunks := 0

		for chunk := range stream {
			loggedStream <- chunk

			if chunk.Error != nil {
				s.logger.Printf("Stream error from %s: %v", providerName, chunk.Error)
				return
			}

			if chunk.Done {
				duration := time.Since(start)
				s.logger.Printf("Stream chat request to %s completed in %v (%d chunks)",
					providerName, duration, totalChunks)
				return
			}

			totalChunks++
		}
	}()

	return loggedStream, nil
}

// GetProviderModels returns available models for a provider
func (s *Service) GetProviderModels(ctx context.Context, providerName string) ([]string, error) {
	provider, err := s.GetProvider(providerName)
	if err != nil {
		return nil, err
	}

	return provider.GetAvailableModels(ctx)
}
