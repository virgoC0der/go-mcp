package transport

import (
	"context"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/virgoC0der/go-mcp/types"
)

func TestWebSocketHandler_ServeHTTP(t *testing.T) {
	// Create a mock server
	mockServer := &MockServer{
		initializeFunc: func(ctx context.Context, options interface{}) error {
			return nil
		},
	}

	// Create WebSocket handler
	handler := NewWebSocketHandler(mockServer)

	// Create test server
	server := httptest.NewServer(handler)
	defer server.Close()

	// Replace "http" with "ws" in the URL
	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")

	// Create WebSocket client
	dialer := websocket.Dialer{
		HandshakeTimeout: 5 * time.Second,
	}

	conn, _, err := dialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("Failed to connect to WebSocket server: %v", err)
	}
	defer conn.Close()

	// Send initialize request
	initRequest := map[string]interface{}{
		"type":      "initialize",
		"messageId": "1",
		"data": map[string]interface{}{
			"serverName":    "test-server",
			"serverVersion": "1.0.0",
		},
	}

	err = conn.WriteJSON(initRequest)
	if err != nil {
		t.Fatalf("Failed to send initialize request: %v", err)
	}

	// Read response
	var response map[string]interface{}
	err = conn.ReadJSON(&response)
	if err != nil {
		t.Fatalf("Failed to read response: %v", err)
	}

	// Check response
	if respType, ok := response["type"].(string); !ok || respType != "response" {
		t.Errorf("Expected response type 'response', got '%v'", respType)
	}

	if messageID, ok := response["messageId"].(string); !ok || messageID != "1" {
		t.Errorf("Expected messageId '1', got '%v'", messageID)
	}

	if success, ok := response["success"].(bool); !ok || !success {
		t.Error("Expected success to be true")
	}
}

