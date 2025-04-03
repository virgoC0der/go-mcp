package mock

import (
	"context"

	"github.com/virgoC0der/go-mcp/internal/types"
)

// MockServer implements Server interface for testing
type MockServer struct {
	InitializeFunc    func(ctx context.Context, options any) error
	StartFunc         func() error
	ShutdownFunc      func(ctx context.Context) error
	ListPromptsFunc   func(ctx context.Context) ([]types.Prompt, error)
	GetPromptFunc     func(ctx context.Context, name string, args map[string]any) (*types.GetPromptResult, error)
	ListToolsFunc     func(ctx context.Context) ([]types.Tool, error)
	CallToolFunc      func(ctx context.Context, name string, args map[string]any) (*types.CallToolResult, error)
	ListResourcesFunc func(ctx context.Context) ([]types.Resource, error)
	ReadResourceFunc  func(ctx context.Context, name string) ([]byte, string, error)
}

// Initialize implements the Server interface for testing initialization
func (m *MockServer) Initialize(ctx context.Context, options any) error {
	if m.InitializeFunc != nil {
		return m.InitializeFunc(ctx, options)
	}
	return nil
}

// Start implements the Server interface for testing server start
func (m *MockServer) Start() error {
	if m.StartFunc != nil {
		return m.StartFunc()
	}
	return nil
}

// Shutdown implements the Server interface for testing server shutdown
func (m *MockServer) Shutdown(ctx context.Context) error {
	if m.ShutdownFunc != nil {
		return m.ShutdownFunc(ctx)
	}
	return nil
}

// ListPrompts implements the Server interface for testing prompt listing
func (m *MockServer) ListPrompts(ctx context.Context) ([]types.Prompt, error) {
	if m.ListPromptsFunc != nil {
		return m.ListPromptsFunc(ctx)
	}
	return nil, nil
}

// GetPrompt implements the Server interface for testing prompt retrieval
func (m *MockServer) GetPrompt(ctx context.Context, name string, args map[string]any) (*types.GetPromptResult, error) {
	if m.GetPromptFunc != nil {
		return m.GetPromptFunc(ctx, name, args)
	}
	return nil, nil
}

// ListTools implements the Server interface for testing tool listing
func (m *MockServer) ListTools(ctx context.Context) ([]types.Tool, error) {
	if m.ListToolsFunc != nil {
		return m.ListToolsFunc(ctx)
	}
	return nil, nil
}

// CallTool implements the Server interface for testing tool invocation
func (m *MockServer) CallTool(ctx context.Context, name string, args map[string]any) (*types.CallToolResult, error) {
	if m.CallToolFunc != nil {
		return m.CallToolFunc(ctx, name, args)
	}
	return nil, nil
}

// ListResources implements the Server interface for testing resource listing
func (m *MockServer) ListResources(ctx context.Context) ([]types.Resource, error) {
	if m.ListResourcesFunc != nil {
		return m.ListResourcesFunc(ctx)
	}
	return nil, nil
}

// ReadResource implements the Server interface for testing resource reading
func (m *MockServer) ReadResource(ctx context.Context, name string) ([]byte, string, error) {
	if m.ReadResourceFunc != nil {
		return m.ReadResourceFunc(ctx, name)
	}
	return nil, "", nil
}
