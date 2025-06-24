package transport

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sync"

	"github.com/virgoC0der/go-mcp/internal/response"
	"github.com/virgoC0der/go-mcp/internal/types"
)

// StdioServer wraps the stdio transport layer for an MCP server
type StdioServer struct {
	srv      types.MCPService
	reader   io.Reader
	writer   io.Writer
	incoming chan []byte
	done     chan struct{}
	mu       sync.Mutex
}

// StdioMessage alias, using the common JSON-RPC request structure
type StdioMessage = response.JSONRPCRequest

// InitializeResponse represents the structure of an initialization response
type InitializeResponse struct {
	Meta            interface{}          `json:"meta,omitempty"`
	Capabilities    *Capabilities        `json:"capabilities"`
	Instructions    interface{}          `json:"instructions,omitempty"`
	ProtocolVersion string               `json:"protocolVersion"`
	ServerInfo      ServerImplementation `json:"serverInfo"`
}

// ServerImplementation represents server implementation information
type ServerImplementation struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

// Capabilities represents server capabilities
type Capabilities struct {
	Tools     *types.ToolCapabilities     `json:"tools,omitempty"`
	Prompts   *types.PromptCapabilities   `json:"prompts,omitempty"`
	Resources *types.ResourceCapabilities `json:"resources,omitempty"`
}

// NewStdioServer creates a new stdio server instance
func NewStdioServer(srv types.MCPService) *StdioServer {
	return NewStdioServerWithIO(srv, os.Stdin, os.Stdout)
}

// NewStdioServerWithIO creates a new stdio server instance with custom I/O
func NewStdioServerWithIO(srv types.MCPService, reader io.Reader, writer io.Writer) *StdioServer {
	return &StdioServer{
		srv:      srv,
		reader:   reader,
		writer:   writer,
		incoming: make(chan []byte),
		done:     make(chan struct{}),
	}
}

// Start starts the stdio server
func (s *StdioServer) Start() error {
	// Start reading from stdin
	go s.readLoop()

	// Process messages
	for {
		select {
		case data := <-s.incoming:
			go s.handleMessage(data)
		case <-s.done:
			return nil
		}
	}
}

// Stop stops the stdio server
func (s *StdioServer) Stop() error {
	close(s.done)
	return nil
}

// readLoop reads messages from stdin
func (s *StdioServer) readLoop() {
	scanner := bufio.NewScanner(s.reader)
	for scanner.Scan() {
		line := scanner.Bytes()
		s.incoming <- line
	}
}

// TryParseClaudeMessage tries to extract useful information from Claude's message format
func TryParseClaudeMessage(data []byte) (*StdioMessage, error) {
	var rawMsg map[string]interface{}
	if err := json.Unmarshal(data, &rawMsg); err != nil {
		return nil, err
	}

	// Log the raw message
	fmt.Printf("Raw message from Claude: %+v\n", rawMsg)

	msg := &StdioMessage{
		JSONRPC: "2.0",
		ID:      1, // Default ID
	}

	// Extract ID
	if id, ok := rawMsg["id"]; ok && id != nil {
		switch v := id.(type) {
		case float64:
			msg.ID = int(v)
		case int:
			msg.ID = v
		case string:
			if v != "" {
				msg.ID = v
			} else {
				msg.ID = 1
			}
			fmt.Printf("Using string ID: %s\n", v)
		default:
			msg.ID = 1
		}
	} else {
		msg.ID = 1
	}

	if jsonrpc, ok := rawMsg["jsonrpc"].(string); ok {
		msg.JSONRPC = jsonrpc
	}

	// 1. 优先使用 method 字段
	if method, ok := rawMsg["method"].(string); ok && method != "" {
		msg.Method = method
		if params, ok := rawMsg["params"]; ok {
			paramsBytes, _ := json.Marshal(params)
			msg.Params = paramsBytes
		} else {
			msg.Params = json.RawMessage([]byte("{}"))
		}
		return msg, nil
	}

	// 2. Claude 风格推断
	if role, ok := rawMsg["role"].(string); ok {
		if content, hasContent := rawMsg["content"]; hasContent {
			toolName := "default"
			if role == "user" {
				msg.Method = "callTool"
				toolParams := map[string]interface{}{
					"name": toolName,
					"args": map[string]interface{}{
						"content":    content,
						"rawMessage": rawMsg,
					},
				}
				paramsBytes, _ := json.Marshal(toolParams)
				msg.Params = paramsBytes
				return msg, nil
			} else if role == "assistant" {
				msg.Method = "getPrompt"
				promptParams := map[string]interface{}{
					"name": toolName,
					"args": map[string]interface{}{
						"content":    content,
						"rawMessage": rawMsg,
					},
				}
				paramsBytes, _ := json.Marshal(promptParams)
				msg.Params = paramsBytes
				return msg, nil
			}
		}
	}

	// 3. 空对象推断 listTools
	if len(rawMsg) == 0 {
		msg.Method = "listTools"
		msg.Params = json.RawMessage([]byte("{}"))
		return msg, nil
	}

	// 4. 兼容 params 但无 method 的情况（如 Claude 初始化）
	if _, hasId := rawMsg["id"]; hasId {
		if _, hasParams := rawMsg["params"]; hasParams {
			msg.Method = "initialize"
			msg.Params = json.RawMessage([]byte("{}"))
			return msg, nil
		}
	}

	// 5. fallback
	msg.Method = "initialize"
	msg.Params = json.RawMessage([]byte("{}"))
	return msg, nil
}

