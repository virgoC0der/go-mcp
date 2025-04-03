package sse

import "net"

// import (
//
//	"bytes"
//	"context"
//	"encoding/json"
//	"fmt"
//	"io"
//	"net"
//	"net/http"
//	"net/http/httptest"
//	"testing"
//	"time"
//
//	"bufio"
//
//	"github.com/gin-gonic/gin"
//	"github.com/stretchr/testify/assert"
//	"github.com/virgoC0der/go-mcp/internal/types"
//
// )
//
//	func init() {
//		gin.SetMode(gin.TestMode)
//	}
//
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

//
//func TestSSEServer_Events(t *testing.T) {
//	mockServer := &MockServer{
//		startFunc: func() error {
//			return nil
//		},
//		shutdownFunc: func(ctx context.Context) error {
//			return nil
//		},
//	}
//
//	server := NewSSEServer(mockServer, ":0")
//	assert.NotNil(t, server)
//
//	// 创建一个测试请求和响应记录器
//	w := httptest.NewRecorder()
//	ctx, _ := gin.CreateTestContext(w)
//
//	// 创建一个带取消的上下文
//	reqCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
//	defer cancel()
//
//	// 创建一个测试请求
//	req := httptest.NewRequest("GET", "/events", nil)
//	req = req.WithContext(reqCtx)
//	ctx.Request = req
//
//	// 创建一个通道来等待测试完成
//	done := make(chan struct{})
//
//	// 在后台处理 SSE 请求
//	go func() {
//		server.handleSSE(ctx)
//		close(done)
//	}()
//
//	// 等待连接建立
//	time.Sleep(100 * time.Millisecond)
//
//	// 广播一条测试消息
//	testData := map[string]interface{}{
//		"message": "test",
//	}
//	server.Broadcast("test", testData)
//
//	// 等待消息发送
//	time.Sleep(100 * time.Millisecond)
//
//	// 取消请求
//	cancel()
//
//	// 等待请求处理完成
//	select {
//	case <-done:
//		// 测试成功完成
//	case <-time.After(3 * time.Second):
//		t.Fatal("Test timed out")
//	}
//
//	// 验证响应
//	resp := w.Body.String()
//	assert.Contains(t, resp, "data: ")
//	assert.Contains(t, resp, "test")
//}
//
//func TestSSEServer_API_Endpoints(t *testing.T) {
//	mockServer := &MockServer{
//		listPromptsFunc: func(ctx context.Context) ([]types.Prompt, error) {
//			return []types.Prompt{{Name: "test", Description: "Test Prompt"}}, nil
//		},
//		getPromptFunc: func(ctx context.Context, name string, args map[string]any) (*types.GetPromptResult, error) {
//			return &types.GetPromptResult{Content: "Test Content"}, nil
//		},
//		listToolsFunc: func(ctx context.Context) ([]types.Tool, error) {
//			return []types.Tool{{Name: "test", Description: "Test Tool"}}, nil
//		},
//		callToolFunc: func(ctx context.Context, name string, args map[string]any) (*types.CallToolResult, error) {
//			return &types.CallToolResult{Output: "Test Output"}, nil
//		},
//		listResourcesFunc: func(ctx context.Context) ([]types.Resource, error) {
//			return []types.Resource{{Name: "test", Description: "Test Resource"}}, nil
//		},
//		readResourceFunc: func(ctx context.Context, name string) ([]byte, string, error) {
//			return []byte("Test Content"), "text/plain", nil
//		},
//	}
//
//	server := NewSSEServer(mockServer, ":0")
//
//	tests := []struct {
//		name     string
//		method   string
//		path     string
//		body     interface{}
//		status   int
//		validate func(t *testing.T, response *httptest.ResponseRecorder)
//	}{
//		{
//			name:   "List Prompts",
//			method: "GET",
//			path:   "/api/prompts",
//			status: http.StatusOK,
//			validate: func(t *testing.T, w *httptest.ResponseRecorder) {
//				var response struct {
//					Result []types.Prompt `json:"result"`
//				}
//				err := json.Unmarshal(w.Body.Bytes(), &response)
//				assert.NoError(t, err)
//				assert.Len(t, response.Result, 1)
//				assert.Equal(t, "test", response.Result[0].Name)
//			},
//		},
//		{
//			name:   "Get Prompt",
//			method: "GET",
//			path:   "/api/prompts/test",
//			body:   map[string]interface{}{"message": "test message"},
//			status: http.StatusOK,
//			validate: func(t *testing.T, w *httptest.ResponseRecorder) {
//				var response struct {
//					Result *types.GetPromptResult `json:"result"`
//				}
//				err := json.Unmarshal(w.Body.Bytes(), &response)
//				assert.NoError(t, err)
//				assert.NotNil(t, response.Result)
//				assert.Equal(t, "Test Content", response.Result.Content)
//			},
//		},
//		{
//			name:   "List Tools",
//			method: "GET",
//			path:   "/api/tools",
//			status: http.StatusOK,
//			validate: func(t *testing.T, w *httptest.ResponseRecorder) {
//				var response struct {
//					Result []types.Tool `json:"result"`
//				}
//				err := json.Unmarshal(w.Body.Bytes(), &response)
//				assert.NoError(t, err)
//				assert.Len(t, response.Result, 1)
//				assert.Equal(t, "test", response.Result[0].Name)
//			},
//		},
//		{
//			name:   "Call Tool",
//			method: "POST",
//			path:   "/api/tools/test",
//			body:   map[string]interface{}{"arg": "value"},
//			status: http.StatusOK,
//			validate: func(t *testing.T, w *httptest.ResponseRecorder) {
//				var response struct {
//					Result *types.CallToolResult `json:"result"`
//				}
//				err := json.Unmarshal(w.Body.Bytes(), &response)
//				assert.NoError(t, err)
//				assert.Equal(t, "Test Output", response.Result.Output)
//			},
//		},
//		{
//			name:   "List Resources",
//			method: "GET",
//			path:   "/api/resources",
//			status: http.StatusOK,
//			validate: func(t *testing.T, w *httptest.ResponseRecorder) {
//				var response struct {
//					Result []types.Resource `json:"result"`
//				}
//				err := json.Unmarshal(w.Body.Bytes(), &response)
//				assert.NoError(t, err)
//				assert.Len(t, response.Result, 1)
//				assert.Equal(t, "test", response.Result[0].Name)
//			},
//		},
//		{
//			name:   "Read Resource",
//			method: "GET",
//			path:   "/api/resources/test",
//			status: http.StatusOK,
//			validate: func(t *testing.T, w *httptest.ResponseRecorder) {
//				assert.Equal(t, "text/plain", w.Header().Get("Content-Type"))
//				assert.Equal(t, "Test Content", w.Body.String())
//			},
//		},
//	}
//
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			w := httptest.NewRecorder()
//			req := httptest.NewRequest(tt.method, tt.path, nil)
//			if tt.body != nil {
//				bodyBytes, _ := json.Marshal(tt.body)
//				req.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
//				req.Header.Set("Content-Type", "application/json")
//			}
//
//			server.engine.ServeHTTP(w, req)
//
//			assert.Equal(t, tt.status, w.Code)
//			if tt.validate != nil {
//				tt.validate(t, w)
//			}
//		})
//	}
//}
//
//func TestSSEServer_Error_Handling(t *testing.T) {
//	mockServer := &MockServer{
//		listPromptsFunc: func(ctx context.Context) ([]types.Prompt, error) {
//			return nil, types.NewError("test_error", "Test error message")
//		},
//	}
//
//	server := NewSSEServer(mockServer, ":8080")
//
//	w := httptest.NewRecorder()
//	ctx, _ := gin.CreateTestContext(w)
//
//	req := httptest.NewRequest("GET", "/api/prompts", nil)
//	ctx.Request = req
//
//	server.engine.ServeHTTP(w, req)
//
//	assert.Equal(t, http.StatusBadRequest, w.Code)
//
//	var response struct {
//		Error struct {
//			Code    string `json:"code"`
//			Message string `json:"message"`
//		} `json:"error"`
//	}
//
//	err := json.Unmarshal(w.Body.Bytes(), &response)
//	assert.NoError(t, err)
//	assert.Equal(t, "test_error", response.Error.Code)
//	assert.Equal(t, "Test error message", response.Error.Message)
//}
//
//func TestSSEServer_Lifecycle(t *testing.T) {
//	// 获取一个可用端口
//	port, err := getFreePort()
//	if err != nil {
//		t.Fatal(err)
//	}
//
//	t.Logf("Using port: %d", port)
//
//	// 创建一个带状态的 mock 服务器
//	mockServer := &MockServer{
//		startFunc: func() error {
//			return nil
//		},
//		shutdownFunc: func(ctx context.Context) error {
//			return nil
//		},
//	}
//
//	// 创建 SSE 服务器
//	server := NewSSEServer(mockServer, fmt.Sprintf(":%d", port))
//	assert.NotNil(t, server)
//
//	// 在后台启动服务器
//	serverStarted := make(chan struct{})
//	go func() {
//		t.Log("Starting server...")
//		if err := server.Start(); err != nil && err != http.ErrServerClosed {
//			t.Errorf("Server error: %v", err)
//		}
//	}()
//
//	// 等待服务器启动
//	go func() {
//		// 尝试连接直到成功
//		for {
//			t.Log("Attempting to connect to server...")
//			conn, err := net.DialTimeout("tcp", fmt.Sprintf("localhost:%d", port), 100*time.Millisecond)
//			if err == nil {
//				t.Log("Server is up!")
//				conn.Close()
//				close(serverStarted)
//				return
//			}
//			t.Logf("Connection attempt failed: %v", err)
//			time.Sleep(100 * time.Millisecond)
//		}
//	}()
//
//	select {
//	case <-serverStarted:
//		t.Log("Server started successfully")
//	case <-time.After(5 * time.Second):
//		t.Fatal("Server did not start within timeout")
//	}
//
//	// 创建一个客户端连接
//	client := &http.Client{
//		Timeout: 2 * time.Second, // 减少超时时间，因为我们期望立即收到初始化消息
//	}
//
//	t.Log("Establishing SSE connection...")
//	req, err := http.NewRequest("GET", fmt.Sprintf("http://localhost:%d/events", port), nil)
//	assert.NoError(t, err)
//
//	resp, err := client.Do(req)
//	if err != nil {
//		t.Fatalf("Failed to connect to SSE: %v", err)
//	}
//	defer resp.Body.Close()
//
//	t.Log("SSE connection established")
//
//	// 验证响应头
//	assert.Equal(t, "text/event-stream", resp.Header.Get("Content-Type"))
//	assert.Equal(t, "no-cache", resp.Header.Get("Cache-Control"))
//	assert.Equal(t, "keep-alive", resp.Header.Get("Connection"))
//
//	// 创建一个新的客户端用于长连接
//	longClient := &http.Client{}
//	longReq, err := http.NewRequest("GET", fmt.Sprintf("http://localhost:%d/events", port), nil)
//	assert.NoError(t, err)
//
//	longResp, err := longClient.Do(longReq)
//	if err != nil {
//		t.Fatalf("Failed to establish long SSE connection: %v", err)
//	}
//	defer longResp.Body.Close()
//
//	// 广播一条测试消息
//	t.Log("Broadcasting test message...")
//	server.Broadcast("test", map[string]interface{}{
//		"message": "test message",
//	})
//
//	// 读取并验证消息
//	reader := bufio.NewReader(longResp.Body)
//
//	// 首先读取初始化消息
//	line, err := reader.ReadString('\n')
//	if err != nil {
//		t.Fatalf("Failed to read initial message: %v", err)
//	}
//	assert.Contains(t, line, ": connected")
//
//	// 然后读取测试消息
//	line, err = reader.ReadString('\n')
//	if err != nil {
//		t.Fatalf("Failed to read message: %v", err)
//	}
//	assert.Contains(t, line, "data: ")
//	assert.Contains(t, line, "test message")
//	t.Log("Test message received successfully")
//
//	// 关闭服务器
//	t.Log("Shutting down server...")
//	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
//	defer shutdownCancel()
//	assert.NoError(t, server.Shutdown(shutdownCtx))
//
//	// 验证服务器已关闭
//	t.Log("Verifying server shutdown...")
//	conn, err := net.DialTimeout("tcp", fmt.Sprintf("localhost:%d", port), 100*time.Millisecond)
//	if err == nil {
//		conn.Close()
//		t.Fatal("Server is still accepting connections")
//	}
//	t.Log("Server shutdown confirmed")
//
//	// 验证客户端连接已关闭
//	_, err = reader.ReadString('\n')
//	assert.Error(t, err)
//	t.Log("Client connection closed as expected")
//}
