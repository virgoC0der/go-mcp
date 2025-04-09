package http

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/virgoC0der/go-mcp/internal/response"
	"github.com/virgoC0der/go-mcp/internal/types"
)

// Server defines the interface for server implementations
type Server interface {
	Initialize(ctx context.Context, options any) error
	ListPrompts(ctx context.Context, cursor string) (*types.PromptListResult, error)
	GetPrompt(ctx context.Context, name string, args map[string]any) (*types.PromptResult, error)
	ListTools(ctx context.Context, cursor string) (*types.ToolListResult, error)
	CallTool(ctx context.Context, name string, args map[string]any) (*types.CallToolResult, error)
	ListResources(ctx context.Context, cursor string) (*types.ResourceListResult, error)
	ReadResource(ctx context.Context, uri string) (*types.ResourceContent, error)
	ListResourceTemplates(ctx context.Context) ([]types.ResourceTemplate, error)
	SubscribeToResource(ctx context.Context, uri string) error
}

// HTTPHandler handles HTTP requests for the MCP server
type HTTPHandler struct {
	service      types.MCPService
	engine       *gin.Engine
	capabilities *types.ServerCapabilities
}

// Use common JSON-RPC request/response structures
type jsonRPCRequest = response.JSONRPCRequest
type jsonRPCResponse = response.JSONRPCResponse

// Helper function to create JSON-RPC response
func createJSONRPCResponse(id interface{}, result interface{}) jsonRPCResponse {
	return response.NewJSONRPCResponse(id, result)
}

// Helper function to create JSON-RPC error response
func createJSONRPCErrorResponse(id interface{}, code int, message string, data interface{}) jsonRPCResponse {
	return response.NewJSONRPCErrorResponse(id, code, message, data)
}

// newHTTPHandler creates a new HTTP handler
func newHTTPHandler(service types.MCPService, capabilities *types.ServerCapabilities) *HTTPHandler {
	h := &HTTPHandler{
		service:      service,
		engine:       gin.Default(),
		capabilities: capabilities,
	}
	h.setupRoutes()
	return h
}

func (h *HTTPHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.engine.ServeHTTP(w, r)
}

func (h *HTTPHandler) setupRoutes() {
	// JSON-RPC endpoint
	h.engine.POST("/jsonrpc", h.handleJSONRPC)

	// Legacy REST API routes for backwards compatibility
	h.engine.GET("/prompts", h.handlePrompts)
	h.engine.POST("/prompts/:name", h.handlePrompt)
	h.engine.GET("/tools", h.handleTools)
	h.engine.POST("/tools/:name", h.handleTool)
	h.engine.GET("/resources", h.handleResources)
	h.engine.GET("/resources/:uri", h.handleResource)
}

func (h *HTTPHandler) handleJSONRPC(c *gin.Context) {
	var req jsonRPCRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, createJSONRPCErrorResponse(nil, -32700, "Parse error", nil))
		return
	}

	// Ensure JSONRPC version is 2.0
	if req.JSONRPC != "2.0" {
		c.JSON(http.StatusBadRequest, createJSONRPCErrorResponse(req.ID, -32600, "Invalid request", nil))
		return
	}

	// Handle different methods
	switch req.Method {
	case "prompts/list":
		h.handlePromptsList(c, req)
	case "prompts/get":
		h.handlePromptsGet(c, req)
	case "tools/list":
		h.handleToolsList(c, req)
	case "tools/call":
		h.handleToolsCall(c, req)
	case "resources/list":
		h.handleResourcesList(c, req)
	case "resources/read":
		h.handleResourcesRead(c, req)
	case "resources/templates/list":
		h.handleResourceTemplatesList(c, req)
	case "resources/subscribe":
		h.handleResourcesSubscribe(c, req)
	default:
		c.JSON(http.StatusOK, createJSONRPCErrorResponse(req.ID, -32601, "Method not found", nil))
	}
}

func (h *HTTPHandler) handlePromptsList(c *gin.Context, req jsonRPCRequest) {
	// Extract cursor from params
	var params struct {
		Cursor string `json:"cursor"`
	}

	// Parse params (they could be a map or a struct)
	if req.Params != nil {
		paramBytes, err := json.Marshal(req.Params)
		if err != nil {
			h.sendJSONRPCError(c, req.ID, -32602, "Invalid params", nil)
			return
		}

		if err := json.Unmarshal(paramBytes, &params); err != nil {
			h.sendJSONRPCError(c, req.ID, -32602, "Invalid params", nil)
			return
		}
	}

	result, err := h.service.ListPrompts(c.Request.Context(), params.Cursor)
	if err != nil {
		h.handleJSONRPCError(c, req.ID, err)
		return
	}

	c.JSON(http.StatusOK, createJSONRPCResponse(req.ID, result))
}

