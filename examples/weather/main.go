package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/virgoC0der/go-mcp"
	"github.com/virgoC0der/go-mcp/internal/types"
)

// WeatherServer implements a weather service
type WeatherServer struct {
	prompts   []types.Prompt
	tools     []types.Tool
	resources []types.Resource
	apiKey    string
}

// NewWeatherServer creates a new weather server instance
func NewWeatherServer(apiKey string) *WeatherServer {
	s := &WeatherServer{
		apiKey: apiKey,
		prompts: []types.Prompt{
			{
				Name:        "weather",
				Description: "Get weather information for a city",
				Arguments: []types.PromptArgument{
					{
						Name:        "city",
						Description: "City name",
						Required:    true,
					},
				},
			},
		},
		tools: []types.Tool{
			{
				Name:        "getWeather",
				Description: "Get weather information for a city",
				InputSchema: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"city": map[string]interface{}{
							"type":        "string",
							"description": "City name",
						},
					},
					"required": []string{"city"},
				},
			},
		},
		resources: []types.Resource{
			{
				URI:         "cities",
				Name:        "cities",
				Description: "List of supported cities",
				MimeType:    "application/json",
			},
		},
	}
	return s
}

// Initialize implements the Server interface
func (s *WeatherServer) Initialize(ctx context.Context, options any) error {
	if s.apiKey == "" {
		return fmt.Errorf("API key is required")
	}
	return nil
}

// Start implements the Server interface
func (s *WeatherServer) Start() error {
	return nil
}

// Shutdown implements the Server interface
func (s *WeatherServer) Shutdown(ctx context.Context) error {
	return nil
}

// ListPrompts implements the Server interface
func (s *WeatherServer) ListPrompts(ctx context.Context, cursor string) (*types.PromptListResult, error) {
	return &types.PromptListResult{
		Prompts:    s.prompts,
		NextCursor: "",
	}, nil
}

// GetPrompt implements the Server interface
func (s *WeatherServer) GetPrompt(ctx context.Context, name string, args map[string]any) (*types.PromptResult, error) {
	if name != "weather" {
		return nil, fmt.Errorf("unknown prompt: %s", name)
	}

	city, ok := args["city"].(string)
	if !ok {
		return nil, fmt.Errorf("missing required argument: city")
	}

	// Get weather data using the tool
	result, err := s.CallTool(ctx, "getWeather", map[string]any{"city": city})
	if err != nil {
		return nil, err
	}

	// Get the text content from the tool result
	if len(result.Content) == 0 || result.Content[0].Type != "text" {
		return nil, fmt.Errorf("invalid weather data format")
	}

	// Use the text content directly
	responseText := result.Content[0].Text

	return &types.PromptResult{
		Description: "Weather information",
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
func (s *WeatherServer) ListTools(ctx context.Context, cursor string) (*types.ToolListResult, error) {
	return &types.ToolListResult{
		Tools:      s.tools,
		NextCursor: "",
	}, nil
}

// CallTool implements the Server interface
func (s *WeatherServer) CallTool(ctx context.Context, name string, args map[string]any) (*types.CallToolResult, error) {
	if name != "getWeather" {
		return nil, fmt.Errorf("unknown tool: %s", name)
	}

	city, ok := args["city"].(string)
	if !ok {
		return nil, fmt.Errorf("missing or invalid argument: city")
	}

	// Call OpenWeatherMap API
	url := fmt.Sprintf(
		"http://api.openweathermap.org/data/2.5/weather?q=%s&appid=%s&units=metric",
		city, s.apiKey,
	)

	resp, err := http.Get(url)
	if err != nil {
		// 返回错误结果
		return &types.CallToolResult{
			Content: []types.ToolContent{
				{
					Type: "text",
					Text: fmt.Sprintf("Failed to get weather data: %v", err),
				},
			},
			IsError: true,
		}, nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		// 返回错误结果
		return &types.CallToolResult{
			Content: []types.ToolContent{
				{
					Type: "text",
					Text: fmt.Sprintf("Weather API error: %s", resp.Status),
				},
			},
			IsError: true,
		}, nil
	}

	var data struct {
		Main struct {
			Temp float64 `json:"temp"`
		} `json:"main"`
		Weather []struct {
			Description string `json:"description"`
		} `json:"weather"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		// 返回错误结果
		return &types.CallToolResult{
			Content: []types.ToolContent{
				{
					Type: "text",
					Text: fmt.Sprintf("Failed to parse weather data: %v", err),
				},
			},
			IsError: true,
		}, nil
	}

	return &types.CallToolResult{
		Content: []types.ToolContent{
			{
				Type: "text",
				Text: fmt.Sprintf("Current weather in %s:\nTemperature: %.1f°C\nConditions: %s",
					city, data.Main.Temp, data.Weather[0].Description),
			},
		},
		IsError: false,
	}, nil
}

// ListResources implements the Server interface
func (s *WeatherServer) ListResources(ctx context.Context, cursor string) (*types.ResourceListResult, error) {
	return &types.ResourceListResult{
		Resources:  s.resources,
		NextCursor: "",
	}, nil
}

// ReadResource implements the Server interface
func (s *WeatherServer) ReadResource(ctx context.Context, uri string) (*types.ResourceContent, error) {
	if uri != "cities" {
		return nil, fmt.Errorf("unknown resource: %s", uri)
	}

	cities := []string{
		"London",
		"New York",
		"Tokyo",
		"Paris",
		"Beijing",
	}

	content, err := json.Marshal(cities)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal cities: %w", err)
	}

	return &types.ResourceContent{
		URI:      uri,
		MimeType: "application/json",
		Text:     string(content),
	}, nil
}

// ListResourceTemplates implements the Server interface
func (s *WeatherServer) ListResourceTemplates(ctx context.Context) ([]types.ResourceTemplate, error) {
	return []types.ResourceTemplate{}, nil
}

// SubscribeToResource implements the Server interface
func (s *WeatherServer) SubscribeToResource(ctx context.Context, uri string) error {
	return fmt.Errorf("subscription not supported")
}

func main() {
	apiKey := os.Getenv("OPENWEATHERMAP_API_KEY")
	if apiKey == "" {
		log.Fatal("OPENWEATHERMAP_API_KEY environment variable is required")
	}

	// Create weather service
	service := NewWeatherServer(apiKey)

	// Create HTTP server with capabilities
	httpServer, err := mcp.NewServer(service, &types.ServerOptions{
		Address: ":8080",
		Capabilities: &types.ServerCapabilities{
			Tools: &types.ToolCapabilities{
				ListChanged: true,
			},
			Prompts: &types.PromptCapabilities{
				ListChanged: true,
			},
			Resources: &types.ResourceCapabilities{
				ListChanged: true,
			},
		},
	})
	if err != nil {
		log.Fatal(err)
	}

	// Initialize servers
	ctx := context.Background()
	if err := httpServer.Initialize(ctx, nil); err != nil {
		log.Fatal(err)
	}

	// Create context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start servers in goroutines
	go func() {
		log.Printf("Starting HTTP server on :8080")
		if err := httpServer.Start(); err != nil {
			log.Printf("HTTP server error: %v", err)
			cancel()
		}
	}()

	// Wait for interrupt signal
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	select {
	case <-sigCh:
		log.Println("Received shutdown signal")
	case <-ctx.Done():
		log.Println("Server error occurred")
	}

	// Graceful shutdown
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		log.Printf("Error shutting down HTTP server: %v", err)
	}

	log.Println("Servers stopped")
}
