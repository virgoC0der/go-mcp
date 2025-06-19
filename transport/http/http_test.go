package http

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/virgoC0der/go-mcp/internal/response"
	"github.com/virgoC0der/go-mcp/internal/types"
)

// --- Mock MCP Service for http tests ---
type callRecord struct {
	Method string
	Params interface{}
}

type mockHttpService struct {
	mu          sync.Mutex
	calls       []callRecord
	nextResult  interface{}
	nextError   error
	initialized bool
	initOptions any
}

func (m *mockHttpService) SetNextResponse(result interface{}, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.nextResult = result
	m.nextError = err
}

func (m *mockHttpService) GetCalls() []callRecord {
	m.mu.Lock()
	defer m.mu.Unlock()
	callsCopy := make([]callRecord, len(m.calls))
	copy(callsCopy, m.calls)
	return callsCopy
}

func (m *mockHttpService) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.calls = nil
	m.nextResult = nil
	m.nextError = nil
	m.initialized = false
	m.initOptions = nil
}

func (m *mockHttpService) recordCall(method string, params interface{}) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.calls = append(m.calls, callRecord{Method: method, Params: params})
}

// Implement types.Server for Initialize check
func (m *mockHttpService) Initialize(ctx context.Context, options any) error {
	m.recordCall("Initialize", options)
	m.mu.Lock()
	defer m.mu.Unlock()
	m.initOptions = options
	if m.nextError != nil {
		return m.nextError
	}
	m.initialized = true
	return nil
}

func (m *mockHttpService) Start() error { m.recordCall("Start", nil); return nil }
func (m *mockHttpService) Shutdown(ctx context.Context) error {
	m.recordCall("Shutdown", nil)
	return nil
}

// Implement types.MCPService
func (m *mockHttpService) ListPrompts(ctx context.Context, cursor string) (*types.PromptListResult, error) {
	m.recordCall("ListPrompts", cursor)
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.nextError != nil {
		return nil, m.nextError
	}
	if res, ok := m.nextResult.(*types.PromptListResult); ok {
		return res, nil
	}
	return nil, errors.New("mock type mismatch")
}

func (m *mockHttpService) GetPrompt(ctx context.Context, name string, args map[string]any) (*types.PromptResult, error) {
	m.recordCall("GetPrompt", map[string]interface{}{"name": name, "args": args})
	// ... similar implementation for other MCPService methods ...
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.nextError != nil {
		return nil, m.nextError
	}
	if res, ok := m.nextResult.(*types.PromptResult); ok {
		return res, nil
	}
	return nil, errors.New("mock type mismatch")
}

func (m *mockHttpService) ListTools(ctx context.Context, cursor string) (*types.ToolListResult, error) {
	m.recordCall("ListTools", cursor)
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.nextError != nil {
		return nil, m.nextError
	}
	if res, ok := m.nextResult.(*types.ToolListResult); ok {
		return res, nil
	}
	return nil, errors.New("mock type mismatch")
}

func (m *mockHttpService) CallTool(ctx context.Context, name string, args map[string]any) (*types.CallToolResult, error) {
	m.recordCall("CallTool", map[string]interface{}{"name": name, "args": args})
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.nextError != nil {
		return nil, m.nextError
	}
	if res, ok := m.nextResult.(*types.CallToolResult); ok {
		return res, nil
	}
	return nil, errors.New("mock type mismatch")
}

func (m *mockHttpService) ListResources(ctx context.Context, cursor string) (*types.ResourceListResult, error) {
	m.recordCall("ListResources", cursor)
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.nextError != nil {
		return nil, m.nextError
	}
	if res, ok := m.nextResult.(*types.ResourceListResult); ok {
		return res, nil
	}
	return nil, errors.New("mock type mismatch")
}

func (m *mockHttpService) ReadResource(ctx context.Context, uri string) (*types.ResourceContent, error) {
	m.recordCall("ReadResource", uri)
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.nextError != nil {
		return nil, m.nextError
	}
	if res, ok := m.nextResult.(*types.ResourceContent); ok {
		return res, nil
	}
	return nil, errors.New("mock type mismatch")
}