func (h *HTTPHandler) handlePromptsGet(c *gin.Context, req jsonRPCRequest) {
	var params struct {
		Name      string         `json:"name"`
		Arguments map[string]any `json:"arguments"`
	}

	paramBytes, err := json.Marshal(req.Params)
	if err != nil {
		h.sendJSONRPCError(c, req.ID, -32602, "Invalid params", nil)
		return
	}

	if err := json.Unmarshal(paramBytes, &params); err != nil {
		h.sendJSONRPCError(c, req.ID, -32602, "Invalid params", nil)
		return
	}

	if params.Name == "" {
		h.sendJSONRPCError(c, req.ID, -32602, "Missing required parameter: name", nil)
		return
	}

	result, err := h.service.GetPrompt(c.Request.Context(), params.Name, params.Arguments)
	if err != nil {
		h.handleJSONRPCError(c, req.ID, err)
		return
	}

	c.JSON(http.StatusOK, createJSONRPCResponse(req.ID, result))
}

func (h *HTTPHandler) handleToolsList(c *gin.Context, req jsonRPCRequest) {
	var params struct {
		Cursor string `json:"cursor"`
	}

	if req.Params != nil {
		paramBytes, err := json.Marshal(req.Params)
		if err != nil {
			h.sendJSONRPCError(c, req.ID, -32602, "Invalid params", nil)
			return
		}

		if err := json.Unmarshal(paramBytes, &params); err != nil {
			h.sendJSONRPCError(c, req.ID, -32602, "Invalid params", nil)
			return
		}
	}

	result, err := h.service.ListTools(c.Request.Context(), params.Cursor)
	if err != nil {
		h.handleJSONRPCError(c, req.ID, err)
		return
	}

	c.JSON(http.StatusOK, createJSONRPCResponse(req.ID, result))
}

func (h *HTTPHandler) handleToolsCall(c *gin.Context, req jsonRPCRequest) {
	var params struct {
		Name      string         `json:"name"`
		Arguments map[string]any `json:"arguments"`
	}

	paramBytes, err := json.Marshal(req.Params)
	if err != nil {
		h.sendJSONRPCError(c, req.ID, -32602, "Invalid params", nil)
		return
	}

	if err := json.Unmarshal(paramBytes, &params); err != nil {
		h.sendJSONRPCError(c, req.ID, -32602, "Invalid params", nil)
		return
	}

	if params.Name == "" {
		h.sendJSONRPCError(c, req.ID, -32602, "Missing required parameter: name", nil)
		return
	}

	result, err := h.service.CallTool(c.Request.Context(), params.Name, params.Arguments)
	if err != nil {
		h.handleJSONRPCError(c, req.ID, err)
		return
	}

	// 按照MCP规范格式化工具调用响应
	c.JSON(http.StatusOK, createJSONRPCResponse(req.ID, result))
}

func (h *HTTPHandler) handleResourcesList(c *gin.Context, req jsonRPCRequest) {
	var params struct {
		Cursor string `json:"cursor"`
	}

	if req.Params != nil {
		paramBytes, err := json.Marshal(req.Params)
		if err != nil {
			h.sendJSONRPCError(c, req.ID, -32602, "Invalid params", nil)
			return
		}

		if err := json.Unmarshal(paramBytes, &params); err != nil {
			h.sendJSONRPCError(c, req.ID, -32602, "Invalid params", nil)
			return
		}
	}

	result, err := h.service.ListResources(c.Request.Context(), params.Cursor)
	if err != nil {
		h.handleJSONRPCError(c, req.ID, err)
		return
	}

	c.JSON(http.StatusOK, createJSONRPCResponse(req.ID, result))
}

func (h *HTTPHandler) handleResourcesRead(c *gin.Context, req jsonRPCRequest) {
	var params struct {
		URI string `json:"uri"`
	}

	paramBytes, err := json.Marshal(req.Params)
	if err != nil {
		h.sendJSONRPCError(c, req.ID, -32602, "Invalid params", nil)
		return
	}

	if err := json.Unmarshal(paramBytes, &params); err != nil {
		h.sendJSONRPCError(c, req.ID, -32602, "Invalid params", nil)
		return
	}

	if params.URI == "" {
		h.sendJSONRPCError(c, req.ID, -32602, "Missing required parameter: uri", nil)
		return
	}

	result, err := h.service.ReadResource(c.Request.Context(), params.URI)
	if err != nil {
		h.handleJSONRPCError(c, req.ID, err)
		return
	}

	c.JSON(http.StatusOK, createJSONRPCResponse(req.ID, result))
}