func TestWebSocketHandler_HandleRequest(t *testing.T) {
	// Create a mock server
	mockServer := &MockServer{
		listPromptsFunc: func(ctx context.Context) ([]types.Prompt, error) {
			return []types.Prompt{
				{Name: "prompt1", Description: "Prompt 1"},
				{Name: "prompt2", Description: "Prompt 2"},
			}, nil
		},
		getPromptFunc: func(ctx context.Context, name string, args map[string]interface{}) (*types.GetPromptResult, error) {
			return &types.GetPromptResult{
				Description: "Test prompt",
				Message:     "Hello, world!",
			}, nil
		},
		listToolsFunc: func(ctx context.Context) ([]types.Tool, error) {
			return []types.Tool{
				{Name: "tool1", Description: "Tool 1"},
				{Name: "tool2", Description: "Tool 2"},
			}, nil
		},
		callToolFunc: func(ctx context.Context, name string, args map[string]interface{}) (*types.CallToolResult, error) {
			return &types.CallToolResult{
				Content: map[string]interface{}{
					"message": "Tool executed successfully",
				},
			}, nil
		},
		listResourcesFunc: func(ctx context.Context) ([]types.Resource, error) {
			return []types.Resource{
				{Name: "resource1", Description: "Resource 1", MimeType: "text/plain"},
				{Name: "resource2", Description: "Resource 2", MimeType: "application/json"},
			}, nil
		},
		readResourceFunc: func(ctx context.Context, name string) ([]byte, string, error) {
			return []byte("resource content"), "text/plain", nil
		},
	}

	// Create WebSocketHandler
	handler := NewWebSocketHandler(mockServer)

	// Create WebSocket connection
	server := httptest.NewServer(handler)
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")

	dialer := websocket.Dialer{
		HandshakeTimeout: 5 * time.Second,
	}

	conn, _, err := dialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("Failed to connect to WebSocket server: %v", err)
	}
	defer conn.Close()

	// Test listPrompts
	testCases := []struct {
		name          string
		request       map[string]interface{}
		checkResponse func(t *testing.T, response map[string]interface{})
	}{
		{
			name: "listPrompts",
			request: map[string]interface{}{
				"type":      "listPrompts",
				"messageId": "1",
			},
			checkResponse: func(t *testing.T, response map[string]interface{}) {
				if success, ok := response["success"].(bool); !ok || !success {
					t.Error("Expected success to be true")
				}

				if result, ok := response["result"].([]interface{}); !ok {
					t.Error("Expected result array in response")
				} else if len(result) != 2 {
					t.Errorf("Expected 2 prompts, got %d", len(result))
				}
			},
		},
		{
			name: "getPrompt",
			request: map[string]interface{}{
				"type":      "getPrompt",
				"messageId": "2",
				"name":      "test-prompt",
				"args":      map[string]interface{}{"arg1": "value1"},
			},
			checkResponse: func(t *testing.T, response map[string]interface{}) {
				if success, ok := response["success"].(bool); !ok || !success {
					t.Error("Expected success to be true")
				}

				if result, ok := response["result"].(map[string]interface{}); !ok {
					t.Error("Expected result object in response")
				} else {
					if msg, ok := result["message"].(string); !ok || msg != "Hello, world!" {
						t.Errorf("Expected message 'Hello, world!', got '%v'", msg)
					}
				}
			},
		},
		{
			name: "listTools",
			request: map[string]interface{}{
				"type":      "listTools",
				"messageId": "3",
			},
			checkResponse: func(t *testing.T, response map[string]interface{}) {
				if success, ok := response["success"].(bool); !ok || !success {
					t.Error("Expected success to be true")
				}

				if result, ok := response["result"].([]interface{}); !ok {
					t.Error("Expected result array in response")
				} else if len(result) != 2 {
					t.Errorf("Expected 2 tools, got %d", len(result))
				}
			},
		},
		{
			name: "callTool",
			request: map[string]interface{}{
				"type":      "callTool",
				"messageId": "4",
				"name":      "test-tool",
				"args":      map[string]interface{}{"arg1": "value1"},
			},
			checkResponse: func(t *testing.T, response map[string]interface{}) {
				if success, ok := response["success"].(bool); !ok || !success {
					t.Error("Expected success to be true")
				}

				if result, ok := response["result"].(map[string]interface{}); !ok {
					t.Error("Expected result object in response")
				} else {
					if content, ok := result["content"].(map[string]interface{}); !ok {
						t.Error("Expected content object in result")
					} else {
						if msg, ok := content["message"].(string); !ok || msg != "Tool executed successfully" {
							t.Errorf("Expected message 'Tool executed successfully', got '%v'", msg)
						}
					}
				}
			},
		},
		{
			name: "listResources",
			request: map[string]interface{}{
				"type":      "listResources",
				"messageId": "5",
			},
			checkResponse: func(t *testing.T, response map[string]interface{}) {
				if success, ok := response["success"].(bool); !ok || !success {
					t.Error("Expected success to be true")
				}

				if result, ok := response["result"].([]interface{}); !ok {
					t.Error("Expected result array in response")
				} else if len(result) != 2 {
					t.Errorf("Expected 2 resources, got %d", len(result))
				}
			},
		},
		{
			name: "readResource",
			request: map[string]interface{}{
				"type":      "readResource",
				"messageId": "6",
				"name":      "test-resource",
			},
			checkResponse: func(t *testing.T, response map[string]interface{}) {
				if success, ok := response["success"].(bool); !ok || !success {
					t.Error("Expected success to be true")
				}

				if result, ok := response["result"].(map[string]interface{}); !ok {
					t.Error("Expected result object in response")
				} else {
					if content, ok := result["content"].(string); !ok || content != "cmVzb3VyY2UgY29udGVudA==" { // base64 of "resource content"
						t.Errorf("Expected base64 encoded content, got '%v'", content)
					}

					if mimeType, ok := result["mimeType"].(string); !ok || mimeType != "text/plain" {
						t.Errorf("Expected mimeType 'text/plain', got '%v'", mimeType)
					}
				}
			},
		},
		{
			name: "invalidType",
			request: map[string]interface{}{
				"type":      "invalidType",
				"messageId": "7",
			},
			checkResponse: func(t *testing.T, response map[string]interface{}) {
				if success, ok := response["success"].(bool); !ok || success {
					t.Error("Expected success to be false")
				}

				if errObj, ok := response["error"].(map[string]interface{}); !ok {
					t.Error("Expected error object in response")
				} else {
					if code, ok := errObj["code"].(string); !ok || code != "invalid_request" {
						t.Errorf("Expected error code 'invalid_request', got '%v'", code)
					}
				}
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Send request
			err := conn.WriteJSON(tc.request)
			if err != nil {
				t.Fatalf("Failed to send request: %v", err)
			}

			// Read response
			var response map[string]interface{}
			err = conn.ReadJSON(&response)
			if err != nil {
				t.Fatalf("Failed to read response: %v", err)
			}

			// Check response
			if respType, ok := response["type"].(string); !ok || respType != "response" {
				t.Errorf("Expected response type 'response', got '%v'", respType)
			}

			if messageID, ok := response["messageId"].(string); !ok || messageID != tc.request["messageId"].(string) {
				t.Errorf("Expected messageId '%s', got '%v'", tc.request["messageId"].(string), messageID)
			}

			// Run case-specific checks
			tc.checkResponse(t, response)
		})
	}
}
