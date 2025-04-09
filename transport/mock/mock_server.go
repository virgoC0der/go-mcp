package mock

import (
	"context"

	"github.com/virgoC0der/go-mcp/internal/types"
)

// MockServer implements Server interface for testing
type MockServer struct {
	InitializeFunc            func(ctx context.Context, options any) error
	StartFunc                 func() error
	ShutdownFunc              func(ctx context.Context) error
	ListPromptsFunc           func(ctx context.Context, cursor string) (*types.PromptListResult, error)
	GetPromptFunc             func(ctx context.Context, name string, args map[string]any) (*types.PromptResult, error)
	ListToolsFunc             func(ctx context.Context, cursor string) (*types.ToolListResult, error)
	CallToolFunc              func(ctx context.Context, name string, args map[string]any) (*types.CallToolResult, error)
	ListResourcesFunc         func(ctx context.Context, cursor string) (*types.ResourceListResult, error)
	ReadResourceFunc          func(ctx context.Context, uri string) (*types.ResourceContent, error)
	ListResourceTemplatesFunc func(ctx context.Context) ([]types.ResourceTemplate, error)
	SubscribeToResourceFunc   func(ctx context.Context, uri string) error
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
func (m *MockServer) ListPrompts(ctx context.Context, cursor string) (*types.PromptListResult, error) {
	if m.ListPromptsFunc != nil {
		return m.ListPromptsFunc(ctx, cursor)
	}
	return &types.PromptListResult{Prompts: []types.Prompt{}, NextCursor: ""}, nil
}

// GetPrompt implements the Server interface for testing prompt retrieval
func (m *MockServer) GetPrompt(ctx context.Context, name string, args map[string]any) (*types.PromptResult, error) {
	if m.GetPromptFunc != nil {
		return m.GetPromptFunc(ctx, name, args)
	}
	return nil, nil
}

// ListTools implements the Server interface for testing tool listing
func (m *MockServer) ListTools(ctx context.Context, cursor string) (*types.ToolListResult, error) {
	if m.ListToolsFunc != nil {
		return m.ListToolsFunc(ctx, cursor)
	}
	return &types.ToolListResult{Tools: []types.Tool{}, NextCursor: ""}, nil
}

// CallTool implements the Server interface for testing tool invocation
func (m *MockServer) CallTool(ctx context.Context, name string, args map[string]any) (*types.CallToolResult, error) {
	if m.CallToolFunc != nil {
		return m.CallToolFunc(ctx, name, args)
	}
	return nil, nil
}

// ListResources implements the Server interface for testing resource listing
func (m *MockServer) ListResources(ctx context.Context, cursor string) (*types.ResourceListResult, error) {
	if m.ListResourcesFunc != nil {
		return m.ListResourcesFunc(ctx, cursor)
	}
	return &types.ResourceListResult{Resources: []types.Resource{}, NextCursor: ""}, nil
}

// ReadResource implements the Server interface for testing resource reading
func (m *MockServer) ReadResource(ctx context.Context, uri string) (*types.ResourceContent, error) {
	if m.ReadResourceFunc != nil {
		return m.ReadResourceFunc(ctx, uri)
	}
	return &types.ResourceContent{URI: uri, MimeType: "text/plain"}, nil
}

// ListResourceTemplates implements the Server interface for testing resource template listing
func (m *MockServer) ListResourceTemplates(ctx context.Context) ([]types.ResourceTemplate, error) {
	if m.ListResourceTemplatesFunc != nil {
		return m.ListResourceTemplatesFunc(ctx)
	}
	return []types.ResourceTemplate{}, nil
}

// SubscribeToResource implements the Server interface for testing resource subscription
func (m *MockServer) SubscribeToResource(ctx context.Context, uri string) error {
	if m.SubscribeToResourceFunc != nil {
		return m.SubscribeToResourceFunc(ctx, uri)
	}
	return nil
}
