package transport

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/virgoC0der/go-mcp/internal/response"
	"github.com/virgoC0der/go-mcp/internal/types"
)

// callRecord holds information about a method call
type callRecord struct {
	Method string
	Params interface{}
}

// recordingMockServer is a mock implementation of types.Server that records calls
type recordingMockServer struct {
	mu          sync.Mutex
	calls       []callRecord
	nextResult  interface{}
	nextError   error
	initialized bool
}

// --- Helper methods for mock control ---

func (m *recordingMockServer) SetNextResponse(result interface{}, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.nextResult = result
	m.nextError = err
}

func (m *recordingMockServer) GetCalls() []callRecord {
	m.mu.Lock()
	defer m.mu.Unlock()
	// Return a copy to prevent race conditions
	callsCopy := make([]callRecord, len(m.calls))
	copy(callsCopy, m.calls)
	return callsCopy
}

func (m *recordingMockServer) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.calls = nil
	m.nextResult = nil
	m.nextError = nil
	m.initialized = false
}

func (m *recordingMockServer) recordCall(method string, params interface{}) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.calls = append(m.calls, callRecord{Method: method, Params: params})
}

// --- Implementation of types.Server ---

func (m *recordingMockServer) Initialize(ctx context.Context, options any) error {
	m.recordCall("Initialize", options)
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.nextError != nil {
		return m.nextError
	}
	m.initialized = true
	return nil
}

func (m *recordingMockServer) Start() error {
	m.recordCall("Start", nil)
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.nextError
}

func (m *recordingMockServer) Shutdown(ctx context.Context) error {
	m.recordCall("Shutdown", nil)
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.nextError
}

// --- Implementation of types.MCPService ---

func (m *recordingMockServer) ListPrompts(ctx context.Context, cursor string) (*types.PromptListResult, error) {
	m.recordCall("ListPrompts", map[string]interface{}{"cursor": cursor})
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.nextError != nil {
		return nil, m.nextError
	}
	if result, ok := m.nextResult.(*types.PromptListResult); ok {
		return result, nil
	}
	return nil, errors.New("mock configuration error: nextResult type mismatch for ListPrompts")
}

func (m *recordingMockServer) GetPrompt(ctx context.Context, name string, args map[string]any) (*types.PromptResult, error) {
	m.recordCall("GetPrompt", map[string]interface{}{"name": name, "args": args})
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.nextError != nil {
		return nil, m.nextError
	}
	if result, ok := m.nextResult.(*types.PromptResult); ok {
		return result, nil
	}
	return nil, errors.New("mock configuration error: nextResult type mismatch for GetPrompt")
}

func (m *recordingMockServer) ListTools(ctx context.Context, cursor string) (*types.ToolListResult, error) {
	m.recordCall("ListTools", map[string]interface{}{"cursor": cursor})
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.nextError != nil {
		return nil, m.nextError
	}
	if result, ok := m.nextResult.(*types.ToolListResult); ok {
		return result, nil
	}
	return nil, errors.New("mock configuration error: nextResult type mismatch for ListTools")
}

func (m *recordingMockServer) CallTool(ctx context.Context, name string, args map[string]any) (*types.CallToolResult, error) {
	m.recordCall("CallTool", map[string]interface{}{"name": name, "args": args})
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.nextError != nil {
		return nil, m.nextError
	}
	if result, ok := m.nextResult.(*types.CallToolResult); ok {
		return result, nil
	}
	return nil, errors.New("mock configuration error: nextResult type mismatch for CallTool")
}

func (m *recordingMockServer) ListResources(ctx context.Context, cursor string) (*types.ResourceListResult, error) {
	m.recordCall("ListResources", map[string]interface{}{"cursor": cursor})
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.nextError != nil {
		return nil, m.nextError
	}
	if result, ok := m.nextResult.(*types.ResourceListResult); ok {
		return result, nil
	}
	return nil, errors.New("mock configuration error: nextResult type mismatch for ListResources")
}

func (m *recordingMockServer) ReadResource(ctx context.Context, uri string) (*types.ResourceContent, error) {
	m.recordCall("ReadResource", map[string]interface{}{"uri": uri})
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.nextError != nil {
		return nil, m.nextError
	}
	if result, ok := m.nextResult.(*types.ResourceContent); ok {
		return result, nil
	}
	return nil, errors.New("mock configuration error: nextResult type mismatch for ReadResource")
}

func (m *recordingMockServer) ListResourceTemplates(ctx context.Context) ([]types.ResourceTemplate, error) {
	m.recordCall("ListResourceTemplates", nil)
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.nextError != nil {
		return nil, m.nextError
	}
	if result, ok := m.nextResult.([]types.ResourceTemplate); ok {
		return result, nil
	}
	return nil, errors.New("mock configuration error: nextResult type mismatch for ListResourceTemplates")
}

