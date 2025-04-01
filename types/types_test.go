package types

import (
	"encoding/json"
	"testing"
)

func TestPromptSerialization(t *testing.T) {
	// Create a prompt
	prompt := Prompt{
		Name:        "test_prompt",
		Description: "A test prompt",
		Arguments: []PromptArgument{
			{
				Name:        "arg1",
				Description: "Argument 1",
				Required:    true,
			},
			{
				Name:        "arg2",
				Description: "Argument 2",
				Required:    false,
			},
		},
	}

	// Serialize to JSON
	jsonData, err := json.Marshal(prompt)
	if err != nil {
		t.Fatalf("Failed to marshal prompt to JSON: %v", err)
	}

	// Deserialize back
	var deserializedPrompt Prompt
	err = json.Unmarshal(jsonData, &deserializedPrompt)
	if err != nil {
		t.Fatalf("Failed to unmarshal prompt from JSON: %v", err)
	}

	// Check fields
	if prompt.Name != deserializedPrompt.Name {
		t.Errorf("Name mismatch: got %s, want %s", deserializedPrompt.Name, prompt.Name)
	}
	if prompt.Description != deserializedPrompt.Description {
		t.Errorf("Description mismatch: got %s, want %s", deserializedPrompt.Description, prompt.Description)
	}
	if len(prompt.Arguments) != len(deserializedPrompt.Arguments) {
		t.Errorf("Arguments length mismatch: got %d, want %d", len(deserializedPrompt.Arguments), len(prompt.Arguments))
	}
}

func TestToolSerialization(t *testing.T) {
	// Create a tool
	tool := Tool{
		Name:        "test_tool",
		Description: "A test tool",
		Arguments: []PromptArgument{
			{
				Name:        "arg1",
				Description: "Argument 1",
				Required:    true,
			},
		},
		Schema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"arg1": map[string]any{
					"type":        "string",
					"description": "Argument 1",
				},
			},
		},
	}

	// Serialize to JSON
	jsonData, err := json.Marshal(tool)
	if err != nil {
		t.Fatalf("Failed to marshal tool to JSON: %v", err)
	}

	// Deserialize back
	var deserializedTool Tool
	err = json.Unmarshal(jsonData, &deserializedTool)
	if err != nil {
		t.Fatalf("Failed to unmarshal tool from JSON: %v", err)
	}

	// Check fields
	if tool.Name != deserializedTool.Name {
		t.Errorf("Name mismatch: got %s, want %s", deserializedTool.Name, tool.Name)
	}
	if tool.Description != deserializedTool.Description {
		t.Errorf("Description mismatch: got %s, want %s", deserializedTool.Description, tool.Description)
	}
	if len(tool.Arguments) != len(deserializedTool.Arguments) {
		t.Errorf("Arguments length mismatch: got %d, want %d", len(deserializedTool.Arguments), len(tool.Arguments))
	}
}

func TestResourceSerialization(t *testing.T) {
	// Create a resource
	resource := Resource{
		Name:        "test_resource",
		Description: "A test resource",
		MimeType:    "text/plain",
	}

	// Serialize to JSON
	jsonData, err := json.Marshal(resource)
	if err != nil {
		t.Fatalf("Failed to marshal resource to JSON: %v", err)
	}

	// Deserialize back
	var deserializedResource Resource
	err = json.Unmarshal(jsonData, &deserializedResource)
	if err != nil {
		t.Fatalf("Failed to unmarshal resource from JSON: %v", err)
	}

	// Check fields
	if resource.Name != deserializedResource.Name {
		t.Errorf("Name mismatch: got %s, want %s", deserializedResource.Name, resource.Name)
	}
	if resource.Description != deserializedResource.Description {
		t.Errorf("Description mismatch: got %s, want %s", deserializedResource.Description, resource.Description)
	}
	if resource.MimeType != deserializedResource.MimeType {
		t.Errorf("MimeType mismatch: got %s, want %s", deserializedResource.MimeType, resource.MimeType)
	}
}

func TestGetPromptResultSerialization(t *testing.T) {
	// Create a GetPromptResult
	result := GetPromptResult{
		Description: "Test result",
		Message:     "Hello, world!",
		Messages: []Message{
			{
				Role: "user",
				Content: Content{
					Type: "text",
					Text: "Hello",
				},
			},
		},
	}

	// Serialize to JSON
	jsonData, err := json.Marshal(result)
	if err != nil {
		t.Fatalf("Failed to marshal GetPromptResult to JSON: %v", err)
	}

	// Deserialize back
	var deserializedResult GetPromptResult
	err = json.Unmarshal(jsonData, &deserializedResult)
	if err != nil {
		t.Fatalf("Failed to unmarshal GetPromptResult from JSON: %v", err)
	}

	// Check fields
	if result.Description != deserializedResult.Description {
		t.Errorf("Description mismatch: got %s, want %s", deserializedResult.Description, result.Description)
	}
	if result.Message != deserializedResult.Message {
		t.Errorf("Message mismatch: got %s, want %s", deserializedResult.Message, result.Message)
	}
	if len(result.Messages) != len(deserializedResult.Messages) {
		t.Errorf("Messages length mismatch: got %d, want %d", len(deserializedResult.Messages), len(result.Messages))
	}
}

func TestCallToolResultSerialization(t *testing.T) {
	// Create a CallToolResult
	result := CallToolResult{
		Content: map[string]any{
			"key":    "value",
			"number": 42,
		},
	}

	// Serialize to JSON
	jsonData, err := json.Marshal(result)
	if err != nil {
		t.Fatalf("Failed to marshal CallToolResult to JSON: %v", err)
	}

	// Deserialize back
	var deserializedResult CallToolResult
	err = json.Unmarshal(jsonData, &deserializedResult)
	if err != nil {
		t.Fatalf("Failed to unmarshal CallToolResult from JSON: %v", err)
	}

	// Check Content field is present
	if deserializedResult.Content == nil {
		t.Errorf("Content field is nil, expected non-nil")
	}
}
