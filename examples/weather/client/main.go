package main

import (
	"context"
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

	fmt.Println("Available tools list:")
	for _, t := range tools.Tools {
		fmt.Printf("- %s: %s\n", t.Name, t.Description)
	}

	// Prepare parameters for tool call
	// Default to city name "Beijing"
	city := "Beijing"

	// If city name is provided in command line, use it instead
	if len(os.Args) >= 2 {
		city = os.Args[1]
	}

	args := map[string]any{
		"city": city,
	}

	fmt.Printf("Querying weather information for city: %s...\n", city)

	// Call weather tool
	result, err := c.CallTool(context.Background(), "getWeather", args)
	if err != nil {
		log.Fatalf("Failed to call weather tool: %v", err)
	}

	// Display the result
	fmt.Println("Weather information:")

	// Check if there was an error
	if result.IsError {
		fmt.Println("Error occurred:")
		for _, content := range result.Content {
			if content.Type == "text" {
				fmt.Println(content.Text)
			}
		}
		return
	}

	// Display the content
	for _, content := range result.Content {
		switch content.Type {
		case "text":
			fmt.Println(content.Text)
		case "image":
			fmt.Println("[Image data available]")
		case "audio":
			fmt.Println("[Audio data available]")
		case "resource":
			if content.Resource != nil {
				fmt.Printf("[Resource available: %s]\n", content.Resource.URI)
			}
		default:
			fmt.Printf("[Unknown content type: %s]\n", content.Type)
		}
	}
}
