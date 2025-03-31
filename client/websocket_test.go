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
	
	// Create test server that handles WebSocket connections
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
			var request map[string]interface{}
			err = json.Unmarshal(message, &request)
			if err != nil {
				t.Fatalf("Failed to parse message: %v", err)
				continue
			}
			
			// Check request type
			requestType, ok := request["type"].(string)
			if !ok {
				t.Error("Missing type field in request")
				continue
			}
			
			// Check message ID
			messageID, ok := request["messageId"].(string)
			if !ok {
				t.Error("Missing messageId field in request")
				continue
			}
			
			// Prepare response
			response := map[string]interface{}{
				"type":      "response",
				"messageId": messageID,
				"success":   true,
			}
			
			// Handle different request types
			switch requestType {
			case "initialize":
				// Check initialize parameters
				options, ok := request["data"].(map[string]interface{})
				if !ok {
					t.Error("Missing data field in initialize request")
					continue
				}
				
				serverName, ok := options["serverName"].(string)
				if !ok || serverName != "test-server" {
					t.Errorf("Expected serverName 'test-server', got '%v'", options["serverName"])
				}
				
				serverVersion, ok := options["serverVersion"].(string)
				if !ok || serverVersion != "1.0.0" {
					t.Errorf("Expected serverVersion '1.0.0', got '%v'", options["serverVersion"])
				}
				
			case "getPrompt":
				// Check getPrompt parameters
				name, ok := request["name"].(string)
				if !ok || name != "test-prompt" {
					t.Errorf("Expected name 'test-prompt', got '%v'", request["name"])
				}
				
				args, ok := request["args"].(map[string]interface{})
				if !ok {
					t.Error("Missing args field in getPrompt request")
					continue
				}
				
				arg1, ok := args["arg1"].(string)
				if !ok || arg1 != "value1" {
					t.Errorf("Expected arg1 'value1', got '%v'", args["arg1"])
				}
				
				// Add result to response
				response["result"] = map[string]interface{}{
					"description": "Test prompt",
					"message":     "Hello, world!",
				}
				
			case "callTool":
				// Check callTool parameters
				name, ok := request["name"].(string)
				if !ok || name != "test-tool" {
					t.Errorf("Expected name 'test-tool', got '%v'", request["name"])
				}
				
				args, ok := request["args"].(map[string]interface{})
				if !ok {
					t.Error("Missing args field in callTool request")
					continue
				}
				
				arg1, ok := args["arg1"].(string)
				if !ok || arg1 != "value1" {
					t.Errorf("Expected arg1 'value1', got '%v'", args["arg1"])
				}
				
				// Add result to response
				response["result"] = map[string]interface{}{
					"content": map[string]interface{}{
						"message": "Tool executed successfully",
					},
				}
				
			case "listPrompts":
				// Add result to response
				response["result"] = []map[string]interface{}{
					{"name": "prompt1", "description": "Prompt 1"},
					{"name": "prompt2", "description": "Prompt 2"},
				}
				
			case "listTools":
				// Add result to response
				response["result"] = []map[string]interface{}{
					{"name": "tool1", "description": "Tool 1"},
					{"name": "tool2", "description": "Tool 2"},
				}
				
			case "listResources":
				// Add result to response
				response["result"] = []map[string]interface{}{
					{"name": "resource1", "description": "Resource 1", "mimeType": "text/plain"},
					{"name": "resource2", "description": "Resource 2", "mimeType": "application/json"},
				}
				
			case "readResource":
				// Check readResource parameters
				name, ok := request["name"].(string)
				if !ok || name != "test-resource" {
					t.Errorf("Expected name 'test-resource', got '%v'", request["name"])
				}
				
				// Add result to response
				response["result"] = map[string]interface{}{
					"content":  "cmVzb3VyY2UgY29udGVudA==", // base64 of "resource content"
					"mimeType": "text/plain",
				}
				
			default:
				t.Errorf("Unexpected request type: %s", requestType)
				response["success"] = false
				response["error"] = map[string]interface{}{
					"code":    "invalid_request",
					"message": "Invalid request type",
				}
			}
			
			// Send response
			err = conn.WriteJSON(response)
			if err != nil {
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
				
				// Check result is nil (initialize has no result)
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
			name:        "CallTool",
			requestType: "callTool",
			params: map[string]interface{}{
				"name": "test-tool",
				"args": map[string]interface{}{
					"arg1": "value1",
				},
			},
			validate: func(t *testing.T, result interface{}, err error) {
				if err != nil {
					t.Fatalf("SendRequest returned error: %v", err)
				}
				
				// Check result type
				r, ok := result.(*types.CallToolResult)
				if !ok {
					t.Fatalf("Expected result type *types.CallToolResult, got %T", result)
				}
				
				// Check result content
				content, ok := r.Content.(map[string]interface{})
				if !ok {
					t.Fatalf("Expected content type map[string]interface{}, got %T", r.Content)
				}
				
				message, ok := content["message"].(string)
				if !ok {
					t.Fatalf("Expected message type string, got %T", content["message"])
				}
				
				if message != "Tool executed successfully" {
					t.Errorf("Expected message 'Tool executed successfully', got '%s'", message)
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
				
				// Check prompt fields
				if prompts[0].Name != "prompt1" {
					t.Errorf("Expected prompt 0 name 'prompt1', got '%s'", prompts[0].Name)
				}
				
				if prompts[1].Name != "prompt2" {
					t.Errorf("Expected prompt 1 name 'prompt2', got '%s'", prompts[1].Name)
				}
			},
		},
		{
			name:        "ListTools",
			requestType: "listTools",
			params:      map[string]interface{}{},
			validate: func(t *testing.T, result interface{}, err error) {
				if err != nil {
					t.Fatalf("SendRequest returned error: %v", err)
				}
				
				// Check result type
				tools, ok := result.([]types.Tool)
				if !ok {
					t.Fatalf("Expected result type []types.Tool, got %T", result)
				}
				
				// Check result length
				if len(tools) != 2 {
					t.Errorf("Expected 2 tools, got %d", len(tools))
				}
				
				// Check tool fields
				if tools[0].Name != "tool1" {
					t.Errorf("Expected tool 0 name 'tool1', got '%s'", tools[0].Name)
				}
				
				if tools[1].Name != "tool2" {
					t.Errorf("Expected tool 1 name 'tool2', got '%s'", tools[1].Name)
				}
			},
		},
		{
			name:        "ListResources",
			requestType: "listResources",
			params:      map[string]interface{}{},
			validate: func(t *testing.T, result interface{}, err error) {
				if err != nil {
					t.Fatalf("SendRequest returned error: %v", err)
				}
				
				// Check result type
				resources, ok := result.([]types.Resource)
				if !ok {
					t.Fatalf("Expected result type []types.Resource, got %T", result)
				}
				
				// Check result length
				if len(resources) != 2 {
					t.Errorf("Expected 2 resources, got %d", len(resources))
				}
				
				// Check resource fields
				if resources[0].Name != "resource1" {
					t.Errorf("Expected resource 0 name 'resource1', got '%s'", resources[0].Name)
				}
				
				if resources[1].Name != "resource2" {
					t.Errorf("Expected resource 1 name 'resource2', got '%s'", resources[1].Name)
				}
			},
		},
		{
			name:        "ReadResource",
			requestType: "readResource",
			params: map[string]interface{}{
				"name": "test-resource",
			},
			validate: func(t *testing.T, result interface{}, err error) {
				if err != nil {
					t.Fatalf("SendRequest returned error: %v", err)
				}
				
				// Check result type
				resource, ok := result.(map[string]interface{})
				if !ok {
					t.Fatalf("Expected result type map[string]interface{}, got %T", result)
				}
				
				// Check content and mimeType
				content, ok := resource["content"].(string)
				if !ok {
					t.Fatalf("Expected content type string, got %T", resource["content"])
				}
				
				if content != "cmVzb3VyY2UgY29udGVudA==" {
					t.Errorf("Expected content 'cmVzb3VyY2UgY29udGVudA==', got '%s'", content)
				}
				
				mimeType, ok := resource["mimeType"].(string)
				if !ok {
					t.Fatalf("Expected mimeType type string, got %T", resource["mimeType"])
				}
				
				if mimeType != "text/plain" {
					t.Errorf("Expected mimeType 'text/plain', got '%s'", mimeType)
				}
			},
		},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Send request
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			
			result, err := transport.SendRequest(ctx, tc.requestType, tc.params)
			
			// Validate result
			tc.validate(t, result, err)
		})
	}
}