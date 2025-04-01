package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/virgoC0der/go-mcp/client"
)

func main() {
	// Create HTTP client
	c := client.NewHTTPClient("http://localhost:8080")

	// Initialize connection
	err := c.Initialize(context.Background())
	if err != nil {
		log.Fatalf("Failed to initialize client: %v", err)
	}

	// List available tools
	tools, err := c.ListTools(context.Background())
	if err != nil {
		log.Fatalf("Failed to get tool list: %v", err)
	}

	fmt.Println("Available tools list:")
	for _, t := range tools {
		fmt.Printf("- %s: %s\n", t.Name, t.Description)
	}

	// Prepare parameters for tool call
	// Default to Beijing coordinates: longitude 116.407526, latitude 39.90403
	longitude := 116.407526
	latitude := 39.90403

	// If coordinates are provided in command line, use those instead
	if len(os.Args) >= 3 {
		fmt.Sscanf(os.Args[1], "%f", &longitude)
		fmt.Sscanf(os.Args[2], "%f", &latitude)
	}

	args := map[string]any{
		"longitude": longitude,
		"latitude":  latitude,
		"language":  "zh_CN",
	}

	fmt.Printf("Querying weather information for location (longitude: %.6f, latitude: %.6f)...\n", longitude, latitude)

	// Call weather tool
	result, err := c.CallTool(context.Background(), "getWeather", args)
	if err != nil {
		log.Fatalf("Failed to call weather tool: %v", err)
	}

	// Pretty print the result
	prettyResult, _ := json.MarshalIndent(result.Content, "", "  ")
	fmt.Println("Weather information:")
	fmt.Println(string(prettyResult))

	// Parse and display important weather information
	if resultMap, ok := result.Content.(map[string]any); ok {
		if resultData, ok := resultMap["result"].(map[string]any); ok {
			// Get location information
			if location, ok := resultData["location"].(map[string]any); ok {
				fmt.Printf("\nLocation: longitude %.6f, latitude %.6f\n",
					location["longitude"].(float64),
					location["latitude"].(float64))
			}

			// Get current weather
			if realtime, ok := resultData["realtime"].(map[string]any); ok {
				fmt.Println("\nCurrent weather:")
				fmt.Printf("Temperature: %.1f°C\n", realtime["temperature"].(float64))
				fmt.Printf("Apparent temperature: %.1f°C\n", realtime["apparent_temperature"].(float64))
				fmt.Printf("Humidity: %.0f%%\n", realtime["humidity"].(float64)*100)
				fmt.Printf("Skycon: %s\n", realtime["skycon"].(string))
				fmt.Printf("Air quality index (AQI): %.0f\n", realtime["air_quality"].(map[string]any)["aqi"].(map[string]any)["chn"].(float64))
				fmt.Printf("Precipitation intensity: %.1f mm/h\n", realtime["precipitation"].(map[string]any)["local"].(map[string]any)["intensity"].(float64))
				fmt.Printf("Wind speed: %.1f km/h\n", realtime["wind"].(map[string]any)["speed"].(float64))
			}

			// Get weather alerts
			if alert, ok := resultData["alert"].(map[string]any); ok {
				if content, ok := alert["content"].([]any); ok && len(content) > 0 {
					fmt.Println("\nWeather alerts:")
					for _, c := range content {
						alertInfo := c.(map[string]any)
						fmt.Printf("- %s: %s\n",
							alertInfo["title"].(string),
							alertInfo["description"].(string))
					}
				}
			}

			// Get forecast summary for the next 24 hours
			if hourly, ok := resultData["hourly"].(map[string]any); ok {
				if temperature, ok := hourly["temperature"].([]any); ok && len(temperature) > 0 {
					fmt.Println("\nForecast for the next 24 hours:")
					for i := 0; i < 6 && i < len(temperature); i++ {
						temp := temperature[i].(map[string]any)
						fmt.Printf("%s: %.1f°C\n",
							temp["datetime"].(string)[11:16],
							temp["value"].(float64))
					}
					fmt.Println("...")
				}
			}
		}
	}
}
