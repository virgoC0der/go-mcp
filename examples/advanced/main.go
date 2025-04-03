package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/virgoC0der/go-mcp"
	"github.com/virgoC0der/go-mcp/internal/types"
)

// CalculatorInput defines the input for the calculator tool
type CalculatorInput struct {
	Operation string  `json:"operation" jsonschema:"required,description=The operation to perform (add, subtract, multiply, divide)"`
	A         float64 `json:"a" jsonschema:"required,description=First operand"`
	B         float64 `json:"b" jsonschema:"required,description=Second operand"`
}

// AdvancedServer implements a server with typed handlers and advanced features
type AdvancedServer struct {
	prompts   []types.Prompt
	tools     []types.Tool
	resources []types.Resource
}

// NewAdvancedServer creates a new advanced server
func NewAdvancedServer() *AdvancedServer {
	s := &AdvancedServer{
		prompts: []types.Prompt{
			{
				Name:        "greet",
				Description: "Greet a person",
				Template:    "{{if .formal}}{{if eq .language \"en\"}}Good day{{else if eq .language \"es\"}}Buenos días{{else if eq .language \"fr\"}}Bonjour{{end}}{{else}}{{if eq .language \"en\"}}Hi{{else if eq .language \"es\"}}¡Hola{{else if eq .language \"fr\"}}Salut{{end}}{{end}}, {{.name}}{{if .formal}}.{{else}}!{{end}}",
				Metadata: map[string]interface{}{
					"required": []string{"name"},
					"schema": map[string]interface{}{
						"type": "object",
						"properties": map[string]interface{}{
							"name": map[string]interface{}{
								"type":        "string",
								"description": "Name to greet",
							},
							"formal": map[string]interface{}{
								"type":        "boolean",
								"description": "Whether to use formal greeting",
							},
							"language": map[string]interface{}{
								"type":        "string",
								"description": "Language code for the greeting (en, es, fr)",
								"enum":        []string{"en", "es", "fr"},
							},
						},
					},
				},
			},
		},
		tools: []types.Tool{
			{
				Name:        "calculator",
				Description: "Perform basic arithmetic operations",
				Parameters: map[string]interface{}{
					"operation": map[string]interface{}{
						"type":        "string",
						"description": "Operation to perform",
						"enum":        []string{"add", "subtract", "multiply", "divide"},
					},
					"a": map[string]interface{}{
						"type":        "number",
						"description": "First operand",
					},
					"b": map[string]interface{}{
						"type":        "number",
						"description": "Second operand",
					},
				},
			},
		},
		resources: []types.Resource{
			{
				Name:        "help",
				Description: "Help documentation",
				Type:        "text/markdown",
			},
		},
	}
	return s
}

// Initialize implements the Server interface
func (s *AdvancedServer) Initialize(ctx context.Context, options any) error {
	return nil
}

// ListPrompts implements the Server interface
func (s *AdvancedServer) ListPrompts(ctx context.Context) ([]types.Prompt, error) {
	return s.prompts, nil
}

// GetPrompt implements the Server interface
func (s *AdvancedServer) GetPrompt(ctx context.Context, name string, args map[string]any) (*types.GetPromptResult, error) {
	if name != "greet" {
		return nil, fmt.Errorf("unknown prompt: %s", name)
	}

	// Extract and validate input
	name, ok := args["name"].(string)
	if !ok {
		return nil, fmt.Errorf("missing required argument: name")
	}

	formal, _ := args["formal"].(bool)
	language, ok := args["language"].(string)
	if !ok {
		language = "en"
	}

	// Generate greeting
	var greeting string
	switch language {
	case "en":
		if formal {
			greeting = fmt.Sprintf("Good day, %s.", name)
		} else {
			greeting = fmt.Sprintf("Hi, %s!", name)
		}
	case "es":
		if formal {
			greeting = fmt.Sprintf("Buenos días, %s.", name)
		} else {
			greeting = fmt.Sprintf("¡Hola, %s!", name)
		}
	case "fr":
		if formal {
			greeting = fmt.Sprintf("Bonjour, %s.", name)
		} else {
			greeting = fmt.Sprintf("Salut, %s!", name)
		}
	default:
		greeting = fmt.Sprintf("Hello, %s!", name)
	}

	return &types.GetPromptResult{
		Content: greeting,
	}, nil
}