func (m *recordingMockServer) SubscribeToResource(ctx context.Context, uri string) error {
	m.recordCall("SubscribeToResource", map[string]interface{}{"uri": uri})
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.nextError
}

func TestTryParseClaudeMessage(t *testing.T) {
	tests := []struct {
		name        string
		inputJSON   string
		expectMsg   *response.JSONRPCRequest // Using alias StdioMessage might cause import cycle if test is in response pkg
		expectError bool
	}{
		{
			name:      "standard json-rpc initialize",
			inputJSON: `{"jsonrpc": "2.0", "id": 1, "method": "initialize", "params": {"capabilities": {}}}`, // Standard JSON-RPC should ideally be handled by direct unmarshal, but this tests the fallback logic
			expectMsg: &response.JSONRPCRequest{
				JSONRPC: "2.0",
				ID:      1,
				Method:  "initialize",
				Params:  json.RawMessage(`{"capabilities": {}}`),
			},
			expectError: false,
		},
		{
			name:      "claude initialize without method",
			inputJSON: `{"id": 2, "params": {"clientName": "claude"}}`, // Claude-like init
			expectMsg: &response.JSONRPCRequest{
				JSONRPC: "2.0", // Defaulted
				ID:      2,
				Method:  "initialize",
				Params:  json.RawMessage(`{}`), // Params are currently ignored/reset for inferred initialize
			},
			expectError: false,
		},
		{
			name:      "claude message with content (tool call inference)",
			inputJSON: `{"role": "user", "content": [{"type":"text", "text":"call the tool"}]}`, // Claude message that might be inferred as a tool call
			expectMsg: &response.JSONRPCRequest{
				JSONRPC: "2.0",
				ID:      1, // Default ID
				Method:  "callTool",
				Params:  json.RawMessage(`{"args":{"content":[{"type":"text","text":"call the tool"}],"rawMessage":{"content":[{"type":"text","text":"call the tool"}],"role":"user"}},"name":"default"}`),
			},
			expectError: false,
		},
		{
			name:      "claude assistant message (getPrompt inference)",
			inputJSON: `{"role": "assistant", "content": "some response"}`, // Claude assistant response might infer getPrompt
			expectMsg: &response.JSONRPCRequest{
				JSONRPC: "2.0",
				ID:      1, // Default ID
				Method:  "getPrompt",
				Params:  json.RawMessage(`{"args":{"content":"some response","rawMessage":{"content":"some response","role":"assistant"}},"name":"default"}`), // Params structure is complex due to inference
			},
			expectError: false,
		},
		{
			name:      "empty json",
			inputJSON: `{}`, // Empty object, should default to listTools
			expectMsg: &response.JSONRPCRequest{
				JSONRPC: "2.0",
				ID:      1, // Default ID
				Method:  "listTools",
				Params:  json.RawMessage(`{}`),
			},
			expectError: false,
		},
		{
			name:        "invalid json",
			inputJSON:   `{invalid`, // Malformed JSON
			expectMsg:   nil,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg, err := tryParseClaudeMessage([]byte(tt.inputJSON))

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, msg)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, msg)
				// Compare relevant fields, Params might be tricky due to marshaling differences
				assert.Equal(t, tt.expectMsg.JSONRPC, msg.JSONRPC)
				assert.Equal(t, tt.expectMsg.ID, msg.ID)
				assert.Equal(t, tt.expectMsg.Method, msg.Method)
				assert.JSONEq(t, string(tt.expectMsg.Params), string(msg.Params))
			}
		})
	}
}

