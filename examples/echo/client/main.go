package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

func main() {
	// Test HTTP endpoints
	testHTTPEndpoints()

	// Test WebSocket endpoints
	testWebSocketEndpoints()
}

func testHTTPEndpoints() {
	fmt.Println("Testing HTTP endpoints...")

	// Test getting prompts list
	resp, err := http.Get("http://localhost:8080/prompts")
	if err != nil {
		log.Printf("Failed to get prompts: %v", err)
		return
	}
	defer resp.Body.Close()

	var prompts []struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&prompts); err != nil {
		log.Printf("Failed to decode prompts: %v", err)
		return
	}

	fmt.Println("Available prompts:")
	for _, p := range prompts {
		fmt.Printf("- %s: %s\n", p.Name, p.Description)
	}

	// Test calling a prompt
	promptReq := struct {
		Name      string            `json:"name"`
		Arguments map[string]string `json:"arguments"`
	}{
		Name: "echo",
		Arguments: map[string]string{
			"message": "Hello from HTTP client!",
		},
	}

	promptBody, _ := json.Marshal(promptReq)
	resp, err = http.Post("http://localhost:8080/prompt", "application/json", bytes.NewBuffer(promptBody))
	if err != nil {
		log.Printf("Failed to call prompt: %v", err)
		return
	}
	defer resp.Body.Close()

	var promptResult struct {
		Description string `json:"description"`
		Messages    []struct {
			Role    string `json:"role"`
			Content struct {
				Text string `json:"text"`
			} `json:"content"`
		} `json:"messages"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&promptResult); err != nil {
		log.Printf("Failed to decode prompt result: %v", err)
		return
	}

	fmt.Printf("Prompt response: %s\n", promptResult.Messages[0].Content.Text)
}

func testWebSocketEndpoints() {
	fmt.Println("\nTesting WebSocket endpoints...")

	// Connect to WebSocket server
	conn, _, err := websocket.DefaultDialer.Dial("ws://localhost:8081/ws", nil)
	if err != nil {
		log.Printf("Failed to connect to WebSocket server: %v", err)
		return
	}
	defer conn.Close()

	// Create a channel for receiving responses
	responses := make(chan map[string]interface{}, 1)
	go func() {
		for {
			var response map[string]interface{}
			if err := conn.ReadJSON(&response); err != nil {
				log.Printf("Failed to read response: %v", err)
				return
			}
			responses <- response
		}
	}()

	// Test getting tools list
	err = conn.WriteJSON(map[string]interface{}{
		"type":   "request",
		"id":     "1",
		"method": "listTools",
	})
	if err != nil {
		log.Printf("Failed to send listTools request: %v", err)
		return
	}

	// Wait for response
	select {
	case response := <-responses:
		if tools, ok := response["result"].([]interface{}); ok {
			fmt.Println("Available tools:")
			for _, t := range tools {
				if tool, ok := t.(map[string]interface{}); ok {
					fmt.Printf("- %s: %s\n", tool["name"], tool["description"])
				}
			}
		}
	case <-time.After(5 * time.Second):
		log.Println("Timeout waiting for listTools response")
		return
	}

	// Test calling a tool
	err = conn.WriteJSON(map[string]interface{}{
		"type":   "request",
		"id":     "2",
		"method": "callTool",
		"arguments": map[string]interface{}{
			"name": "echo",
			"arguments": map[string]interface{}{
				"message": "Hello from WebSocket client!",
			},
		},
	})
	if err != nil {
		log.Printf("Failed to send callTool request: %v", err)
		return
	}

	// Wait for response
	select {
	case response := <-responses:
		if result, ok := response["result"].(string); ok {
			fmt.Printf("Tool response: %s\n", result)
		}
	case <-time.After(5 * time.Second):
		log.Println("Timeout waiting for callTool response")
		return
	}
}
