package server

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/virgoC0der/go-mcp/internal/types"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func TestSSEServer_Lifecycle(t *testing.T) {
	mockServer := &MockServer{
		startFunc: func() error {
			return nil
		},
		shutdownFunc: func(ctx context.Context) error {
			return nil
		},
	}

	server := NewSSEServer(mockServer, ":8080")
	assert.NotNil(t, server)

	// 创建一个通道来同步服务器的启动和关闭
	serverStarted := make(chan struct{})
	serverStopped := make(chan struct{})

	// 在后台启动服务器
	go func() {
		time.Sleep(100 * time.Millisecond)
		close(serverStarted)

		err := server.Start()
		if err != nil && err.Error() != "http: Server closed" {
			t.Errorf("Server error: %v", err)
		}
		close(serverStopped)
	}()

	// 等待服务器启动
	select {
	case <-serverStarted:
		// 服务器已启动
	case <-time.After(2 * time.Second):
		t.Fatal("Server did not start within timeout")
	}

	// 创建一个带超时的上下文
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 关闭服务器
	err := server.Shutdown(ctx)
	assert.NoError(t, err)

	// 等待服务器完全停止
	select {
	case <-serverStopped:
		// 服务器已停止
	case <-time.After(5 * time.Second):
		t.Error("Server did not stop within timeout")
	}
}

func TestSSEServer_Events(t *testing.T) {
	mockServer := &MockServer{
		listPromptsFunc: func(ctx context.Context) ([]types.Prompt, error) {
			return []types.Prompt{
				{Name: "prompt1", Description: "Test Prompt 1"},
				{Name: "prompt2", Description: "Test Prompt 2"},
			}, nil
		},
	}

	server := NewSSEServer(mockServer, ":8080")

	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)

	// 模拟SSE请求
	req := httptest.NewRequest("GET", "/events", nil)
	ctx.Request = req

	// 创建一个通道来接收SSE处理完成的信号
	done := make(chan bool)

	go func() {
		server.handleSSE(ctx)
		done <- true
	}()

	// 广播一些测试数据
	server.Broadcast("test", map[string]interface{}{
		"message": "Hello, SSE!",
	})

	// 等待一小段时间让数据发送
	time.Sleep(100 * time.Millisecond)

	// 验证响应
	assert.Equal(t, "text/event-stream", w.Header().Get("Content-Type"))
	assert.Equal(t, "no-cache", w.Header().Get("Cache-Control"))
	assert.Equal(t, "keep-alive", w.Header().Get("Connection"))

	// 验证事件数据
	body := w.Body.String()
	assert.Contains(t, body, "event: message")
	assert.Contains(t, body, `"type":"test"`)
	assert.Contains(t, body, `"message":"Hello, SSE!"`)
}

func TestSSEServer_API_Endpoints(t *testing.T) {
	mockServer := &MockServer{
		listPromptsFunc: func(ctx context.Context) ([]types.Prompt, error) {
			return []types.Prompt{{Name: "test", Description: "Test Prompt"}}, nil
		},
		getPromptFunc: func(ctx context.Context, name string, args map[string]any) (*types.GetPromptResult, error) {
			return &types.GetPromptResult{Content: "Test Content"}, nil
		},
		listToolsFunc: func(ctx context.Context) ([]types.Tool, error) {
			return []types.Tool{{Name: "test", Description: "Test Tool"}}, nil
		},
		callToolFunc: func(ctx context.Context, name string, args map[string]any) (*types.CallToolResult, error) {
			return &types.CallToolResult{Output: "Test Output"}, nil
		},
		listResourcesFunc: func(ctx context.Context) ([]types.Resource, error) {
			return []types.Resource{{Name: "test", Description: "Test Resource"}}, nil
		},
		readResourceFunc: func(ctx context.Context, name string) ([]byte, string, error) {
			return []byte("Test Content"), "text/plain", nil
		},
	}

	server := NewSSEServer(mockServer, ":8080")

	tests := []struct {
		name     string
		method   string
		path     string
		body     interface{}
		status   int
		validate func(t *testing.T, response *httptest.ResponseRecorder)
	}{
		{
			name:   "List Prompts",
			method: "GET",
			path:   "/api/prompts",
			status: http.StatusOK,
			validate: func(t *testing.T, w *httptest.ResponseRecorder) {
				var prompts []types.Prompt
				err := json.Unmarshal(w.Body.Bytes(), &prompts)
				assert.NoError(t, err)
				assert.Len(t, prompts, 1)
				assert.Equal(t, "test", prompts[0].Name)
			},
		},
		{
			name:   "Get Prompt",
			method: "GET",
			path:   "/api/prompts/test",
			status: http.StatusOK,
			validate: func(t *testing.T, w *httptest.ResponseRecorder) {
				var result types.GetPromptResult
				err := json.Unmarshal(w.Body.Bytes(), &result)
				assert.NoError(t, err)
				assert.Equal(t, "Test Content", result.Content)
			},
		},
		{
			name:   "List Tools",
			method: "GET",
			path:   "/api/tools",
			status: http.StatusOK,
			validate: func(t *testing.T, w *httptest.ResponseRecorder) {
				var tools []types.Tool
				err := json.Unmarshal(w.Body.Bytes(), &tools)
				assert.NoError(t, err)
				assert.Len(t, tools, 1)
				assert.Equal(t, "test", tools[0].Name)
			},
		},
		{
			name:   "Call Tool",
			method: "POST",
			path:   "/api/tools/test",
			body:   map[string]interface{}{"arg": "value"},
			status: http.StatusOK,
			validate: func(t *testing.T, w *httptest.ResponseRecorder) {
				var result types.CallToolResult
				err := json.Unmarshal(w.Body.Bytes(), &result)
				assert.NoError(t, err)
				assert.Equal(t, "Test Output", result.Output)
			},
		},
		{
			name:   "List Resources",
			method: "GET",
			path:   "/api/resources",
			status: http.StatusOK,
			validate: func(t *testing.T, w *httptest.ResponseRecorder) {
				var resources []types.Resource
				err := json.Unmarshal(w.Body.Bytes(), &resources)
				assert.NoError(t, err)
				assert.Len(t, resources, 1)
				assert.Equal(t, "test", resources[0].Name)
			},
		},
		{
			name:   "Read Resource",
			method: "GET",
			path:   "/api/resources/test",
			status: http.StatusOK,
			validate: func(t *testing.T, w *httptest.ResponseRecorder) {
				assert.Equal(t, "text/plain", w.Header().Get("Content-Type"))
				assert.Equal(t, "Test Content", w.Body.String())
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(w)

			req := httptest.NewRequest(tt.method, tt.path, nil)
			if tt.body != nil {
				bodyBytes, _ := json.Marshal(tt.body)
				req.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
				req.Header.Set("Content-Type", "application/json")
			}
			ctx.Request = req

			server.engine.ServeHTTP(w, req)

			assert.Equal(t, tt.status, w.Code)
			if tt.validate != nil {
				tt.validate(t, w)
			}
		})
	}
}

func TestSSEServer_Error_Handling(t *testing.T) {
	mockServer := &MockServer{
		listPromptsFunc: func(ctx context.Context) ([]types.Prompt, error) {
			return nil, types.NewError("test_error", "Test error message")
		},
	}

	server := NewSSEServer(mockServer, ":8080")

	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)

	req := httptest.NewRequest("GET", "/api/prompts", nil)
	ctx.Request = req

	server.engine.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response struct {
		Error struct {
			Code    string `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}

	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "test_error", response.Error.Code)
	assert.Equal(t, "Test error message", response.Error.Message)
}
