package types

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPrompt(t *testing.T) {
	prompt := Prompt{
		Name:        "test",
		Description: "Test prompt",
		Template:    "Hello, {{.name}}!",
		Metadata: map[string]interface{}{
			"required": []string{"name"},
		},
	}

	assert.Equal(t, "test", prompt.Name)
	assert.Equal(t, "Test prompt", prompt.Description)
	assert.Equal(t, "Hello, {{.name}}!", prompt.Template)
	assert.NotNil(t, prompt.Metadata)
}

func TestTool(t *testing.T) {
	tool := Tool{
		Name:        "test",
		Description: "Test tool",
		Parameters: map[string]interface{}{
			"name": map[string]interface{}{
				"type":        "string",
				"description": "Name parameter",
			},
		},
	}

	assert.Equal(t, "test", tool.Name)
	assert.Equal(t, "Test tool", tool.Description)
	assert.NotNil(t, tool.Parameters)
}

func TestResource(t *testing.T) {
	resource := Resource{
		Name:        "test",
		Description: "Test resource",
		Type:        "text/plain",
	}

	assert.Equal(t, "test", resource.Name)
	assert.Equal(t, "Test resource", resource.Description)
	assert.Equal(t, "text/plain", resource.Type)
}

func TestGetPromptResult(t *testing.T) {
	result := GetPromptResult{
		Content: "Hello, test!",
	}

	assert.Equal(t, "Hello, test!", result.Content)
}

func TestCallToolResult(t *testing.T) {
	result := CallToolResult{
		Output: map[string]interface{}{
			"result": "test result",
		},
	}

	assert.NotNil(t, result.Output)
	assert.Equal(t, "test result", result.Output.(map[string]interface{})["result"])
}

func TestServerOptions(t *testing.T) {
	options := ServerOptions{
		Address: ":8080",
	}

	assert.Equal(t, ":8080", options.Address)
}

func TestClientOptions(t *testing.T) {
	options := ClientOptions{
		ServerAddress: "localhost:8080",
		Type:          "http",
	}

	assert.Equal(t, "localhost:8080", options.ServerAddress)
	assert.Equal(t, "http", options.Type)
}
