package server

import (
	"context"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
	"github.com/virgoC0der/go-mcp/internal/types"
)

func TestWebSocketServer_Lifecycle(t *testing.T) {
	mockServer := &MockServer{
		startFunc: func() error {
			return nil
		},
		shutdownFunc: func(ctx context.Context) error {
			return nil
		},
	}

	server := NewWSServer(mockServer, ":8080")
	assert.NotNil(t, server)

	// Test Start (in goroutine since it blocks)
	go func() {
		err := server.Start()
		assert.NoError(t, err)
	}()

	// Give server time to start
	time.Sleep(100 * time.Millisecond)

	// Test Shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := server.Shutdown(ctx)
	assert.NoError(t, err)
}

func TestWebSocketHandler_Initialize(t *testing.T) {
	mockServer := &MockServer{
		initializeFunc: func(ctx context.Context, options any) error {
			return nil
		},
	}

	handler := NewWebSocketHandler(mockServer)
	server := httptest.NewServer(handler)
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")
	dialer := websocket.Dialer{
		HandshakeTimeout: 5 * time.Second,
	}

	conn, _, err := dialer.Dial(wsURL, nil)
	assert.NoError(t, err)
	defer conn.Close()

	// Test successful initialization
	initRequest := map[string]any{
		"type":      "initialize",
		"messageId": "1",
		"data": map[string]any{
			"serverName":    "test-server",
			"serverVersion": "1.0.0",
		},
	}

	err = conn.WriteJSON(initRequest)
	assert.NoError(t, err)

	var response map[string]any
	err = conn.ReadJSON(&response)
	assert.NoError(t, err)

	assert.Equal(t, "response", response["type"])
	assert.Equal(t, "1", response["messageId"])
	assert.True(t, response["success"].(bool))

	// Test initialization error
	mockServer.initializeFunc = func(ctx context.Context, options any) error {
		return types.NewError("initialize_error", "Failed to initialize")
	}

	initRequest["messageId"] = "2"
	err = conn.WriteJSON(initRequest)
	assert.NoError(t, err)

	err = conn.ReadJSON(&response)
	assert.NoError(t, err)

	assert.Equal(t, "response", response["type"])
	assert.Equal(t, "2", response["messageId"])
	assert.False(t, response["success"].(bool))

	errObj := response["error"].(map[string]any)
	assert.Equal(t, "initialize_error", errObj["code"])
}

