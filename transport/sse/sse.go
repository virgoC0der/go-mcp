package sse

//
//import (
//	"context"
//	"encoding/json"
//	"net/http"
//	"strings"
//	"sync"
//	"time"
//
//	"github.com/gin-gonic/gin"
//	"github.com/virgoC0der/go-mcp/internal/types"
//)
//
//// SSEServer wraps the Server-Sent Events transport layer for an MCP server
//type SSEServer struct {
//	server Server
//	addr   string
//	srv    *http.Server
//	engine *gin.Engine
//	mu     sync.RWMutex
//
//	// clients management
//	clients    map[chan []byte]bool
//	clientsMux sync.RWMutex
//}
//
//// NewSSEServer creates a new SSE server instance
//func NewSSEServer(mcpServer Server, addr string) *SSEServer {
//	engine := gin.Default()
//	s := &SSEServer{
//		server:  mcpServer,
//		addr:    addr,
//		engine:  engine,
//		clients: make(map[chan []byte]bool),
//	}
//
//	// Setup routes
//	s.setupRoutes()
//	return s
//}
//
//// setupRoutes configures the Gin routes
//func (s *SSEServer) setupRoutes() {
//	// SSE endpoint
//	s.engine.GET("/events", s.handleSSE)
//
//	// API endpoints
//	api := s.engine.Group("/api")
//	{
//		api.GET("/prompts", s.handlePrompts)
//		api.GET("/prompts/:name", s.handleGetPrompt)
//		api.GET("/tools", s.handleTools)
//		api.POST("/tools/:name", s.handleCallTool)
//		api.GET("/resources", s.handleResources)
//		api.GET("/resources/:name", s.handleReadResource)
//	}
//}
//
//// Start starts the SSE server
//func (s *SSEServer) Start() error {
//	s.mu.Lock()
//	defer s.mu.Unlock()
//
//	s.srv = &http.Server{
//		Addr:              s.addr,
//		Handler:           s.engine,
//		ReadHeaderTimeout: 10 * time.Second,
//	}
//	return s.srv.ListenAndServe()
//}
//
//// Shutdown gracefully shuts down the SSE server
//func (s *SSEServer) Shutdown(ctx context.Context) error {
//	// 首先关闭所有客户端连接
//	s.clientsMux.Lock()
//	for clientChan := range s.clients {
//		close(clientChan)
//		delete(s.clients, clientChan)
//	}
//	s.clientsMux.Unlock()
//
//	// 然后关闭 HTTP 服务器
//	s.mu.Lock()
//	if s.srv == nil {
//		s.mu.Unlock()
//		return nil
//	}
//	srv := s.srv
//	s.srv = nil // 防止重复关闭
//	s.mu.Unlock()
//
//	// 最后等待所有连接关闭
//	return srv.Shutdown(ctx)
//}
//
//// handleSSE handles the SSE connection
//func (s *SSEServer) handleSSE(c *gin.Context) {
//	// Set headers for SSE
//	c.Header("Content-Type", "text/event-stream")
//	c.Header("Cache-Control", "no-cache")
//	c.Header("Connection", "keep-alive")
//	c.Header("Access-Control-Allow-Origin", "*")
//	c.Header("X-Accel-Buffering", "no") // Disable buffering for nginx
//
//	// Create a channel for this client
//	messageChan := make(chan []byte, 10) // Buffer size of 10
//
//	// Register client
//	s.clientsMux.Lock()
//	s.clients[messageChan] = true
//	s.clientsMux.Unlock()
//
//	// Clean up on client disconnect
//	defer func() {
//		s.clientsMux.Lock()
//		if _, exists := s.clients[messageChan]; exists {
//			delete(s.clients, messageChan)
//			close(messageChan)
//		}
//		s.clientsMux.Unlock()
//	}()
//
//	// 获取请求的上下文
//	ctx := c.Request.Context()
//	flusher, ok := c.Writer.(http.Flusher)
//	if !ok {
//		c.AbortWithStatus(http.StatusInternalServerError)
//		return
//	}
//
//	// 立即发送一个初始化消息
//	if _, err := c.Writer.Write([]byte(": connected\n\n")); err != nil {
//		return
//	}
//	flusher.Flush()
//
//	// 创建一个定时器用于心跳
//	heartbeat := time.NewTicker(30 * time.Second)
//	defer heartbeat.Stop()
//
//	// 使用无限循环来处理消息
//	for {
//		select {
//		case <-ctx.Done():
//			return
//		case msg, ok := <-messageChan:
//			if !ok {
//				return
//			}
//			// 写入消息
//			if _, err := c.Writer.Write([]byte("data: ")); err != nil {
//				return
//			}
//			if _, err := c.Writer.Write(msg); err != nil {
//				return
//			}
//			if _, err := c.Writer.Write([]byte("\n\n")); err != nil {
//				return
//			}
//			flusher.Flush()
//		case <-heartbeat.C:
//			// 发送心跳
//			if _, err := c.Writer.Write([]byte(": keepalive\n\n")); err != nil {
//				return
//			}
//			flusher.Flush()
//		}
//	}
//}
//
//// Broadcast sends a message to all connected clients
//func (s *SSEServer) Broadcast(eventType string, data interface{}) {
//	response := types.NewSuccessResponse(map[string]interface{}{
//		"type": eventType,
//		"data": data,
//	})
//
//	jsonData, err := json.Marshal(response)
//	if err != nil {
//		return
//	}
//
//	s.clientsMux.RLock()
//	for clientChan := range s.clients {
//		select {
//		case clientChan <- jsonData:
//		default:
//			// Skip if client's buffer is full
//		}
//	}
//	s.clientsMux.RUnlock()
//}
//
//// BroadcastError sends an error message to all connected clients
//func (s *SSEServer) BroadcastError(err error) {
//	var response types.Response
//	if mcpErr, ok := err.(*types.Error); ok {
//		response = types.NewMCPErrorResponse(mcpErr)
//	} else {
//		response = types.NewErrorResponse("internal_error", err.Error())
//	}
//
//	jsonData, err := json.Marshal(response)
//	if err != nil {
//		return
//	}
//
//	s.clientsMux.RLock()
//	for clientChan := range s.clients {
//		select {
//		case clientChan <- jsonData:
//		default:
//			// Skip if client's buffer is full
//		}
//	}
//	s.clientsMux.RUnlock()
//}
//
//// handlePrompts handles the prompts endpoint
//func (s *SSEServer) handlePrompts(c *gin.Context) {
//	prompts, err := s.server.ListPrompts(c.Request.Context())
//	if err != nil {
//		s.handleError(c, err)
//		return
//	}
//	s.Broadcast("prompts", prompts)
//	c.JSON(http.StatusOK, types.NewSuccessResponse(prompts))
//}
//
//// handleGetPrompt handles getting a specific prompt
//func (s *SSEServer) handleGetPrompt(c *gin.Context) {
//	var args map[string]interface{}
//	if err := c.ShouldBindJSON(&args); err != nil {
//		s.handleError(c, types.NewError("invalid_request", err.Error()))
//		return
//	}
//
//	result, err := s.server.GetPrompt(c.Request.Context(), c.Param("name"), args)
//	if err != nil {
//		s.handleError(c, err)
//		return
//	}
//	s.Broadcast("prompt", result)
//	c.JSON(http.StatusOK, types.NewSuccessResponse(result))
//}
//
//// handleTools handles the tools endpoint
//func (s *SSEServer) handleTools(c *gin.Context) {
//	tools, err := s.server.ListTools(c.Request.Context())
//	if err != nil {
//		s.handleError(c, err)
//		return
//	}
//	s.Broadcast("tools", tools)
//	c.JSON(http.StatusOK, types.NewSuccessResponse(tools))
//}
//
//// handleCallTool handles calling a specific tool
//func (s *SSEServer) handleCallTool(c *gin.Context) {
//	var args map[string]interface{}
//	if err := c.ShouldBindJSON(&args); err != nil {
//		s.handleError(c, types.NewError("invalid_request", err.Error()))
//		return
//	}
//
//	result, err := s.server.CallTool(c.Request.Context(), c.Param("name"), args)
//	if err != nil {
//		s.handleError(c, err)
//		return
//	}
//	s.Broadcast("tool_result", result)
//	c.JSON(http.StatusOK, types.NewSuccessResponse(result))
//}
//
//// handleResources handles the resources endpoint
//func (s *SSEServer) handleResources(c *gin.Context) {
//	resources, err := s.server.ListResources(c.Request.Context())
//	if err != nil {
//		s.handleError(c, err)
//		return
//	}
//	s.Broadcast("resources", resources)
//	c.JSON(http.StatusOK, types.NewSuccessResponse(resources))
//}
//
//// handleReadResource handles reading a specific resource
//func (s *SSEServer) handleReadResource(c *gin.Context) {
//	content, mimeType, err := s.server.ReadResource(c.Request.Context(), c.Param("name"))
//	if err != nil {
//		s.handleError(c, err)
//		return
//	}
//
//	resourceData := map[string]interface{}{
//		"name":     c.Param("name"),
//		"content":  string(content),
//		"mimeType": mimeType,
//	}
//	s.Broadcast("resource", resourceData)
//
//	// 对于文本内容，直接返回原始内容
//	if strings.HasPrefix(mimeType, "text/") {
//		c.Data(http.StatusOK, mimeType, content)
//		return
//	}
//
//	// 对于其他类型，返回 JSON 响应
//	c.JSON(http.StatusOK, types.NewSuccessResponse(resourceData))
//}
//
//// handleError handles error responses
//func (s *SSEServer) handleError(c *gin.Context, err error) {
//	var response types.Response
//	if mcpErr, ok := err.(*types.Error); ok {
//		response = types.NewMCPErrorResponse(mcpErr)
//		c.JSON(http.StatusBadRequest, response)
//	} else {
//		response = types.NewErrorResponse("internal_error", err.Error())
//		c.JSON(http.StatusInternalServerError, response)
//	}
//	s.BroadcastError(err)
//}
