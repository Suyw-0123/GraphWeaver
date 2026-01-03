package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/suyw-0123/graphweaver/internal/entity"
	"github.com/suyw-0123/graphweaver/internal/repository"
	"github.com/suyw-0123/graphweaver/pkg/llm"
	"github.com/suyw-0123/graphweaver/pkg/parser"
)

// IngestionService defines the logic for processing uploaded files.
type IngestionService interface {
	ProcessUpload(ctx context.Context, file multipart.File, header *multipart.FileHeader, notebookID *int64) (*entity.Document, error)
}

type ingestionService struct {
	docRepo   repository.DocumentRepository
	graphRepo repository.GraphRepository
	llmClient llm.Client
	uploadDir string
}

// NewIngestionService creates a new IngestionService.
func NewIngestionService(docRepo repository.DocumentRepository, graphRepo repository.GraphRepository, llmClient llm.Client, uploadDir string) IngestionService {
	// Ensure upload directory exists
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		// In a real app, we might want to handle this error more gracefully or panic at startup
		fmt.Printf("Warning: failed to create upload dir: %v\n", err)
	}
	return &ingestionService{
		docRepo:   docRepo,
		graphRepo: graphRepo,
		llmClient: llmClient,
		uploadDir: uploadDir,
	}
}

func (s *ingestionService) ProcessUpload(ctx context.Context, file multipart.File, header *multipart.FileHeader, notebookID *int64) (*entity.Document, error) {
	// 0. Check if notebook already has a document
	if notebookID != nil {
		existingDocs, err := s.docRepo.List(ctx, 1, 0, notebookID)
		if err != nil {
			return nil, fmt.Errorf("failed to check existing documents: %w", err)
		}
		if len(existingDocs) > 0 {
			return nil, fmt.Errorf("notebook already contains a document. Only one document per notebook is allowed.")
		}
	}

	// 1. Save file to local storage
	filename := fmt.Sprintf("%d_%s", time.Now().Unix(), header.Filename)
	filePath := filepath.Join(s.uploadDir, filename)

	dst, err := os.Create(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to create destination file: %w", err)
	}
	defer dst.Close()

	if _, err := io.Copy(dst, file); err != nil {
		return nil, fmt.Errorf("failed to save file: %w", err)
	}

	// 2. Create Document Metadata
	doc := &entity.Document{
		Filename:   header.Filename,
		FilePath:   filePath,
		MimeType:   header.Header.Get("Content-Type"),
		FileSize:   header.Size,
		Status:     "processing",
		NotebookID: notebookID,
	}

	if err := s.docRepo.Create(ctx, doc); err != nil {
		return nil, fmt.Errorf("failed to create document record: %w", err)
	}

	// 3. Trigger Async Processing (Goroutine)
	// In a production system, this should be a job queue (e.g., Redis/Kafka)
	go s.processDocumentAsync(doc.ID, filePath)

	return doc, nil
}

