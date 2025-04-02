package server

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/virgoC0der/go-mcp/internal/types"
)

// SSEServer wraps the Server-Sent Events transport layer for an MCP server
type SSEServer struct {
	server Server
	addr   string
	srv    *http.Server
	engine *gin.Engine
	mu     sync.RWMutex

	// clients management
	clients    map[chan []byte]bool
	clientsMux sync.RWMutex
}

// NewSSEServer creates a new SSE server instance
func NewSSEServer(mcpServer Server, addr string) *SSEServer {
	engine := gin.Default()
	s := &SSEServer{
		server:  mcpServer,
		addr:    addr,
		engine:  engine,
		clients: make(map[chan []byte]bool),
	}

	// Setup routes
	s.setupRoutes()
	return s
}

// setupRoutes configures the Gin routes
func (s *SSEServer) setupRoutes() {
	// SSE endpoint
	s.engine.GET("/events", s.handleSSE)

	// API endpoints
	api := s.engine.Group("/api")
	{
		api.GET("/prompts", s.handlePrompts)
		api.GET("/prompts/:name", s.handleGetPrompt)
		api.GET("/tools", s.handleTools)
		api.POST("/tools/:name", s.handleCallTool)
		api.GET("/resources", s.handleResources)
		api.GET("/resources/:name", s.handleReadResource)
	}
}

// Start starts the SSE server
func (s *SSEServer) Start() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.srv = &http.Server{
		Addr:              s.addr,
		Handler:           s.engine,
		ReadHeaderTimeout: 10 * time.Second,
	}
	return s.srv.ListenAndServe()
}

// Shutdown gracefully shuts down the SSE server
func (s *SSEServer) Shutdown(ctx context.Context) error {
	s.mu.Lock()
	if s.srv == nil {
		s.mu.Unlock()
		return nil
	}
	srv := s.srv
	s.mu.Unlock()

	// Close all client connections
	s.clientsMux.Lock()
	for clientChan := range s.clients {
		close(clientChan)
		delete(s.clients, clientChan)
	}
	s.clientsMux.Unlock()

	return srv.Shutdown(ctx)
}

// handleSSE handles the SSE connection
func (s *SSEServer) handleSSE(c *gin.Context) {
	// Set headers for SSE
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("Access-Control-Allow-Origin", "*")
	c.Header("X-Accel-Buffering", "no") // Disable buffering for nginx

	// Create a channel for this client
	messageChan := make(chan []byte, 10) // Buffer size of 10

	// Register client
	s.clientsMux.Lock()
	s.clients[messageChan] = true
	s.clientsMux.Unlock()

	// Clean up on client disconnect
	defer func() {
		s.clientsMux.Lock()
		delete(s.clients, messageChan)
		close(messageChan)
		s.clientsMux.Unlock()
	}()

	c.Stream(func(w io.Writer) bool {
		select {
		case <-c.Done():
			return false
		case msg, ok := <-messageChan:
			if !ok {
				return false
			}
			c.SSEvent("message", string(msg))
			return true
		case <-time.After(30 * time.Second):
			// Send keep-alive
			c.SSEvent("", "")
			return true
		}
	})
}

// Broadcast sends a message to all connected clients
func (s *SSEServer) Broadcast(eventType string, data interface{}) {
	response := types.NewSuccessResponse(map[string]interface{}{
		"type": eventType,
		"data": data,
	})

	jsonData, err := json.Marshal(response)
	if err != nil {
		return
	}

	s.clientsMux.RLock()
	for clientChan := range s.clients {
		select {
		case clientChan <- jsonData:
		default:
			// Skip if client's buffer is full
		}
	}
	s.clientsMux.RUnlock()
}

