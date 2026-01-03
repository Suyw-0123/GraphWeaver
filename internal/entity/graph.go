package entity

import "time"

// Node represents a node in the knowledge graph.
type Node struct {
	ID         int64     `db:"id" json:"id"`
	DocumentID int64     `db:"document_id" json:"document_id"`
	Label      string    `db:"label" json:"label"` // e.g., "Person", "Location"
	Name       string    `db:"name" json:"name"`   // e.g., "Alice", "New York"
	Properties string    `db:"properties" json:"properties"` // JSON string
	CreatedAt  time.Time `db:"created_at" json:"created_at"`
}

// Edge represents a relationship between two nodes.
type Edge struct {
	ID           int64     `db:"id" json:"id"`
	DocumentID   int64     `db:"document_id" json:"document_id"`
	SourceNodeID int64     `db:"source_node_id" json:"source_node_id"`
	TargetNodeID int64     `db:"target_node_id" json:"target_node_id"`
	RelationType string    `db:"relation_type" json:"relation_type"` // e.g., "LIVES_IN"
	Properties   string    `db:"properties" json:"properties"` // JSON string
	CreatedAt    time.Time `db:"created_at" json:"created_at"`
}