// handleMessage processes a received message
func (s *StdioServer) handleMessage(data []byte) {
	var msg StdioMessage

	// Try to parse as standard JSON-RPC message
	err := json.Unmarshal(data, &msg)

	// Check if the message format is valid
	if err != nil || msg.Method == "" {
		// Try to parse as Claude format message
		claudeMsg, parseErr := TryParseClaudeMessage(data)
		if parseErr != nil {
			// 尝试从原始数据中提取ID用于错误响应
			var rawMsg map[string]interface{}
			errorID := interface{}(0) // 默认ID为0，符合测试期望
			if json.Unmarshal(data, &rawMsg) == nil {
				if id, ok := rawMsg["id"]; ok && id != nil {
					errorID = id
				}
			}
			s.sendError(errorID, fmt.Sprintf("Invalid message format: %v", err))
			return
		}

		// Use the parsed Claude message
		msg = *claudeMsg
	}

	// Check if it's a JSON-RPC 2.0 request
	if msg.JSONRPC != "2.0" && msg.JSONRPC != "" {
		s.sendError(msg.ID, "Invalid JSON-RPC version. Expected '2.0'")
		return
	}

	switch msg.Method {
	case "initialize":
		s.handleInitialize(msg)
	case "prompts/get":
		s.handleGetPrompt(msg)
	case "prompts/list":
		s.handleListPrompts(msg)
	case "tools/list":
		s.handleListTools(msg)
	case "tools/call":
		s.handleCallTool(msg)
	case "resources/list":
		s.handleListResources(msg)
	case "resources/read":
		s.handleReadResource(msg)
	case "resources/templates/list":
		s.handleListResourceTemplates(msg)
	case "resources/subscribe":
		s.handleSubscribeToResource(msg)
	default:
		s.sendError(msg.ID, fmt.Sprintf("Unknown method: %s", msg.Method))
	}
}

