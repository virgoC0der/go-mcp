package server

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

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
}

// newHTTPHandler creates a new HTTP handler
func newHTTPHandler(service types.MCPService) *HTTPHandler {
	return &HTTPHandler{
		service: service,
	}
}

func (h *HTTPHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/prompts":
		h.handlePrompts(w, r)
	case "/prompt":
		h.handlePrompt(w, r)
	case "/tools":
		h.handleTools(w, r)
	case "/tool":
		h.handleTool(w, r)
	case "/resources":
		h.handleResources(w, r)
	case "/resource":
		h.handleResource(w, r)
	default:
		http.NotFound(w, r)
	}
}

func (h *HTTPHandler) handlePrompts(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	prompts, err := h.service.ListPrompts(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := json.NewEncoder(w).Encode(prompts); err != nil {
		log.Printf("Failed to encode response: %v", err)
	}
}

func (h *HTTPHandler) handlePrompt(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Name      string         `json:"name"`
		Arguments map[string]any `json:"arguments"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	result, err := h.service.GetPrompt(r.Context(), req.Name, req.Arguments)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := json.NewEncoder(w).Encode(result); err != nil {
		log.Printf("Failed to encode response: %v", err)
	}
}

func (h *HTTPHandler) handleTools(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	tools, err := h.service.ListTools(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := json.NewEncoder(w).Encode(tools); err != nil {
		log.Printf("Failed to encode response: %v", err)
	}
}

func (h *HTTPHandler) handleTool(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Name      string         `json:"name"`
		Arguments map[string]any `json:"arguments"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	result, err := h.service.CallTool(r.Context(), req.Name, req.Arguments)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := json.NewEncoder(w).Encode(result); err != nil {
		log.Printf("Failed to encode response: %v", err)
	}
}

func (h *HTTPHandler) handleResources(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	resources, err := h.service.ListResources(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := json.NewEncoder(w).Encode(resources); err != nil {
		log.Printf("Failed to encode response: %v", err)
	}
}

func (h *HTTPHandler) handleResource(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	name := r.URL.Query().Get("name")
	if name == "" {
		http.Error(w, "Missing resource name", http.StatusBadRequest)
		return
	}

	content, mimeType, err := h.service.ReadResource(r.Context(), name)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", mimeType)
	if _, err := w.Write(content); err != nil {
		log.Printf("Failed to write response: %v", err)
	}
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
