package entity

import (
	"time"
)

// Document represents a file uploaded to the system.
type Document struct {
	ID           int64     `db:"id" json:"id"`
	Filename     string    `db:"filename" json:"filename"`
	FilePath     string    `db:"file_path" json:"file_path"`
	MimeType     string    `db:"mime_type" json:"mime_type"`
	FileSize     int64     `db:"file_size" json:"file_size"`
	Status       string    `db:"status" json:"status"` // pending, processing, completed, failed
	ErrorMessage *string   `db:"error_message" json:"error_message,omitempty"`
	Summary      *string   `db:"summary" json:"summary,omitempty"`
	NotebookID   *int64    `db:"notebook_id" json:"notebook_id,omitempty"`
	CreatedAt    time.Time `db:"created_at" json:"created_at"`
	UpdatedAt    time.Time `db:"updated_at" json:"updated_at"`
	IsDeleted    bool      `db:"is_deleted" json:"is_deleted"`
}

// ProcessingJob represents a stage in the document processing pipeline.
type ProcessingJob struct {
	ID          int64      `db:"id" json:"id"`
	DocumentID  int64      `db:"document_id" json:"document_id"`
	Stage       string     `db:"stage" json:"stage"` // extraction, embedding, graph_sync
	Status      string     `db:"status" json:"status"`
	StartedAt   *time.Time `db:"started_at" json:"started_at,omitempty"`
	CompletedAt *time.Time `db:"completed_at" json:"completed_at,omitempty"`
	CreatedAt   time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt   time.Time  `db:"updated_at" json:"updated_at"`
}
