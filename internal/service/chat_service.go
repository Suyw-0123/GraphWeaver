package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/suyw-0123/graphweaver/internal/repository"
	"github.com/suyw-0123/graphweaver/pkg/embedding"
	"github.com/suyw-0123/graphweaver/pkg/llm"
)

type ChatService interface {
	Chat(ctx context.Context, notebookID int64, query string) (string, error)
}

type chatService struct {
	docRepo         repository.DocumentRepository
	graphRepo       repository.GraphRepository
	vectorRepo      repository.VectorRepository
	llmClient       llm.Client
	embeddingClient embedding.Client
}

func NewChatService(
	docRepo repository.DocumentRepository,
	graphRepo repository.GraphRepository,
	vectorRepo repository.VectorRepository,
	llmClient llm.Client,
	embeddingClient embedding.Client,
) ChatService {
	return &chatService{
		docRepo:         docRepo,
		graphRepo:       graphRepo,
		vectorRepo:      vectorRepo,
		llmClient:       llmClient,
		embeddingClient: embeddingClient,
	}
}

func (s *chatService) Chat(ctx context.Context, notebookID int64, query string) (string, error) {
	// 0. Hybrid Retrieval Setup
	var relevantDocIDs []int64
	var textContextBuilder strings.Builder

	useVectorSearch := s.embeddingClient != nil && s.vectorRepo != nil

	if useVectorSearch {
		// A. Generate Query Embedding
		queryVector, err := s.embeddingClient.EmbedText(ctx, query)
		if err != nil {
			fmt.Printf("Warning: Failed to embed query, falling back to full scan: %v\n", err)
			useVectorSearch = false
		} else {
			// B. Vector Search
			// We search in "documents" collection. In real world, filter by notebookID if possible (payload filter).
			// Currently our VectorRepo Search doesn't support filters, so we filtered by payload manually or accept global search?
			// The proposal didn't specify adding filtering support to Repo yet.
			// Ideally we should filter by notebookID.
			// For now, we search globally and filter results in code (inefficient but works for MVP).
			results, err := s.vectorRepo.Search(ctx, "documents", queryVector, 20, 0.6)
			if err != nil {
				fmt.Printf("Warning: Vector search failed: %v\n", err)
				useVectorSearch = false
			} else {
				// C. Process Results
				docIDMap := make(map[int64]bool)
				count := 0

				textContextBuilder.WriteString("Relevant Text Segments:\n")

				for _, res := range results {
					if count >= 5 {
						break
					} // Limit to top 5 valid checks

					// Extract fields
					docIDVal, ok1 := res.Payload["document_id"]
					contentVal, ok2 := res.Payload["content"]

					if !ok1 || !ok2 {
						continue
					}

					docID := int64(0)
					switch v := docIDVal.(type) {
					case int64:
						docID = v
					case int:
						docID = int64(v)
					case float64:
						docID = int64(v) // JSON often parses numbers as floats
					default:
						continue
					}

					// Verify this doc belongs to the notebook?
					// We need to fetch the doc to check notebookID.
					// Optimization: Cache or trust?
					// Or just check only docs belonging to current notebook.
					// The Chat function signature has notebookID.

					// We'll filter in memory: verify doc belongs to notebook
					doc, err := s.docRepo.GetByID(ctx, docID)
					if err != nil || doc.NotebookID == nil || *doc.NotebookID != notebookID {
						continue
					}

					if !docIDMap[docID] {
						docIDMap[docID] = true
						relevantDocIDs = append(relevantDocIDs, docID)
					}

					textContextBuilder.WriteString(fmt.Sprintf("- ...%s...\n", contentVal))
					count++
				}

				if len(relevantDocIDs) == 0 {
					fmt.Println("No relevant chunks found via vector search for this notebook. Falling back to all documents.")
					useVectorSearch = false
				}
			}
		}
	}

	// Fallback or Basic Retrieval
	if !useVectorSearch || len(relevantDocIDs) == 0 {
		// 1. Get all documents for the notebook
		docs, err := s.docRepo.List(ctx, 100, 0, &notebookID)
		if err != nil {
			return "", fmt.Errorf("failed to list documents: %w", err)
		}
		if len(docs) == 0 {
			return "This notebook has no documents. Please upload some documents first.", nil
		}
		for _, d := range docs {
			relevantDocIDs = append(relevantDocIDs, d.ID)
		}
	}

	// 2. Collect context from Graph (for relevant documents)
	var contextBuilder strings.Builder
	contextBuilder.WriteString("Context information is below.\n---------------------\n")

	if textContextBuilder.Len() > 0 {
		contextBuilder.WriteString(textContextBuilder.String())
		contextBuilder.WriteString("\n")
	}

	contextBuilder.WriteString("Knowledge Graph:\n")

	for _, docID := range relevantDocIDs {
		// Get Doc info for Source name
		doc, err := s.docRepo.GetByID(ctx, docID)
		filename := "Unknown"
		if err == nil {
			filename = doc.Filename
		}

		nodes, err := s.graphRepo.GetNodesByDocumentID(ctx, docID)
		if err != nil {
			continue
		}
		edges, err := s.graphRepo.GetEdgesByDocumentID(ctx, docID)
		if err != nil {
			continue
		}

		if len(nodes) == 0 {
			continue
		}

		// Build Node Map for Edge resolution
		nodeMap := make(map[int64]string)
		for _, node := range nodes {
			nodeMap[node.ID] = node.Name
		}

		contextBuilder.WriteString(fmt.Sprintf("\nSource: %s\n", filename))
		contextBuilder.WriteString("Entities:\n")
		for _, node := range nodes {
			contextBuilder.WriteString(fmt.Sprintf("- %s (%s)\n", node.Name, node.Label))
		}
		contextBuilder.WriteString("Relationships:\n")
		for _, edge := range edges {
			sourceName, ok1 := nodeMap[edge.SourceNodeID]
			targetName, ok2 := nodeMap[edge.TargetNodeID]
			if ok1 && ok2 {
				contextBuilder.WriteString(fmt.Sprintf("- %s --[%s]--> %s\n", sourceName, edge.RelationType, targetName))
			}
		}
	}

	contextBuilder.WriteString("---------------------\n")

	// 3. Construct Prompt
	prompt := fmt.Sprintf(`You are a helpful assistant for a Knowledge Graph application.
Use the following Context to answer the User's Question.
The Context consists of Text Segments and Graph Entities/Relationships from documents.
If the answer is not in the context, say you don't know.

Context:
%s

User Question: %s

Answer:`, contextBuilder.String(), query)

	// 4. Call LLM
	return s.llmClient.GenerateContent(ctx, prompt)
}
