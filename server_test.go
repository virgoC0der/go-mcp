package mcp

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/virgoC0der/go-mcp/internal/types"
)

func TestServerLifecycle(t *testing.T) {
	service := NewMockService()

	server, err := NewServer(service, &types.ServerOptions{
		Address: ":8081",
	})
	assert.NoError(t, err)
	assert.NotNil(t, server)

	// Test Initialize
	err = server.Initialize(context.Background(), nil)
	assert.NoError(t, err)

	// Create error channel for goroutine
	errChan := make(chan error, 1)
	go func() {
		errChan <- server.Start()
	}()

	// Give server time to start
	time.Sleep(100 * time.Millisecond)

	// Test Shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = server.Shutdown(ctx)
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

func TestServerMethods(t *testing.T) {
	service := NewMockService()

	server, err := NewServer(service, &types.ServerOptions{
		Address: ":8082",
	})
	assert.NoError(t, err)

	ctx := context.Background()

	// Test ListPrompts
	prompts, err := server.ListPrompts(ctx)
	assert.NoError(t, err)
	assert.NotEmpty(t, prompts)

	// Test GetPrompt
	result, err := server.GetPrompt(ctx, "test_prompt", map[string]any{"name": "test"})
	assert.NoError(t, err)
	assert.NotNil(t, result)

	// Test ListTools
	tools, err := server.ListTools(ctx)
	assert.NoError(t, err)
	assert.NotEmpty(t, tools)

	// Test CallTool
	toolResult, err := server.CallTool(ctx, "test_tool", map[string]any{"name": "test"})
	assert.NoError(t, err)
	assert.NotNil(t, toolResult)

	// Test ListResources
	resources, err := server.ListResources(ctx)
	assert.NoError(t, err)
	assert.NotEmpty(t, resources)

	// Test ReadResource
	content, contentType, err := server.ReadResource(ctx, "test_resource")
	assert.NoError(t, err)
	assert.NotEmpty(t, content)
	assert.Equal(t, "text/plain", contentType)
}
