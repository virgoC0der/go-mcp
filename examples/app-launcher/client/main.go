package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/virgoC0der/go-mcp/transport/http"
)

func main() {
	// Create HTTP client
	c := http.NewHTTPClient("http://localhost:8080")

	// List available tools
	tools, err := c.ListTools(context.Background(), "")
	if err != nil {
		log.Fatalf("Failed to get tool list: %v", err)
	}

	fmt.Println("Available tools:")
	for _, t := range tools.Tools {
		fmt.Printf("- %s: %s\n", t.Name, t.Description)
	}

	// Get list of available apps
	resource, err := c.ReadResource(context.Background(), "apps")
	if err != nil {
		log.Fatalf("Failed to get apps list: %v", err)
	}

	var apps []string
	if err := json.Unmarshal([]byte(resource.Text), &apps); err != nil {
		log.Fatalf("Failed to parse apps list: %v", err)
	}

	fmt.Println("\nAvailable apps:")
	for _, app := range apps {
		fmt.Printf("- %s\n", app)
	}

	// Prepare parameters for tool call
	// Default to app name "Calculator"
	appName := "Calculator"

	// If app name is provided in command line, use it instead
	if len(os.Args) >= 2 {
		appName = os.Args[1]
	}

	fmt.Printf("\nAttempting to open app: %s\n", appName)

	args := map[string]any{
		"appName": appName,
	}

	// Call the openApp tool
	result, err := c.CallTool(context.Background(), "openApp", args)
	if err != nil {
		log.Fatalf("Failed to call tool: %v", err)
	}

	// Print the result
	if result.IsError {
		fmt.Printf("Error: %s\n", result.Content[0].Text)
	} else {
		fmt.Printf("Success: %s\n", result.Content[0].Text)
	}
}
