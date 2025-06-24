package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"sync"
	"syscall"

	"github.com/virgoC0der/go-mcp"
	"github.com/virgoC0der/go-mcp/internal/types"
)

// AppLauncherServer implements the Server interface
type AppLauncherServer struct {
	prompts   []types.Prompt
	tools     []types.Tool
	resources []types.Resource
}

// Initialize implements the Server interface
func (s *AppLauncherServer) Initialize(ctx context.Context, options any) error {
	// Check if running on macOS
	if runtime.GOOS != "darwin" {
		return fmt.Errorf("this server is designed to run on macOS only, current OS: %s", runtime.GOOS)
	}
	return nil
}

// Start implements Server interface
func (s *AppLauncherServer) Start() error {
	return nil
}

// Shutdown implements the Server interface
func (s *AppLauncherServer) Shutdown(ctx context.Context) error {
	return nil
}

// NewAppLauncherServer creates a new app launcher server instance
func NewAppLauncherServer() *AppLauncherServer {
	s := &AppLauncherServer{
		prompts: []types.Prompt{
			{
				Name:        "openApp",
				Description: "Open a macOS application",
				Arguments: []types.PromptArgument{
					{
						Name:        "appName",
						Description: "Name of the application to open",
						Required:    true,
					},
				},
			},
			{
				Name:        "closeApp",
				Description: "Close a macOS application",
				Arguments: []types.PromptArgument{
					{
						Name:        "appName",
						Description: "Name of the application to close",
						Required:    true,
					},
				},
			},
		},
		tools: []types.Tool{
			{
				Name:        "openApp",
				Description: "Open a macOS application",
				InputSchema: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"appName": map[string]interface{}{
							"type":        "string",
							"description": "Name of the application to open",
						},
					},
					"required": []string{"appName"},
				},
			},
			{
				Name:        "closeApp",
				Description: "Close a macOS application",
				InputSchema: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"appName": map[string]interface{}{
							"type":        "string",
							"description": "Name of the application to close",
						},
					},
					"required": []string{"appName"},
				},
			},
		},
		resources: []types.Resource{
			{
				URI:         "apps",
				Name:        "apps",
				Description: "List of common macOS applications",
				MimeType:    "application/json",
			},
		},
	}
	return s
}

// ListPrompts implements the Server interface
func (s *AppLauncherServer) ListPrompts(ctx context.Context, cursor string) (*types.PromptListResult, error) {
	return &types.PromptListResult{
		Prompts:    s.prompts,
		NextCursor: "",
	}, nil
}

// GetPrompt implements the Server interface
func (s *AppLauncherServer) GetPrompt(ctx context.Context, name string, args map[string]any) (*types.PromptResult, error) {
	if name != "openApp" && name != "closeApp" {
		return nil, fmt.Errorf("unknown prompt: %s", name)
	}

	appName, ok := args["appName"].(string)
	if !ok || appName == "" {
		return nil, fmt.Errorf("missing or invalid argument: appName")
	}

	// Call the corresponding tool
	result, err := s.CallTool(ctx, name, args)
	if err != nil {
		return nil, err
	}

	// Create a response message
	var responseText string
	if result.IsError {
		if name == "openApp" {
			responseText = fmt.Sprintf("Failed to open application '%s': %s", appName, result.Content[0].Text)
		} else {
			responseText = fmt.Sprintf("Failed to close application '%s': %s", appName, result.Content[0].Text)
		}
	} else {
		responseText = result.Content[0].Text
	}

	return &types.PromptResult{
		Description: "Application launcher result",
		Messages: []types.Message{
			{
				Role: "assistant",
				Content: types.Content{
					Type: "text",
					Text: responseText,
				},
			},
		},
	}, nil
}

// ListTools implements the Server interface
func (s *AppLauncherServer) ListTools(ctx context.Context, cursor string) (*types.ToolListResult, error) {
	return &types.ToolListResult{
		Tools:      s.tools,
		NextCursor: "",
	}, nil
}

