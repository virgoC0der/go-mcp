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
				Arguments: []types.PromptArgument{
					{
						Name:        "message",
						Description: "Message to echo",
						Required:    true,
					},
				},
			},
		},
		tools: []types.Tool{
			{
				Name:        "echo",
				Description: "Echo a message",
				InputSchema: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"message": map[string]interface{}{
							"type":        "string",
							"description": "Message to echo",
						},
					},
					"required": []string{"message"},
				},
			},
		},
		resources: []types.Resource{
			{
				URI:         "echo",
				Name:        "echo",
				Description: "Echo resource",
				MimeType:    "text/plain",
			},
		},
	}
	return s
}

// Initialize implements the Server interface
func (s *EchoServer) Initialize(ctx context.Context, options any) error {
	return nil
}

// Start implements the Server interface
func (s *EchoServer) Start() error {
	return nil
}

// Shutdown implements the Server interface
func (s *EchoServer) Shutdown(ctx context.Context) error {
	return nil
}

// ListPrompts implements the Server interface
func (s *EchoServer) ListPrompts(ctx context.Context, cursor string) (*types.PromptListResult, error) {
	return &types.PromptListResult{
		Prompts:    s.prompts,
		NextCursor: "",
	}, nil
}

// GetPrompt implements the Server interface
func (s *EchoServer) GetPrompt(ctx context.Context, name string, args map[string]any) (*types.PromptResult, error) {
	if name != "echo" {
		return nil, fmt.Errorf("unknown prompt: %s", name)
	}

	message, ok := args["message"].(string)
	if !ok {
		return nil, fmt.Errorf("missing required argument: message")
	}

	// Create response with the expected Message format
	return &types.PromptResult{
		Description: "Echo response",
		Messages: []types.Message{
			{
				Role: "assistant",
				Content: types.Content{
					Type: "text",
					Text: fmt.Sprintf("Echo: %s", message),
				},
			},
		},
	}, nil
}

// ListTools implements the Server interface
func (s *EchoServer) ListTools(ctx context.Context, cursor string) (*types.ToolListResult, error) {
	return &types.ToolListResult{
		Tools:      s.tools,
		NextCursor: "",
	}, nil
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
		Content: []types.ToolContent{
			{
				Type: "text",
				Text: fmt.Sprintf("Echo: %s", message),
			},
		},
		IsError: false,
	}, nil
}

// ListResources implements the Server interface
func (s *EchoServer) ListResources(ctx context.Context, cursor string) (*types.ResourceListResult, error) {
	return &types.ResourceListResult{
		Resources:  s.resources,
		NextCursor: "",
	}, nil
}

// ReadResource implements the Server interface
func (s *EchoServer) ReadResource(ctx context.Context, uri string) (*types.ResourceContent, error) {
	if uri != "echo" {
		return nil, fmt.Errorf("unknown resource: %s", uri)
	}

	return &types.ResourceContent{
		URI:      uri,
		MimeType: "text/plain",
		Text:     "This is an echo resource",
	}, nil
}

// ListResourceTemplates implements the Server interface
func (s *EchoServer) ListResourceTemplates(ctx context.Context) ([]types.ResourceTemplate, error) {
	return []types.ResourceTemplate{}, nil
}

// SubscribeToResource implements the Server interface
func (s *EchoServer) SubscribeToResource(ctx context.Context, uri string) error {
	return fmt.Errorf("subscription not supported")
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
