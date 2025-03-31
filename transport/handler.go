package transport

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/virgoC0der/go-mcp/types"
)

// Server defines the interface for server implementations
type Server interface {
	Initialize(ctx context.Context, options interface{}) error
	ListPrompts(ctx context.Context) ([]types.Prompt, error)
	GetPrompt(ctx context.Context, name string, args map[string]interface{}) (*types.GetPromptResult, error)
	ListTools(ctx context.Context) ([]types.Tool, error)
	CallTool(ctx context.Context, name string, args map[string]interface{}) (*types.CallToolResult, error)
	ListResources(ctx context.Context) ([]types.Resource, error)
	ReadResource(ctx context.Context, name string) ([]byte, string, error)
}

// HTTPHandler handles HTTP requests for the MCP server
type HTTPHandler struct {
	server Server
}

// NewHTTPHandler creates a new HTTP handler
func NewHTTPHandler(server Server) *HTTPHandler {
	return &HTTPHandler{
		server: server,
	}
}

// ServeHTTP implements the http.Handler interface
func (h *HTTPHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Parse path to determine what action to take
	path := strings.TrimPrefix(r.URL.Path, "/")
	pathParts := strings.Split(path, "/")

	var response map[string]interface{}

	switch {
	case path == "initialize" && r.Method == http.MethodPost:
		response = h.handleInitialize(r)
	case path == "prompts" && r.Method == http.MethodGet:
		response = h.handleListPrompts(r)
	case len(pathParts) == 2 && pathParts[0] == "prompts" && r.Method == http.MethodPost:
		response = h.handleGetPrompt(r, pathParts[1])
	case path == "tools" && r.Method == http.MethodGet:
		response = h.handleListTools(r)
	case len(pathParts) == 2 && pathParts[0] == "tools" && r.Method == http.MethodPost:
		response = h.handleCallTool(r, pathParts[1])
	case path == "resources" && r.Method == http.MethodGet:
		response = h.handleListResources(r)
	case len(pathParts) == 2 && pathParts[0] == "resources" && r.Method == http.MethodGet:
		response = h.handleReadResource(r, pathParts[1])
	default:
		response = h.createErrorResponse("not_found", fmt.Sprintf("Path not found: %s", path))
	}

	// Write response
	json.NewEncoder(w).Encode(response)
}

func (h *HTTPHandler) handleInitialize(r *http.Request) map[string]interface{} {
	var options map[string]interface{}
	err := json.NewDecoder(r.Body).Decode(&options)
	if err != nil {
		return h.createErrorResponse("invalid_request", fmt.Sprintf("Failed to parse request body: %v", err))
	}

	err = h.server.Initialize(r.Context(), options)
	if err != nil {
		if mcpErr, ok := err.(*types.Error); ok {
			return h.createErrorResponse(mcpErr.Code, mcpErr.Message)
		}
		return h.createErrorResponse("unknown_error", err.Error())
	}

	return h.createSuccessResponse(nil)
}

func (h *HTTPHandler) handleListPrompts(r *http.Request) map[string]interface{} {
	prompts, err := h.server.ListPrompts(r.Context())
	if err != nil {
		if mcpErr, ok := err.(*types.Error); ok {
			return h.createErrorResponse(mcpErr.Code, mcpErr.Message)
		}
		return h.createErrorResponse("unknown_error", err.Error())
	}

	return h.createSuccessResponse(prompts)
}

func (h *HTTPHandler) handleGetPrompt(r *http.Request, promptName string) map[string]interface{} {
	var req struct {
		Args map[string]interface{} `json:"args"`
	}

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		return h.createErrorResponse("invalid_request", fmt.Sprintf("Failed to parse request body: %v", err))
	}

	result, err := h.server.GetPrompt(r.Context(), promptName, req.Args)
	if err != nil {
		if mcpErr, ok := err.(*types.Error); ok {
			return h.createErrorResponse(mcpErr.Code, mcpErr.Message)
		}
		return h.createErrorResponse("unknown_error", err.Error())
	}

	return h.createSuccessResponse(result)
}

func (h *HTTPHandler) handleListTools(r *http.Request) map[string]interface{} {
	tools, err := h.server.ListTools(r.Context())
	if err != nil {
		if mcpErr, ok := err.(*types.Error); ok {
			return h.createErrorResponse(mcpErr.Code, mcpErr.Message)
		}
		return h.createErrorResponse("unknown_error", err.Error())
	}

	return h.createSuccessResponse(tools)
}

func (h *HTTPHandler) handleCallTool(r *http.Request, toolName string) map[string]interface{} {
	var req struct {
		Args map[string]interface{} `json:"args"`
	}

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		return h.createErrorResponse("invalid_request", fmt.Sprintf("Failed to parse request body: %v", err))
	}

	result, err := h.server.CallTool(r.Context(), toolName, req.Args)
	if err != nil {
		if mcpErr, ok := err.(*types.Error); ok {
			return h.createErrorResponse(mcpErr.Code, mcpErr.Message)
		}
		return h.createErrorResponse("unknown_error", err.Error())
	}

	return h.createSuccessResponse(result)
}

func (h *HTTPHandler) handleListResources(r *http.Request) map[string]interface{} {
	resources, err := h.server.ListResources(r.Context())
	if err != nil {
		if mcpErr, ok := err.(*types.Error); ok {
			return h.createErrorResponse(mcpErr.Code, mcpErr.Message)
		}
		return h.createErrorResponse("unknown_error", err.Error())
	}

	return h.createSuccessResponse(resources)
}

func (h *HTTPHandler) handleReadResource(r *http.Request, resourceName string) map[string]interface{} {
	content, mimeType, err := h.server.ReadResource(r.Context(), resourceName)
	if err != nil {
		if mcpErr, ok := err.(*types.Error); ok {
			return h.createErrorResponse(mcpErr.Code, mcpErr.Message)
		}
		return h.createErrorResponse("unknown_error", err.Error())
	}

	// Base64 encode the content
	encodedContent := base64.StdEncoding.EncodeToString(content)

	return h.createSuccessResponse(map[string]interface{}{
		"content":  encodedContent,
		"mimeType": mimeType,
	})
}

func (h *HTTPHandler) createSuccessResponse(result interface{}) map[string]interface{} {
	response := map[string]interface{}{
		"success": true,
	}

	if result != nil {
		response["result"] = result
	}

	return response
}

func (h *HTTPHandler) createErrorResponse(code, message string) map[string]interface{} {
	return map[string]interface{}{
		"success": false,
		"error": map[string]interface{}{
			"code":    code,
			"message": message,
		},
	}
}
