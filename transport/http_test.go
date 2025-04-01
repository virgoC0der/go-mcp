package transport

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/virgoC0der/go-mcp/types"
)

func TestHTTPHandler_Initialize(t *testing.T) {
	// Create a mock server
	mockServer := &MockServer{
		initializeFunc: func(ctx context.Context, options any) error {
			return nil
		},
	}

	// Create HTTP handler
	handler := NewHTTPHandler(mockServer)

	// Create request
	body := `{"serverName": "test-server", "serverVersion": "1.0.0"}`
	req := httptest.NewRequest(http.MethodPost, "/initialize", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	// Create response recorder
	res := httptest.NewRecorder()

	// Handle request
	handler.ServeHTTP(res, req)

	// Check status code
	if res.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, res.Code)
	}

	// Check response body
	var response map[string]any
	err := json.Unmarshal(res.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if success, ok := response["success"].(bool); !ok || !success {
		t.Error("Expected success to be true")
	}

	// Test with initialize error
	mockServer.initializeFunc = func(ctx context.Context, options any) error {
		return types.NewError("initialize_error", "Failed to initialize")
	}

	// Create a new request (since the old one has been consumed)
	req = httptest.NewRequest(http.MethodPost, "/initialize", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	// Reset response recorder
	res = httptest.NewRecorder()

	// Handle request
	handler.ServeHTTP(res, req)

	// Check status code (should still be 200 with error in body)
	if res.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, res.Code)
	}

	// Check response body
	err = json.Unmarshal(res.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if success, ok := response["success"].(bool); !ok || success {
		t.Error("Expected success to be false")
	}

	if errObj, ok := response["error"].(map[string]any); !ok {
		t.Error("Expected error object in response")
	} else {
		if code, ok := errObj["code"].(string); !ok || code != "initialize_error" {
			t.Errorf("Expected error code 'initialize_error', got '%v'", code)
		}
	}
}

func TestHTTPHandler_ListPrompts(t *testing.T) {
	// Create mock prompts
	prompts := []types.Prompt{
		{Name: "prompt1", Description: "Prompt 1"},
		{Name: "prompt2", Description: "Prompt 2"},
	}

	// Create a mock server
	mockServer := &MockServer{
		listPromptsFunc: func(ctx context.Context) ([]types.Prompt, error) {
			return prompts, nil
		},
	}

	// Create HTTP handler
	handler := NewHTTPHandler(mockServer)

	// Create request
	req := httptest.NewRequest(http.MethodGet, "/prompts", http.NoBody)

	// Create response recorder
	res := httptest.NewRecorder()

	// Handle request
	handler.ServeHTTP(res, req)

	// Check status code
	if res.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, res.Code)
	}

	// Check response body
	var response map[string]any
	err := json.Unmarshal(res.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if success, ok := response["success"].(bool); !ok || !success {
		t.Error("Expected success to be true")
	}

	if result, ok := response["result"].([]any); !ok {
		t.Error("Expected result array in response")
	} else if len(result) != 2 {
		t.Errorf("Expected 2 prompts, got %d", len(result))
	}

	// Test with list prompts error
	mockServer.listPromptsFunc = func(ctx context.Context) ([]types.Prompt, error) {
		return nil, types.NewError("list_prompts_error", "Failed to list prompts")
	}

	// Reset response recorder
	res = httptest.NewRecorder()

	// Handle request
	handler.ServeHTTP(res, req)

	// Check status code
	if res.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, res.Code)
	}

	// Check response body
	err = json.Unmarshal(res.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if success, ok := response["success"].(bool); !ok || success {
		t.Error("Expected success to be false")
	}

	if errObj, ok := response["error"].(map[string]any); !ok {
		t.Error("Expected error object in response")
	} else {
		if code, ok := errObj["code"].(string); !ok || code != "list_prompts_error" {
			t.Errorf("Expected error code 'list_prompts_error', got '%v'", code)
		}
	}
}

