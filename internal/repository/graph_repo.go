package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/suyw-0123/graphweaver/internal/entity"
)

// GraphRepository defines the interface for graph data persistence.
type GraphRepository interface {
	CreateNode(ctx context.Context, node *entity.Node) error
	CreateEdge(ctx context.Context, edge *entity.Edge) error
	GetNodesByDocumentID(ctx context.Context, docID int64) ([]*entity.Node, error)
	GetEdgesByDocumentID(ctx context.Context, docID int64) ([]*entity.Edge, error)
	FindNode(ctx context.Context, docID int64, name, label string) (*entity.Node, error)
}

// PostgresGraphRepository implements GraphRepository using PostgreSQL.
type PostgresGraphRepository struct {
	db *sqlx.DB
}

// NewPostgresGraphRepository creates a new PostgresGraphRepository.
func NewPostgresGraphRepository(db *sqlx.DB) *PostgresGraphRepository {
	return &PostgresGraphRepository{db: db}
}

func (r *PostgresGraphRepository) CreateNode(ctx context.Context, node *entity.Node) error {
	query := `
		INSERT INTO nodes (document_id, label, name, properties, created_at)
		VALUES (:document_id, :label, :name, :properties, :created_at)
		RETURNING id
	`
	node.CreatedAt = time.Now()

	rows, err := r.db.NamedQueryContext(ctx, query, node)
	if err != nil {
		return fmt.Errorf("failed to create node: %w", err)
	}
	defer rows.Close()

	if rows.Next() {
		if err := rows.Scan(&node.ID); err != nil {
			return fmt.Errorf("failed to scan node id: %w", err)
		}
	}
	return nil
}

func (r *PostgresGraphRepository) CreateEdge(ctx context.Context, edge *entity.Edge) error {
	query := `
		INSERT INTO edges (document_id, source_node_id, target_node_id, relation_type, properties, created_at)
		VALUES (:document_id, :source_node_id, :target_node_id, :relation_type, :properties, :created_at)
		RETURNING id
	`
	edge.CreatedAt = time.Now()

	rows, err := r.db.NamedQueryContext(ctx, query, edge)
	if err != nil {
		return fmt.Errorf("failed to create edge: %w", err)
	}
	defer rows.Close()

	if rows.Next() {
		if err := rows.Scan(&edge.ID); err != nil {
			return fmt.Errorf("failed to scan edge id: %w", err)
		}
	}
	return nil
}

func (r *PostgresGraphRepository) GetNodesByDocumentID(ctx context.Context, docID int64) ([]*entity.Node, error) {
	nodes := []*entity.Node{}
	query := `SELECT * FROM nodes WHERE document_id = $1`
	if err := r.db.SelectContext(ctx, &nodes, query, docID); err != nil {
		return nil, fmt.Errorf("failed to list nodes: %w", err)
	}
	return nodes, nil
}

func (r *PostgresGraphRepository) GetEdgesByDocumentID(ctx context.Context, docID int64) ([]*entity.Edge, error) {
	edges := []*entity.Edge{}
	query := `SELECT * FROM edges WHERE document_id = $1`
	if err := r.db.SelectContext(ctx, &edges, query, docID); err != nil {
		return nil, fmt.Errorf("failed to list edges: %w", err)
	}
	return edges, nil
}

func (r *PostgresGraphRepository) FindNode(ctx context.Context, docID int64, name, label string) (*entity.Node, error) {
	var node entity.Node
	query := `SELECT * FROM nodes WHERE document_id = $1 AND name = $2 AND label = $3 LIMIT 1`
	if err := r.db.GetContext(ctx, &node, query, docID, name, label); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find node: %w", err)
	}
	return &node, nil
}