// handleInitialize handles initialize request
func (s *StdioServer) handleInitialize(msg StdioMessage) {
	// Parse initialization options
	var opts map[string]any
	if err := json.Unmarshal(msg.Params, &opts); err != nil {
		s.sendError(msg.ID, fmt.Sprintf("Invalid initialization parameters: %v", err))
		return
	}

	// Call the server's Initialize method
	if srv, ok := s.srv.(interface {
		Initialize(ctx context.Context, options any) error
	}); ok {
		if err := srv.Initialize(context.Background(), opts); err != nil {
			s.sendError(msg.ID, fmt.Sprintf("Initialization failed: %v", err))
			return
		}
	}

	// 构建标准的初始化响应
	serverInfo := ServerImplementation{
		Name:    "go-mcp stdio server",
		Version: "1.0.0",
	}

	// 如果提供了 ServerOptions，从中获取更详细的信息
	var capabilities *Capabilities = nil

	// 检查服务器是否支持特定能力
	capabilities = &Capabilities{
		Tools:     &types.ToolCapabilities{ListChanged: true},
		Prompts:   &types.PromptCapabilities{ListChanged: true},
		Resources: &types.ResourceCapabilities{ListChanged: true},
	}

	initResponse := InitializeResponse{
		Meta:            nil,
		Capabilities:    capabilities,
		Instructions:    nil,
		ProtocolVersion: "2024-11-05",
		ServerInfo:      serverInfo,
	}

	// 转换为 JSON
	content, err := json.Marshal(initResponse)
	if err != nil {
		s.sendError(msg.ID, fmt.Sprintf("Failed to marshal initialization response: %v", err))
		return
	}

	// 发送成功响应
	s.sendResponse(msg.ID, true, content, "")
}

// handleGetPrompt handles getPrompt request
func (s *StdioServer) handleGetPrompt(msg StdioMessage) {
	// Parse parameters
	var params struct {
		Name      string         `json:"name"`
		Arguments map[string]any `json:"arguments"`
	}
	if err := json.Unmarshal(msg.Params, &params); err != nil {
		s.sendError(msg.ID, fmt.Sprintf("Invalid getPrompt parameters: %v", err))
		return
	}

	// Call the server's GetPrompt method
	result, err := s.srv.GetPrompt(context.Background(), params.Name, params.Arguments)
	if err != nil {
		s.sendError(msg.ID, fmt.Sprintf("GetPrompt failed: %v", err))
		return
	}

	// Convert result to JSON
	content, err := json.Marshal(result)
	if err != nil {
		s.sendError(msg.ID, fmt.Sprintf("Failed to marshal result: %v", err))
		return
	}

	// Send success response
	s.sendResponse(msg.ID, true, content, "")
}

// handleListPrompts handles listPrompts request
func (s *StdioServer) handleListPrompts(msg StdioMessage) {
	// Parse parameters
	var params struct {
		Cursor string `json:"cursor"`
	}
	if err := json.Unmarshal(msg.Params, &params); err != nil {
		// Default to empty cursor if not provided
		params.Cursor = ""
	}

	// Call the server's ListPrompts method
	result, err := s.srv.ListPrompts(context.Background(), params.Cursor)
	if err != nil {
		s.sendError(msg.ID, fmt.Sprintf("ListPrompts failed: %v", err))
		return
	}

	// Convert result to JSON
	content, err := json.Marshal(result)
	if err != nil {
		s.sendError(msg.ID, fmt.Sprintf("Failed to marshal result: %v", err))
		return
	}

	// Send success response
	s.sendResponse(msg.ID, true, content, "")
}

// handleCallTool handles callTool request
func (s *StdioServer) handleCallTool(msg StdioMessage) {
	// Parse parameters
	var params struct {
		Name      string         `json:"name"`
		Arguments map[string]any `json:"arguments"`
	}
	if err := json.Unmarshal(msg.Params, &params); err != nil {
		s.sendError(msg.ID, fmt.Sprintf("Invalid callTool parameters: %v", err))
		return
	}

	// Call the server's CallTool method
	result, err := s.srv.CallTool(context.Background(), params.Name, params.Arguments)
	if err != nil {
		s.sendError(msg.ID, fmt.Sprintf("CallTool failed: %v", err))
		return
	}

	// Convert result to JSON
	content, err := json.Marshal(result)
	if err != nil {
		s.sendError(msg.ID, fmt.Sprintf("Failed to marshal result: %v", err))
		return
	}

	// Send success response
	s.sendResponse(msg.ID, true, content, "")
}

