package server

import (
	"context"

	"github.com/virgoC0der/go-mcp/internal/types"
)

// MockServer implements Server interface for testing
type MockServer struct {
	initializeFunc    func(ctx context.Context, options any) error
	startFunc         func() error
	shutdownFunc      func(ctx context.Context) error
	listPromptsFunc   func(ctx context.Context) ([]types.Prompt, error)
	getPromptFunc     func(ctx context.Context, name string, args map[string]any) (*types.GetPromptResult, error)
	listToolsFunc     func(ctx context.Context) ([]types.Tool, error)
	callToolFunc      func(ctx context.Context, name string, args map[string]any) (*types.CallToolResult, error)
	listResourcesFunc func(ctx context.Context) ([]types.Resource, error)
	readResourceFunc  func(ctx context.Context, name string) ([]byte, string, error)
}

// Initialize implements the Server interface for testing initialization
func (m *MockServer) Initialize(ctx context.Context, options any) error {
	if m.initializeFunc != nil {
		return m.initializeFunc(ctx, options)
	}
	return nil
}

// Start implements the Server interface for testing server start
func (m *MockServer) Start() error {
	if m.startFunc != nil {
		return m.startFunc()
	}
	return nil
}

// Shutdown implements the Server interface for testing server shutdown
func (m *MockServer) Shutdown(ctx context.Context) error {
	if m.shutdownFunc != nil {
		return m.shutdownFunc(ctx)
	}
	return nil
}

// ListPrompts implements the Server interface for testing prompt listing
func (m *MockServer) ListPrompts(ctx context.Context) ([]types.Prompt, error) {
	if m.listPromptsFunc != nil {
		return m.listPromptsFunc(ctx)
	}
	return nil, nil
}

// GetPrompt implements the Server interface for testing prompt retrieval
func (m *MockServer) GetPrompt(ctx context.Context, name string, args map[string]any) (*types.GetPromptResult, error) {
	if m.getPromptFunc != nil {
		return m.getPromptFunc(ctx, name, args)
	}
	return nil, nil
}

// ListTools implements the Server interface for testing tool listing
func (m *MockServer) ListTools(ctx context.Context) ([]types.Tool, error) {
	if m.listToolsFunc != nil {
		return m.listToolsFunc(ctx)
	}
	return nil, nil
}

// CallTool implements the Server interface for testing tool invocation
func (m *MockServer) CallTool(ctx context.Context, name string, args map[string]any) (*types.CallToolResult, error) {
	if m.callToolFunc != nil {
		return m.callToolFunc(ctx, name, args)
	}
	return nil, nil
}

// ListResources implements the Server interface for testing resource listing
func (m *MockServer) ListResources(ctx context.Context) ([]types.Resource, error) {
	if m.listResourcesFunc != nil {
		return m.listResourcesFunc(ctx)
	}
	return nil, nil
}

// ReadResource implements the Server interface for testing resource reading
func (m *MockServer) ReadResource(ctx context.Context, name string) ([]byte, string, error) {
	if m.readResourceFunc != nil {
		return m.readResourceFunc(ctx, name)
	}
	return nil, "", nil
}