func (s *ingestionService) processDocumentAsync(docID int64, filePath string) {
	// Create a background context for the async job
	ctx := context.Background()

	// Update status to processing (already set, but good practice)
	_ = s.docRepo.UpdateStatus(ctx, docID, "processing", nil)

	// 1. Parse File
	text, err := parser.ParseFile(filePath)
	if err != nil {
		errMsg := fmt.Sprintf("parsing failed: %v", err)
		_ = s.docRepo.UpdateStatus(ctx, docID, "failed", &errMsg)
		return
	}

	// 2. Call LLM for Entity Extraction
	prompt := fmt.Sprintf(`
You are a knowledge graph expert. Extract entities and relationships from the following text.
Return ONLY a valid JSON object with the following structure:
{
  "summary": "A brief summary of the text (max 50 words)",
  "entities": [
    {"name": "Entity Name", "label": "Person/Location/Organization/Concept", "description": "Brief description"}
  ],
  "relations": [
    {"source": "Entity Name", "target": "Entity Name", "type": "RELATION_TYPE", "description": "Context of relation"}
  ]
}

Text to analyze:
%s
`, text)

	if len(text) > 10000 {
		prompt = fmt.Sprintf(`
You are a knowledge graph expert. Extract entities and relationships from the following text.
Return ONLY a valid JSON object with the following structure:
{
  "summary": "A brief summary of the text (max 50 words)",
  "entities": [
    {"name": "Entity Name", "label": "Person/Location/Organization/Concept", "description": "Brief description"}
  ],
  "relations": [
    {"source": "Entity Name", "target": "Entity Name", "type": "RELATION_TYPE", "description": "Context of relation"}
  ]
}

Text to analyze:
%s... (truncated)
`, text[:10000])
	}

	response, err := s.llmClient.GenerateContent(ctx, prompt)
	if err != nil {
		errMsg := fmt.Sprintf("llm generation failed: %v", err)
		_ = s.docRepo.UpdateStatus(ctx, docID, "failed", &errMsg)
		return
	}

	// Clean up response (remove markdown code blocks if present)
	jsonStr := strings.TrimSpace(response)
	if strings.HasPrefix(jsonStr, "```json") {
		jsonStr = strings.TrimPrefix(jsonStr, "```json")
		jsonStr = strings.TrimSuffix(jsonStr, "```")
	} else if strings.HasPrefix(jsonStr, "```") {
		jsonStr = strings.TrimPrefix(jsonStr, "```")
		jsonStr = strings.TrimSuffix(jsonStr, "```")
	}
	jsonStr = strings.TrimSpace(jsonStr)

	// Parse JSON
	var result struct {
		Summary  string `json:"summary"`
		Entities []struct {
			Name  string `json:"name"`
			Label string `json:"label"`
			Desc  string `json:"description"`
		} `json:"entities"`
		Relations []struct {
			Source string `json:"source"`
			Target string `json:"target"`
			Type   string `json:"type"`
			Desc   string `json:"description"`
		} `json:"relations"`
	}

	if err := json.Unmarshal([]byte(jsonStr), &result); err != nil {
		// Fallback: try to save just the raw response as summary if JSON parsing fails
		// But for now, let's mark as failed to debug prompt issues
		errMsg := fmt.Sprintf("failed to parse LLM response: %v. Response: %s", err, response)
		_ = s.docRepo.UpdateStatus(ctx, docID, "failed", &errMsg)
		return
	}

	// 3. Save Summary
	if err := s.docRepo.UpdateSummary(ctx, docID, result.Summary); err != nil {
		errMsg := fmt.Sprintf("failed to save summary: %v", err)
		_ = s.docRepo.UpdateStatus(ctx, docID, "failed", &errMsg)
		return
	}

	// 4. Save Graph Data
	// Save Entities (Nodes)
	nodeMap := make(map[string]int64) // Name -> ID
	for _, e := range result.Entities {
		// Check if node already exists for this document (or globally? For now, per document scope or simple dedupe)
		// We'll use FindNode to check within this document context
		existingNode, _ := s.graphRepo.FindNode(ctx, docID, e.Name, e.Label)
		if existingNode != nil {
			nodeMap[e.Name] = existingNode.ID
			continue
		}

		node := &entity.Node{
			DocumentID: docID,
			Label:      e.Label,
			Name:       e.Name,
			Properties: fmt.Sprintf(`{"description": "%s"}`, strings.ReplaceAll(e.Desc, "\"", "\\\"")),
		}
		if err := s.graphRepo.CreateNode(ctx, node); err != nil {
			fmt.Printf("Error creating node %s: %v\n", e.Name, err)
			continue
		}
		nodeMap[e.Name] = node.ID
	}

	// Save Relations (Edges)
	for _, r := range result.Relations {
		sourceID, ok1 := nodeMap[r.Source]
		targetID, ok2 := nodeMap[r.Target]

		if !ok1 || !ok2 {
			// Skip if nodes not found (maybe LLM hallucinated a relation with a non-extracted entity)
			continue
		}

		edge := &entity.Edge{
			DocumentID:   docID,
			SourceNodeID: sourceID,
			TargetNodeID: targetID,
			RelationType: r.Type,
			Properties:   fmt.Sprintf(`{"description": "%s"}`, strings.ReplaceAll(r.Desc, "\"", "\\\"")),
		}
		if err := s.graphRepo.CreateEdge(ctx, edge); err != nil {
			fmt.Printf("Error creating edge %s->%s: %v\n", r.Source, r.Target, err)
		}
	}

	// 5. Mark as Completed
	_ = s.docRepo.UpdateStatus(ctx, docID, "completed", nil)
}