func TestStdioServer_HandleMessage(t *testing.T) {
	mockSrv := &recordingMockServer{}
	var reader, writer bytes.Buffer

	stdioServer := NewStdioServerWithIO(mockSrv, &reader, &writer)

	tests := []struct {
		name            string
		inputMsg        string      // JSON message to send
		mockResult      interface{} // Result to preset in mock
		mockError       error       // Error to preset in mock
		expectedMethod  string      // Expected method called on mock
		expectedParams  interface{} // Expected params for the method call
		expectOutputSub string      // Substring expected in the output JSON
		expectOutputID  interface{} // Expected ID in the output JSON
		expectErrorResp bool        // Expect an error JSON-RPC response
		resetMock       bool        // Reset mock before test
	}{
		{
			name:            "handle initialize success",
			inputMsg:        `{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"clientName":"testClient"}}`,
			mockResult:      nil, // Initialize returns error only
			mockError:       nil,
			expectedMethod:  "Initialize",
			expectedParams:  map[string]interface{}{"clientName": "testClient"},
			expectOutputSub: `"result":{"capabilities":`, // Check for capabilities in result
			expectOutputID:  1,
			expectErrorResp: false,
			resetMock:       true,
		},
		{
			name:            "handle initialize error",
			inputMsg:        `{"jsonrpc":"2.0","id":2,"method":"initialize","params":{}}`,
			mockResult:      nil,
			mockError:       errors.New("init failed"),
			expectedMethod:  "Initialize",
			expectedParams:  map[string]interface{}{},
			expectOutputSub: `"error":{"code":-32603,"message":"Initialization failed: init failed"`, // Check for error message
			expectOutputID:  2,
			expectErrorResp: true,
			resetMock:       true,
		},
		{
			name:            "handle listTools success",
			inputMsg:        `{"jsonrpc":"2.0","id":3,"method":"tools/list","params":{"cursor":"abc"}}`,
			mockResult:      &types.ToolListResult{Tools: []types.Tool{{Name: "tool1"}}, NextCursor: "def"},
			mockError:       nil,
			expectedMethod:  "ListTools",
			expectedParams:  map[string]interface{}{"cursor": "abc"},
			expectOutputSub: `"result":{"tools":[{"name":"tool1"}],"nextCursor":"def"}`, // Check result structure
			expectOutputID:  3,
			expectErrorResp: false,
			resetMock:       true,
		},
		{
			name:            "handle callTool error",
			inputMsg:        `{"jsonrpc":"2.0","id":4,"method":"tools/call","params":{"name":"badTool"}}`,
			mockResult:      nil,
			mockError:       errors.New("tool execution failed"),
			expectedMethod:  "CallTool",
			expectedParams:  map[string]interface{}{"name": "badTool", "args": map[string]interface{}(nil)}, // Args default to nil map if not provided
			expectOutputSub: `"error":{"code":-32603,"message":"CallTool failed: tool execution failed"`,    // Check error message
			expectOutputID:  4,
			expectErrorResp: true,
			resetMock:       true,
		},
		{
			name:            "invalid method",
			inputMsg:        `{"jsonrpc":"2.0","id":5,"method":"unknown/method"}`,
			mockResult:      nil,
			mockError:       nil,
			expectedMethod:  "", // No call to mock server
			expectedParams:  nil,
			expectOutputSub: `"error":{"code":-32603,"message":"Unknown method: unknown/method"`, // Check error message
			expectOutputID:  5,
			expectErrorResp: true,
			resetMock:       true,
		},
		{
			name:            "invalid json format",
			inputMsg:        `{invalid json`, // Malformed JSON
			mockResult:      nil,
			mockError:       nil,
			expectedMethod:  "", // No call to mock server
			expectedParams:  nil,
			expectOutputSub: `"error":{"code":-32603,"message":"Invalid message format:`, // Check error message
			expectOutputID:  0,                                                           // Default ID for parse errors
			expectErrorResp: true,
			resetMock:       true,
		},
		{
			name:            "handle subscribe success",
			inputMsg:        `{"jsonrpc":"2.0","id":6,"method":"resources/subscribe","params":{"uri":"file:///a.txt"}}`,
			mockResult:      nil, // Subscribe returns error only
			mockError:       nil,
			expectedMethod:  "SubscribeToResource",
			expectedParams:  map[string]interface{}{"uri": "file:///a.txt"},
			expectOutputSub: `"result":{}`, // Empty result object for success
			expectOutputID:  6,
			expectErrorResp: false,
			resetMock:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.resetMock {
				mockSrv.Reset()
				mockSrv.SetNextResponse(tt.mockResult, tt.mockError)
			}
			writer.Reset() // Clear output buffer

			// Simulate receiving the message
			stdioServer.handleMessage([]byte(tt.inputMsg))

			// Check mock server calls
			calls := mockSrv.GetCalls()
			if tt.expectedMethod != "" {
				assert.Len(t, calls, 1, "Expected exactly one call to mock server")
				assert.Equal(t, tt.expectedMethod, calls[0].Method)
				// Use assert.Equal for params check; requires careful construction of expected map
				assert.Equal(t, tt.expectedParams, calls[0].Params)
			} else {
				assert.Empty(t, calls, "Expected no calls to mock server")
			}

			// Check output written to writer
			outputBytes := writer.Bytes()
			assert.NotEmpty(t, outputBytes, "Expected output to be written")

			var outputResp response.JSONRPCResponse
			err := json.Unmarshal(outputBytes, &outputResp)
			assert.NoError(t, err, "Failed to unmarshal output response JSON")

			assert.Equal(t, "2.0", outputResp.JSONRPC)
			assert.Equal(t, tt.expectOutputID, outputResp.ID)

			if tt.expectErrorResp {
				assert.NotNil(t, outputResp.Error, "Expected error field in response")
				assert.Nil(t, outputResp.Result, "Expected nil result field in error response")
				assert.Contains(t, string(outputBytes), tt.expectOutputSub, "Expected error substring not found")
			} else {
				assert.Nil(t, outputResp.Error, "Expected nil error field in success response")
				assert.NotNil(t, outputResp.Result, "Expected result field in success response")
				assert.Contains(t, string(outputBytes), tt.expectOutputSub, "Expected result substring not found")
			}
		})
	}
}