// BroadcastError sends an error message to all connected clients
func (s *SSEServer) BroadcastError(err error) {
	var response types.Response
	if mcpErr, ok := err.(*types.Error); ok {
		response = types.NewMCPErrorResponse(mcpErr)
	} else {
		response = types.NewErrorResponse("internal_error", err.Error())
	}

	jsonData, err := json.Marshal(response)
	if err != nil {
		return
	}

	s.clientsMux.RLock()
	for clientChan := range s.clients {
		select {
		case clientChan <- jsonData:
		default:
			// Skip if client's buffer is full
		}
	}
	s.clientsMux.RUnlock()
}

// handlePrompts handles the prompts endpoint
func (s *SSEServer) handlePrompts(c *gin.Context) {
	prompts, err := s.server.ListPrompts(c.Request.Context())
	if err != nil {
		s.handleError(c, err)
		return
	}
	s.Broadcast("prompts", prompts)
	c.JSON(http.StatusOK, types.NewSuccessResponse(prompts))
}

// handleGetPrompt handles getting a specific prompt
func (s *SSEServer) handleGetPrompt(c *gin.Context) {
	var args map[string]interface{}
	if err := c.ShouldBindJSON(&args); err != nil {
		s.handleError(c, types.NewError("invalid_request", err.Error()))
		return
	}

	result, err := s.server.GetPrompt(c.Request.Context(), c.Param("name"), args)
	if err != nil {
		s.handleError(c, err)
		return
	}
	s.Broadcast("prompt", result)
	c.JSON(http.StatusOK, types.NewSuccessResponse(result))
}

// handleTools handles the tools endpoint
func (s *SSEServer) handleTools(c *gin.Context) {
	tools, err := s.server.ListTools(c.Request.Context())
	if err != nil {
		s.handleError(c, err)
		return
	}
	s.Broadcast("tools", tools)
	c.JSON(http.StatusOK, types.NewSuccessResponse(tools))
}

// handleCallTool handles calling a specific tool
func (s *SSEServer) handleCallTool(c *gin.Context) {
	var args map[string]interface{}
	if err := c.ShouldBindJSON(&args); err != nil {
		s.handleError(c, types.NewError("invalid_request", err.Error()))
		return
	}

	result, err := s.server.CallTool(c.Request.Context(), c.Param("name"), args)
	if err != nil {
		s.handleError(c, err)
		return
	}
	s.Broadcast("tool_result", result)
	c.JSON(http.StatusOK, types.NewSuccessResponse(result))
}

// handleResources handles the resources endpoint
func (s *SSEServer) handleResources(c *gin.Context) {
	resources, err := s.server.ListResources(c.Request.Context())
	if err != nil {
		s.handleError(c, err)
		return
	}
	s.Broadcast("resources", resources)
	c.JSON(http.StatusOK, types.NewSuccessResponse(resources))
}

// handleReadResource handles reading a specific resource
func (s *SSEServer) handleReadResource(c *gin.Context) {
	content, mimeType, err := s.server.ReadResource(c.Request.Context(), c.Param("name"))
	if err != nil {
		s.handleError(c, err)
		return
	}

	resourceData := map[string]interface{}{
		"name":     c.Param("name"),
		"content":  string(content),
		"mimeType": mimeType,
	}
	s.Broadcast("resource", resourceData)

	// For binary data or specific mime types, use Data instead of JSON
	if strings.HasPrefix(mimeType, "text/") || mimeType == "application/json" {
		c.JSON(http.StatusOK, types.NewSuccessResponse(resourceData))
	} else {
		c.Data(http.StatusOK, mimeType, content)
	}
}

// handleError handles error responses
func (s *SSEServer) handleError(c *gin.Context, err error) {
	var response types.Response
	if mcpErr, ok := err.(*types.Error); ok {
		response = types.NewMCPErrorResponse(mcpErr)
		c.JSON(http.StatusBadRequest, response)
	} else {
		response = types.NewErrorResponse("internal_error", err.Error())
		c.JSON(http.StatusInternalServerError, response)
	}
	s.BroadcastError(err)
}
