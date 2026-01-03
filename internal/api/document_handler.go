package api

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/suyw-0123/graphweaver/internal/service"
)

// DocumentHandler handles HTTP requests for documents.
type DocumentHandler struct {
	docService       service.DocumentService
	ingestionService service.IngestionService
}

// NewDocumentHandler creates a new DocumentHandler.
func NewDocumentHandler(docService service.DocumentService, ingestionService service.IngestionService) *DocumentHandler {
	return &DocumentHandler{
		docService:       docService,
		ingestionService: ingestionService,
	}
}

// RegisterRoutes registers the document routes.
func (h *DocumentHandler) RegisterRoutes(router *gin.Engine) {
	v1 := router.Group("/api/v1")
	{
		v1.POST("/documents/upload", h.UploadDocument)
		v1.GET("/documents", h.ListDocuments)
		v1.GET("/documents/:id", h.GetDocument)
		v1.GET("/documents/:id/graph", h.GetDocumentGraph)
	}
}

// UploadDocument handles file upload and ingestion triggering.
func (h *DocumentHandler) UploadDocument(c *gin.Context) {
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file is required"})
		return
	}
	defer file.Close()

	notebookIDStr := c.Request.FormValue("notebook_id")
	fmt.Printf("UploadDocument: Received upload request for notebook_id=%s, filename=%s\n", notebookIDStr, header.Filename)

	var notebookID *int64
	if notebookIDStr != "" {
		id, err := strconv.ParseInt(notebookIDStr, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid notebook_id"})
			return
		}
		notebookID = &id
	}

	doc, err := h.ingestionService.ProcessUpload(c.Request.Context(), file, header, notebookID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, doc)
}

// GetDocument handles retrieving a single document.
func (h *DocumentHandler) GetDocument(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	doc, err := h.docService.GetDocument(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, doc)
}

// GetDocumentGraph handles retrieving the graph data for a document.
func (h *DocumentHandler) GetDocumentGraph(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	graph, err := h.docService.GetGraph(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, graph)
}

// ListDocuments handles listing documents.
func (h *DocumentHandler) ListDocuments(c *gin.Context) {
	pageStr := c.DefaultQuery("page", "1")
	pageSizeStr := c.DefaultQuery("page_size", "10")
	notebookIDStr := c.Query("notebook_id")

	page, _ := strconv.Atoi(pageStr)
	pageSize, _ := strconv.Atoi(pageSizeStr)

	var notebookID *int64
	if notebookIDStr != "" {
		id, err := strconv.ParseInt(notebookIDStr, 10, 64)
		if err == nil {
			notebookID = &id
		}
	}

	docs, err := h.docService.ListDocuments(c.Request.Context(), page, pageSize, notebookID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, docs)
}