// CallTool implements the Server interface
func (s *AppLauncherServer) CallTool(ctx context.Context, name string, args map[string]any) (*types.CallToolResult, error) {
	if name != "openApp" && name != "closeApp" {
		return nil, fmt.Errorf("unknown tool: %s", name)
	}

	appName, ok := args["appName"].(string)
	if !ok || appName == "" {
		return nil, fmt.Errorf("missing or invalid argument: appName")
	}

	var cmd *exec.Cmd
	var actionText string

	if name == "openApp" {
		// Use the 'open' command to open the application
		cmd = exec.Command("open", "-a", appName)
		actionText = "opened"
	} else { // closeApp
		// Use AppleScript to quit the application
		script := fmt.Sprintf("tell application \"%s\" to quit", appName)
		cmd = exec.Command("osascript", "-e", script)
		actionText = "closed"
	}

	err := cmd.Run()

	if err != nil {
		// Return error result
		var errorMsg string
		if name == "openApp" {
			errorMsg = fmt.Sprintf("Error opening application: %v", err)
		} else {
			errorMsg = fmt.Sprintf("Error closing application: %v", err)
		}

		return &types.CallToolResult{
			Content: []types.ToolContent{
				{
					Type: "text",
					Text: errorMsg,
				},
			},
			IsError: true,
		}, nil
	}

	// Return success result
	return &types.CallToolResult{
		Content: []types.ToolContent{
			{
				Type: "text",
				Text: fmt.Sprintf("Successfully %s application: %s", actionText, appName),
			},
		},
		IsError: false,
	}, nil
}

// ListResources implements the Server interface
func (s *AppLauncherServer) ListResources(ctx context.Context, cursor string) (*types.ResourceListResult, error) {
	return &types.ResourceListResult{
		Resources:  s.resources,
		NextCursor: "",
	}, nil
}

// ReadResource implements the Server interface
func (s *AppLauncherServer) ReadResource(ctx context.Context, uri string) (*types.ResourceContent, error) {
	if uri != "apps" {
		return nil, fmt.Errorf("unknown resource: %s", uri)
	}

	// List of common macOS applications
	apps := []string{
		"Safari",
		"Mail",
		"Calendar",
		"Notes",
		"Maps",
		"Photos",
		"Messages",
		"FaceTime",
		"Music",
		"App Store",
		"System Settings",
		"Terminal",
		"Calculator",
		"TextEdit",
		"Preview",
		"GoLand",
		"Microsoft Edge",
		"Cursor",
		"Warp",
		"iTerm 2",
	}

	content, err := json.Marshal(apps)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal apps list: %w", err)
	}

	return &types.ResourceContent{
		URI:      uri,
		MimeType: "application/json",
		Text:     string(content),
	}, nil
}

// ListResourceTemplates implements the Server interface
func (s *AppLauncherServer) ListResourceTemplates(ctx context.Context) ([]types.ResourceTemplate, error) {
	return []types.ResourceTemplate{}, nil
}

// SubscribeToResource implements the Server interface
func (s *AppLauncherServer) SubscribeToResource(ctx context.Context, uri string) error {
	return fmt.Errorf("resource subscription not supported")
}

func main() {
	// Create app launcher service
	service := NewAppLauncherServer()

	// Create context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 用工厂方法创建 server
	server, err := mcp.NewServer(service, nil)
	if err != nil {
		log.Fatalf("Failed to create server: %v", err)
	}

	// 用 WaitGroup 管理 goroutine
	wg := sync.WaitGroup{}
	wg.Add(1)

	// 启动 server
	go func() {
		defer wg.Done()
		log.Printf("Starting server")
		if err := server.Start(); err != nil {
			log.Printf("Server error: %v", err)
			cancel()
		}
	}()

	// Handle interrupt signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Wait for shutdown signal or error
	shutdownCh := make(chan struct{})
	go func() {
		select {
		case <-sigChan:
			log.Println("Received shutdown signal")
		case <-ctx.Done():
			log.Println("Server error occurred")
		}
		close(shutdownCh)
	}()

	<-shutdownCh

	// Graceful shutdown
	log.Println("Shutting down servers...")

	// 用 Shutdown(ctx) 替代 Stop()
	if err := server.Shutdown(ctx); err != nil {
		log.Printf("Server shutdown error: %v", err)
	}

	// Wait for servers to finish
	wg.Wait()
	log.Println("All servers stopped")
}
