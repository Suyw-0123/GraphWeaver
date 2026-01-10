package repository

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/suyw-0123/graphweaver/internal/entity"
)

func TestQdrantVectorRepository_Integration(t *testing.T) {
	// Skip if no Qdrant available (manual toggle or check connection)
	// For CI/CD, we would ensuring Qdrant is up.
	// Here we try to connect, if fail we skip.

	ctx := context.Background()
	repo, err := NewQdrantVectorRepository("localhost", 6334)
	if err != nil {
		t.Skipf("Skipping integration test: failed to connect to qdrant: %v", err)
	}
	defer repo.Close()

	// Verify connection by listing collections (or creating one)
	collectionName := "test_integration_" + uuid.New().String()
	vectorSize := 4 // Small size for testing

	// 1. Create Collection
	err = repo.CreateCollection(ctx, collectionName, vectorSize)
	if err != nil {
		t.Skipf("Skipping integration test: failed to create collection (is qdrant running?): %v", err)
	}
	// Cleanup at end
	defer repo.DeleteCollection(ctx, collectionName)

	// 2. Upsert Points
	points := []*entity.VectorPoint{
		{
			ID:      uuid.New().String(),
			Vector:  []float32{0.1, 0.2, 0.3, 0.4},
			Payload: map[string]interface{}{"type": "test", "index": 1},
		},
		{
			ID:      uuid.New().String(),
			Vector:  []float32{0.9, 0.8, 0.7, 0.6},
			Payload: map[string]interface{}{"type": "test", "index": 2},
		},
	}

	err = repo.Upsert(ctx, collectionName, points)
	if err != nil {
		t.Fatalf("Upsert failed: %v", err)
	}

	// Give Qdrant a moment to index
	time.Sleep(1 * time.Second)

	// 3. Search
	// Query close to first point
	query := []float32{0.1, 0.2, 0.3, 0.4}
	results, err := repo.Search(ctx, collectionName, query, 5, 0.0)
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}

	if len(results) == 0 {
		t.Fatal("Expected results, got 0")
	}

	// First result should be the first point
	if results[0].Score < 0.9 {
		t.Errorf("Expected high score for exact match, got %f", results[0].Score)
	}

	// Check payload
	if val, ok := results[0].Payload["type"]; !ok || val != "test" {
		t.Errorf("Expected payload 'type'='test', got %v", val)
	}

	// 4. Delete
	idsToDelete := []string{points[0].ID}
	err = repo.Delete(ctx, collectionName, idsToDelete)
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	// Verify deleted
	time.Sleep(1 * time.Second)
	results, err = repo.Search(ctx, collectionName, query, 5, 0.0)
	if err != nil {
		t.Fatalf("Search after delete failed: %v", err)
	}

	// Should find the other point (maybe) but definitely not the deleted one
	for _, res := range results {
		if res.ID == points[0].ID {
			t.Error("Deleted point still found")
		}
	}
}