func (h *HTTPHandler) handleResourceTemplatesList(c *gin.Context, req jsonRPCRequest) {
	result, err := h.service.ListResourceTemplates(c.Request.Context())
	if err != nil {
		h.handleJSONRPCError(c, req.ID, err)
		return
	}

	c.JSON(http.StatusOK, createJSONRPCResponse(req.ID, map[string]interface{}{
		"resourceTemplates": result,
	}))
}

func (h *HTTPHandler) handleResourcesSubscribe(c *gin.Context, req jsonRPCRequest) {
	var params struct {
		URI string `json:"uri"`
	}

	paramBytes, err := json.Marshal(req.Params)
	if err != nil {
		h.sendJSONRPCError(c, req.ID, -32602, "Invalid params", nil)
		return
	}

	if err := json.Unmarshal(paramBytes, &params); err != nil {
		h.sendJSONRPCError(c, req.ID, -32602, "Invalid params", nil)
		return
	}

	if params.URI == "" {
		h.sendJSONRPCError(c, req.ID, -32602, "Missing required parameter: uri", nil)
		return
	}

	err = h.service.SubscribeToResource(c.Request.Context(), params.URI)
	if err != nil {
		h.handleJSONRPCError(c, req.ID, err)
		return
	}

	c.JSON(http.StatusOK, createJSONRPCResponse(req.ID, true))
}

// Legacy REST API handlers for backwards compatibility

func (h *HTTPHandler) handlePrompts(c *gin.Context) {
	cursor := c.Query("cursor")
	prompts, err := h.service.ListPrompts(c.Request.Context(), cursor)
	if err != nil {
		h.handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, types.NewSuccessResponse(prompts))
}

func (h *HTTPHandler) handlePrompt(c *gin.Context) {
	var req struct {
		Arguments map[string]any `json:"arguments"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, types.NewErrorResponse("invalid_request", err.Error()))
		return
	}

	result, err := h.service.GetPrompt(c.Request.Context(), c.Param("name"), req.Arguments)
	if err != nil {
		h.handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, types.NewSuccessResponse(result))
}

func (h *HTTPHandler) handleTools(c *gin.Context) {
	cursor := c.Query("cursor")
	tools, err := h.service.ListTools(c.Request.Context(), cursor)
	if err != nil {
		h.handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, types.NewSuccessResponse(tools))
}

func (h *HTTPHandler) handleTool(c *gin.Context) {
	var req struct {
		Arguments map[string]any `json:"arguments"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, types.NewErrorResponse("invalid_request", err.Error()))
		return
	}

	result, err := h.service.CallTool(c.Request.Context(), c.Param("name"), req.Arguments)
	if err != nil {
		h.handleError(c, err)
		return
	}

	// 返回工具调用结果
	c.JSON(http.StatusOK, types.NewSuccessResponse(result))
}

func (h *HTTPHandler) handleResources(c *gin.Context) {
	cursor := c.Query("cursor")
	resources, err := h.service.ListResources(c.Request.Context(), cursor)
	if err != nil {
		h.handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, types.NewSuccessResponse(resources))
}

func (h *HTTPHandler) handleResource(c *gin.Context) {
	uri := c.Param("uri")
	resource, err := h.service.ReadResource(c.Request.Context(), uri)
	if err != nil {
		h.handleError(c, err)
		return
	}

	if resource.MimeType == "application/json" {
		c.JSON(http.StatusOK, types.NewSuccessResponse(resource.Text))
	} else if resource.Text != "" {
		c.Data(http.StatusOK, resource.MimeType, []byte(resource.Text))
	} else {
		// 解码Base64的二进制数据
		c.Data(http.StatusOK, resource.MimeType, []byte(resource.Blob))
	}
}

// Error handling helpers

func (h *HTTPHandler) handleError(c *gin.Context, err error) {
	if mcpErr, ok := err.(*types.Error); ok {
		c.JSON(http.StatusBadRequest, types.NewMCPErrorResponse(mcpErr))
	} else {
		c.JSON(http.StatusInternalServerError, types.NewErrorResponse("internal_error", err.Error()))
	}
}

func (h *HTTPHandler) handleJSONRPCError(c *gin.Context, id interface{}, err error) {
	// Use common error handling function
	c.JSON(http.StatusOK, response.HandleJSONRPCError(id, err))
}

func (h *HTTPHandler) sendJSONRPCError(c *gin.Context, id interface{}, code int, message string, data interface{}) {
	c.JSON(http.StatusOK, response.NewJSONRPCErrorResponse(id, code, message, data))
}