func (m *mockHttpService) ListResourceTemplates(ctx context.Context) ([]types.ResourceTemplate, error) {
	m.recordCall("ListResourceTemplates", nil)
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.nextError != nil {
		return nil, m.nextError
	}
	if res, ok := m.nextResult.([]types.ResourceTemplate); ok {
		return res, nil
	}
	return nil, errors.New("mock type mismatch")
}

func (m *mockHttpService) SubscribeToResource(ctx context.Context, uri string) error {
	m.recordCall("SubscribeToResource", uri)
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.nextError
}

// --- Tests ---

func TestNewHTTPServer(t *testing.T) {
	mockService := &mockHttpService{}
	addr := "localhost:12345"
	options := &types.ServerOptions{
		Address: addr,
		Capabilities: &types.ServerCapabilities{
			Completion: true,
		},
	}

	httpServer := NewHTTPServer(mockService, options)

	assert.NotNil(t, httpServer)
	assert.Equal(t, mockService, httpServer.service)
	assert.NotNil(t, httpServer.server)
	assert.Equal(t, addr, httpServer.server.Addr)
	assert.Equal(t, options.Capabilities, httpServer.capabilities)
	assert.NotNil(t, httpServer.server.Handler, "Handler should be set")

	// Check if the handler implements the http.Handler interface
	assert.Implements(t, (*http.Handler)(nil), httpServer.server.Handler)

	// We cannot directly check the internal state (service, capabilities)
	// of the unexported httpHandler without modifications or reflection.
	// We trust that newHTTPHandler correctly initializes the handler internally.
	/*
		// Check handler's internal state if possible/necessary
		if handler, ok := httpServer.server.Handler.(*httpHandler); ok {
			assert.Equal(t, mockService, handler.service)
			assert.Equal(t, options.Capabilities, handler.capabilities)
		}
	*/
}

func TestHTTPServer_Initialize(t *testing.T) {
	mockService := &mockHttpService{}
	options := &types.ServerOptions{Address: "localhost:8080"}
	httpServer := NewHTTPServer(mockService, options)

	initParams := map[string]interface{}{"param1": "value1"}

	// Test successful initialization
	mockService.Reset()
	mockService.SetNextResponse(nil, nil)
	err := httpServer.Initialize(context.Background(), initParams)
	assert.NoError(t, err)
	calls := mockService.GetCalls()
	assert.Len(t, calls, 1)
	assert.Equal(t, "Initialize", calls[0].Method)
	assert.Equal(t, initParams, calls[0].Params)
	assert.True(t, mockService.initialized)

	// Test initialization error
	mockService.Reset()
	mockService.SetNextResponse(nil, errors.New("init fail"))
	err = httpServer.Initialize(context.Background(), initParams)
	assert.Error(t, err)
	assert.Equal(t, "init fail", err.Error())
	calls = mockService.GetCalls()
	assert.Len(t, calls, 1)
	assert.Equal(t, "Initialize", calls[0].Method)
	assert.False(t, mockService.initialized)
}

// Test MCPService method delegation
func TestHTTPServer_MethodDelegation(t *testing.T) {
	mockService := &mockHttpService{}
	options := &types.ServerOptions{Address: "localhost:8080"}
	httpServer := NewHTTPServer(mockService, options)
	ctx := context.Background()

	// --- Test ListPrompts ---
	mockService.Reset()
	expectedResultLP := &types.PromptListResult{Prompts: []types.Prompt{{Name: "p1"}}}
	mockService.SetNextResponse(expectedResultLP, nil)

	resultLP, errLP := httpServer.ListPrompts(ctx, "cursorLP")

	assert.NoError(t, errLP)
	assert.Equal(t, expectedResultLP, resultLP)
	callsLP := mockService.GetCalls()
	assert.Len(t, callsLP, 1)
	assert.Equal(t, "ListPrompts", callsLP[0].Method)
	assert.Equal(t, "cursorLP", callsLP[0].Params)

	// --- Test CallTool ---
	mockService.Reset()
	expectedResultCT := &types.CallToolResult{Content: []types.ToolContent{{Type: "text", Text: "tool result"}}}
	mockService.SetNextResponse(expectedResultCT, nil)
	argsCT := map[string]any{"arg1": 1}

	resultCT, errCT := httpServer.CallTool(ctx, "toolX", argsCT)

	assert.NoError(t, errCT)
	assert.Equal(t, expectedResultCT, resultCT)
	callsCT := mockService.GetCalls()
	assert.Len(t, callsCT, 1)
	assert.Equal(t, "CallTool", callsCT[0].Method)
	expectedParamsCT := map[string]interface{}{"name": "toolX", "args": argsCT}
	assert.Equal(t, expectedParamsCT, callsCT[0].Params)

	// Add similar tests for other delegated methods (GetPrompt, ListTools, ListResources, etc.)
}

