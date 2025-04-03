package http

import (
	"context"
	"net/http"
	"time"

	"github.com/virgoC0der/go-mcp/internal/types"
)

// HTTPServer implements the Server interface using HTTP
type HTTPServer struct {
	service types.MCPService
	server  *http.Server
}

// NewHTTPServer creates a new HTTP server instance
func NewHTTPServer(service types.MCPService, address string) *HTTPServer {
	s := &HTTPServer{
		service: service,
		server: &http.Server{
			Addr:              address,
			ReadHeaderTimeout: 10 * time.Second,
		},
	}
	s.server.Handler = newHTTPHandler(service)
	return s
}

// Initialize implements the Server interface
func (s *HTTPServer) Initialize(ctx context.Context, options any) error {
	if initializer, ok := s.service.(types.Server); ok {
		return initializer.Initialize(ctx, options)
	}
	return nil
}

// ListPrompts implements the Server interface
func (s *HTTPServer) ListPrompts(ctx context.Context) ([]types.Prompt, error) {
	return s.service.ListPrompts(ctx)
}

// GetPrompt implements the Server interface
func (s *HTTPServer) GetPrompt(ctx context.Context, name string, args map[string]any) (*types.GetPromptResult, error) {
	return s.service.GetPrompt(ctx, name, args)
}

// ListTools implements the Server interface
func (s *HTTPServer) ListTools(ctx context.Context) ([]types.Tool, error) {
	return s.service.ListTools(ctx)
}

// CallTool implements the Server interface
func (s *HTTPServer) CallTool(ctx context.Context, name string, args map[string]any) (*types.CallToolResult, error) {
	return s.service.CallTool(ctx, name, args)
}

// ListResources implements the Server interface
func (s *HTTPServer) ListResources(ctx context.Context) ([]types.Resource, error) {
	return s.service.ListResources(ctx)
}

// ReadResource implements the Server interface
func (s *HTTPServer) ReadResource(ctx context.Context, name string) ([]byte, string, error) {
	return s.service.ReadResource(ctx, name)
}

// Start starts the HTTP server
func (s *HTTPServer) Start() error {
	return s.server.ListenAndServe()
}

// Shutdown gracefully shuts down the HTTP server
func (s *HTTPServer) Shutdown(ctx context.Context) error {
	return s.server.Shutdown(ctx)
}
