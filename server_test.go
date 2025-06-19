package mcp

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/virgoC0der/go-mcp/internal/types"
)

// mockMCPService is a simple mock implementation of types.MCPService
type mockMCPService struct{}

func (m *mockMCPService) ListPrompts(ctx context.Context, cursor string) (*types.PromptListResult, error) {
	return nil, nil
}
func (m *mockMCPService) GetPrompt(ctx context.Context, name string, args map[string]any) (*types.PromptResult, error) {
	return nil, nil
}
func (m *mockMCPService) ListTools(ctx context.Context, cursor string) (*types.ToolListResult, error) {
	return nil, nil
}
func (m *mockMCPService) CallTool(ctx context.Context, name string, args map[string]any) (*types.CallToolResult, error) {
	return nil, nil
}
func (m *mockMCPService) ListResources(ctx context.Context, cursor string) (*types.ResourceListResult, error) {
	return nil, nil
}
func (m *mockMCPService) ReadResource(ctx context.Context, uri string) (*types.ResourceContent, error) {
	return nil, nil
}
func (m *mockMCPService) ListResourceTemplates(ctx context.Context) ([]types.ResourceTemplate, error) {
	return nil, nil
}
func (m *mockMCPService) SubscribeToResource(ctx context.Context, uri string) error {
	return nil
}

func TestWithAddress(t *testing.T) {
	options := &types.ServerOptions{}
	testAddress := "localhost:8080"
	option := WithAddress(testAddress)
	option(options)
	assert.Equal(t, testAddress, options.Address)
}

func TestNewServer(t *testing.T) {
	mockService := &mockMCPService{}
	options := &types.ServerOptions{
		Address: "localhost:9090",
	}

	server, err := NewServer(mockService, options)

	assert.NoError(t, err)
	assert.NotNil(t, server)
	// Verify the returned object implements the types.Server interface
	assert.Implements(t, (*types.Server)(nil), server)
}
