package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os/exec"
	"time"

	"github.com/virgoC0der/go-mcp/client"
)

// Test both HTTP and WebSocket endpoints
func main() {
	// Test the typed HTTP client
	fmt.Println("Testing HTTP client:")
	testHTTPClient()

	fmt.Println("\nTesting WebSocket client:")
	testWebSocketClient()

	fmt.Println("\nTesting stdio client:")
	testStdioClient()
}

func testHTTPClient() {
	// Create an HTTP client
	c := client.NewHTTPClient("http://localhost:8080")

	// List prompts
	prompts, err := c.ListPrompts(context.Background())
	if err != nil {
		log.Printf("Error listing prompts: %v", err)
		return
	}

	fmt.Println("Available prompts:")
	for _, p := range prompts {
		fmt.Printf("- %s: %s\n", p.Name, p.Description)
	}

	// Call a prompt
	greetArgs := map[string]any{
		"name":     "World",
		"formal":   true,
		"language": "fr",
	}

	promptResult, err := c.GetPrompt(context.Background(), "greet", greetArgs)
	if err != nil {
		log.Printf("Error calling prompt: %v", err)
		return
	}

	fmt.Printf("Prompt result: %s\n", promptResult.Message)

	// List tools
	tools, err := c.ListTools(context.Background())
	if err != nil {
		log.Printf("Error listing tools: %v", err)
		return
	}

	fmt.Println("Available tools:")
	for _, t := range tools {
		fmt.Printf("- %s: %s\n", t.Name, t.Description)
	}

	// Call a tool
	calcArgs := map[string]any{
		"operation": "multiply",
		"a":         10.5,
		"b":         2.0,
	}

	toolResult, err := c.CallTool(context.Background(), "calculator", calcArgs)
	if err != nil {
		log.Printf("Error calling tool: %v", err)
		return
	}

	// Pretty print the result
	resultBytes, _ := json.MarshalIndent(toolResult.Content, "", "  ")
	fmt.Printf("Tool result: %s\n", string(resultBytes))

	// List resources
	resources, err := c.ListResources(context.Background())
	if err != nil {
		log.Printf("Error listing resources: %v", err)
		return
	}

	fmt.Println("Available resources:")
	for _, r := range resources {
		fmt.Printf("- %s: %s (%s)\n", r.Name, r.Description, r.MimeType)
	}

	// Read a resource
	content, mimeType, err := c.ReadResource(context.Background(), "help")
	if err != nil {
		log.Printf("Error reading resource: %v", err)
		return
	}

	fmt.Printf("Resource content (%s):\n%s\n", mimeType, string(content))
}

func testWebSocketClient() {
	// Create a WebSocket client
	c, err := client.NewWebSocketClient("ws://localhost:8081/ws")
	if err != nil {
		log.Printf("Error creating WebSocket client: %v", err)
		return
	}

	// Initialize the client
	if err := c.Initialize(context.Background()); err != nil {
		log.Printf("Error initializing client: %v", err)
		return
	}

	// List tools
	tools, err := c.ListTools(context.Background())
	if err != nil {
		log.Printf("Error listing tools: %v", err)
		return
	}

	fmt.Println("Available tools (WebSocket):")
	for _, t := range tools {
		fmt.Printf("- %s: %s\n", t.Name, t.Description)
	}

	// Call a tool
	calcArgs := map[string]any{
		"operation": "add",
		"a":         42,
		"b":         58,
	}

	toolResult, err := c.CallTool(context.Background(), "calculator", calcArgs)
	if err != nil {
		log.Printf("Error calling tool: %v", err)
		return
	}

	// Pretty print the result
	resultBytes, _ := json.MarshalIndent(toolResult.Content, "", "  ")
	fmt.Printf("Tool result (WebSocket): %s\n", string(resultBytes))

	// Close the client
	c.Close()
}

func testStdioClient() {
	// Start the server as a subprocess
	cmd := exec.Command("go", "run", "../main.go")

	// Get stdin and stdout pipes
	stdin, err := cmd.StdinPipe()
	if err != nil {
		log.Fatalf("Failed to get stdin pipe: %v", err)
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Fatalf("Failed to get stdout pipe: %v", err)
	}

	// Start the server
	if err := cmd.Start(); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}

	// Ensure the server is terminated when done
	defer cmd.Process.Kill()

	// Create a stdio client
	c := client.NewStdioClient(stdout, stdin)

	// Wait a moment for the server to start
	time.Sleep(500 * time.Millisecond)

	// Initialize the client
	if err := c.Initialize(context.Background()); err != nil {
		log.Printf("Error initializing stdio client: %v", err)
		return
	}

	// List prompts
	prompts, err := c.ListPrompts(context.Background())
	if err != nil {
		log.Printf("Error listing prompts (stdio): %v", err)
		return
	}

	fmt.Println("Available prompts (stdio):")
	for _, p := range prompts {
		fmt.Printf("- %s: %s\n", p.Name, p.Description)
	}

	// Call a prompt
	greetArgs := map[string]any{
		"name":     "Stdio",
		"language": "en",
		"formal":   false,
	}

	promptResult, err := c.GetPrompt(context.Background(), "greet", greetArgs)
	if err != nil {
		log.Printf("Error calling prompt (stdio): %v", err)
		return
	}

	fmt.Printf("Prompt result (stdio): %s\n", promptResult.Message)
}
