package server

import (
	"context"
	"testing"

	"github.com/virgoC0der/go-mcp/types"
)

func TestNewBaseServer(t *testing.T) {
	// Create server
	srv := NewBaseServer("test-server", "1.0.0")

	// Check fields
	if srv.name != "test-server" {
		t.Errorf("Expected server name to be 'test-server', got '%s'", srv.name)
	}

	if srv.version != "1.0.0" {
		t.Errorf("Expected server version to be '1.0.0', got '%s'", srv.version)
	}

	// Check maps are initialized
	if srv.prompts == nil {
		t.Error("Prompts map not initialized")
	}

	if srv.tools == nil {
		t.Error("Tools map not initialized")
	}

	if srv.resources == nil {
		t.Error("Resources map not initialized")
	}

	if srv.promptHandlers == nil {
		t.Error("Prompt handlers map not initialized")
	}

	if srv.toolHandlers == nil {
		t.Error("Tool handlers map not initialized")
	}

	if srv.notifications == nil {
		t.Error("Notifications registry not initialized")
	}

	if srv.schemaGen == nil {
		t.Error("Schema generator not initialized")
	}
}

func TestBaseServer_Initialize(t *testing.T) {
	// Create server
	srv := NewBaseServer("old-name", "0.0.1")

	// Initialize with options
	err := srv.Initialize(context.Background(), types.InitializationOptions{
		ServerName:    "new-name",
		ServerVersion: "1.0.0",
		Capabilities: types.ServerCapabilities{
			Prompts:   true,
			Tools:     true,
			Resources: true,
		},
	})

	// Check error
	if err != nil {
		t.Fatalf("Initialize returned error: %v", err)
	}

	// Check fields were updated
	if srv.name != "new-name" {
		t.Errorf("Expected server name to be updated to 'new-name', got '%s'", srv.name)
	}

	if srv.version != "1.0.0" {
		t.Errorf("Expected server version to be updated to '1.0.0', got '%s'", srv.version)
	}

	// Test with map[string]any
	srv = NewBaseServer("old-name", "0.0.1")

	// Initialize with map
	err = srv.Initialize(context.Background(), map[string]any{
		"serverName":    "map-name",
		"serverVersion": "2.0.0",
	})

	// Check error
	if err != nil {
		t.Fatalf("Initialize with map returned error: %v", err)
	}

	// Check fields were updated
	if srv.name != "map-name" {
		t.Errorf("Expected server name to be updated to 'map-name', got '%s'", srv.name)
	}

	if srv.version != "2.0.0" {
		t.Errorf("Expected server version to be updated to '2.0.0', got '%s'", srv.version)
	}

	// Test with invalid type
	srv = NewBaseServer("old-name", "0.0.1")

	// Initialize with invalid type
	err = srv.Initialize(context.Background(), "not an options object")

	// Check error
	if err == nil {
		t.Errorf("Expected error when initializing with invalid type, got nil")
	}
}

func TestBaseServer_RegisterPrompt(t *testing.T) {
	// Create server
	srv := NewBaseServer("test-server", "1.0.0")

	// Register prompt
	prompt := types.Prompt{
		Name:        "test-prompt",
		Description: "A test prompt",
		Arguments: []types.PromptArgument{
			{
				Name:        "arg1",
				Description: "Argument 1",
				Required:    true,
			},
		},
	}

	srv.RegisterPrompt(prompt)

	// Check prompt was registered
	registeredPrompt, ok := srv.prompts["test-prompt"]
	if !ok {
		t.Fatal("Prompt not registered")
	}

	// Check fields
	if registeredPrompt.Name != prompt.Name {
		t.Errorf("Prompt name mismatch: expected '%s', got '%s'", prompt.Name, registeredPrompt.Name)
	}

	if registeredPrompt.Description != prompt.Description {
		t.Errorf("Prompt description mismatch: expected '%s', got '%s'", prompt.Description, registeredPrompt.Description)
	}

	if len(registeredPrompt.Arguments) != len(prompt.Arguments) {
		t.Errorf("Prompt arguments length mismatch: expected %d, got %d", len(prompt.Arguments), len(registeredPrompt.Arguments))
	}
}

