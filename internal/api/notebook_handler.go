package api

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/suyw-0123/graphweaver/internal/service"
)

type NotebookHandler struct {
	notebookService *service.NotebookService
}

func NewNotebookHandler(notebookService *service.NotebookService) *NotebookHandler {
	return &NotebookHandler{notebookService: notebookService}
}

func (h *NotebookHandler) RegisterRoutes(r *gin.Engine) {
	v1 := r.Group("/api/v1")
	{
		v1.GET("/notebooks", h.ListNotebooks)
		v1.POST("/notebooks", h.CreateNotebook)
		v1.GET("/notebooks/:id", h.GetNotebook)
		v1.DELETE("/notebooks/:id", h.DeleteNotebook)
	}
}

func (h *NotebookHandler) ListNotebooks(c *gin.Context) {
	notebooks, err := h.notebookService.ListNotebooks(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, notebooks)
}

func (h *NotebookHandler) CreateNotebook(c *gin.Context) {
	var req struct {
		Title       string `json:"title" binding:"required"`
		Description string `json:"description"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	notebook, err := h.notebookService.CreateNotebook(c.Request.Context(), req.Title, req.Description)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, notebook)
}

func (h *NotebookHandler) GetNotebook(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	notebook, err := h.notebookService.GetNotebook(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Notebook not found"})
		return
	}
	c.JSON(http.StatusOK, notebook)
}

func (h *NotebookHandler) DeleteNotebook(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	if err := h.notebookService.DeleteNotebook(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Notebook deleted"})
}
