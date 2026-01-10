package repository

import (
	"context"

	"github.com/suyw-0123/graphweaver/internal/entity"
)

// SearchResult represents a single result from a vector search
type SearchResult struct {
	ID      string                 `json:"id"`
	Score   float32                `json:"score"`
	Payload map[string]interface{} `json:"payload"`
	Vector  []float32              `json:"vector,omitempty"`
}

// VectorRepository defines the interface for vector database operations
type VectorRepository interface {
	// CreateCollection creates a new collection if it doesn't exist
	CreateCollection(ctx context.Context, name string, vectorSize int) error

	// DeleteCollection deletes a collection
	DeleteCollection(ctx context.Context, name string) error

	// Upsert stores or updates vectors in a collection
	Upsert(ctx context.Context, collection string, points []*entity.VectorPoint) error

	// Search finds the nearest neighbors for a query vector
	Search(ctx context.Context, collection string, vector []float32, limit int, scoreThreshold float32) ([]SearchResult, error)

	// Delete removes points by ID
	Delete(ctx context.Context, collection string, ids []string) error
}
