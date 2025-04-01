package client

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/virgoC0der/go-mcp/types"
)

// MockTransport implements the Transport interface for testing
type MockTransport struct {
	sendRequestFunc func(ctx context.Context, requestType string, params map[string]any) (any, error)
	closeFunc       func() error
}

func (m *MockTransport) SendRequest(ctx context.Context, requestType string, params map[string]any) (any, error) {
	return m.sendRequestFunc(ctx, requestType, params)
}

func (m *MockTransport) Close() error {
	return m.closeFunc()
}

func TestClient_Initialize(t *testing.T) {
	// Create mock transport
	mockTransport := &MockTransport{
		sendRequestFunc: func(ctx context.Context, requestType string, params map[string]any) (any, error) {
			// Check request type
			if requestType != "initialize" {
				t.Errorf("Expected request type 'initialize', got '%s'", requestType)
			}

			// Check parameters
			if params["serverName"] != "client" {
				t.Errorf("Expected serverName 'client', got '%v'", params["serverName"])
			}

			if params["serverVersion"] != "1.0.0" {
				t.Errorf("Expected serverVersion '1.0.0', got '%v'", params["serverVersion"])
			}

			return nil, nil
		},
	}

	// Create client
	client := NewClient(mockTransport)

	// Initialize client
	err := client.Initialize(context.Background())

	// Check error
	if err != nil {
		t.Fatalf("Initialize returned error: %v", err)
	}
}

func TestClient_ListPrompts(t *testing.T) {
	// Create mock prompts
	mockPrompts := []types.Prompt{
		{Name: "prompt1", Description: "Prompt 1"},
		{Name: "prompt2", Description: "Prompt 2"},
	}

	// Create mock transport
	mockTransport := &MockTransport{
		sendRequestFunc: func(ctx context.Context, requestType string, params map[string]any) (any, error) {
			// Check request type
			if requestType != "listPrompts" {
				t.Errorf("Expected request type 'listPrompts', got '%s'", requestType)
			}

			return mockPrompts, nil
		},
	}

	// Create client
	client := NewClient(mockTransport)

	// List prompts
	prompts, err := client.ListPrompts(context.Background())

	// Check error
	if err != nil {
		t.Fatalf("ListPrompts returned error: %v", err)
	}

	// Check prompts
	if len(prompts) != len(mockPrompts) {
		t.Errorf("Expected %d prompts, got %d", len(mockPrompts), len(prompts))
	}

	for i, prompt := range prompts {
		if prompt.Name != mockPrompts[i].Name {
			t.Errorf("Prompt %d name mismatch: expected '%s', got '%s'", i, mockPrompts[i].Name, prompt.Name)
		}

		if prompt.Description != mockPrompts[i].Description {
			t.Errorf("Prompt %d description mismatch: expected '%s', got '%s'", i, mockPrompts[i].Description, prompt.Description)
		}
	}
}

func TestClient_GetPrompt(t *testing.T) {
	// Create mock result
	mockResult := &types.GetPromptResult{
		Description: "Test prompt",
		Message:     "Hello, world!",
	}

	// Create mock transport
	mockTransport := &MockTransport{
		sendRequestFunc: func(ctx context.Context, requestType string, params map[string]any) (any, error) {
			// Check request type
			if requestType != "getPrompt" {
				t.Errorf("Expected request type 'getPrompt', got '%s'", requestType)
			}

			// Check parameters
			if params["name"] != "test-prompt" {
				t.Errorf("Expected name 'test-prompt', got '%v'", params["name"])
			}

			if args, ok := params["args"].(map[string]any); !ok {
				t.Error("Expected args parameter to be map[string]any")
			} else if args["arg1"] != "value1" {
				t.Errorf("Expected arg1 'value1', got '%v'", args["arg1"])
			}

			return mockResult, nil
		},
	}

	// Create client
	client := NewClient(mockTransport)

	// Get prompt
	result, err := client.GetPrompt(context.Background(), "test-prompt", map[string]any{
		"arg1": "value1",
	})

	// Check error
	if err != nil {
		t.Fatalf("GetPrompt returned error: %v", err)
	}

	// Check result
	if result.Description != mockResult.Description {
		t.Errorf("Result description mismatch: expected '%s', got '%s'", mockResult.Description, result.Description)
	}

	if result.Message != mockResult.Message {
		t.Errorf("Result message mismatch: expected '%s', got '%s'", mockResult.Message, result.Message)
	}
}

func TestClient_ListTools(t *testing.T) {
	// Create mock tools
	mockTools := []types.Tool{
		{Name: "tool1", Description: "Tool 1"},
		{Name: "tool2", Description: "Tool 2"},
	}

	// Create mock transport
	mockTransport := &MockTransport{
		sendRequestFunc: func(ctx context.Context, requestType string, params map[string]any) (any, error) {
			// Check request type
			if requestType != "listTools" {
				t.Errorf("Expected request type 'listTools', got '%s'", requestType)
			}

			return mockTools, nil
		},
	}

	// Create client
	client := NewClient(mockTransport)

	// List tools
	tools, err := client.ListTools(context.Background())

	// Check error
	if err != nil {
		t.Fatalf("ListTools returned error: %v", err)
	}

	// Check tools
	if len(tools) != len(mockTools) {
		t.Errorf("Expected %d tools, got %d", len(mockTools), len(tools))
	}

	for i, tool := range tools {
		if tool.Name != mockTools[i].Name {
			t.Errorf("Tool %d name mismatch: expected '%s', got '%s'", i, mockTools[i].Name, tool.Name)
		}

		if tool.Description != mockTools[i].Description {
			t.Errorf("Tool %d description mismatch: expected '%s', got '%s'", i, mockTools[i].Description, tool.Description)
		}
	}
}