// --- HTTP Handler Tests ---

func performRequest(handler http.Handler, method, path string, body io.Reader) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, body)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	return w
}

func TestHTTPHandler_JSONRPC_PromptsList(t *testing.T) {
	mockService := &mockHttpService{}
	capabilities := &types.ServerCapabilities{}
	h := newHTTPHandler(mockService, capabilities) // Use the actual constructor

	// Test case 1: Successful prompts list
	mockService.Reset()
	expectedResult := &types.PromptListResult{
		Prompts:    []types.Prompt{{Name: "prompt1"}},
		NextCursor: "next123",
	}
	mockService.SetNextResponse(expectedResult, nil)

	jsonReq := response.NewJSONRPCRequest(1, "prompts/list", map[string]string{"cursor": "abc"})
	bodyBytes, _ := json.Marshal(jsonReq)

	w := performRequest(h, "POST", "/jsonrpc", bytes.NewReader(bodyBytes))

	assert.Equal(t, http.StatusOK, w.Code)

	// Check mock calls
	calls := mockService.GetCalls()
	assert.Len(t, calls, 1)
	assert.Equal(t, "ListPrompts", calls[0].Method)
	assert.Equal(t, "abc", calls[0].Params.(map[string]interface{})["cursor"])

	// Check response body
	var resp response.JSONRPCResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, jsonReq.ID, resp.ID)
	assert.Nil(t, resp.Error)
	assert.NotNil(t, resp.Result)

	// Deep compare result after marshaling/unmarshaling
	resultBytes, _ := json.Marshal(resp.Result)
	var actualResult types.PromptListResult
	err = json.Unmarshal(resultBytes, &actualResult)
	assert.NoError(t, err)
	assert.Equal(t, *expectedResult, actualResult)

	// Test case 2: Error from service
	mockService.Reset()
	mockService.SetNextResponse(nil, errors.New("service failure"))

	jsonReqErr := response.NewJSONRPCRequest(2, "prompts/list", nil) // No params
	bodyBytesErr, _ := json.Marshal(jsonReqErr)

	wErr := performRequest(h, "POST", "/jsonrpc", bytes.NewReader(bodyBytesErr))

	assert.Equal(t, http.StatusOK, wErr.Code) // JSON-RPC errors return 200 OK

	// Check mock calls
	callsErr := mockService.GetCalls()
	assert.Len(t, callsErr, 1)
	assert.Equal(t, "ListPrompts", callsErr[0].Method)
	// Check cursor was empty string when params are nil
	assert.Equal(t, "", callsErr[0].Params.(map[string]interface{})["cursor"])

	// Check response body
	var respErr response.JSONRPCResponse
	err = json.Unmarshal(wErr.Body.Bytes(), &respErr)
	assert.NoError(t, err)
	assert.Equal(t, jsonReqErr.ID, respErr.ID)
	assert.Nil(t, respErr.Result)
	assert.NotNil(t, respErr.Error)
	assert.Equal(t, -32603, respErr.Error.Code) // Internal error code for generic service error
	assert.Contains(t, respErr.Error.Message, "service failure")

	// Test case 3: Invalid JSON in request body
	wInvalidJSON := performRequest(h, "POST", "/jsonrpc", strings.NewReader(`{invalid`))
	assert.Equal(t, http.StatusBadRequest, wInvalidJSON.Code)
	var respInvalidJSON response.JSONRPCResponse
	err = json.Unmarshal(wInvalidJSON.Body.Bytes(), &respInvalidJSON)
	assert.NoError(t, err)
	assert.Nil(t, respInvalidJSON.ID)
	assert.NotNil(t, respInvalidJSON.Error)
	assert.Equal(t, -32700, respInvalidJSON.Error.Code) // Parse error

	// Test case 4: Method not found
	jsonReqNotFound := response.NewJSONRPCRequest(4, "nonexistent/method", nil)
	bodyBytesNotFound, _ := json.Marshal(jsonReqNotFound)
	wNotFound := performRequest(h, "POST", "/jsonrpc", bytes.NewReader(bodyBytesNotFound))
	assert.Equal(t, http.StatusOK, wNotFound.Code)
	var respNotFound response.JSONRPCResponse
	err = json.Unmarshal(wNotFound.Body.Bytes(), &respNotFound)
	assert.NoError(t, err)
	assert.Equal(t, jsonReqNotFound.ID, respNotFound.ID)
	assert.NotNil(t, respNotFound.Error)
	assert.Equal(t, -32601, respNotFound.Error.Code) // Method not found
}

