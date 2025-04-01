package server

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/virgoC0der/go-mcp/internal/types"
)

// Server defines the interface for server implementations
type Server interface {
	Initialize(ctx context.Context, options any) error
	ListPrompts(ctx context.Context) ([]types.Prompt, error)
	GetPrompt(ctx context.Context, name string, args map[string]any) (*types.GetPromptResult, error)
	ListTools(ctx context.Context) ([]types.Tool, error)
	CallTool(ctx context.Context, name string, args map[string]any) (*types.CallToolResult, error)
	ListResources(ctx context.Context) ([]types.Resource, error)
	ReadResource(ctx context.Context, name string) ([]byte, string, error)
}

// HTTPHandler handles HTTP requests for the MCP server
type HTTPHandler struct {
	service types.MCPService
	engine  *gin.Engine
}

// newHTTPHandler creates a new HTTP handler
func newHTTPHandler(service types.MCPService) *HTTPHandler {
	h := &HTTPHandler{
		service: service,
		engine:  gin.Default(),
	}
	h.setupRoutes()
	return h
}

func (h *HTTPHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.engine.ServeHTTP(w, r)
}

func (h *HTTPHandler) setupRoutes() {
	h.engine.GET("/prompts", h.handlePrompts)
	h.engine.POST("/prompts/:name", h.handlePrompt)
	h.engine.GET("/tools", h.handleTools)
	h.engine.POST("/tools/:name", h.handleTool)
	h.engine.GET("/resources", h.handleResources)
	h.engine.GET("/resources/:name", h.handleResource)
}

func (h *HTTPHandler) handlePrompts(c *gin.Context) {
	prompts, err := h.service.ListPrompts(c.Request.Context())
	if err != nil {
		if mcpErr, ok := err.(*types.Error); ok {
			c.JSON(200, h.createErrorResponse(mcpErr.Code, mcpErr.Message))
		} else {
			c.JSON(200, h.createErrorResponse("unknown_error", err.Error()))
		}
		return
	}
	c.JSON(200, h.createSuccessResponse(prompts))
}

func (h *HTTPHandler) handlePrompt(c *gin.Context) {
	var req struct {
		Arguments map[string]any `json:"arguments"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(200, h.createErrorResponse("invalid_request", err.Error()))
		return
	}

	result, err := h.service.GetPrompt(c.Request.Context(), c.Param("name"), req.Arguments)
	if err != nil {
		if mcpErr, ok := err.(*types.Error); ok {
			c.JSON(200, h.createErrorResponse(mcpErr.Code, mcpErr.Message))
		} else {
			c.JSON(200, h.createErrorResponse("unknown_error", err.Error()))
		}
		return
	}
	c.JSON(200, h.createSuccessResponse(result))
}

func (h *HTTPHandler) handleTools(c *gin.Context) {
	tools, err := h.service.ListTools(c.Request.Context())
	if err != nil {
		if mcpErr, ok := err.(*types.Error); ok {
			c.JSON(200, h.createErrorResponse(mcpErr.Code, mcpErr.Message))
		} else {
			c.JSON(200, h.createErrorResponse("unknown_error", err.Error()))
		}
		return
	}
	c.JSON(200, h.createSuccessResponse(tools))
}

func (h *HTTPHandler) handleTool(c *gin.Context) {
	var req struct {
		Arguments map[string]any `json:"arguments"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(200, h.createErrorResponse("invalid_request", err.Error()))
		return
	}

	result, err := h.service.CallTool(c.Request.Context(), c.Param("name"), req.Arguments)
	if err != nil {
		if mcpErr, ok := err.(*types.Error); ok {
			c.JSON(200, h.createErrorResponse(mcpErr.Code, mcpErr.Message))
		} else {
			c.JSON(200, h.createErrorResponse("unknown_error", err.Error()))
		}
		return
	}
	c.JSON(200, h.createSuccessResponse(result))
}

func (h *HTTPHandler) handleResources(c *gin.Context) {
	resources, err := h.service.ListResources(c.Request.Context())
	if err != nil {
		if mcpErr, ok := err.(*types.Error); ok {
			c.JSON(200, h.createErrorResponse(mcpErr.Code, mcpErr.Message))
		} else {
			c.JSON(200, h.createErrorResponse("unknown_error", err.Error()))
		}
		return
	}
	c.JSON(200, h.createSuccessResponse(resources))
}

func (h *HTTPHandler) handleResource(c *gin.Context) {
	content, mimeType, err := h.service.ReadResource(c.Request.Context(), c.Param("name"))
	if err != nil {
		if mcpErr, ok := err.(*types.Error); ok {
			c.JSON(200, h.createErrorResponse(mcpErr.Code, mcpErr.Message))
		} else {
			c.JSON(200, h.createErrorResponse("unknown_error", err.Error()))
		}
		return
	}

	c.Data(200, mimeType, content)
}

func (h *HTTPHandler) createSuccessResponse(result any) map[string]any {
	response := map[string]any{
		"success": true,
	}
	if result != nil {
		response["result"] = result
	}
	return response
}

func (h *HTTPHandler) createErrorResponse(code, message string) map[string]any {
	return map[string]any{
		"success": false,
		"error": map[string]any{
			"code":    code,
			"message": message,
		},
	}
}
