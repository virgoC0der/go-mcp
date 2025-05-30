package http

import (
	"context"
	"net/http"
	"time"

	"github.com/virgoC0der/go-mcp/internal/types"
)

// HTTPServer implements the Server interface using HTTP
type HTTPServer struct {
	service      types.MCPService
	server       *http.Server
	capabilities *types.ServerCapabilities
}

// NewHTTPServer creates a new HTTP server instance
func NewHTTPServer(service types.MCPService, options *types.ServerOptions) *HTTPServer {
	s := &HTTPServer{
		service: service,
		server: &http.Server{
			Addr:              options.Address,
			ReadHeaderTimeout: 10 * time.Second,
		},
		capabilities: options.Capabilities,
	}
	s.server.Handler = newHTTPHandler(service, options.Capabilities)
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
func (s *HTTPServer) ListPrompts(ctx context.Context, cursor string) (*types.PromptListResult, error) {
	return s.service.ListPrompts(ctx, cursor)
}

// GetPrompt implements the Server interface
func (s *HTTPServer) GetPrompt(ctx context.Context, name string, args map[string]any) (*types.PromptResult, error) {
	return s.service.GetPrompt(ctx, name, args)
}

// ListTools implements the Server interface
func (s *HTTPServer) ListTools(ctx context.Context, cursor string) (*types.ToolListResult, error) {
	return s.service.ListTools(ctx, cursor)
}

// CallTool implements the Server interface
func (s *HTTPServer) CallTool(ctx context.Context, name string, args map[string]any) (*types.CallToolResult, error) {
	return s.service.CallTool(ctx, name, args)
}

// ListResources implements the Server interface
func (s *HTTPServer) ListResources(ctx context.Context, cursor string) (*types.ResourceListResult, error) {
	return s.service.ListResources(ctx, cursor)
}

// ReadResource implements the Server interface
func (s *HTTPServer) ReadResource(ctx context.Context, uri string) (*types.ResourceContent, error) {
	return s.service.ReadResource(ctx, uri)
}

// ListResourceTemplates implements the Server interface
func (s *HTTPServer) ListResourceTemplates(ctx context.Context) ([]types.ResourceTemplate, error) {
	return s.service.ListResourceTemplates(ctx)
}

// SubscribeToResource implements the Server interface
func (s *HTTPServer) SubscribeToResource(ctx context.Context, uri string) error {
	return s.service.SubscribeToResource(ctx, uri)
}

// Start starts the HTTP server
func (s *HTTPServer) Start() error {
	return s.server.ListenAndServe()
}

// Shutdown gracefully shuts down the HTTP server
func (s *HTTPServer) Shutdown(ctx context.Context) error {
	return s.server.Shutdown(ctx)
}
