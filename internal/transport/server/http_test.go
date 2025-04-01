package server

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/virgoC0der/go-mcp/internal/types"
)

func TestHTTPServer_Lifecycle(t *testing.T) {
	mockServer := &MockServer{
		initializeFunc: func(ctx context.Context, options any) error {
			return nil
		},
		startFunc: func() error {
			return nil
		},
		shutdownFunc: func(ctx context.Context) error {
			return nil
		},
	}

	server := NewHTTPServer(mockServer, ":8080")
	assert.NotNil(t, server)

	// Test Initialize
	err := server.Initialize(context.Background(), nil)
	assert.NoError(t, err)

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

	err = server.Shutdown(ctx)
	assert.NoError(t, err)
}

func TestHTTPHandler_Initialize(t *testing.T) {
	mockServer := &MockServer{
		initializeFunc: func(ctx context.Context, options any) error {
			return nil
		},
	}

	handler := newHTTPHandler(mockServer)

	// Test successful initialization
	body := `{"serverName": "test-server", "serverVersion": "1.0.0"}`
	req := httptest.NewRequest(http.MethodPost, "/initialize", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	res := httptest.NewRecorder()

	handler.ServeHTTP(res, req)
	assert.Equal(t, http.StatusOK, res.Code)

	var response map[string]any
	err := json.Unmarshal(res.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.True(t, response["success"].(bool))

	// Test initialization error
	mockServer.initializeFunc = func(ctx context.Context, options any) error {
		return types.NewError("initialize_error", "Failed to initialize")
	}

	req = httptest.NewRequest(http.MethodPost, "/initialize", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	res = httptest.NewRecorder()

	handler.ServeHTTP(res, req)
	assert.Equal(t, http.StatusOK, res.Code)

	err = json.Unmarshal(res.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.False(t, response["success"].(bool))

	errObj := response["error"].(map[string]any)
	assert.Equal(t, "initialize_error", errObj["code"])
}

func TestHTTPHandler_ListPrompts(t *testing.T) {
	prompts := []types.Prompt{
		{Name: "prompt1", Description: "Prompt 1"},
		{Name: "prompt2", Description: "Prompt 2"},
	}

	mockServer := &MockServer{
		listPromptsFunc: func(ctx context.Context) ([]types.Prompt, error) {
			return prompts, nil
		},
	}

	handler := newHTTPHandler(mockServer)

	// Test successful listing
	req := httptest.NewRequest(http.MethodGet, "/prompts", http.NoBody)
	res := httptest.NewRecorder()

	handler.ServeHTTP(res, req)
	assert.Equal(t, http.StatusOK, res.Code)

	var response map[string]any
	err := json.Unmarshal(res.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.True(t, response["success"].(bool))

	result := response["result"].([]any)
	assert.Len(t, result, 2)

	// Test listing error
	mockServer.listPromptsFunc = func(ctx context.Context) ([]types.Prompt, error) {
		return nil, types.NewError("list_prompts_error", "Failed to list prompts")
	}

	req = httptest.NewRequest(http.MethodGet, "/prompts", http.NoBody)
	res = httptest.NewRecorder()

	handler.ServeHTTP(res, req)
	assert.Equal(t, http.StatusOK, res.Code)

	err = json.Unmarshal(res.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.False(t, response["success"].(bool))

	errObj := response["error"].(map[string]any)
	assert.Equal(t, "list_prompts_error", errObj["code"])
}

func TestHTTPHandler_GetPrompt(t *testing.T) {
	mockServer := &MockServer{
		getPromptFunc: func(ctx context.Context, name string, args map[string]any) (*types.GetPromptResult, error) {
			return &types.GetPromptResult{
				Content: "Hello, world!",
			}, nil
		},
	}

	handler := newHTTPHandler(mockServer)

	// Test successful prompt retrieval
	body := `{"args": {"name": "test"}}`
	req := httptest.NewRequest(http.MethodPost, "/prompts/test-prompt", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	res := httptest.NewRecorder()

	handler.ServeHTTP(res, req)
	assert.Equal(t, http.StatusOK, res.Code)

	var response map[string]any
	err := json.Unmarshal(res.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.True(t, response["success"].(bool))

	result := response["result"].(map[string]any)
	assert.Equal(t, "Hello, world!", result["content"])

	// Test prompt error
	mockServer.getPromptFunc = func(ctx context.Context, name string, args map[string]any) (*types.GetPromptResult, error) {
		return nil, types.NewError("get_prompt_error", "Failed to get prompt")
	}

	req = httptest.NewRequest(http.MethodPost, "/prompts/test-prompt", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	res = httptest.NewRecorder()

	handler.ServeHTTP(res, req)
	assert.Equal(t, http.StatusOK, res.Code)

	err = json.Unmarshal(res.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.False(t, response["success"].(bool))

	errObj := response["error"].(map[string]any)
	assert.Equal(t, "get_prompt_error", errObj["code"])
}

func TestHTTPHandler_ListTools(t *testing.T) {
	tools := []types.Tool{
		{Name: "tool1", Description: "Tool 1"},
		{Name: "tool2", Description: "Tool 2"},
	}

	mockServer := &MockServer{
		listToolsFunc: func(ctx context.Context) ([]types.Tool, error) {
			return tools, nil
		},
	}

	handler := newHTTPHandler(mockServer)

	// Test successful listing
	req := httptest.NewRequest(http.MethodGet, "/tools", http.NoBody)
	res := httptest.NewRecorder()

	handler.ServeHTTP(res, req)
	assert.Equal(t, http.StatusOK, res.Code)

	var response map[string]any
	err := json.Unmarshal(res.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.True(t, response["success"].(bool))

	result := response["result"].([]any)
	assert.Len(t, result, 2)

	// Test listing error
	mockServer.listToolsFunc = func(ctx context.Context) ([]types.Tool, error) {
		return nil, types.NewError("list_tools_error", "Failed to list tools")
	}

	req = httptest.NewRequest(http.MethodGet, "/tools", http.NoBody)
	res = httptest.NewRecorder()

	handler.ServeHTTP(res, req)
	assert.Equal(t, http.StatusOK, res.Code)

	err = json.Unmarshal(res.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.False(t, response["success"].(bool))

	errObj := response["error"].(map[string]any)
	assert.Equal(t, "list_tools_error", errObj["code"])
}

func TestHTTPHandler_CallTool(t *testing.T) {
	mockServer := &MockServer{
		callToolFunc: func(ctx context.Context, name string, args map[string]any) (*types.CallToolResult, error) {
			return &types.CallToolResult{
				Output: map[string]interface{}{
					"result": "success",
				},
			}, nil
		},
	}

	handler := newHTTPHandler(mockServer)

	// Test successful tool call
	body := `{"args": {"param": "value"}}`
	req := httptest.NewRequest(http.MethodPost, "/tools/test-tool", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	res := httptest.NewRecorder()

	handler.ServeHTTP(res, req)
	assert.Equal(t, http.StatusOK, res.Code)

	var response map[string]any
	err := json.Unmarshal(res.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.True(t, response["success"].(bool))

	result := response["result"].(map[string]any)
	assert.Equal(t, "success", result["output"].(map[string]any)["result"])

	// Test tool error
	mockServer.callToolFunc = func(ctx context.Context, name string, args map[string]any) (*types.CallToolResult, error) {
		return nil, types.NewError("call_tool_error", "Failed to call tool")
	}

	req = httptest.NewRequest(http.MethodPost, "/tools/test-tool", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	res = httptest.NewRecorder()

	handler.ServeHTTP(res, req)
	assert.Equal(t, http.StatusOK, res.Code)

	err = json.Unmarshal(res.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.False(t, response["success"].(bool))

	errObj := response["error"].(map[string]any)
	assert.Equal(t, "call_tool_error", errObj["code"])
}

func TestHTTPHandler_ListResources(t *testing.T) {
	resources := []types.Resource{
		{Name: "resource1", Description: "Resource 1", Type: "text/plain"},
		{Name: "resource2", Description: "Resource 2", Type: "application/json"},
	}

	mockServer := &MockServer{
		listResourcesFunc: func(ctx context.Context) ([]types.Resource, error) {
			return resources, nil
		},
	}

	handler := newHTTPHandler(mockServer)

	// Test successful listing
	req := httptest.NewRequest(http.MethodGet, "/resources", http.NoBody)
	res := httptest.NewRecorder()

	handler.ServeHTTP(res, req)
	assert.Equal(t, http.StatusOK, res.Code)

	var response map[string]any
	err := json.Unmarshal(res.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.True(t, response["success"].(bool))

	result := response["result"].([]any)
	assert.Len(t, result, 2)

	// Test listing error
	mockServer.listResourcesFunc = func(ctx context.Context) ([]types.Resource, error) {
		return nil, types.NewError("list_resources_error", "Failed to list resources")
	}

	req = httptest.NewRequest(http.MethodGet, "/resources", http.NoBody)
	res = httptest.NewRecorder()

	handler.ServeHTTP(res, req)
	assert.Equal(t, http.StatusOK, res.Code)

	err = json.Unmarshal(res.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.False(t, response["success"].(bool))

	errObj := response["error"].(map[string]any)
	assert.Equal(t, "list_resources_error", errObj["code"])
}

func TestHTTPHandler_ReadResource(t *testing.T) {
	mockServer := &MockServer{
		readResourceFunc: func(ctx context.Context, name string) ([]byte, string, error) {
			return []byte("test content"), "text/plain", nil
		},
	}

	handler := newHTTPHandler(mockServer)

	// Test successful resource read
	req := httptest.NewRequest(http.MethodGet, "/resources/test-resource", http.NoBody)
	res := httptest.NewRecorder()

	handler.ServeHTTP(res, req)
	assert.Equal(t, http.StatusOK, res.Code)
	assert.Equal(t, "text/plain", res.Header().Get("Content-Type"))
	assert.Equal(t, "test content", res.Body.String())

	// Test resource error
	mockServer.readResourceFunc = func(ctx context.Context, name string) ([]byte, string, error) {
		return nil, "", types.NewError("read_resource_error", "Failed to read resource")
	}

	req = httptest.NewRequest(http.MethodGet, "/resources/test-resource", http.NoBody)
	res = httptest.NewRecorder()

	handler.ServeHTTP(res, req)
	assert.Equal(t, http.StatusOK, res.Code)

	var response map[string]any
	err := json.Unmarshal(res.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.False(t, response["success"].(bool))

	errObj := response["error"].(map[string]any)
	assert.Equal(t, "read_resource_error", errObj["code"])
}