// handleListTools handles listTools request
func (s *StdioServer) handleListTools(msg StdioMessage) {
	// Parse parameters
	var params struct {
		Cursor string `json:"cursor"`
	}
	if err := json.Unmarshal(msg.Params, &params); err != nil {
		// Default to empty cursor if not provided
		params.Cursor = ""
	}

	// Call the server's ListTools method
	result, err := s.srv.ListTools(context.Background(), params.Cursor)
	if err != nil {
		s.sendError(msg.ID, fmt.Sprintf("ListTools failed: %v", err))
		return
	}

	// Convert result to JSON
	content, err := json.Marshal(result)
	if err != nil {
		s.sendError(msg.ID, fmt.Sprintf("Failed to marshal result: %v", err))
		return
	}

	// Send success response
	s.sendResponse(msg.ID, true, content, "")
}

// handleReadResource handles readResource request
func (s *StdioServer) handleReadResource(msg StdioMessage) {
	// Parse parameters
	var params struct {
		URI string `json:"uri"`
	}
	if err := json.Unmarshal(msg.Params, &params); err != nil {
		s.sendError(msg.ID, fmt.Sprintf("Invalid readResource parameters: %v", err))
		return
	}

	// Call the server's ReadResource method
	result, err := s.srv.ReadResource(context.Background(), params.URI)
	if err != nil {
		s.sendError(msg.ID, fmt.Sprintf("ReadResource failed: %v", err))
		return
	}

	// Convert result to JSON
	content, err := json.Marshal(result)
	if err != nil {
		s.sendError(msg.ID, fmt.Sprintf("Failed to marshal result: %v", err))
		return
	}

	// Send success response
	s.sendResponse(msg.ID, true, content, "")
}

// handleListResources handles listResources request
func (s *StdioServer) handleListResources(msg StdioMessage) {
	// Parse parameters
	var params struct {
		Cursor string `json:"cursor"`
	}
	if err := json.Unmarshal(msg.Params, &params); err != nil {
		// Default to empty cursor if not provided
		params.Cursor = ""
	}

	// Call the server's ListResources method
	result, err := s.srv.ListResources(context.Background(), params.Cursor)
	if err != nil {
		s.sendError(msg.ID, fmt.Sprintf("ListResources failed: %v", err))
		return
	}

	// Convert result to JSON
	content, err := json.Marshal(result)
	if err != nil {
		s.sendError(msg.ID, fmt.Sprintf("Failed to marshal result: %v", err))
		return
	}

	// Send success response
	s.sendResponse(msg.ID, true, content, "")
}

// handleListResourceTemplates handles listResourceTemplates request
func (s *StdioServer) handleListResourceTemplates(msg StdioMessage) {
	// Call the server's ListResourceTemplates method
	result, err := s.srv.ListResourceTemplates(context.Background())
	if err != nil {
		s.sendError(msg.ID, fmt.Sprintf("ListResourceTemplates failed: %v", err))
		return
	}

	// Convert result to JSON
	content, err := json.Marshal(result)
	if err != nil {
		s.sendError(msg.ID, fmt.Sprintf("Failed to marshal result: %v", err))
		return
	}

	// Send success response
	s.sendResponse(msg.ID, true, content, "")
}

// handleSubscribeToResource handles subscribeToResource request
func (s *StdioServer) handleSubscribeToResource(msg StdioMessage) {
	// Parse parameters
	var params struct {
		URI string `json:"uri"`
	}
	if err := json.Unmarshal(msg.Params, &params); err != nil {
		s.sendError(msg.ID, fmt.Sprintf("Invalid subscribeToResource parameters: %v", err))
		return
	}

	// Call the server's SubscribeToResource method
	err := s.srv.SubscribeToResource(context.Background(), params.URI)
	if err != nil {
		s.sendError(msg.ID, fmt.Sprintf("SubscribeToResource failed: %v", err))
		return
	}

	// Send success response
	s.sendResponse(msg.ID, true, nil, "")
}

