package transport

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/virgoC0der/go-mcp/server"
)

// HTTPServer wraps the HTTP transport layer for an MCP server
type HTTPServer struct {
	server  server.Server
	addr    string
	handler *HTTPHandler
	srv     *http.Server
}

// NewHTTPServer creates a new HTTP server instance
func NewHTTPServer(mcpServer server.Server, addr string) *HTTPServer {
	return &HTTPServer{
		server:  mcpServer,
		addr:    addr,
		handler: NewHTTPHandler(mcpServer),
	}
}

// Start starts the HTTP server
func (s *HTTPServer) Start() error {
	s.srv = &http.Server{
		Addr:              s.addr,
		Handler:           s.handler,
		ReadHeaderTimeout: 10 * time.Second,
	}
	return s.srv.ListenAndServe()
}

// Shutdown gracefully shuts down the HTTP server
func (s *HTTPServer) Shutdown(ctx context.Context) error {
	if s.srv != nil {
		return s.srv.Shutdown(ctx)
	}
	return nil
}

// Legacy HTTP handlers kept for backwards compatibility

// Deprecated: Legacy handler kept for backwards compatibility
func (s *HTTPServer) handlePrompts(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	prompts, err := s.server.ListPrompts(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := json.NewEncoder(w).Encode(prompts); err != nil {
		log.Printf("Failed to encode response: %v", err)
	}
}

// Deprecated: Legacy handler kept for backwards compatibility
func (s *HTTPServer) handlePrompt(w http.ResponseWriter, r *http.Request) {
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

	result, err := s.server.GetPrompt(r.Context(), req.Name, req.Arguments)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := json.NewEncoder(w).Encode(result); err != nil {
		log.Printf("Failed to encode response: %v", err)
	}
}

// Deprecated: Legacy handler kept for backwards compatibility
func (s *HTTPServer) handleTools(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	tools, err := s.server.ListTools(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := json.NewEncoder(w).Encode(tools); err != nil {
		log.Printf("Failed to encode response: %v", err)
	}
}

// Deprecated: Legacy handler kept for backwards compatibility
func (s *HTTPServer) handleTool(w http.ResponseWriter, r *http.Request) {
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

	result, err := s.server.CallTool(r.Context(), req.Name, req.Arguments)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := json.NewEncoder(w).Encode(result); err != nil {
		log.Printf("Failed to encode response: %v", err)
	}
}

// Deprecated: Legacy handler kept for backwards compatibility
func (s *HTTPServer) handleResources(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	resources, err := s.server.ListResources(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := json.NewEncoder(w).Encode(resources); err != nil {
		log.Printf("Failed to encode response: %v", err)
	}
}

// Deprecated: Legacy handler kept for backwards compatibility
func (s *HTTPServer) handleResource(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	name := r.URL.Query().Get("name")
	if name == "" {
		http.Error(w, "Missing resource name", http.StatusBadRequest)
		return
	}

	content, mimeType, err := s.server.ReadResource(r.Context(), name)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", mimeType)
	if _, err := w.Write(content); err != nil {
		log.Printf("Failed to write response: %v", err)
	}
}
