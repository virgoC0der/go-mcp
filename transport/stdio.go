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
	srv      types.Server
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
func NewStdioServer(srv types.Server) *StdioServer {
	return NewStdioServerWithIO(srv, os.Stdin, os.Stdout)
}

// NewStdioServerWithIO creates a new stdio server instance with custom I/O
func NewStdioServerWithIO(srv types.Server, reader io.Reader, writer io.Writer) *StdioServer {
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

// tryParseClaudeMessage tries to extract useful information from Claude's message format
func tryParseClaudeMessage(data []byte) (*StdioMessage, error) {
	// First try to parse as raw JSON
	var rawMsg map[string]interface{}
	if err := json.Unmarshal(data, &rawMsg); err != nil {
		return nil, err
	}

	// Log the raw message
	fmt.Printf("Raw message from Claude: %+v\n", rawMsg)

	// Create a new StdioMessage
	msg := &StdioMessage{
		JSONRPC: "2.0",
		ID:      1, // Default ID
	}

	// Extract ID
	if id, ok := rawMsg["id"]; ok {
		switch v := id.(type) {
		case float64:
			msg.ID = int(v)
		case int:
			msg.ID = v
		case string:
			// 使用一个默认值
			fmt.Printf("Using default ID for string ID: %s\n", v)
		}
	}

	// 如果消息包含 jsonrpc 字段，则保留它
	if jsonrpc, ok := rawMsg["jsonrpc"].(string); ok {
		msg.JSONRPC = jsonrpc
	}

	// 如果消息中有 method 字段，则使用它
	if method, ok := rawMsg["method"].(string); ok && method != "" {
		msg.Method = method

		// 如果是 initialize 方法，确保正确解析
		if method == "initialize" || method == "Initialize" {
			msg.Method = "initialize"
			if params, ok := rawMsg["params"]; ok {
				paramsBytes, _ := json.Marshal(params)
				msg.Params = paramsBytes
			} else {
				// 对于初始化请求，如果没有参数，提供一个空对象
				msg.Params = json.RawMessage([]byte("{}"))
			}
			return msg, nil
		}
	} else {
		// 尝试推断方法类型
		// 如果消息中有 id 但没有 method，可能是个初始化请求
		_, hasId := rawMsg["id"]
		if hasId && !ok {
			msg.Method = "initialize"
			msg.Params = json.RawMessage([]byte("{}"))
			return msg, nil
		}
	}

	// 根据消息内容确定方法
	if content, hasContent := rawMsg["content"]; hasContent {
		msg.Method = "callTool"

		// 构造工具调用参数
		toolName := "default" // 默认工具名称

		// 尝试从消息中提取工具名称，这取决于 Claude 的消息格式
		if role, ok := rawMsg["role"].(string); ok && role == "assistant" {
			msg.Method = "getPrompt"
		}

		toolParams := map[string]interface{}{
			"name": toolName,
			"args": map[string]interface{}{
				"content":    content,
				"rawMessage": rawMsg,
			},
		}

		paramsBytes, _ := json.Marshal(toolParams)
		msg.Params = paramsBytes
	} else if msg.Method == "" {
		// 如果方法为空，默认为 listTools
		msg.Method = "listTools"
		msg.Params = json.RawMessage([]byte("{}"))
	} else if params, ok := rawMsg["params"]; ok {
		// 如果有参数字段，使用它
		paramsBytes, _ := json.Marshal(params)
		msg.Params = paramsBytes
	} else {
		// 如果没有参数，提供一个空对象
		msg.Params = json.RawMessage([]byte("{}"))
	}

	fmt.Printf("Transformed message: Method=%s, ID=%d\n", msg.Method, msg.ID)
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
		claudeMsg, parseErr := tryParseClaudeMessage(data)
		if parseErr != nil {
			s.sendError(0, fmt.Sprintf("Invalid message format: %v", err))
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
	if err := s.srv.Initialize(context.Background(), opts); err != nil {
		s.sendError(msg.ID, fmt.Sprintf("Initialization failed: %v", err))
		return
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
