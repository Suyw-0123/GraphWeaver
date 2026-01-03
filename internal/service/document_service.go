package service

import (
	"context"
	"fmt"

	"github.com/suyw-0123/graphweaver/internal/entity"
	"github.com/suyw-0123/graphweaver/internal/repository"
)

type GraphData struct {
	Nodes []*entity.Node `json:"nodes"`
	Edges []*entity.Edge `json:"edges"`
}

// DocumentService defines the business logic for documents.
type DocumentService interface {
	UploadDocument(ctx context.Context, filename, filePath, mimeType string, fileSize int64) (*entity.Document, error)
	GetDocument(ctx context.Context, id int64) (*entity.Document, error)
	ListDocuments(ctx context.Context, page, pageSize int, notebookID *int64) ([]*entity.Document, error)
	GetGraph(ctx context.Context, docID int64) (*GraphData, error)
}

// documentService implements DocumentService.
type documentService struct {
	repo      repository.DocumentRepository
	graphRepo repository.GraphRepository
}

// NewDocumentService creates a new DocumentService.
func NewDocumentService(repo repository.DocumentRepository, graphRepo repository.GraphRepository) DocumentService {
	return &documentService{repo: repo, graphRepo: graphRepo}
}

// UploadDocument handles the metadata creation for a new document.
// Note: The actual file upload logic (saving to disk/S3) should be handled by the handler or another service before calling this.
func (s *documentService) UploadDocument(ctx context.Context, filename, filePath, mimeType string, fileSize int64) (*entity.Document, error) {
	doc := &entity.Document{
		Filename: filename,
		FilePath: filePath,
		MimeType: mimeType,
		FileSize: fileSize,
		Status:   "pending",
	}

	if err := s.repo.Create(ctx, doc); err != nil {
		return nil, fmt.Errorf("service: failed to create document: %w", err)
	}

	return doc, nil
}

// GetDocument retrieves a document by ID.
func (s *documentService) GetDocument(ctx context.Context, id int64) (*entity.Document, error) {
	doc, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("service: failed to get document: %w", err)
	}
	if doc == nil {
		return nil, fmt.Errorf("service: document not found")
	}
	return doc, nil
}

// GetGraph retrieves the graph data for a document.
func (s *documentService) GetGraph(ctx context.Context, docID int64) (*GraphData, error) {
	nodes, err := s.graphRepo.GetNodesByDocumentID(ctx, docID)
	if err != nil {
		return nil, fmt.Errorf("service: failed to get nodes: %w", err)
	}
	edges, err := s.graphRepo.GetEdgesByDocumentID(ctx, docID)
	if err != nil {
		return nil, fmt.Errorf("service: failed to get edges: %w", err)
	}
	return &GraphData{Nodes: nodes, Edges: edges}, nil
}

// ListDocuments retrieves a paginated list of documents.
func (s *documentService) ListDocuments(ctx context.Context, page, pageSize int, notebookID *int64) ([]*entity.Document, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}
	offset := (page - 1) * pageSize

	docs, err := s.repo.List(ctx, pageSize, offset, notebookID)
	if err != nil {
		return nil, fmt.Errorf("service: failed to list documents: %w", err)
	}
	return docs, nil
}
