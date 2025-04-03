package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/virgoC0der/go-mcp"
	"github.com/virgoC0der/go-mcp/internal/types"
)

// EchoServer implements a simple echo server for demonstration
type EchoServer struct {
	prompts   []types.Prompt
	tools     []types.Tool
	resources []types.Resource
}

// NewEchoServer creates a new echo server instance
func NewEchoServer() *EchoServer {
	s := &EchoServer{
		prompts: []types.Prompt{
			{
				Name:        "echo",
				Description: "Echo a message",
				Template:    "Echo: {{.message}}",
				Metadata: map[string]interface{}{
					"required": []string{"message"},
				},
			},
		},
		tools: []types.Tool{
			{
				Name:        "echo",
				Description: "Echo a message",
				Parameters: map[string]interface{}{
					"message": map[string]interface{}{
						"type":        "string",
						"description": "Message to echo",
					},
				},
			},
		},
		resources: []types.Resource{
			{
				Name:        "echo",
				Type:        "text/plain",
				Description: "Echo resource",
			},
		},
	}
	return s
}

// Initialize implements the Server interface
func (s *EchoServer) Initialize(ctx context.Context, options any) error {
	return nil
}

// ListPrompts implements the Server interface
func (s *EchoServer) ListPrompts(ctx context.Context) ([]types.Prompt, error) {
	return s.prompts, nil
}

// GetPrompt implements the Server interface
func (s *EchoServer) GetPrompt(ctx context.Context, name string, args map[string]any) (*types.GetPromptResult, error) {
	if name != "echo" {
		return nil, fmt.Errorf("unknown prompt: %s", name)
	}

	message, ok := args["message"].(string)
	if !ok {
		return nil, fmt.Errorf("missing required argument: message")
	}

	return &types.GetPromptResult{
		Content: fmt.Sprintf("Echo: %s", message),
	}, nil
}

// ListTools implements the Server interface
func (s *EchoServer) ListTools(ctx context.Context) ([]types.Tool, error) {
	return s.tools, nil
}

// CallTool implements the Server interface
func (s *EchoServer) CallTool(ctx context.Context, name string, args map[string]any) (*types.CallToolResult, error) {
	if name != "echo" {
		return nil, fmt.Errorf("unknown tool: %s", name)
	}

	message, ok := args["message"].(string)
	if !ok {
		return nil, fmt.Errorf("missing or invalid argument: message")
	}

	return &types.CallToolResult{
		Output: fmt.Sprintf("Echo: %s", message),
	}, nil
}

// ListResources implements the Server interface
func (s *EchoServer) ListResources(ctx context.Context) ([]types.Resource, error) {
	return s.resources, nil
}

// ReadResource implements the Server interface
func (s *EchoServer) ReadResource(ctx context.Context, name string) ([]byte, string, error) {
	if name != "echo" {
		return nil, "", fmt.Errorf("unknown resource: %s", name)
	}

	content := []byte("This is an echo resource")
	return content, "text/plain", nil
}

func main() {
	srv := NewEchoServer()

	// Create HTTP server
	httpServer, err := mcp.NewServer(srv, &types.ServerOptions{
		Address: ":8080",
	})
	if err != nil {
		log.Fatalf("Failed to create HTTP server: %v", err)
	}

	// Create WebSocket server
	wsServer, err := mcp.NewServer(srv, &types.ServerOptions{
		Address: ":8081",
	})
	if err != nil {
		log.Fatalf("Failed to create WebSocket server: %v", err)
	}

	// Create context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Initialize servers
	if err := httpServer.Initialize(ctx, nil); err != nil {
		log.Fatalf("Failed to initialize HTTP server: %v", err)
	}
	if err := wsServer.Initialize(ctx, nil); err != nil {
		log.Fatalf("Failed to initialize WebSocket server: %v", err)
	}

	// Start servers
	go func() {
		log.Printf("Starting HTTP server on :8080")
		if err := httpServer.Start(); err != nil {
			log.Printf("HTTP server error: %v", err)
			cancel()
		}
	}()

	go func() {
		log.Printf("Starting WebSocket server on :8081")
		if err := wsServer.Start(); err != nil {
			log.Printf("WebSocket server error: %v", err)
			cancel()
		}
	}()

	// Handle interrupt signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	select {
	case <-sigChan:
		log.Println("Received shutdown signal")
	case <-ctx.Done():
		log.Println("Server error occurred")
	}

	// Graceful shutdown
	log.Println("Shutting down servers...")
	if err := httpServer.Shutdown(ctx); err != nil {
		log.Printf("HTTP server shutdown error: %v", err)
	}
	if err := wsServer.Shutdown(ctx); err != nil {
		log.Printf("WebSocket server shutdown error: %v", err)
	}
	log.Println("Servers stopped")
}