func TestHTTPHandler_JSONRPC_ToolsCall(t *testing.T) {
	mockService := &mockHttpService{}
	capabilities := &types.ServerCapabilities{}
	h := newHTTPHandler(mockService, capabilities)

	// Test case 1: Successful tool call
	mockService.Reset()
	expectedResult := &types.CallToolResult{
		Content: []types.ToolContent{{Type: "text", Text: "Success!"}},
	}
	mockService.SetNextResponse(expectedResult, nil)

	args := map[string]any{"input": "data"}
	jsonReq := response.NewJSONRPCRequest("tool-req-1", "tools/call", map[string]any{
		"name":      "myTool",
		"arguments": args,
	})
	bodyBytes, _ := json.Marshal(jsonReq)

	w := performRequest(h, "POST", "/jsonrpc", bytes.NewReader(bodyBytes))

	assert.Equal(t, http.StatusOK, w.Code)

	// Check mock calls
	calls := mockService.GetCalls()
	assert.Len(t, calls, 1)
	assert.Equal(t, "CallTool", calls[0].Method)
	expectedCallParams := map[string]interface{}{"name": "myTool", "args": args}
	assert.Equal(t, expectedCallParams, calls[0].Params)

	// Check response body
	var resp response.JSONRPCResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, jsonReq.ID, resp.ID)
	assert.Nil(t, resp.Error)
	assert.NotNil(t, resp.Result)

	// Deep compare result
	resultBytes, _ := json.Marshal(resp.Result)
	var actualResult types.CallToolResult
	err = json.Unmarshal(resultBytes, &actualResult)
	assert.NoError(t, err)
	assert.Equal(t, *expectedResult, actualResult)

	// Test case 2: Missing tool name parameter
	mockService.Reset()
	jsonReqMissingName := response.NewJSONRPCRequest("tool-req-2", "tools/call", map[string]any{"arguments": args}) // Missing name
	bodyBytesMissingName, _ := json.Marshal(jsonReqMissingName)

	wMissingName := performRequest(h, "POST", "/jsonrpc", bytes.NewReader(bodyBytesMissingName))

	assert.Equal(t, http.StatusOK, wMissingName.Code)

	// Check mock calls (should be none)
	callsMissingName := mockService.GetCalls()
	assert.Empty(t, callsMissingName)

	// Check response body for error
	var respMissingName response.JSONRPCResponse
	err = json.Unmarshal(wMissingName.Body.Bytes(), &respMissingName)
	assert.NoError(t, err)
	assert.Equal(t, jsonReqMissingName.ID, respMissingName.ID)
	assert.Nil(t, respMissingName.Result)
	assert.NotNil(t, respMissingName.Error)
	assert.Equal(t, -32602, respMissingName.Error.Code) // Invalid params
	assert.Contains(t, respMissingName.Error.Message, "Missing required parameter: name")

	// Test case 3: Error from service
	mockService.Reset()
	mockService.SetNextResponse(nil, &types.Error{Code: "tool_error", Message: "Tool failed execution"})

	jsonReqSvcErr := response.NewJSONRPCRequest("tool-req-3", "tools/call", map[string]any{"name": "myTool"})
	bodyBytesSvcErr, _ := json.Marshal(jsonReqSvcErr)
	wSvcErr := performRequest(h, "POST", "/jsonrpc", bytes.NewReader(bodyBytesSvcErr))
	assert.Equal(t, http.StatusOK, wSvcErr.Code)

	// Check mock calls
	callsSvcErr := mockService.GetCalls()
	assert.Len(t, callsSvcErr, 1)

	// Check response body
	var respSvcErr response.JSONRPCResponse
	err = json.Unmarshal(wSvcErr.Body.Bytes(), &respSvcErr)
	assert.NoError(t, err)
	assert.Equal(t, jsonReqSvcErr.ID, respSvcErr.ID)
	assert.Nil(t, respSvcErr.Result)
	assert.NotNil(t, respSvcErr.Error)
	// Check if MCP error code was mapped correctly
	expectedCode := response.MapMCPErrorToJSONRPCCode("tool_error") // Should map to internal error -32603
	assert.Equal(t, expectedCode, respSvcErr.Error.Code)
	assert.Equal(t, "Tool failed execution", respSvcErr.Error.Message)
	// Check if original MCP error is in data
	dataBytes, _ := json.Marshal(respSvcErr.Error.Data)
	var dataErr types.Error
	err = json.Unmarshal(dataBytes, &dataErr)
	assert.NoError(t, err)
	assert.Equal(t, "tool_error", dataErr.Code)
}