func TestHTTPHandler_GetPrompt(t *testing.T) {
	// Create a mock server
	mockServer := &MockServer{
		getPromptFunc: func(ctx context.Context, name string, args map[string]any) (*types.GetPromptResult, error) {
			return &types.GetPromptResult{
				Description: "Test prompt",
				Message:     "Hello, world!",
			}, nil
		},
	}

	// Create HTTP handler
	handler := NewHTTPHandler(mockServer)

	// Create request
	body := `{"args": {"arg1": "value1"}}`
	req := httptest.NewRequest(http.MethodPost, "/prompts/test-prompt", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	// Create response recorder
	res := httptest.NewRecorder()

	// Handle request
	handler.ServeHTTP(res, req)

	// Check status code
	if res.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, res.Code)
	}

	// Check response body
	var response map[string]any
	err := json.Unmarshal(res.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if success, ok := response["success"].(bool); !ok || !success {
		t.Error("Expected success to be true")
	}

	if result, ok := response["result"].(map[string]any); !ok {
		t.Error("Expected result object in response")
	} else {
		if msg, ok := result["message"].(string); !ok || msg != "Hello, world!" {
			t.Errorf("Expected message 'Hello, world!', got '%v'", msg)
		}
	}

	// Test with get prompt error
	mockServer.getPromptFunc = func(ctx context.Context, name string, args map[string]any) (*types.GetPromptResult, error) {
		return nil, types.NewError("get_prompt_error", "Failed to get prompt")
	}

	// Create a new request (since the old one has been consumed)
	req = httptest.NewRequest(http.MethodPost, "/prompts/test-prompt", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	// Reset response recorder
	res = httptest.NewRecorder()

	// Handle request
	handler.ServeHTTP(res, req)

	// Check response body
	err = json.Unmarshal(res.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if success, ok := response["success"].(bool); !ok || success {
		t.Error("Expected success to be false")
	}

	if errObj, ok := response["error"].(map[string]any); !ok {
		t.Error("Expected error object in response")
	} else {
		if code, ok := errObj["code"].(string); !ok || code != "get_prompt_error" {
			t.Errorf("Expected error code 'get_prompt_error', got '%v'", code)
		}
	}
}

func TestHTTPHandler_ListTools(t *testing.T) {
	// Create mock tools
	tools := []types.Tool{
		{Name: "tool1", Description: "Tool 1"},
		{Name: "tool2", Description: "Tool 2"},
	}

	// Create a mock server
	mockServer := &MockServer{
		listToolsFunc: func(ctx context.Context) ([]types.Tool, error) {
			return tools, nil
		},
	}

	// Create HTTP handler
	handler := NewHTTPHandler(mockServer)

	// Create request
	req := httptest.NewRequest(http.MethodGet, "/tools", http.NoBody)

	// Create response recorder
	res := httptest.NewRecorder()

	// Handle request
	handler.ServeHTTP(res, req)

	// Check status code
	if res.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, res.Code)
	}

	// Check response body
	var response map[string]any
	err := json.Unmarshal(res.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if success, ok := response["success"].(bool); !ok || !success {
		t.Error("Expected success to be true")
	}

	if result, ok := response["result"].([]any); !ok {
		t.Error("Expected result array in response")
	} else if len(result) != 2 {
		t.Errorf("Expected 2 tools, got %d", len(result))
	}
}

func TestHTTPHandler_CallTool(t *testing.T) {
	// Create a mock server
	mockServer := &MockServer{
		callToolFunc: func(ctx context.Context, name string, args map[string]any) (*types.CallToolResult, error) {
			return &types.CallToolResult{
				Content: map[string]any{
					"message": "Tool executed successfully",
				},
			}, nil
		},
	}

	// Create HTTP handler
	handler := NewHTTPHandler(mockServer)

	// Create request
	body := `{"args": {"arg1": "value1"}}`
	req := httptest.NewRequest(http.MethodPost, "/tools/test-tool", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	// Create response recorder
	res := httptest.NewRecorder()

	// Handle request
	handler.ServeHTTP(res, req)

	// Check status code
	if res.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, res.Code)
	}

	// Check response body
	var response map[string]any
	err := json.Unmarshal(res.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if success, ok := response["success"].(bool); !ok || !success {
		t.Error("Expected success to be true")
	}

	if result, ok := response["result"].(map[string]any); !ok {
		t.Error("Expected result object in response")
	} else {
		if content, ok := result["content"].(map[string]any); !ok {
			t.Error("Expected content object in result")
		} else {
			if msg, ok := content["message"].(string); !ok || msg != "Tool executed successfully" {
				t.Errorf("Expected message 'Tool executed successfully', got '%v'", msg)
			}
		}
	}
}
