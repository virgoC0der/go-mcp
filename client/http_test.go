package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/virgoC0der/go-mcp/types"
)

func TestHTTPTransport_SendRequest(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Handle different HTTP methods
		switch r.Method {
		case http.MethodPost:
			// Check Content-Type header
			contentType := r.Header.Get("Content-Type")
			if contentType != "application/json" {
				t.Errorf("Expected Content-Type 'application/json', got '%s'", contentType)
			}

			// Read request body
			var requestBody map[string]interface{}
			decoder := json.NewDecoder(r.Body)
			err := decoder.Decode(&requestBody)
			if err != nil {
				t.Fatalf("Failed to decode request body: %v", err)
			}

			// Check request parameters for different endpoints
			switch r.URL.Path {
			case "/initialize":
				// Check initialize parameters
				if requestBody["serverName"] != "test-server" {
					t.Errorf("Expected serverName 'test-server', got '%v'", requestBody["serverName"])
				}

				if requestBody["serverVersion"] != "1.0.0" {
					t.Errorf("Expected serverVersion '1.0.0', got '%v'", requestBody["serverVersion"])
				}

				// Return success response
				w.WriteHeader(http.StatusOK)
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(map[string]interface{}{
					"success": true,
				})

			case "/prompts/test-prompt":
				// Check getPrompt parameters
				if args, ok := requestBody["args"].(map[string]interface{}); !ok {
					t.Error("Expected args parameter to be map[string]interface{}")
				} else if args["arg1"] != "value1" {
					t.Errorf("Expected arg1 'value1', got '%v'", args["arg1"])
				}

				// Return success response with result
				w.WriteHeader(http.StatusOK)
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(map[string]interface{}{
					"success": true,
					"result": map[string]interface{}{
						"description": "Test prompt",
						"message":     "Hello, world!",
					},
				})

			case "/tools/test-tool":
				// Check callTool parameters
				if args, ok := requestBody["args"].(map[string]interface{}); !ok {
					t.Error("Expected args parameter to be map[string]interface{}")
				} else if args["arg1"] != "value1" {
					t.Errorf("Expected arg1 'value1', got '%v'", args["arg1"])
				}

				// Return success response with result
				w.WriteHeader(http.StatusOK)
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(map[string]interface{}{
					"success": true,
					"result": map[string]interface{}{
						"content": map[string]interface{}{
							"message": "Tool executed successfully",
						},
					},
				})

			default:
				t.Errorf("Unexpected URL path for POST: %s", r.URL.Path)
				w.WriteHeader(http.StatusNotFound)
			}

		case http.MethodGet:
			// Handle GET requests
			switch r.URL.Path {
			case "/prompts":
				// Return success response with prompts
				w.WriteHeader(http.StatusOK)
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(map[string]interface{}{
					"success": true,
					"result": []map[string]interface{}{
						{"name": "prompt1", "description": "Prompt 1"},
						{"name": "prompt2", "description": "Prompt 2"},
					},
				})

			case "/tools":
				// Return success response with tools
				w.WriteHeader(http.StatusOK)
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(map[string]interface{}{
					"success": true,
					"result": []map[string]interface{}{
						{"name": "tool1", "description": "Tool 1"},
						{"name": "tool2", "description": "Tool 2"},
					},
				})

			case "/resources":
				// Return success response with resources
				w.WriteHeader(http.StatusOK)
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(map[string]interface{}{
					"success": true,
					"result": []map[string]interface{}{
						{"name": "resource1", "description": "Resource 1", "mimeType": "text/plain"},
						{"name": "resource2", "description": "Resource 2", "mimeType": "application/json"},
					},
				})

			case "/resources/test-resource":
				// Return success response with resource content
				w.WriteHeader(http.StatusOK)
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(map[string]interface{}{
					"success": true,
					"result": map[string]interface{}{
						"content":  "cmVzb3VyY2UgY29udGVudA==", // base64 of "resource content"
						"mimeType": "text/plain",
					},
				})

			default:
				t.Errorf("Unexpected URL path for GET: %s", r.URL.Path)
				w.WriteHeader(http.StatusNotFound)
			}

		default:
			t.Errorf("Unexpected HTTP method: %s", r.Method)
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	}))
	defer server.Close()

	// Create HTTP transport
	transport := NewHTTPTransport(server.URL)

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
			result, err := transport.SendRequest(context.Background(), tc.requestType, tc.params)

			// Validate result
			tc.validate(t, result, err)
		})
	}
}