// --- HTTP Client Tests ---

func TestNewHTTPClient(t *testing.T) {
	addr := "http://localhost:8080"
	client := NewHTTPClient(addr)

	assert.NotNil(t, client)
	assert.Equal(t, addr, client.serverURL)
	assert.NotNil(t, client.client)
	assert.True(t, client.jsonRPC)       // Default should be true
	assert.Equal(t, 0, client.requestID) // Initial ID
}

func TestHTTPClient_Connect(t *testing.T) {
	// Test valid URL
	clientValid := NewHTTPClient("http://valid.url:1234")
	errValid := clientValid.Connect(context.Background())
	assert.NoError(t, errValid)

	// Test invalid URL
	clientInvalid := NewHTTPClient("://invalid url")
	errInvalid := clientInvalid.Connect(context.Background())
	assert.Error(t, errInvalid)
}

func TestHTTPClient_Service(t *testing.T) {
	client := NewHTTPClient("http://test.com")
	service := client.Service()
	assert.Equal(t, client, service) // Should return itself
}

func TestHTTPClient_nextRequestID(t *testing.T) {
	client := NewHTTPClient("http://test.com")
	assert.Equal(t, 1, client.nextRequestID())
	assert.Equal(t, 2, client.nextRequestID())
}

func TestHTTPClient_ListPrompts_JSONRPC(t *testing.T) {
	// --- Test Server Setup ---
	var receivedReq response.JSONRPCRequest

	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/jsonrpc", r.URL.Path)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

		body, _ := io.ReadAll(r.Body)
		err := json.Unmarshal(body, &receivedReq)
		assert.NoError(t, err)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		// Prepare response based on received method
		var resp response.JSONRPCResponse
		if receivedReq.Method == "prompts/list" {
			// Simulate successful response
			result := types.PromptListResult{
				Prompts:    []types.Prompt{{Name: "server_prompt"}},
				NextCursor: "server_cursor",
			}
			resp = response.NewJSONRPCResponse(receivedReq.ID, result)
		} else {
			// Simulate method not found for other methods in this specific test
			resp = response.NewJSONRPCErrorResponse(receivedReq.ID, -32601, "Method not found", nil)
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer mockServer.Close()

	// --- Client Test ---
	client := NewHTTPClient(mockServer.URL) // Use test server URL
	ctx := context.Background()
	testCursor := "client_cursor"

	result, err := client.ListPrompts(ctx, testCursor)

	// --- Assertions ---
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result.Prompts, 1)
	assert.Equal(t, "server_prompt", result.Prompts[0].Name)
	assert.Equal(t, "server_cursor", result.NextCursor)

	// Verify request received by server
	assert.Equal(t, "2.0", receivedReq.JSONRPC)
	assert.Equal(t, "prompts/list", receivedReq.Method)
	assert.NotNil(t, receivedReq.ID)
	// Check params
	var params map[string]interface{}
	err = json.Unmarshal(receivedReq.Params, &params)
	assert.NoError(t, err)
	assert.Equal(t, testCursor, params["cursor"])
}

