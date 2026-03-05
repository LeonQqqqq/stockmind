package handler

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"stockmind-go/internal/model"
	"stockmind-go/internal/service"
	"stockmind-go/internal/store"
)

type Handler struct {
	chatSvc *service.ChatService
	store   *store.SQLiteStore
}

func NewHandler(chatSvc *service.ChatService, store *store.SQLiteStore) *Handler {
	return &Handler{chatSvc: chatSvc, store: store}
}

func (h *Handler) RegisterRoutes(r *gin.Engine) {
	api := r.Group("/api/v1")
	{
		api.POST("/chat/stream", h.ChatStream)
		api.GET("/sessions", h.ListSessions)
		api.POST("/sessions", h.CreateSession)
		api.DELETE("/sessions/:id", h.DeleteSession)
		api.GET("/sessions/:id/messages", h.GetMessages)
		api.GET("/experiences", h.ListExperiences)
		api.POST("/experiences", h.CreateExperience)
		api.PUT("/experiences/:id", h.UpdateExperience)
		api.DELETE("/experiences/:id", h.DeleteExperience)
		api.GET("/experiences/search", h.SearchExperiences)
		api.GET("/opinions", h.ListOpinions)
		api.POST("/opinions", h.CreateOpinion)
		api.DELETE("/opinions/:id", h.DeleteOpinion)
		api.GET("/opinions/authors", h.ListAuthors)
		api.GET("/opinions/search", h.SearchOpinions)
	}
}

// writeSSE writes a raw SSE event without JSON encoding.
func writeSSE(w io.Writer, event, data string) {
	fmt.Fprintf(w, "event: %s\n", event)
	for _, line := range strings.Split(data, "\n") {
		fmt.Fprintf(w, "data: %s\n", line)
	}
	fmt.Fprintf(w, "\n")
}

// ChatStream handles SSE streaming chat
func (h *Handler) ChatStream(c *gin.Context) {
	var req model.ChatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.SessionID == "" {
		req.SessionID = uuid.New().String()
		title := req.Message
		if len(title) > 50 {
			title = title[:50] + "..."
		}
		h.store.CreateSession(req.SessionID, title)
	}

	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("X-Accel-Buffering", "no")

	writeSSE(c.Writer, "session", req.SessionID)
	c.Writer.Flush()

	textCh := make(chan string, 10)
	errCh := make(chan error, 1)
	go func() {
		errCh <- h.chatSvc.Chat(req.SessionID, req.Message, textCh)
	}()

	for text := range textCh {
		writeSSE(c.Writer, "message", text)
		c.Writer.Flush()
	}

	if err := <-errCh; err != nil {
		log.Printf("Chat error: %v", err)
		writeSSE(c.Writer, "error", err.Error())
		c.Writer.Flush()
	}

	writeSSE(c.Writer, "done", "[DONE]")
	c.Writer.Flush()
}

// === Sessions ===

func (h *Handler) ListSessions(c *gin.Context) {
	sessions, err := h.store.ListSessions()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": sessions})
}

func (h *Handler) CreateSession(c *gin.Context) {
	var body struct {
		Title string `json:"title"`
	}
	c.ShouldBindJSON(&body)
	id := uuid.New().String()
	if body.Title == "" {
		body.Title = "新对话"
	}
	if err := h.store.CreateSession(id, body.Title); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": model.Session{ID: id, Title: body.Title}})
}

func (h *Handler) DeleteSession(c *gin.Context) {
	id := c.Param("id")
	if err := h.store.DeleteSession(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "ok"})
}

func (h *Handler) GetMessages(c *gin.Context) {
	id := c.Param("id")
	msgs, err := h.store.GetMessages(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": msgs})
}

// === Experiences ===

func (h *Handler) ListExperiences(c *gin.Context) {
	exps, err := h.store.ListExperiences()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": exps})
}

func (h *Handler) CreateExperience(c *gin.Context) {
	var body struct {
		Type    string `json:"type"`
		Title   string `json:"title"`
		Content string `json:"content"`
		Tags    string `json:"tags"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	id, err := h.store.CreateExperience(body.Type, body.Title, body.Content, body.Tags)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": gin.H{"id": id}})
}

func (h *Handler) UpdateExperience(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var body struct {
		Title   string `json:"title"`
		Content string `json:"content"`
		Tags    string `json:"tags"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.store.UpdateExperience(id, body.Title, body.Content, body.Tags); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "ok"})
}

func (h *Handler) DeleteExperience(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	if err := h.store.DeleteExperience(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "ok"})
}

func (h *Handler) SearchExperiences(c *gin.Context) {
	keyword := c.Query("keyword")
	if keyword == "" {
		h.ListExperiences(c)
		return
	}
	exps, err := h.store.SearchExperiences(keyword)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": exps})
}

// === Opinions ===

func (h *Handler) ListOpinions(c *gin.Context) {
	author := c.Query("author")
	ops, err := h.store.ListOpinions(author)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": ops})
}

func (h *Handler) CreateOpinion(c *gin.Context) {
	var body struct {
		Author  string `json:"author"`
		Content string `json:"content"`
		Tags    string `json:"tags"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	id, err := h.store.CreateOpinion(body.Author, body.Content, body.Tags)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": gin.H{"id": id}})
}

func (h *Handler) DeleteOpinion(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	if err := h.store.DeleteOpinion(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "ok"})
}

func (h *Handler) ListAuthors(c *gin.Context) {
	authors, err := h.store.ListAuthors()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": authors})
}

func (h *Handler) SearchOpinions(c *gin.Context) {
	keyword := c.Query("keyword")
	if keyword == "" {
		h.ListOpinions(c)
		return
	}
	ops, err := h.store.SearchOpinions(keyword)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": ops})
}

func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	}
}
