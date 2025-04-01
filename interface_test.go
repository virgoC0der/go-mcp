package mcp

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/virgoC0der/go-mcp/internal/types"
)

// MockService implements MCPService for testing
type MockService struct {
	prompts   []types.Prompt
	tools     []types.Tool
	resources []types.Resource
}

func NewMockService() *MockService {
	return &MockService{
		prompts: []types.Prompt{
			{
				Name:        "test_prompt",
				Description: "Test prompt",
				Template:    "Hello, {{.name}}!",
			},
		},
		tools: []types.Tool{
			{
				Name:        "test_tool",
				Description: "Test tool",
				Parameters: map[string]interface{}{
					"name": map[string]interface{}{
						"type":        "string",
						"description": "Name parameter",
					},
				},
			},
		},
		resources: []types.Resource{
			{
				Name:        "test_resource",
				Description: "Test resource",
				Type:        "text/plain",
			},
		},
	}
}

func (m *MockService) Initialize(ctx context.Context, options any) error {
	return nil
}

func (m *MockService) ListPrompts(ctx context.Context) ([]types.Prompt, error) {
	return m.prompts, nil
}

func (m *MockService) GetPrompt(ctx context.Context, name string, args map[string]any) (*types.GetPromptResult, error) {
	return &types.GetPromptResult{
		Content: "Hello, test!",
	}, nil
}

func (m *MockService) ListTools(ctx context.Context) ([]types.Tool, error) {
	return m.tools, nil
}

func (m *MockService) CallTool(ctx context.Context, name string, args map[string]any) (*types.CallToolResult, error) {
	return &types.CallToolResult{
		Output: map[string]interface{}{
			"result": "test result",
		},
	}, nil
}

func (m *MockService) ListResources(ctx context.Context) ([]types.Resource, error) {
	return m.resources, nil
}

func (m *MockService) ReadResource(ctx context.Context, name string) ([]byte, string, error) {
	return []byte("test content"), "text/plain", nil
}

func TestNewServer(t *testing.T) {
	service := NewMockService()

	server, err := NewServer(service, &types.ServerOptions{
		Address: ":8080",
	})

	assert.NoError(t, err)
	assert.NotNil(t, server)
}

func TestNewClient(t *testing.T) {
	client, err := NewClient(&types.ClientOptions{
		ServerAddress: "localhost:8080",
		Type:          "http",
	})

	assert.NoError(t, err)
	assert.NotNil(t, client)
}