func TestHTTPClient_CallTool_JSONRPC_Error(t *testing.T) {
	// --- Test Server Setup for Error Response ---
	var receivedReq response.JSONRPCRequest
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		json.Unmarshal(body, &receivedReq)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK) // JSON-RPC errors use 200 OK

		// Simulate JSON-RPC error response
		errResp := response.NewJSONRPCErrorResponse(receivedReq.ID, -32000, "Tool execution failed on server", map[string]string{"detail": "some reason"})
		json.NewEncoder(w).Encode(errResp)
	}))
	defer mockServer.Close()

	// --- Client Test ---
	client := NewHTTPClient(mockServer.URL)
	ctx := context.Background()
	toolName := "errorTool"
	args := map[string]any{"input": 123}

	result, err := client.CallTool(ctx, toolName, args)

	// --- Assertions ---
	assert.Error(t, err) // Expect an error on the client side
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "JSON-RPC error (code -32000): Tool execution failed on server")

	// Verify request received by server
	assert.Equal(t, "2.0", receivedReq.JSONRPC)
	assert.Equal(t, "tools/call", receivedReq.Method)
	var params map[string]interface{}
	json.Unmarshal(receivedReq.Params, &params)
	assert.Equal(t, toolName, params["name"])
	assert.EqualValues(t, args, params["arguments"])
}

// Test HTTP error (e.g., 500 Internal Server Error)
func TestHTTPClient_CallTool_HTTPError(t *testing.T) {
	// --- Test Server Setup for HTTP Error ---
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Internal Server Error"))
	}))
	defer mockServer.Close()

	// --- Client Test ---
	client := NewHTTPClient(mockServer.URL)
	ctx := context.Background()

	result, err := client.CallTool(ctx, "anyTool", nil)

	// --- Assertions ---
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "server returned HTTP error: 500 Internal Server Error")
}

// Test helper functions
func Test_buildURL(t *testing.T) {
	client := NewHTTPClient("http://base.com/api")
	assert.Equal(t, "http://base.com/api/path1", client.buildURL("/path1"))
	assert.Equal(t, "http://base.com/api/path2", client.buildURL("path2"))   // Should add leading slash
	assert.Equal(t, "http://base.com/api/path3", client.buildURL("//path3")) // Should clean extra slashes

	clientNoSlash := NewHTTPClient("http://base.com")
	assert.Equal(t, "http://base.com/path", clientNoSlash.buildURL("/path"))
}

func Test_isTextMIME(t *testing.T) {
	assert.True(t, isTextMIME("text/plain"))
	assert.True(t, isTextMIME("text/html; charset=utf-8"))
	assert.True(t, isTextMIME("application/json"))
	assert.True(t, isTextMIME("application/xml"))
	assert.True(t, isTextMIME("text/css"))
	assert.True(t, isTextMIME("text/markdown"))
	assert.True(t, isTextMIME("text/javascript")) // Corrected from application/javascript if primary use is text

	assert.False(t, isTextMIME("image/png"))
	assert.False(t, isTextMIME("application/octet-stream"))
	assert.False(t, isTextMIME("application/pdf"))
	assert.False(t, isTextMIME(""))
}
