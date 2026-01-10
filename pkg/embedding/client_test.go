package embedding

import (
	"context"
	"os"
	"testing"
)

// Ensure GeminiClient implements Client interface
var _ Client = (*GeminiClient)(nil)

func TestNewGeminiClient(t *testing.T) {
	ctx := context.Background()
	apiKey := "test-api-key"

	client, err := NewGeminiClient(ctx, apiKey)
	if err != nil {
		t.Fatalf("NewGeminiClient failed: %v", err)
	}
	if client == nil {
		t.Fatal("NewGeminiClient returned nil")
	}
	defer client.Close()
}

func TestEmbedText_Integration(t *testing.T) {
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		t.Skip("Skipping integration test: GEMINI_API_KEY not set")
	}

	ctx := context.Background()
	client, err := NewGeminiClient(ctx, apiKey)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	defer client.Close()

	text := "Hello world"
	embedding, err := client.EmbedText(ctx, text)
	if err != nil {
		t.Fatalf("EmbedText failed: %v", err)
	}

	if len(embedding) == 0 {
		t.Error("Expected non-empty embedding")
	}

	// text-embedding-004 should have 768 dimensions
	if len(embedding) != 768 {
		t.Logf("Warning: Expected 768 dimensions, got %d", len(embedding))
	}
}
