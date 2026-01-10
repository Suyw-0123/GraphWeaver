package repository

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/suyw-0123/graphweaver/internal/entity"
)

// ChunkRepository defines the interface for chunk persistence
type ChunkRepository interface {
	CreateChunks(ctx context.Context, chunks []*entity.Chunk) error
	GetChunksByDocumentID(ctx context.Context, docID int64) ([]*entity.Chunk, error)
}

// PostgresChunkRepository implements ChunkRepository using PostgreSQL
type PostgresChunkRepository struct {
	db *sqlx.DB
}

// NewPostgresChunkRepository creates a new PostgresChunkRepository
func NewPostgresChunkRepository(db *sqlx.DB) *PostgresChunkRepository {
	return &PostgresChunkRepository{db: db}
}

// CreateChunks inserts multiple chunks into the database
func (r *PostgresChunkRepository) CreateChunks(ctx context.Context, chunks []*entity.Chunk) error {
	query := `
		INSERT INTO chunks (id, document_id, chunk_index, content, token_count)
		VALUES (:id, :document_id, :chunk_index, :content, :token_count)
	`

	// Create transaction
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Batch insert possible, but sqlx NamedExec is easiest for slice of structs
	_, err = tx.NamedExecContext(ctx, query, chunks)
	if err != nil {
		return fmt.Errorf("failed to insert chunks: %w", err)
	}

	return tx.Commit()
}

// GetChunksByDocumentID retrieves all chunks for a document
func (r *PostgresChunkRepository) GetChunksByDocumentID(ctx context.Context, docID int64) ([]*entity.Chunk, error) {
	chunks := []*entity.Chunk{}
	query := `SELECT * FROM chunks WHERE document_id = $1 ORDER BY chunk_index ASC`

	if err := r.db.SelectContext(ctx, &chunks, query, docID); err != nil {
		return nil, fmt.Errorf("failed to get chunks: %w", err)
	}
	return chunks, nil
}
