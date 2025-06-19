package transport

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/virgoC0der/go-mcp/internal/types"
	"github.com/virgoC0der/go-mcp/transport/http"
)

// mockMCPService is a simple mock implementation of types.MCPService for factory tests
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

// Implement Server interface methods needed for type assertion if necessary (Initialize, Start, Shutdown)
func (m *mockMCPService) Initialize(ctx context.Context, options any) error { return nil }
func (m *mockMCPService) Start() error                                      { return nil }
func (m *mockMCPService) Shutdown(ctx context.Context) error                { return nil }

func TestNewServer(t *testing.T) {
	mockService := &mockMCPService{}

	// Test with nil options
	t.Run("nil options", func(t *testing.T) {
		server, err := NewServer(mockService, nil)
		assert.NoError(t, err)
		assert.NotNil(t, server)
		// Check if it's an HTTP server (or the expected type)
		_, ok := server.(*http.HTTPServer)
		assert.True(t, ok, "Expected server to be *http.HTTPServer with nil options")
		// Optionally check default address if accessible from HTTPServer
	})

	// Test with specific options
	t.Run("specific options", func(t *testing.T) {
		options := &types.ServerOptions{
			Address: "localhost:9999",
			// Type is not used by factory's NewServer currently
		}
		server, err := NewServer(mockService, options)
		assert.NoError(t, err)
		assert.NotNil(t, server)
		// Check if it's an HTTP server
		_, ok := server.(*http.HTTPServer)
		assert.True(t, ok, "Expected server to be *http.HTTPServer with specific options")
		// Optionally check if the address was correctly passed to HTTPServer if accessible
	})
}

func TestNewClient(t *testing.T) {

	// Test with nil options
	t.Run("nil options", func(t *testing.T) {
		client, err := NewClient(nil)
		assert.NoError(t, err)
		assert.NotNil(t, client)
		// Check default type (currently HTTP)
		_, ok := client.(*http.HTTPClient)
		assert.True(t, ok, "Expected client to be *http.HTTPClient with nil options")
		// Optionally check default address if accessible
	})

	// Test with type http
	t.Run("type http", func(t *testing.T) {
		options := &types.ClientOptions{
			ServerAddress: "127.0.0.1:5000",
			Type:          "http",
		}
		client, err := NewClient(options)
		assert.NoError(t, err)
		assert.NotNil(t, client)
		_, ok := client.(*http.HTTPClient)
		assert.True(t, ok, "Expected client to be *http.HTTPClient for type http")
		// Optionally check address if accessible
	})

	// Test with type websocket (currently falls back to http)
	t.Run("type websocket", func(t *testing.T) {
		options := &types.ClientOptions{
			ServerAddress: "ws://localhost:8080",
			Type:          "websocket",
		}
		client, err := NewClient(options)
		assert.NoError(t, err)
		assert.NotNil(t, client)
		// Factory currently returns HTTPClient as default/fallback
		_, ok := client.(*http.HTTPClient)
		assert.True(t, ok, "Expected client to be *http.HTTPClient as fallback for type websocket")
	})

	// Test with unknown type (currently falls back to http)
	t.Run("type unknown", func(t *testing.T) {
		options := &types.ClientOptions{
			ServerAddress: "somehost:1111",
			Type:          "unknown",
		}
		client, err := NewClient(options)
		assert.NoError(t, err)
		assert.NotNil(t, client)
		// Factory currently returns HTTPClient as default/fallback
		_, ok := client.(*http.HTTPClient)
		assert.True(t, ok, "Expected client to be *http.HTTPClient as fallback for unknown type")
	})
}
