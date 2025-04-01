package transport

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sync"

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

// StdioMessage represents the structure of a stdio message
type StdioMessage struct {
	Type    string          `json:"type"`
	ID      string          `json:"id"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
	Content json.RawMessage `json:"content,omitempty"`
}

// StdioResponse represents the structure of a stdio response
type StdioResponse struct {
	Type    string          `json:"type"`
	ID      string          `json:"id"`
	Success bool            `json:"success"`
	Content json.RawMessage `json:"content,omitempty"`
	Error   string          `json:"error,omitempty"`
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

// handleMessage processes a received message
func (s *StdioServer) handleMessage(data []byte) {
	var msg StdioMessage
	if err := json.Unmarshal(data, &msg); err != nil {
		s.sendError(msg.ID, fmt.Sprintf("Invalid message format: %v", err))
		return
	}

	switch msg.Method {
	case "initialize":
		s.handleInitialize(msg)
	case "getPrompt":
		s.handleGetPrompt(msg)
	case "listPrompts":
		s.handleListPrompts(msg)
	case "callTool":
		s.handleCallTool(msg)
	case "listTools":
		s.handleListTools(msg)
	case "readResource":
		s.handleReadResource(msg)
	case "listResources":
		s.handleListResources(msg)
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

	// Send success response
	s.sendResponse(msg.ID, true, nil, "")
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
	// Call the server's ListPrompts method
	prompts, err := s.srv.ListPrompts(context.Background())
	if err != nil {
		s.sendError(msg.ID, fmt.Sprintf("ListPrompts failed: %v", err))
		return
	}

	// Convert result to JSON
	content, err := json.Marshal(prompts)
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
	// Call the server's ListTools method
	tools, err := s.srv.ListTools(context.Background())
	if err != nil {
		s.sendError(msg.ID, fmt.Sprintf("ListTools failed: %v", err))
		return
	}

	// Convert result to JSON
	content, err := json.Marshal(tools)
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
		Name string `json:"name"`
	}
	if err := json.Unmarshal(msg.Params, &params); err != nil {
		s.sendError(msg.ID, fmt.Sprintf("Invalid readResource parameters: %v", err))
		return
	}

	// Call the server's ReadResource method
	content, mimeType, err := s.srv.ReadResource(context.Background(), params.Name)
	if err != nil {
		s.sendError(msg.ID, fmt.Sprintf("ReadResource failed: %v", err))
		return
	}

	// Create response structure
	response := struct {
		Content  []byte `json:"content"`
		MimeType string `json:"mimeType"`
	}{
		Content:  content,
		MimeType: mimeType,
	}

	// Convert result to JSON
	resultContent, err := json.Marshal(response)
	if err != nil {
		s.sendError(msg.ID, fmt.Sprintf("Failed to marshal result: %v", err))
		return
	}

	// Send success response
	s.sendResponse(msg.ID, true, resultContent, "")
}

// handleListResources handles listResources request
func (s *StdioServer) handleListResources(msg StdioMessage) {
	// Call the server's ListResources method
	resources, err := s.srv.ListResources(context.Background())
	if err != nil {
		s.sendError(msg.ID, fmt.Sprintf("ListResources failed: %v", err))
		return
	}

	// Convert result to JSON
	content, err := json.Marshal(resources)
	if err != nil {
		s.sendError(msg.ID, fmt.Sprintf("Failed to marshal result: %v", err))
		return
	}

	// Send success response
	s.sendResponse(msg.ID, true, content, "")
}

// sendResponse sends a response to stdout
func (s *StdioServer) sendResponse(id string, success bool, content json.RawMessage, errorMsg string) {
	resp := StdioResponse{
		Type:    "response",
		ID:      id,
		Success: success,
		Content: content,
		Error:   errorMsg,
	}

	data, err := json.Marshal(resp)
	if err != nil {
		// If we can't marshal the response, try to send a simplified error
		s.sendError(id, fmt.Sprintf("Failed to marshal response: %v", err))
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	fmt.Fprintln(s.writer, string(data))
}

// sendError sends an error response
func (s *StdioServer) sendError(id, errorMsg string) {
	s.sendResponse(id, false, nil, errorMsg)
}
