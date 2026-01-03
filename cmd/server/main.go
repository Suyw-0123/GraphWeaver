package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/suyw-0123/graphweaver/internal/api"
	"github.com/suyw-0123/graphweaver/internal/repository"
	"github.com/suyw-0123/graphweaver/internal/service"
	"github.com/suyw-0123/graphweaver/pkg/llm"
)

func main() {
	// Load .env file if it exists
	_ = godotenv.Load()

	fmt.Println("GraphWeaver API Server starting...")

	// Database Connection
	dbHost := os.Getenv("DB_HOST")
	if dbHost == "" {
		dbHost = "localhost"
	}
	dbPort := os.Getenv("DB_PORT")
	if dbPort == "" {
		dbPort = "5432"
	}
	dbUser := os.Getenv("DB_USER")
	if dbUser == "" {
		dbUser = "graphweaver"
	}
	dbPass := os.Getenv("DB_PASSWORD")
	if dbPass == "" {
		dbPass = "graphweaver123"
	}
	dbName := os.Getenv("DB_NAME")
	if dbName == "" {
		dbName = "graphweaver"
	}

	dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", dbUser, dbPass, dbHost, dbPort, dbName)
	
	db, err := repository.NewPostgresDB(dsn)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()
	log.Println("Connected to PostgreSQL")

	// LLM Client Initialization
	apiKey := os.Getenv("GEMINI_API_KEY")
	modelName := os.Getenv("GEMINI_MODEL_NAME")
	if apiKey == "" {
		log.Println("Warning: GEMINI_API_KEY is not set. LLM features will fail.")
	}

	llmClient, err := llm.NewGeminiClient(context.Background(), apiKey, modelName)
	if err != nil {
		log.Printf("Warning: Failed to initialize Gemini client: %v", err)
		// We don't fatal here to allow the server to start even if LLM is misconfigured,
		// but ingestion will fail.
	} else {
		defer llmClient.Close()
		log.Printf("Initialized Gemini Client with model: %s", modelName)
	}

	// Dependency Injection
	docRepo := repository.NewPostgresDocumentRepository(db)
	notebookRepo := repository.NewPostgresNotebookRepository(db)
	graphRepo := repository.NewPostgresGraphRepository(db)

	docService := service.NewDocumentService(docRepo, graphRepo)
	notebookService := service.NewNotebookService(notebookRepo)
	ingestionService := service.NewIngestionService(docRepo, graphRepo, llmClient, "uploads")
	chatService := service.NewChatService(docRepo, graphRepo, llmClient)

	docHandler := api.NewDocumentHandler(docService, ingestionService)
	notebookHandler := api.NewNotebookHandler(notebookService)
	chatHandler := api.NewChatHandler(chatService)

	// Router Setup
	r := gin.Default()
	
	// CORS Middleware (Basic)
	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	r.GET("/health", func(c *gin.Context) {
		c.String(http.StatusOK, "OK")
	})

	docHandler.RegisterRoutes(r)
	notebookHandler.RegisterRoutes(r)
	chatHandler.RegisterRoutes(r)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server listening on :%s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
