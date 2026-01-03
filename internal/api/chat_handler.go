package api

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/suyw-0123/graphweaver/internal/service"
)

type ChatHandler struct {
	chatService service.ChatService
}

func NewChatHandler(chatService service.ChatService) *ChatHandler {
	return &ChatHandler{chatService: chatService}
}

func (h *ChatHandler) RegisterRoutes(r *gin.Engine) {
	v1 := r.Group("/api/v1")
	{
		v1.POST("/notebooks/:id/chat", h.Chat)
	}
}

func (h *ChatHandler) Chat(c *gin.Context) {
	idStr := c.Param("id")
	notebookID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid notebook ID"})
		return
	}

	var req struct {
		Query string `json:"query" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Query is required"})
		return
	}

	answer, err := h.chatService.Chat(c.Request.Context(), notebookID, req.Query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"answer": answer})
}