// sendResponse sends a response to stdout (JSON-RPC 2.0 compatible)
func (s *StdioServer) sendResponse(id interface{}, success bool, content json.RawMessage, errorMsg string) {
	var resp response.JSONRPCResponse

	// 确保ID不为nil
	if id == nil {
		id = 1
	}

	// If successful, set the result field
	if success && content != nil {
		// Parse JSON content as interface{}
		var result interface{}
		if err := json.Unmarshal(content, &result); err == nil {
			resp = response.NewJSONRPCResponse(id, result)
		} else {
			// If parsing fails, use the raw content directly
			resp = response.NewJSONRPCResponse(id, string(content))
		}
	} else if success {
		// Claude may expect to always have a result object
		resp = response.NewJSONRPCResponse(id, map[string]interface{}{})
	} else if errorMsg != "" {
		// Error response
		resp = response.NewJSONRPCErrorResponse(id, -32603, errorMsg, nil)
	}

	data, err := response.MarshalResponse(resp)
	if err != nil {
		// If we can't serialize the response, try to send a simplified error
		s.sendSimplifiedError(id, fmt.Sprintf("Failed to marshal response: %v", err))
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	fmt.Fprintln(s.writer, string(data))
}

// sendSimplifiedError sends a simplified error response when the standard response cannot be serialized
func (s *StdioServer) sendSimplifiedError(id interface{}, errorMsg string) {
	// 确保ID不为nil
	if id == nil {
		id = 1
	}

	// Create a minimal response
	simpleResp := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      id,
		"error": map[string]interface{}{
			"code":    -32603,
			"message": errorMsg,
		},
	}

	data, err := json.Marshal(simpleResp)
	if err != nil {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	fmt.Fprintln(s.writer, string(data))
}

// sendError sends an error response
func (s *StdioServer) sendError(id interface{}, errorMsg string) {
	s.sendResponse(id, false, nil, errorMsg)
}

// Server interface implementation - delegate to the wrapped server

// Initialize initializes the server with given options
func (s *StdioServer) Initialize(ctx context.Context, options any) error {
	// StdioServer handles its own initialization
	return nil
}

// Shutdown gracefully shuts down the server
func (s *StdioServer) Shutdown(ctx context.Context) error {
	// Close the done channel to signal shutdown
	select {
	case <-s.done:
		// Already closed
	default:
		close(s.done)
	}
	return nil
}

// MCPService interface implementation - delegate to the wrapped server

// ListPrompts returns a list of available prompts
func (s *StdioServer) ListPrompts(ctx context.Context, cursor string) (*types.PromptListResult, error) {
	return s.srv.ListPrompts(ctx, cursor)
}

// GetPrompt retrieves a specific prompt by name with optional arguments
func (s *StdioServer) GetPrompt(ctx context.Context, name string, args map[string]any) (*types.PromptResult, error) {
	return s.srv.GetPrompt(ctx, name, args)
}

// ListTools returns a list of available tools
func (s *StdioServer) ListTools(ctx context.Context, cursor string) (*types.ToolListResult, error) {
	return s.srv.ListTools(ctx, cursor)
}

// CallTool invokes a specific tool by name with arguments
func (s *StdioServer) CallTool(ctx context.Context, name string, args map[string]any) (*types.CallToolResult, error) {
	return s.srv.CallTool(ctx, name, args)
}

// ListResources returns a list of available resources
func (s *StdioServer) ListResources(ctx context.Context, cursor string) (*types.ResourceListResult, error) {
	return s.srv.ListResources(ctx, cursor)
}

// ReadResource reads the content of a specific resource
func (s *StdioServer) ReadResource(ctx context.Context, uri string) (*types.ResourceContent, error) {
	return s.srv.ReadResource(ctx, uri)
}

// ListResourceTemplates returns a list of available resource templates
func (s *StdioServer) ListResourceTemplates(ctx context.Context) ([]types.ResourceTemplate, error) {
	return s.srv.ListResourceTemplates(ctx)
}

// SubscribeToResource subscribes to changes on a specific resource
func (s *StdioServer) SubscribeToResource(ctx context.Context, uri string) error {
	return s.srv.SubscribeToResource(ctx, uri)
}
