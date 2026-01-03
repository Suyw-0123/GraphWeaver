package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/suyw-0123/graphweaver/internal/repository"
	"github.com/suyw-0123/graphweaver/pkg/llm"
)

type ChatService interface {
	Chat(ctx context.Context, notebookID int64, query string) (string, error)
}

type chatService struct {
	docRepo   repository.DocumentRepository
	graphRepo repository.GraphRepository
	llmClient llm.Client
}

func NewChatService(docRepo repository.DocumentRepository, graphRepo repository.GraphRepository, llmClient llm.Client) ChatService {
	return &chatService{
		docRepo:   docRepo,
		graphRepo: graphRepo,
		llmClient: llmClient,
	}
}

func (s *chatService) Chat(ctx context.Context, notebookID int64, query string) (string, error) {
	// 1. Get all documents for the notebook
	docs, err := s.docRepo.List(ctx, 100, 0, &notebookID)
	if err != nil {
		return "", fmt.Errorf("failed to list documents: %w", err)
	}

	if len(docs) == 0 {
		return "This notebook has no documents. Please upload some documents first.", nil
	}

	// 2. Collect context from Graph
	var contextBuilder strings.Builder
	contextBuilder.WriteString("Knowledge Graph Context:\n")

	for _, doc := range docs {
		nodes, err := s.graphRepo.GetNodesByDocumentID(ctx, doc.ID)
		if err != nil {
			continue
		}
		edges, err := s.graphRepo.GetEdgesByDocumentID(ctx, doc.ID)
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

		contextBuilder.WriteString(fmt.Sprintf("\nSource: %s\n", doc.Filename))
		
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

	// 3. Construct Prompt
	prompt := fmt.Sprintf(`You are a helpful assistant for a Knowledge Graph application.
Use the following Context to answer the User's Question.
The Context consists of Entities and Relationships extracted from documents.

Context:
%s

User Question: %s

Answer:`, contextBuilder.String(), query)

	// 4. Call LLM
	return s.llmClient.GenerateContent(ctx, prompt)
}