func TestWebSocketHandler_HandleRequests(t *testing.T) {
	mockServer := &MockServer{
		listPromptsFunc: func(ctx context.Context) ([]types.Prompt, error) {
			return []types.Prompt{
				{Name: "prompt1", Description: "Prompt 1"},
				{Name: "prompt2", Description: "Prompt 2"},
			}, nil
		},
		getPromptFunc: func(ctx context.Context, name string, args map[string]any) (*types.GetPromptResult, error) {
			return &types.GetPromptResult{
				Content: "Hello, world!",
			}, nil
		},
		listToolsFunc: func(ctx context.Context) ([]types.Tool, error) {
			return []types.Tool{
				{Name: "tool1", Description: "Tool 1"},
				{Name: "tool2", Description: "Tool 2"},
			}, nil
		},
		callToolFunc: func(ctx context.Context, name string, args map[string]any) (*types.CallToolResult, error) {
			return &types.CallToolResult{
				Output: map[string]interface{}{
					"result": "success",
				},
			}, nil
		},
		listResourcesFunc: func(ctx context.Context) ([]types.Resource, error) {
			return []types.Resource{
				{Name: "resource1", Description: "Resource 1", Type: "text/plain"},
				{Name: "resource2", Description: "Resource 2", Type: "application/json"},
			}, nil
		},
		readResourceFunc: func(ctx context.Context, name string) ([]byte, string, error) {
			return []byte("test content"), "text/plain", nil
		},
	}

	handler := NewWebSocketHandler(mockServer)
	server := httptest.NewServer(handler)
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")
	dialer := websocket.Dialer{
		HandshakeTimeout: 5 * time.Second,
	}

	conn, _, err := dialer.Dial(wsURL, nil)
	assert.NoError(t, err)
	defer conn.Close()

	testCases := []struct {
		name    string
		request map[string]any
		check   func(t *testing.T, response map[string]any)
	}{
		{
			name: "listPrompts",
			request: map[string]any{
				"type":      "listPrompts",
				"messageId": "1",
			},
			check: func(t *testing.T, response map[string]any) {
				assert.True(t, response["success"].(bool))
				result := response["result"].([]any)
				assert.Len(t, result, 2)
			},
		},
		{
			name: "getPrompt",
			request: map[string]any{
				"type":      "getPrompt",
				"messageId": "2",
				"name":      "test-prompt",
				"args":      map[string]any{"name": "test"},
			},
			check: func(t *testing.T, response map[string]any) {
				assert.True(t, response["success"].(bool))
				result := response["result"].(map[string]any)
				assert.Equal(t, "Hello, world!", result["content"])
			},
		},
		{
			name: "listTools",
			request: map[string]any{
				"type":      "listTools",
				"messageId": "3",
			},
			check: func(t *testing.T, response map[string]any) {
				assert.True(t, response["success"].(bool))
				result := response["result"].([]any)
				assert.Len(t, result, 2)
			},
		},
		{
			name: "callTool",
			request: map[string]any{
				"type":      "callTool",
				"messageId": "4",
				"name":      "test-tool",
				"args":      map[string]any{"param": "value"},
			},
			check: func(t *testing.T, response map[string]any) {
				assert.True(t, response["success"].(bool))
				result := response["result"].(map[string]any)
				assert.Equal(t, "success", result["output"].(map[string]any)["result"])
			},
		},
		{
			name: "listResources",
			request: map[string]any{
				"type":      "listResources",
				"messageId": "5",
			},
			check: func(t *testing.T, response map[string]any) {
				assert.True(t, response["success"].(bool))
				result := response["result"].([]any)
				assert.Len(t, result, 2)
			},
		},
		{
			name: "readResource",
			request: map[string]any{
				"type":      "readResource",
				"messageId": "6",
				"name":      "test-resource",
			},
			check: func(t *testing.T, response map[string]any) {
				assert.True(t, response["success"].(bool))
				result := response["result"].(map[string]any)
				assert.Equal(t, "test content", result["content"])
				assert.Equal(t, "text/plain", result["type"])
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := conn.WriteJSON(tc.request)
			assert.NoError(t, err)

			var response map[string]any
			err = conn.ReadJSON(&response)
			assert.NoError(t, err)

			assert.Equal(t, "response", response["type"])
			assert.Equal(t, tc.request["messageId"], response["messageId"])
			tc.check(t, response)
		})
	}

	// Test error cases
	errorCases := []struct {
		name          string
		setupMock     func()
		request       map[string]any
		expectedError string
	}{
		{
			name: "listPrompts error",
			setupMock: func() {
				mockServer.listPromptsFunc = func(ctx context.Context) ([]types.Prompt, error) {
					return nil, types.NewError("list_prompts_error", "Failed to list prompts")
				}
			},
			request: map[string]any{
				"type":      "listPrompts",
				"messageId": "7",
			},
			expectedError: "list_prompts_error",
		},
		{
			name: "getPrompt error",
			setupMock: func() {
				mockServer.getPromptFunc = func(ctx context.Context, name string, args map[string]any) (*types.GetPromptResult, error) {
					return nil, types.NewError("get_prompt_error", "Failed to get prompt")
				}
			},
			request: map[string]any{
				"type":      "getPrompt",
				"messageId": "8",
				"name":      "test-prompt",
				"args":      map[string]any{},
			},
			expectedError: "get_prompt_error",
		},
		{
			name: "listTools error",
			setupMock: func() {
				mockServer.listToolsFunc = func(ctx context.Context) ([]types.Tool, error) {
					return nil, types.NewError("list_tools_error", "Failed to list tools")
				}
			},
			request: map[string]any{
				"type":      "listTools",
				"messageId": "9",
			},
			expectedError: "list_tools_error",
		},
		{
			name: "callTool error",
			setupMock: func() {
				mockServer.callToolFunc = func(ctx context.Context, name string, args map[string]any) (*types.CallToolResult, error) {
					return nil, types.NewError("call_tool_error", "Failed to call tool")
				}
			},
			request: map[string]any{
				"type":      "callTool",
				"messageId": "10",
				"name":      "test-tool",
				"args":      map[string]any{},
			},
			expectedError: "call_tool_error",
		},
		{
			name: "listResources error",
			setupMock: func() {
				mockServer.listResourcesFunc = func(ctx context.Context) ([]types.Resource, error) {
					return nil, types.NewError("list_resources_error", "Failed to list resources")
				}
			},
			request: map[string]any{
				"type":      "listResources",
				"messageId": "11",
			},
			expectedError: "list_resources_error",
		},
		{
			name: "readResource error",
			setupMock: func() {
				mockServer.readResourceFunc = func(ctx context.Context, name string) ([]byte, string, error) {
					return nil, "", types.NewError("read_resource_error", "Failed to read resource")
				}
			},
			request: map[string]any{
				"type":      "readResource",
				"messageId": "12",
				"name":      "test-resource",
			},
			expectedError: "read_resource_error",
		},
	}

	for _, tc := range errorCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.setupMock()

			err := conn.WriteJSON(tc.request)
			assert.NoError(t, err)

			var response map[string]any
			err = conn.ReadJSON(&response)
			assert.NoError(t, err)

			assert.Equal(t, "response", response["type"])
			assert.Equal(t, tc.request["messageId"], response["messageId"])
			assert.False(t, response["success"].(bool))

			errObj := response["error"].(map[string]any)
			assert.Equal(t, tc.expectedError, errObj["code"])
		})
	}
}
