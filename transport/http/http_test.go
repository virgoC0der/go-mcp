package http

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/virgoC0der/go-mcp/internal/types"
	"github.com/virgoC0der/go-mcp/transport/mock"
)

func init() {
	gin.SetMode(gin.TestMode)
}

// getFreePort 获取一个可用的端口
func getFreePort() (int, error) {
	addr, err := net.ResolveTCPAddr("tcp", "localhost:0")
	if err != nil {
		return 0, err
	}

	l, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return 0, err
	}
	defer l.Close()
	return l.Addr().(*net.TCPAddr).Port, nil
}

func TestHTTPServer_Lifecycle(t *testing.T) {
	port, err := getFreePort()
	if err != nil {
		t.Fatal(err)
	}

	mockServer := &mock.MockServer{
		InitializeFunc: func(ctx context.Context, options any) error {
			return nil
		},
		StartFunc: func() error {
			return nil
		},
		ShutdownFunc: func(ctx context.Context) error {
			return nil
		},
	}

	server := NewHTTPServer(mockServer, fmt.Sprintf(":%d", port))
	assert.NotNil(t, server)

	// Test Initialize
	err = server.Initialize(context.Background(), nil)
	assert.NoError(t, err)

	// Test Start (in goroutine since it blocks)
	go func() {
		_ = server.Start()
	}()

	// Give server time to start
	time.Sleep(100 * time.Millisecond)

	// Test Shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = server.Shutdown(ctx)
	assert.NoError(t, err)
}

func TestHTTPHandler_ListPrompts(t *testing.T) {
	mockServer := &mock.MockServer{
		ListPromptsFunc: func(ctx context.Context) ([]types.Prompt, error) {
			return []types.Prompt{
				{Name: "prompt1", Description: "Test Prompt 1"},
				{Name: "prompt2", Description: "Test Prompt 2"},
			}, nil
		},
	}

	handler := newHTTPHandler(mockServer)
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = httptest.NewRequest("GET", "/prompts", nil)

	handler.handlePrompts(ctx)

	assert.Equal(t, http.StatusOK, w.Code)

	var response types.Response
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.True(t, response.Success)

	var prompts []types.Prompt
	data, err := json.Marshal(response.Result)
	assert.NoError(t, err)
	err = json.Unmarshal(data, &prompts)
	assert.NoError(t, err)
	assert.Len(t, prompts, 2)
	assert.Equal(t, "prompt1", prompts[0].Name)
}

func TestHTTPHandler_GetPrompt(t *testing.T) {
	mockServer := &mock.MockServer{
		GetPromptFunc: func(ctx context.Context, name string, args map[string]any) (*types.GetPromptResult, error) {
			return &types.GetPromptResult{
				Content: "Test Content",
			}, nil
		},
	}

	handler := newHTTPHandler(mockServer)
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = httptest.NewRequest("POST", "/prompts/prompt1", bytes.NewReader([]byte(`{"arguments":{}}`)))
	ctx.Params = []gin.Param{{Key: "name", Value: "prompt1"}}

	handler.handlePrompt(ctx)

	assert.Equal(t, http.StatusOK, w.Code)

	var response types.Response
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.True(t, response.Success)

	var result types.GetPromptResult
	data, err := json.Marshal(response.Result)
	assert.NoError(t, err)
	err = json.Unmarshal(data, &result)
	assert.NoError(t, err)
	assert.Equal(t, "Test Content", result.Content)
}

func TestHTTPHandler_ListTools(t *testing.T) {
	mockServer := &mock.MockServer{
		ListToolsFunc: func(ctx context.Context) ([]types.Tool, error) {
			return []types.Tool{
				{Name: "tool1", Description: "Test Tool 1"},
				{Name: "tool2", Description: "Test Tool 2"},
			}, nil
		},
	}

	handler := newHTTPHandler(mockServer)
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = httptest.NewRequest("GET", "/tools", nil)

	handler.handleTools(ctx)

	assert.Equal(t, http.StatusOK, w.Code)

	var response types.Response
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.True(t, response.Success)

	var tools []types.Tool
	data, err := json.Marshal(response.Result)
	assert.NoError(t, err)
	err = json.Unmarshal(data, &tools)
	assert.NoError(t, err)
	assert.Len(t, tools, 2)
	assert.Equal(t, "tool1", tools[0].Name)
}

func TestHTTPHandler_CallTool(t *testing.T) {
	mockServer := &mock.MockServer{
		CallToolFunc: func(ctx context.Context, name string, args map[string]any) (*types.CallToolResult, error) {
			return &types.CallToolResult{
				Output: map[string]interface{}{
					"result": "success",
				},
			}, nil
		},
	}

	handler := newHTTPHandler(mockServer)
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)

	body := map[string]interface{}{
		"arguments": map[string]interface{}{
			"param1": "value1",
		},
	}
	bodyBytes, _ := json.Marshal(body)
	ctx.Request = httptest.NewRequest("POST", "/tools/tool1", bytes.NewReader(bodyBytes))
	ctx.Params = []gin.Param{{Key: "name", Value: "tool1"}}

	handler.handleTool(ctx)

	assert.Equal(t, http.StatusOK, w.Code)

	var response types.Response
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.True(t, response.Success)

	var result types.CallToolResult
	data, err := json.Marshal(response.Result)
	assert.NoError(t, err)
	err = json.Unmarshal(data, &result)
	assert.NoError(t, err)

	output, ok := result.Output.(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, "success", output["result"])
}

func TestHTTPHandler_ListResources(t *testing.T) {
	mockServer := &mock.MockServer{
		ListResourcesFunc: func(ctx context.Context) ([]types.Resource, error) {
			return []types.Resource{
				{Name: "resource1", Description: "Resource 1", Type: "text/plain"},
				{Name: "resource2", Description: "Resource 2", Type: "application/json"},
			}, nil
		},
	}

	handler := newHTTPHandler(mockServer)
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = httptest.NewRequest("GET", "/resources", nil)

	handler.handleResources(ctx)

	assert.Equal(t, http.StatusOK, w.Code)

	var response types.Response
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.True(t, response.Success)

	var resources []types.Resource
	data, err := json.Marshal(response.Result)
	assert.NoError(t, err)
	err = json.Unmarshal(data, &resources)
	assert.NoError(t, err)
	assert.Len(t, resources, 2)
	assert.Equal(t, "resource1", resources[0].Name)
}

func TestHTTPHandler_ReadResource(t *testing.T) {
	mockServer := &mock.MockServer{
		ReadResourceFunc: func(ctx context.Context, name string) ([]byte, string, error) {
			return []byte("Test Content"), "text/plain", nil
		},
	}

	handler := newHTTPHandler(mockServer)
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = httptest.NewRequest("GET", "/resources/test", nil)
	ctx.Params = []gin.Param{{Key: "name", Value: "test"}}

	handler.handleResource(ctx)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "text/plain", w.Header().Get("Content-Type"))
	assert.Equal(t, "Test Content", w.Body.String())
}
