package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/suyw-0123/graphweaver/internal/entity"
)

type NotebookRepository interface {
	Create(ctx context.Context, notebook *entity.Notebook) error
	GetByID(ctx context.Context, id int64) (*entity.Notebook, error)
	List(ctx context.Context) ([]*entity.Notebook, error)
	Delete(ctx context.Context, id int64) error
}

type PostgresNotebookRepository struct {
	db *sqlx.DB
}

func NewPostgresNotebookRepository(db *sqlx.DB) *PostgresNotebookRepository {
	return &PostgresNotebookRepository{db: db}
}

func (r *PostgresNotebookRepository) Create(ctx context.Context, notebook *entity.Notebook) error {
	query := `
		INSERT INTO notebooks (title, description, created_at, updated_at)
		VALUES (:title, :description, :created_at, :updated_at)
		RETURNING id
	`
	notebook.CreatedAt = time.Now()
	notebook.UpdatedAt = time.Now()

	rows, err := r.db.NamedQueryContext(ctx, query, notebook)
	if err != nil {
		return fmt.Errorf("failed to create notebook: %w", err)
	}
	defer rows.Close()

	if rows.Next() {
		if err := rows.Scan(&notebook.ID); err != nil {
			return fmt.Errorf("failed to scan notebook id: %w", err)
		}
	}
	return nil
}

func (r *PostgresNotebookRepository) GetByID(ctx context.Context, id int64) (*entity.Notebook, error) {
	query := `SELECT * FROM notebooks WHERE id = $1`
	var notebook entity.Notebook
	if err := r.db.GetContext(ctx, &notebook, query, id); err != nil {
		return nil, fmt.Errorf("failed to get notebook: %w", err)
	}
	return &notebook, nil
}

func (r *PostgresNotebookRepository) List(ctx context.Context) ([]*entity.Notebook, error) {
	query := `SELECT * FROM notebooks ORDER BY created_at DESC`
	notebooks := []*entity.Notebook{} // Initialize as empty slice
	if err := r.db.SelectContext(ctx, &notebooks, query); err != nil {
		return nil, fmt.Errorf("failed to list notebooks: %w", err)
	}
	return notebooks, nil
}

func (r *PostgresNotebookRepository) Delete(ctx context.Context, id int64) error {
	query := `DELETE FROM notebooks WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete notebook: %w", err)
	}
	return nil
}