func TestBaseServer_RegisterTool(t *testing.T) {
	// Create server
	srv := NewBaseServer("test-server", "1.0.0")

	// Register tool
	tool := types.Tool{
		Name:        "test-tool",
		Description: "A test tool",
		Arguments: []types.PromptArgument{
			{
				Name:        "arg1",
				Description: "Argument 1",
				Required:    true,
			},
		},
	}

	srv.RegisterTool(tool)

	// Check tool was registered
	registeredTool, ok := srv.tools["test-tool"]
	if !ok {
		t.Fatal("Tool not registered")
	}

	// Check fields
	if registeredTool.Name != tool.Name {
		t.Errorf("Tool name mismatch: expected '%s', got '%s'", tool.Name, registeredTool.Name)
	}

	if registeredTool.Description != tool.Description {
		t.Errorf("Tool description mismatch: expected '%s', got '%s'", tool.Description, registeredTool.Description)
	}

	if len(registeredTool.Arguments) != len(tool.Arguments) {
		t.Errorf("Tool arguments length mismatch: expected %d, got %d", len(tool.Arguments), len(registeredTool.Arguments))
	}
}

func TestBaseServer_RegisterResource(t *testing.T) {
	// Create server
	srv := NewBaseServer("test-server", "1.0.0")

	// Register resource
	resource := types.Resource{
		Name:        "test-resource",
		Description: "A test resource",
		MimeType:    "text/plain",
	}

	srv.RegisterResource(resource)

	// Check resource was registered
	registeredResource, ok := srv.resources["test-resource"]
	if !ok {
		t.Fatal("Resource not registered")
	}

	// Check fields
	if registeredResource.Name != resource.Name {
		t.Errorf("Resource name mismatch: expected '%s', got '%s'", resource.Name, registeredResource.Name)
	}

	if registeredResource.Description != resource.Description {
		t.Errorf("Resource description mismatch: expected '%s', got '%s'", resource.Description, registeredResource.Description)
	}

	if registeredResource.MimeType != resource.MimeType {
		t.Errorf("Resource mime type mismatch: expected '%s', got '%s'", resource.MimeType, registeredResource.MimeType)
	}
}

func TestBaseServer_ListPrompts(t *testing.T) {
	// Create server
	srv := NewBaseServer("test-server", "1.0.0")

	// Register prompts
	srv.RegisterPrompt(types.Prompt{Name: "prompt1", Description: "Prompt 1"})
	srv.RegisterPrompt(types.Prompt{Name: "prompt2", Description: "Prompt 2"})

	// List prompts
	prompts, err := srv.ListPrompts(context.Background())
	if err != nil {
		t.Fatalf("ListPrompts returned error: %v", err)
	}

	// Check length
	if len(prompts) != 2 {
		t.Errorf("Expected 2 prompts, got %d", len(prompts))
	}

	// Check prompts are in the list
	foundPrompt1 := false
	foundPrompt2 := false
	for _, p := range prompts {
		if p.Name == "prompt1" {
			foundPrompt1 = true
		}
		if p.Name == "prompt2" {
			foundPrompt2 = true
		}
	}

	if !foundPrompt1 {
		t.Error("prompt1 not found in list")
	}

	if !foundPrompt2 {
		t.Error("prompt2 not found in list")
	}
}

func TestBaseServer_ListTools(t *testing.T) {
	// Create server
	srv := NewBaseServer("test-server", "1.0.0")

	// Register tools
	srv.RegisterTool(types.Tool{Name: "tool1", Description: "Tool 1"})
	srv.RegisterTool(types.Tool{Name: "tool2", Description: "Tool 2"})

	// List tools
	tools, err := srv.ListTools(context.Background())
	if err != nil {
		t.Fatalf("ListTools returned error: %v", err)
	}

	// Check length
	if len(tools) != 2 {
		t.Errorf("Expected 2 tools, got %d", len(tools))
	}

	// Check tools are in the list
	foundTool1 := false
	foundTool2 := false
	for _, tool := range tools {
		if tool.Name == "tool1" {
			foundTool1 = true
		}
		if tool.Name == "tool2" {
			foundTool2 = true
		}
	}

	if !foundTool1 {
		t.Error("tool1 not found in list")
	}

	if !foundTool2 {
		t.Error("tool2 not found in list")
	}
}