// ListTools implements the Server interface
func (s *AdvancedServer) ListTools(ctx context.Context) ([]types.Tool, error) {
	return s.tools, nil
}

// CallTool implements the Server interface
func (s *AdvancedServer) CallTool(ctx context.Context, name string, args map[string]any) (*types.CallToolResult, error) {
	if name != "calculator" {
		return nil, fmt.Errorf("unknown tool: %s", name)
	}

	// Extract and validate input
	operation, ok := args["operation"].(string)
	if !ok {
		return nil, fmt.Errorf("missing required argument: operation")
	}

	a, ok := args["a"].(float64)
	if !ok {
		return nil, fmt.Errorf("missing or invalid argument: a")
	}

	b, ok := args["b"].(float64)
	if !ok {
		return nil, fmt.Errorf("missing or invalid argument: b")
	}

	var result float64
	switch operation {
	case "add":
		result = a + b
	case "subtract":
		result = a - b
	case "multiply":
		result = a * b
	case "divide":
		if b == 0 {
			return nil, fmt.Errorf("division by zero")
		}
		result = a / b
	default:
		return nil, fmt.Errorf("unknown operation: %s", operation)
	}

	return &types.CallToolResult{
		Output: map[string]interface{}{
			"result":    result,
			"operation": operation,
			"operands": map[string]float64{
				"a": a,
				"b": b,
			},
		},
	}, nil
}

// ListResources implements the Server interface
func (s *AdvancedServer) ListResources(ctx context.Context) ([]types.Resource, error) {
	return s.resources, nil
}

// ReadResource implements the Server interface
func (s *AdvancedServer) ReadResource(ctx context.Context, name string) ([]byte, string, error) {
	if name != "help" {
		return nil, "", fmt.Errorf("unknown resource: %s", name)
	}

	content := []byte(`# Advanced Server Help
	
## Prompts

- **greet**: Greet a person in different languages
  - Parameters:
    - name: Name to greet (required)
    - formal: Whether to use formal greeting (optional)
    - language: Language code (en, es, fr) (optional, defaults to "en")

## Tools

- **calculator**: Perform basic arithmetic operations
  - Parameters:
    - operation: Operation to perform (add, subtract, multiply, divide) (required)
    - a: First operand (required)
    - b: Second operand (required)
`)

	return content, "text/markdown", nil
}

func main() {
	// Create service
	service := NewAdvancedServer()

	// Create HTTP server
	httpServer, err := mcp.NewServer(service, &types.ServerOptions{
		Address: ":8080",
	})
	if err != nil {
		log.Fatal(err)
	}

	// Create WebSocket server
	wsServer, err := mcp.NewServer(service, &types.ServerOptions{
		Address: ":8081",
	})
	if err != nil {
		log.Fatal(err)
	}

	// Initialize servers
	ctx := context.Background()
	if err := httpServer.Initialize(ctx, nil); err != nil {
		log.Fatal(err)
	}
	if err := wsServer.Initialize(ctx, nil); err != nil {
		log.Fatal(err)
	}

	// Start servers
	if err := httpServer.Start(); err != nil {
		log.Fatal(err)
	}
	if err := wsServer.Start(); err != nil {
		log.Fatal(err)
	}

	// Wait for interrupt signal
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh

	// Graceful shutdown
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		log.Printf("Error shutting down HTTP server: %v", err)
	}
	if err := wsServer.Shutdown(shutdownCtx); err != nil {
		log.Printf("Error shutting down WebSocket server: %v", err)
	}
}
