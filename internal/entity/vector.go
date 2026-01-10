package entity

// VectorPoint represents a data point in the vector database
type VectorPoint struct {
	ID      string                 `json:"id"`
	Vector  []float32              `json:"vector"`
	Payload map[string]interface{} `json:"payload"`
}

// Chunk represents a segment of a document
type Chunk struct {
	ID         string    `json:"id" db:"id"`
	DocumentID int64     `json:"document_id" db:"document_id"`
	Content    string    `json:"content" db:"content"`
	Index      int       `json:"index" db:"chunk_index"`
	TokenCount int       `json:"token_count" db:"token_count"`
	Embedding  []float32 `json:"embedding,omitempty" db:"-"` // Not stored in SQL directly
}
