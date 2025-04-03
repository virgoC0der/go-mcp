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
				Template:    "The weather in {{.city}} is {{.temperature}}°C with {{.description}}",
				Metadata: map[string]interface{}{
					"required": []string{"city"},
				},
			},
		},
		tools: []types.Tool{
			{
				Name:        "getWeather",
				Description: "Get weather information for a city",
				Parameters: map[string]interface{}{
					"city": map[string]interface{}{
						"type":        "string",
						"description": "City name",
					},
				},
			},
		},
		resources: []types.Resource{
			{
				Name:        "cities",
				Type:        "application/json",
				Description: "List of supported cities",
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

// ListPrompts implements the Server interface
func (s *WeatherServer) ListPrompts(ctx context.Context) ([]types.Prompt, error) {
	return s.prompts, nil
}

// GetPrompt implements the Server interface
func (s *WeatherServer) GetPrompt(ctx context.Context, name string, args map[string]any) (*types.GetPromptResult, error) {
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

	// Parse weather data
	data, ok := result.Output.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid weather data format")
	}

	return &types.GetPromptResult{
		Content: fmt.Sprintf(
			"The weather in %s is %.1f°C with %s",
			city,
			data["temperature"].(float64),
			data["description"].(string),
		),
	}, nil
}

// ListTools implements the Server interface
func (s *WeatherServer) ListTools(ctx context.Context) ([]types.Tool, error) {
	return s.tools, nil
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
		return nil, fmt.Errorf("failed to get weather data: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("weather API error: %s", resp.Status)
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
		return nil, fmt.Errorf("failed to parse weather data: %w", err)
	}

	return &types.CallToolResult{
		Output: map[string]interface{}{
			"temperature": data.Main.Temp,
			"description": data.Weather[0].Description,
		},
	}, nil
}

// ListResources implements the Server interface
func (s *WeatherServer) ListResources(ctx context.Context) ([]types.Resource, error) {
	return s.resources, nil
}

// ReadResource implements the Server interface
func (s *WeatherServer) ReadResource(ctx context.Context, name string) ([]byte, string, error) {
	if name != "cities" {
		return nil, "", fmt.Errorf("unknown resource: %s", name)
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
		return nil, "", fmt.Errorf("failed to marshal cities: %w", err)
	}

	return content, "application/json", nil
}

func main() {
	apiKey := os.Getenv("OPENWEATHERMAP_API_KEY")
	if apiKey == "" {
		log.Fatal("OPENWEATHERMAP_API_KEY environment variable is required")
	}

	// Create weather service
	service := NewWeatherServer(apiKey)

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
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5)
	defer cancel()

	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		log.Printf("Error shutting down HTTP server: %v", err)
	}
	if err := wsServer.Shutdown(shutdownCtx); err != nil {
		log.Printf("Error shutting down WebSocket server: %v", err)
	}
}
