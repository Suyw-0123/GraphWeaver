package service

import (
	"context"
	"fmt"

	"github.com/suyw-0123/graphweaver/internal/entity"
	"github.com/suyw-0123/graphweaver/internal/repository"
)

type NotebookService struct {
	repo repository.NotebookRepository
}

func NewNotebookService(repo repository.NotebookRepository) *NotebookService {
	return &NotebookService{repo: repo}
}

func (s *NotebookService) CreateNotebook(ctx context.Context, title, description string) (*entity.Notebook, error) {
	notebook := &entity.Notebook{
		Title:       title,
		Description: description,
	}
	if err := s.repo.Create(ctx, notebook); err != nil {
		return nil, fmt.Errorf("failed to create notebook: %w", err)
	}
	return notebook, nil
}

func (s *NotebookService) ListNotebooks(ctx context.Context) ([]*entity.Notebook, error) {
	return s.repo.List(ctx)
}

func (s *NotebookService) GetNotebook(ctx context.Context, id int64) (*entity.Notebook, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *NotebookService) DeleteNotebook(ctx context.Context, id int64) error {
	// In the future, this will also trigger deletions in Vector DB and Neo4j
	return s.repo.Delete(ctx, id)
}
