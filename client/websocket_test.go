package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/virgoC0der/go-mcp/types"
)

func TestWebSocketTransport_SendRequest(t *testing.T) {
	// Create WebSocket upgrader
	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}

	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Upgrade connection to WebSocket
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			t.Fatalf("Failed to upgrade connection: %v", err)
			return
		}
		defer conn.Close()

		// Handle WebSocket messages
		for {
			// Read message
			_, message, err := conn.ReadMessage()
			if err != nil {
				// Connection closed or error occurred
				break
			}

			// Parse message
			var request struct {
				ID     string          `json:"id"`
				Method string          `json:"method"`
				Params json.RawMessage `json:"params,omitempty"`
			}
			err = json.Unmarshal(message, &request)
			if err != nil {
				t.Fatalf("Failed to parse message: %v", err)
				continue
			}

			// Check request ID
			if request.ID == "" {
				t.Error("Missing id field in request")
				continue
			}

			// Prepare response
			response := struct {
				ID      string          `json:"id"`
				Success bool            `json:"success"`
				Content json.RawMessage `json:"content,omitempty"`
				Error   string          `json:"error,omitempty"`
			}{
				ID:      request.ID,
				Success: true,
			}

			// Handle different request types
			switch request.Method {
			case "initialize":
				// Initialize doesn't need to return any content

			case "getPrompt":
				// Parse getPrompt parameters
				var params struct {
					Name      string                 `json:"name"`
					Arguments map[string]interface{} `json:"arguments"`
				}
				err = json.Unmarshal(request.Params, &params)
				if err != nil {
					t.Fatalf("Failed to parse getPrompt params: %v", err)
					continue
				}

				if params.Name != "test-prompt" {
					t.Errorf("Expected name 'test-prompt', got '%v'", params.Name)
				}

				// Return prompt result
				result := types.GetPromptResult{
					Description: "Test prompt",
					Message:     "Hello, world!",
				}
				resultBytes, _ := json.Marshal(result)
				response.Content = resultBytes

			case "callTool":
				// Parse callTool parameters
				var params struct {
					Name      string                 `json:"name"`
					Arguments map[string]interface{} `json:"arguments"`
				}
				err = json.Unmarshal(request.Params, &params)
				if err != nil {
					t.Fatalf("Failed to parse callTool params: %v", err)
					continue
				}

				if params.Name != "test-tool" {
					t.Errorf("Expected name 'test-tool', got '%v'", params.Name)
				}

				// Return tool result
				result := types.CallToolResult{
					Content: map[string]interface{}{
						"message": "Tool executed successfully",
					},
				}
				resultBytes, _ := json.Marshal(result)
				response.Content = resultBytes

			case "listPrompts":
				// Return prompts list
				prompts := []types.Prompt{
					{Name: "prompt1", Description: "Prompt 1"},
					{Name: "prompt2", Description: "Prompt 2"},
				}
				promptsBytes, _ := json.Marshal(prompts)
				response.Content = promptsBytes

			case "listTools":
				// Return tools list
				tools := []types.Tool{
					{Name: "tool1", Description: "Tool 1"},
					{Name: "tool2", Description: "Tool 2"},
				}
				toolsBytes, _ := json.Marshal(tools)
				response.Content = toolsBytes

			case "listResources":
				// Return resources list
				resources := []types.Resource{
					{Name: "resource1", Description: "Resource 1", MimeType: "text/plain"},
					{Name: "resource2", Description: "Resource 2", MimeType: "application/json"},
				}
				resourcesBytes, _ := json.Marshal(resources)
				response.Content = resourcesBytes

			case "readResource":
				// Parse readResource parameters
				var params struct {
					Name string `json:"name"`
				}
				err = json.Unmarshal(request.Params, &params)
				if err != nil {
					t.Fatalf("Failed to parse readResource params: %v", err)
					continue
				}

				if params.Name != "test-resource" {
					t.Errorf("Expected name 'test-resource', got '%v'", params.Name)
				}

				// Return resource content
				result := map[string]interface{}{
					"content":  "cmVzb3VyY2UgY29udGVudA==", // base64 of "resource content"
					"mimeType": "text/plain",
				}
				resultBytes, _ := json.Marshal(result)
				response.Content = resultBytes

			default:
				t.Errorf("Unexpected request method: %s", request.Method)
				response.Success = false
				response.Error = "Invalid request method"
			}

			// Send response
			responseBytes, _ := json.Marshal(response)
			if err := conn.WriteMessage(websocket.TextMessage, responseBytes); err != nil {
				t.Fatalf("Failed to send response: %v", err)
				break
			}
		}
	}))
	defer server.Close()

	// Replace "http" with "ws" in the URL
	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")

	// Create WebSocket transport
	transport, err := NewWebSocketTransport(wsURL)
	if err != nil {
		t.Fatalf("Failed to create WebSocket transport: %v", err)
	}
	defer transport.Close()

	// Test cases
	testCases := []struct {
		name        string
		requestType string
		params      map[string]interface{}
		validate    func(t *testing.T, result interface{}, err error)
	}{
		{
			name:        "Initialize",
			requestType: "initialize",
			params: map[string]interface{}{
				"serverName":    "test-server",
				"serverVersion": "1.0.0",
			},
			validate: func(t *testing.T, result interface{}, err error) {
				if err != nil {
					t.Fatalf("SendRequest returned error: %v", err)
				}

				// Check if result is nil
				if result != nil {
					t.Errorf("Expected nil result, got %v", result)
				}
			},
		},
		{
			name:        "GetPrompt",
			requestType: "getPrompt",
			params: map[string]interface{}{
				"name": "test-prompt",
				"args": map[string]interface{}{
					"arg1": "value1",
				},
			},
			validate: func(t *testing.T, result interface{}, err error) {
				if err != nil {
					t.Fatalf("SendRequest returned error: %v", err)
				}

				// Check result type
				r, ok := result.(*types.GetPromptResult)
				if !ok {
					t.Fatalf("Expected result type *types.GetPromptResult, got %T", result)
				}

				// Check result fields
				if r.Description != "Test prompt" {
					t.Errorf("Expected description 'Test prompt', got '%s'", r.Description)
				}

				if r.Message != "Hello, world!" {
					t.Errorf("Expected message 'Hello, world!', got '%s'", r.Message)
				}
			},
		},
		{
			name:        "ListPrompts",
			requestType: "listPrompts",
			params:      map[string]interface{}{},
			validate: func(t *testing.T, result interface{}, err error) {
				if err != nil {
					t.Fatalf("SendRequest returned error: %v", err)
				}

				// Check result type
				prompts, ok := result.([]types.Prompt)
				if !ok {
					t.Fatalf("Expected result type []types.Prompt, got %T", result)
				}

				// Check result length
				if len(prompts) != 2 {
					t.Errorf("Expected 2 prompts, got %d", len(prompts))
				}
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Send request
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()

			result, err := transport.SendRequest(ctx, tc.requestType, tc.params)

			// Validate result
			tc.validate(t, result, err)
		})
	}
}
