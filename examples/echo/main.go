package main

import (
	"context"
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

// EchoServer implements a simple echo server for demonstration
type EchoServer struct {
	*server.BaseServer
}

// NewEchoServer creates a new echo server instance
func NewEchoServer() *EchoServer {
	s := &EchoServer{
		BaseServer: server.NewBaseServer("echo", "1.0.0"),
	}

	// Register echo prompt
	s.RegisterPrompt(types.Prompt{
		Name:        "echo",
		Description: "Echo a message",
		Arguments: []types.PromptArgument{
			{
				Name:        "message",
				Description: "Message to echo",
				Required:    true,
			},
		},
	})

	// Register echo tool
	s.RegisterTool(types.Tool{
		Name:        "echo",
		Description: "Echo a message",
	})

	// Register echo resource
	s.RegisterResource(types.Resource{
		Name:        "echo",
		Description: "Echo resource",
		MimeType:    "text/plain",
	})

	return s
}

// GetPrompt implements the prompt handler for the echo server
func (s *EchoServer) GetPrompt(ctx context.Context, name string, arguments map[string]string) (*types.GetPromptResult, error) {
	if name != "echo" {
		return nil, fmt.Errorf("unknown prompt: %s", name)
	}

	message, ok := arguments["message"]
	if !ok {
		return nil, fmt.Errorf("missing required argument: message")
	}

	return &types.GetPromptResult{
		Description: "Echo prompt",
		Messages: []types.Message{
			{
				Role: "user",
				Content: types.Content{
					Type: "text",
					Text: fmt.Sprintf("Echo: %s", message),
				},
			},
		},
	}, nil
}

// CallTool implements the tool handler for the echo server
func (s *EchoServer) CallTool(ctx context.Context, name string, arguments map[string]interface{}) (interface{}, error) {
	if name != "echo" {
		return nil, fmt.Errorf("unknown tool: %s", name)
	}

	message, ok := arguments["message"].(string)
	if !ok {
		return nil, fmt.Errorf("missing or invalid argument: message")
	}

	return fmt.Sprintf("Echo: %s", message), nil
}

// ReadResource implements the resource handler for the echo server
func (s *EchoServer) ReadResource(ctx context.Context, name string) ([]byte, string, error) {
	if name != "echo" {
		return nil, "", fmt.Errorf("unknown resource: %s", name)
	}

	content := []byte("This is an echo resource")
	return content, "text/plain", nil
}

func main() {
	srv := NewEchoServer()

	// Initialize the server
	err := srv.Initialize(context.Background(), types.InitializationOptions{
		ServerName:    "echo",
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

	// Create HTTP and WebSocket servers
	httpServer := transport.NewHTTPServer(srv, ":8080")
	wsServer := transport.NewWSServer(srv, ":8081")

	// Use WaitGroup to wait for all servers to shut down
	var wg sync.WaitGroup
	wg.Add(2)

	// Start HTTP server
	go func() {
		defer wg.Done()
		log.Printf("Starting HTTP server on :8080")
		if err := httpServer.Start(); err != nil {
			log.Printf("HTTP server error: %v", err)
		}
	}()

	// Start WebSocket server
	go func() {
		defer wg.Done()
		log.Printf("Starting WebSocket server on :8081")
		if err := wsServer.Start(); err != nil {
			log.Printf("WebSocket server error: %v", err)
		}
	}()

	// Handle interrupt signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	log.Println("Shutting down servers...")

	// Wait for servers to shut down
	wg.Wait()
	log.Println("Servers stopped")
}
