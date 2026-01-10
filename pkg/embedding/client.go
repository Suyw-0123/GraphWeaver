package embedding

import (
	"context"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

// Client defines the interface for generating embeddings
type Client interface {
	EmbedText(ctx context.Context, text string) ([]float32, error)
	EmbedBatch(ctx context.Context, texts []string) ([][]float32, error)
	Close() error
}

// GeminiClient implements Client using Google Gemini API
type GeminiClient struct {
	client *genai.Client
	model  *genai.EmbeddingModel
}

// NewGeminiClient creates a new embedding client
func NewGeminiClient(ctx context.Context, apiKey string) (*GeminiClient, error) {
	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		return nil, err
	}

	model := client.EmbeddingModel("text-embedding-004")
	return &GeminiClient{
		client: client,
		model:  model,
	}, nil
}

// EmbedText generates embedding for a single text string
func (c *GeminiClient) EmbedText(ctx context.Context, text string) ([]float32, error) {
	res, err := c.model.EmbedContent(ctx, genai.Text(text))
	if err != nil {
		return nil, err
	}
	return res.Embedding.Values, nil
}

// EmbedBatch generates embeddings for multiple text strings
func (c *GeminiClient) EmbedBatch(ctx context.Context, texts []string) ([][]float32, error) {
	batch := c.model.NewBatch()
	for _, text := range texts {
		batch.AddContent(genai.Text(text))
	}

	res, err := c.model.BatchEmbedContents(ctx, batch)
	if err != nil {
		return nil, err
	}

	embeddings := make([][]float32, len(res.Embeddings))
	for i, e := range res.Embeddings {
		embeddings[i] = e.Values
	}
	return embeddings, nil
}

// Close closes the underlying client
func (c *GeminiClient) Close() error {
	return c.client.Close()
}
