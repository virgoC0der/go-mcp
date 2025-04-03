package mcp

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/virgoC0der/go-mcp/internal/types"
)

func TestClientLifecycle(t *testing.T) {
	// Start a test server first
	service := NewMockService()
	server, err := NewServer(service, &types.ServerOptions{
		Address: ":8083",
	})
	assert.NoError(t, err)

	// Initialize and start server
	err = server.Initialize(context.Background(), nil)
	assert.NoError(t, err)

	// Create error channel for goroutine
	errChan := make(chan error, 1)
	go func() {
		errChan <- server.Start()
	}()

	// Give server time to start
	time.Sleep(100 * time.Millisecond)

	// Create and test client
	client, err := NewClient(&types.ClientOptions{
		ServerAddress: "localhost:8083",
		Type:          "http",
	})
	assert.NoError(t, err)
	assert.NotNil(t, client)

	// Test Connect
	err = client.Connect(context.Background())
	assert.NoError(t, err)

	// Test Service methods
	mcpService := client.Service()
	assert.NotNil(t, mcpService)

	ctx := context.Background()

	// Test ListPrompts
	prompts, err := service.ListPrompts(ctx)
	assert.NoError(t, err)
	assert.NotEmpty(t, prompts)

	// Test GetPrompt
	promptResult, err := service.GetPrompt(ctx, "test_prompt", map[string]any{"name": "test"})
	assert.NoError(t, err)
	assert.NotNil(t, promptResult)

	// Test ListTools
	tools, err := service.ListTools(ctx)
	assert.NoError(t, err)
	assert.NotEmpty(t, tools)

	// Test CallTool
	toolResult, err := service.CallTool(ctx, "test_tool", map[string]any{"name": "test"})
	assert.NoError(t, err)
	assert.NotNil(t, toolResult)

	// Test ListResources
	resources, err := service.ListResources(ctx)
	assert.NoError(t, err)
	assert.NotEmpty(t, resources)

	// Test ReadResource
	content, contentType, err := service.ReadResource(ctx, "test_resource")
	assert.NoError(t, err)
	assert.NotEmpty(t, content)
	assert.Equal(t, "text/plain", contentType)

	// Test Close
	err = client.Close()
	assert.NoError(t, err)

	// Cleanup: shutdown server
	err = server.Shutdown(context.Background())
	assert.NoError(t, err)

	// Check for any errors from the server goroutine
	select {
	case err := <-errChan:
		if err != nil && err.Error() != "http: Server closed" {
			t.Errorf("Server goroutine error: %v", err)
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("Timeout waiting for server goroutine to finish")
	}
}