func TestBaseServer_ListResources(t *testing.T) {
	// Create server
	srv := NewBaseServer("test-server", "1.0.0")

	// Register resources
	srv.RegisterResource(types.Resource{Name: "resource1", Description: "Resource 1", MimeType: "text/plain"})
	srv.RegisterResource(types.Resource{Name: "resource2", Description: "Resource 2", MimeType: "application/json"})

	// List resources
	resources, err := srv.ListResources(context.Background())
	if err != nil {
		t.Fatalf("ListResources returned error: %v", err)
	}

	// Check length
	if len(resources) != 2 {
		t.Errorf("Expected 2 resources, got %d", len(resources))
	}

	// Check resources are in the list
	foundResource1 := false
	foundResource2 := false
	for _, r := range resources {
		if r.Name == "resource1" {
			foundResource1 = true
		}
		if r.Name == "resource2" {
			foundResource2 = true
		}
	}

	if !foundResource1 {
		t.Error("resource1 not found in list")
	}

	if !foundResource2 {
		t.Error("resource2 not found in list")
	}
}

func TestBaseServer_GetPrompt(t *testing.T) {
	// Create server
	srv := NewBaseServer("test-server", "1.0.0")

	// Register prompt
	srv.RegisterPrompt(types.Prompt{
		Name:        "test-prompt",
		Description: "A test prompt",
		Arguments: []types.PromptArgument{
			{
				Name:        "arg1",
				Description: "Argument 1",
				Required:    true,
			},
		},
	})

	// Get prompt
	result, err := srv.GetPrompt(context.Background(), "test-prompt", map[string]any{
		"arg1": "value1",
	})

	// Check error
	if err != nil {
		t.Fatalf("GetPrompt returned error: %v", err)
	}

	// Check result
	if result == nil {
		t.Fatal("GetPrompt returned nil result")
	}

	if result.Description != "A test prompt" {
		t.Errorf("Result description mismatch: expected 'A test prompt', got '%s'", result.Description)
	}

	// Test with non-existent prompt
	_, err = srv.GetPrompt(context.Background(), "non-existent", nil)
	if err == nil {
		t.Error("Expected error when getting non-existent prompt, got nil")
	}
}

func TestBaseServer_CallTool(t *testing.T) {
	// Create server
	srv := NewBaseServer("test-server", "1.0.0")

	// Register tool
	srv.RegisterTool(types.Tool{
		Name:        "test-tool",
		Description: "A test tool",
		Arguments: []types.PromptArgument{
			{
				Name:        "arg1",
				Description: "Argument 1",
				Required:    true,
			},
		},
	})

	// Call tool
	result, err := srv.CallTool(context.Background(), "test-tool", map[string]any{
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

	// Test with non-existent tool
	_, err = srv.CallTool(context.Background(), "non-existent", nil)
	if err == nil {
		t.Error("Expected error when calling non-existent tool, got nil")
	}
}

func TestBaseServer_ReadResource(t *testing.T) {
	// Create server
	srv := NewBaseServer("test-server", "1.0.0")

	// Register resource
	srv.RegisterResource(types.Resource{
		Name:        "test-resource",
		Description: "A test resource",
		MimeType:    "text/plain",
	})

	// Read resource
	content, mimeType, err := srv.ReadResource(context.Background(), "test-resource")

	// Check error
	if err != nil {
		t.Fatalf("ReadResource returned error: %v", err)
	}

	// Check content
	if content == nil {
		t.Fatal("ReadResource returned nil content")
	}

	// Check mime type
	if mimeType != "text/plain" {
		t.Errorf("MimeType mismatch: expected 'text/plain', got '%s'", mimeType)
	}

	// Test with non-existent resource
	_, _, err = srv.ReadResource(context.Background(), "non-existent")
	if err == nil {
		t.Error("Expected error when reading non-existent resource, got nil")
	}
}
