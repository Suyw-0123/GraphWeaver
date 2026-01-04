package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/suyw-0123/graphweaver/internal/entity"
)

// DocumentRepository defines the interface for document persistence.
type DocumentRepository interface {
	Create(ctx context.Context, doc *entity.Document) error
	GetByID(ctx context.Context, id int64) (*entity.Document, error)
	List(ctx context.Context, limit, offset int, notebookID *int64) ([]*entity.Document, error)
	UpdateStatus(ctx context.Context, id int64, status string, errorMessage *string) error
	UpdateSummary(ctx context.Context, id int64, summary string) error
}

// PostgresDocumentRepository implements DocumentRepository using PostgreSQL.
type PostgresDocumentRepository struct {
	db *sqlx.DB
}

// NewPostgresDocumentRepository creates a new PostgresDocumentRepository.
func NewPostgresDocumentRepository(db *sqlx.DB) *PostgresDocumentRepository {
	return &PostgresDocumentRepository{db: db}
}

// Create inserts a new document into the database.
func (r *PostgresDocumentRepository) Create(ctx context.Context, doc *entity.Document) error {
	query := `
		INSERT INTO documents (filename, file_path, mime_type, file_size, status, notebook_id, created_at, updated_at)
		VALUES (:filename, :file_path, :mime_type, :file_size, :status, :notebook_id, :created_at, :updated_at)
		RETURNING id
	`

	doc.CreatedAt = time.Now()
	doc.UpdatedAt = time.Now()
	if doc.Status == "" {
		doc.Status = "pending"
	}

	rows, err := r.db.NamedQueryContext(ctx, query, doc)
	if err != nil {
		return fmt.Errorf("failed to create document: %w", err)
	}
	defer rows.Close()

	if rows.Next() {
		if err := rows.Scan(&doc.ID); err != nil {
			return fmt.Errorf("failed to scan document id: %w", err)
		}
	}

	return nil
}

// GetByID retrieves a document by its ID.
func (r *PostgresDocumentRepository) GetByID(ctx context.Context, id int64) (*entity.Document, error) {
	var doc entity.Document
	query := `SELECT * FROM documents WHERE id = $1 AND is_deleted = false`

	if err := r.db.GetContext(ctx, &doc, query, id); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Or return a custom ErrNotFound
		}
		return nil, fmt.Errorf("failed to get document: %w", err)
	}

	return &doc, nil
}

// List retrieves a list of documents with pagination.
func (r *PostgresDocumentRepository) List(ctx context.Context, limit, offset int, notebookID *int64) ([]*entity.Document, error) {
	docs := []*entity.Document{} // Initialize as empty slice
	query := `
		SELECT * FROM documents 
		WHERE is_deleted = false
	`
	args := []interface{}{}
	argIdx := 1

	if notebookID != nil {
		query += fmt.Sprintf(" AND notebook_id = $%d", argIdx)
		args = append(args, *notebookID)
		argIdx++
	}

	query += fmt.Sprintf(" ORDER BY created_at DESC LIMIT $%d OFFSET $%d", argIdx, argIdx+1)
	args = append(args, limit, offset)

	if err := r.db.SelectContext(ctx, &docs, query, args...); err != nil {
		return nil, fmt.Errorf("failed to list documents: %w", err)
	}

	return docs, nil
}

// UpdateStatus updates the status of a document.
func (r *PostgresDocumentRepository) UpdateStatus(ctx context.Context, id int64, status string, errorMessage *string) error {
	query := `
		UPDATE documents 
		SET status = $1, error_message = $2, updated_at = $3 
		WHERE id = $4
	`

	_, err := r.db.ExecContext(ctx, query, status, errorMessage, time.Now(), id)
	if err != nil {
		return fmt.Errorf("failed to update document status: %w", err)
	}

	return nil
}

// UpdateSummary updates the summary of a document.
func (r *PostgresDocumentRepository) UpdateSummary(ctx context.Context, id int64, summary string) error {
	query := `
		UPDATE documents 
		SET summary = $1, updated_at = $2 
		WHERE id = $3
	`

	_, err := r.db.ExecContext(ctx, query, summary, time.Now(), id)
	if err != nil {
		return fmt.Errorf("failed to update document summary: %w", err)
	}
	return nil
}