func TestClient_CallTool(t *testing.T) {
	// Create mock result
	mockResult := &types.CallToolResult{
		Content: map[string]any{
			"message": "Tool executed successfully",
		},
	}

	// Create mock transport
	mockTransport := &MockTransport{
		sendRequestFunc: func(ctx context.Context, requestType string, params map[string]any) (any, error) {
			// Check request type
			if requestType != "callTool" {
				t.Errorf("Expected request type 'callTool', got '%s'", requestType)
			}

			// Check parameters
			if params["name"] != "test-tool" {
				t.Errorf("Expected name 'test-tool', got '%v'", params["name"])
			}

			if args, ok := params["args"].(map[string]any); !ok {
				t.Error("Expected args parameter to be map[string]any")
			} else if args["arg1"] != "value1" {
				t.Errorf("Expected arg1 'value1', got '%v'", args["arg1"])
			}

			return mockResult, nil
		},
	}

	// Create client
	client := NewClient(mockTransport)

	// Call tool
	result, err := client.CallTool(context.Background(), "test-tool", map[string]any{
		"arg1": "value1",
	})

	// Check error
	if err != nil {
		t.Fatalf("CallTool returned error: %v", err)
	}

	// Check result
	if result == nil {
		t.Fatal("CallTool returned nil result")
	}

	contentJSON, err := json.Marshal(result.Content)
	if err != nil {
		t.Fatalf("Failed to marshal result content: %v", err)
	}

	expectedJSON, err := json.Marshal(mockResult.Content)
	if err != nil {
		t.Fatalf("Failed to marshal expected content: %v", err)
	}

	if string(contentJSON) != string(expectedJSON) {
		t.Errorf("Result content mismatch: expected '%s', got '%s'", string(expectedJSON), string(contentJSON))
	}
}

func TestClient_ListResources(t *testing.T) {
	// Create mock resources
	mockResources := []types.Resource{
		{Name: "resource1", Description: "Resource 1", MimeType: "text/plain"},
		{Name: "resource2", Description: "Resource 2", MimeType: "application/json"},
	}

	// Create mock transport
	mockTransport := &MockTransport{
		sendRequestFunc: func(ctx context.Context, requestType string, params map[string]any) (any, error) {
			// Check request type
			if requestType != "listResources" {
				t.Errorf("Expected request type 'listResources', got '%s'", requestType)
			}

			return mockResources, nil
		},
	}

	// Create client
	client := NewClient(mockTransport)

	// List resources
	resources, err := client.ListResources(context.Background())

	// Check error
	if err != nil {
		t.Fatalf("ListResources returned error: %v", err)
	}

	// Check resources
	if len(resources) != len(mockResources) {
		t.Errorf("Expected %d resources, got %d", len(mockResources), len(resources))
	}

	for i, resource := range resources {
		if resource.Name != mockResources[i].Name {
			t.Errorf("Resource %d name mismatch: expected '%s', got '%s'", i, mockResources[i].Name, resource.Name)
		}

		if resource.Description != mockResources[i].Description {
			t.Errorf("Resource %d description mismatch: expected '%s', got '%s'", i, mockResources[i].Description, resource.Description)
		}

		if resource.MimeType != mockResources[i].MimeType {
			t.Errorf("Resource %d mime type mismatch: expected '%s', got '%s'", i, mockResources[i].MimeType, resource.MimeType)
		}
	}
}

func TestClient_ReadResource(t *testing.T) {
	// Create mock result
	mockContent := []byte("resource content")
	mockMimeType := "text/plain"

	// Create mock transport
	mockTransport := &MockTransport{
		sendRequestFunc: func(ctx context.Context, requestType string, params map[string]any) (any, error) {
			// Check request type
			if requestType != "readResource" {
				t.Errorf("Expected request type 'readResource', got '%s'", requestType)
			}

			// Check parameters
			if params["name"] != "test-resource" {
				t.Errorf("Expected name 'test-resource', got '%v'", params["name"])
			}

			return map[string]any{
				"content":  "cmVzb3VyY2UgY29udGVudA==", // base64 of "resource content"
				"mimeType": mockMimeType,
			}, nil
		},
	}

	// Create client
	client := NewClient(mockTransport)

	// Read resource
	content, mimeType, err := client.ReadResource(context.Background(), "test-resource")

	// Check error
	if err != nil {
		t.Fatalf("ReadResource returned error: %v", err)
	}

	// Check content
	if string(content) != string(mockContent) {
		t.Errorf("Content mismatch: expected '%s', got '%s'", string(mockContent), string(content))
	}

	// Check mime type
	if mimeType != mockMimeType {
		t.Errorf("Mime type mismatch: expected '%s', got '%s'", mockMimeType, mimeType)
	}
}

func TestClient_Close(t *testing.T) {
	closed := false

	// Create mock transport
	mockTransport := &MockTransport{
		closeFunc: func() error {
			closed = true
			return nil
		},
	}

	// Create client
	client := NewClient(mockTransport)

	// Close client
	err := client.Close()

	// Check error
	if err != nil {
		t.Fatalf("Close returned error: %v", err)
	}

	// Check if transport was closed
	if !closed {
		t.Error("Transport Close() was not called")
	}
}
