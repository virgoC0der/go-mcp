package sse

//
//import (
//	"context"
//	"encoding/json"
//	"net/http"
//	"net/http/httptest"
//	"testing"
//	"time"
//
//	"github.com/stretchr/testify/assert"
//	"github.com/virgoC0der/go-mcp/internal/types"
//)
//
//func TestSSEClient_Connect(t *testing.T) {
//	done := make(chan struct{})
//
//	// 创建一个测试服务器
//	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
//		w.Header().Set("Content-Type", "text/event-stream")
//		w.Header().Set("Cache-Control", "no-cache")
//		w.Header().Set("Connection", "keep-alive")
//		flusher, ok := w.(http.Flusher)
//		if !ok {
//			t.Fatal("Streaming unsupported")
//		}
//
//		// 发送一个测试事件
//		response := types.NewSuccessResponse(map[string]interface{}{
//			"message": "test event",
//		})
//		data, _ := json.Marshal(response)
//
//		_, err := w.Write([]byte("data: " + string(data) + "\n\n"))
//		if err != nil {
//			t.Fatal(err)
//		}
//		flusher.Flush()
//
//		// 等待测试完成后关闭连接
//		<-done
//	}))
//	defer server.Close()
//
//	// 创建客户端
//	client := NewSSEClient(server.URL)
//
//	// 启动连接
//	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
//	defer cancel()
//
//	err := client.Connect(ctx)
//	assert.NoError(t, err)
//
//	// 等待并验证事件
//	select {
//	case event := <-client.Events():
//		assert.True(t, event.Success)
//		result := event.Result.(map[string]interface{})
//		assert.Equal(t, "test event", result["message"])
//	case <-time.After(1 * time.Second):
//		t.Fatal("Timeout waiting for event")
//	}
//
//	// 测试重复连接
//	err = client.Connect(ctx)
//	assert.Error(t, err)
//	assert.Contains(t, err.Error(), "client already connected")
//
//	// 关闭连接
//	err = client.Close()
//	assert.NoError(t, err)
//
//	// 通知服务器关闭连接
//	close(done)
//
//	// 等待连接完全关闭
//	time.Sleep(100 * time.Millisecond)
//}
//
//func TestSSEClient_API(t *testing.T) {
//	// 创建一个测试服务器来模拟 API 响应
//	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
//		w.Header().Set("Content-Type", "application/json")
//
//		var response types.Response
//		switch r.URL.Path {
//		case "/api/prompts":
//			prompts := []types.Prompt{{Name: "test", Description: "Test Prompt"}}
//			response = types.NewSuccessResponse(prompts)
//		case "/api/prompts/test":
//			result := &types.GetPromptResult{Content: "Test Content"}
//			response = types.NewSuccessResponse(result)
//		case "/api/tools":
//			tools := []types.Tool{{Name: "test", Description: "Test Tool"}}
//			response = types.NewSuccessResponse(tools)
//		case "/api/tools/test":
//			result := &types.CallToolResult{Output: "Test Output"}
//			response = types.NewSuccessResponse(result)
//		case "/api/resources":
//			resources := []types.Resource{{Name: "test", Description: "Test Resource"}}
//			response = types.NewSuccessResponse(resources)
//		case "/api/resources/test":
//			response = types.NewSuccessResponse(map[string]interface{}{
//				"content":      "Test Content",
//				"content_type": "text/plain",
//			})
//		default:
//			response = types.NewErrorResponse("not_found", "Endpoint not found")
//			w.WriteHeader(http.StatusNotFound)
//		}
//
//		json.NewEncoder(w).Encode(response)
//	}))
//	defer server.Close()
//
//	client := NewSSEClient(server.URL)
//	ctx := context.Background()
//
//	// 测试 ListPrompts
//	t.Run("ListPrompts", func(t *testing.T) {
//		prompts, err := client.ListPrompts(ctx)
//		assert.NoError(t, err)
//		assert.Len(t, prompts, 1)
//		assert.Equal(t, "test", prompts[0].Name)
//	})
//
//	// 测试 GetPrompt
//	t.Run("GetPrompt", func(t *testing.T) {
//		result, err := client.GetPrompt(ctx, "test", nil)
//		assert.NoError(t, err)
//		assert.Equal(t, "Test Content", result.Content)
//	})
//
//	// 测试 ListTools
//	t.Run("ListTools", func(t *testing.T) {
//		tools, err := client.ListTools(ctx)
//		assert.NoError(t, err)
//		assert.Len(t, tools, 1)
//		assert.Equal(t, "test", tools[0].Name)
//	})
//
//	// 测试 CallTool
//	t.Run("CallTool", func(t *testing.T) {
//		result, err := client.CallTool(ctx, "test", nil)
//		assert.NoError(t, err)
//		assert.Equal(t, "Test Output", result.Output)
//	})
//
//	// 测试 ListResources
//	t.Run("ListResources", func(t *testing.T) {
//		resources, err := client.ListResources(ctx)
//		assert.NoError(t, err)
//		assert.Len(t, resources, 1)
//		assert.Equal(t, "test", resources[0].Name)
//	})
//
//	// 测试 ReadResource
//	t.Run("ReadResource", func(t *testing.T) {
//		content, contentType, err := client.ReadResource(ctx, "test")
//		assert.NoError(t, err)
//		assert.Equal(t, "Test Content", string(content))
//		assert.Equal(t, "text/plain", contentType)
//	})
//}
//
//func TestSSEClient_Error_Handling(t *testing.T) {
//	// 创建一个总是返回错误的测试服务器
//	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
//		w.Header().Set("Content-Type", "application/json")
//		response := types.NewErrorResponse("test_error", "Test error message")
//		json.NewEncoder(w).Encode(response)
//	}))
//	defer server.Close()
//
//	client := NewSSEClient(server.URL)
//	ctx := context.Background()
//
//	// 测试错误处理
//	_, err := client.ListPrompts(ctx)
//	assert.Error(t, err)
//	assert.Contains(t, err.Error(), "test_error")
//	assert.Contains(t, err.Error(), "Test error message")
//}
