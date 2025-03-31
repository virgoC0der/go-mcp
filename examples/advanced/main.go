package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/virgoC0der/go-mcp/server"
	"github.com/virgoC0der/go-mcp/transport"
	"github.com/virgoC0der/go-mcp/types"
)

// CalculatorInput defines the input for the calculator tool
type CalculatorInput struct {
	Operation string  `json:"operation" jsonschema:"required,description=The operation to perform (add, subtract, multiply, divide)"`
	A         float64 `json:"a" jsonschema:"required,description=First operand"`
	B         float64 `json:"b" jsonschema:"required,description=Second operand"`
}

// AdvancedServer implements a server with typed handlers and advanced features
type AdvancedServer struct {
	*server.BaseServer
}

// NewAdvancedServer creates a new advanced server
func NewAdvancedServer() *AdvancedServer {
	s := &AdvancedServer{
		BaseServer: server.NewBaseServer("advanced-server", "1.0.0"),
	}

	// Register typed prompt
	err := s.RegisterPromptTyped("greet", "Greet a person", s.handleGreetPrompt)
	if err != nil {
		log.Fatalf("Failed to register prompt: %v", err)
	}

	// Register typed tool
	err = s.RegisterToolTyped("calculator", "Perform basic arithmetic operations", s.handleCalculatorTool)
	if err != nil {
		log.Fatalf("Failed to register tool: %v", err)
	}

	// Register resource
	s.RegisterResource(types.Resource{
		Name:        "help",
		Description: "Help documentation",
		MimeType:    "text/markdown",
	})

	return s
}

// GreetInput defines the input for the greet prompt
type GreetInput struct {
	Name    string `json:"name" jsonschema:"required,description=Name to greet"`
	Formal  bool   `json:"formal" jsonschema:"description=Whether to use formal greeting"`
	Language string `json:"language" jsonschema:"description=Language code for the greeting (en, es, fr)"`
}

// handleGreetPrompt handles the greet prompt
func (s *AdvancedServer) handleGreetPrompt(input GreetInput) (*types.GetPromptResult, error) {
	// Default to English
	if input.Language == "" {
		input.Language = "en"
	}

	var greeting string
	switch input.Language {
	case "en":
		if input.Formal {
			greeting = fmt.Sprintf("Good day, %s.", input.Name)
		} else {
			greeting = fmt.Sprintf("Hi, %s!", input.Name)
		}
	case "es":
		if input.Formal {
			greeting = fmt.Sprintf("Buenos días, %s.", input.Name)
		} else {
			greeting = fmt.Sprintf("¡Hola, %s!", input.Name)
		}
	case "fr":
		if input.Formal {
			greeting = fmt.Sprintf("Bonjour, %s.", input.Name)
		} else {
			greeting = fmt.Sprintf("Salut, %s!", input.Name)
		}
	default:
		greeting = fmt.Sprintf("Hello, %s!", input.Name)
	}

	return &types.GetPromptResult{
		Message: greeting,
	}, nil
}

// handleCalculatorTool handles the calculator tool
func (s *AdvancedServer) handleCalculatorTool(input CalculatorInput) (*types.CallToolResult, error) {
	var result float64

	switch input.Operation {
	case "add":
		result = input.A + input.B
	case "subtract":
		result = input.A - input.B
	case "multiply":
		result = input.A * input.B
	case "divide":
		if input.B == 0 {
			return nil, fmt.Errorf("division by zero")
		}
		result = input.A / input.B
	default:
		return nil, fmt.Errorf("unknown operation: %s", input.Operation)
	}

	content := map[string]interface{}{
		"result": result,
		"operation": input.Operation,
		"operands": map[string]float64{
			"a": input.A,
			"b": input.B,
		},
	}

	return &types.CallToolResult{
		Content: content,
	}, nil
}

// ReadResource implements the resource handler
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
	// Create server
	srv := NewAdvancedServer()

	// Initialize server
	err := srv.Initialize(context.Background(), types.InitializationOptions{
		ServerName:    "advanced-server",
		ServerVersion: "1.0.0",
		Capabilities: types.ServerCapabilities{
			Prompts:   true,
			Tools:     true,
			Resources: true,
		},
	})
	if err != nil {
		log.Fatalf("Failed to initialize server: %v", err)
	}

	// Create transport layers
	httpServer := transport.NewHTTPServer(srv, ":8080")
	wsServer := transport.NewWSServer(srv, ":8081")
	stdioServer := transport.NewStdioServer(srv)

	// Use WaitGroup to wait for all servers to shut down
	var wg sync.WaitGroup
	wg.Add(3)

	// Start HTTP server
	go func() {
		defer wg.Done()
		fmt.Println("Starting HTTP server on :8080")
		if err := httpServer.Start(); err != nil {
			log.Printf("HTTP server error: %v", err)
		}
	}()

	// Start WebSocket server
	go func() {
		defer wg.Done()
		fmt.Println("Starting WebSocket server on :8081")
		if err := wsServer.Start(); err != nil {
			log.Printf("WebSocket server error: %v", err)
		}
	}()

	// Start stdio server
	go func() {
		defer wg.Done()
		fmt.Println("Starting stdio server")
		if err := stdioServer.Start(); err != nil {
			log.Printf("Stdio server error: %v", err)
		}
	}()

	// Handle interrupt signals
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM)
	<-signalChan

	fmt.Println("Shutting down servers...")
	httpServer.Stop()
	wsServer.Stop()
	stdioServer.Stop()

	// Wait for servers to shut down
	wg.Wait()
	fmt.Println("All servers shut down")
}