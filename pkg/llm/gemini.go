package llm

import (
	"context"
	"fmt"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

// Client defines the interface for LLM interactions.
type Client interface {
	GenerateContent(ctx context.Context, prompt string) (string, error)
	Close() error
}

// GeminiClient implements Client using Google's Gemini API.
type GeminiClient struct {
	client *genai.Client
	model  *genai.GenerativeModel
}

// NewGeminiClient creates a new GeminiClient.
func NewGeminiClient(ctx context.Context, apiKey string, modelName string) (*GeminiClient, error) {
	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		return nil, fmt.Errorf("failed to create gemini client: %w", err)
	}

	// Use the requested model, defaulting to gemini-1.5-flash if empty
	if modelName == "" {
		modelName = "gemini-1.5-flash"
	}

	// Note: If the user specifically requested "gemini-2.5-flash-lite",
	// we trust the env var. If it fails at runtime, the error will be clear.
	model := client.GenerativeModel(modelName)

	return &GeminiClient{
		client: client,
		model:  model,
	}, nil
}

// GenerateContent sends a prompt to the model and returns the text response.
func (c *GeminiClient) GenerateContent(ctx context.Context, prompt string) (string, error) {
	resp, err := c.model.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		return "", fmt.Errorf("failed to generate content: %w", err)
	}

	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		return "", fmt.Errorf("no content generated")
	}

	// Extract text from the first part of the first candidate
	var result string
	for _, part := range resp.Candidates[0].Content.Parts {
		if txt, ok := part.(genai.Text); ok {
			result += string(txt)
		}
	}

	return result, nil
}

// Close closes the underlying client.
func (c *GeminiClient) Close() error {
	return c.client.Close()
}
