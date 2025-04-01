package client

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/virgoC0der/go-mcp/types"
)

// WebSocketClient is a WebSocket-based MCP client
type WebSocketClient struct {
	conn        *websocket.Conn
	url         string
	requestID   int
	responses   map[string]chan []byte
	lock        sync.Mutex
	initialized bool
}

// NewWebSocketClient creates a new WebSocket client
func NewWebSocketClient(url string) (*WebSocketClient, error) {
	// Connect to the WebSocket server
	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to WebSocket server: %w", err)
	}

	c := &WebSocketClient{
		conn:      conn,
		url:       url,
		requestID: 0,
		responses: make(map[string]chan []byte),
	}

	// Start reading responses
	go c.readResponses()

	return c, nil
}

// readResponses reads and processes responses from the WebSocket connection
func (c *WebSocketClient) readResponses() {
	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			// Connection closed or error
			return
		}

		// Parse the response
		var response struct {
			Type      string          `json:"type"`
			MessageId string          `json:"messageId"`
			ID        string          `json:"id"`
			Success   bool            `json:"success"`
			Result    json.RawMessage `json:"result"`
			Error     struct {
				Code    string `json:"code"`
				Message string `json:"message"`
			} `json:"error"`
		}

		if err := json.Unmarshal(message, &response); err != nil {
			fmt.Printf("Error parsing response: %v\n", err)
			continue
		}

		// Determine the ID to use
		respID := response.ID
		if respID == "" {
			respID = response.MessageId
		}

		// Get the response channel for this ID
		c.lock.Lock()
		ch, ok := c.responses[respID]
		c.lock.Unlock()

		if ok {
			// Send the response back to the waiting goroutine
			select {
			case ch <- message:
				// Message sent successfully
			default:
				// Channel is full or closed, ignore
			}

			// Delete the channel since we don't need it anymore
			c.lock.Lock()
			delete(c.responses, respID)
			c.lock.Unlock()
		}
	}
}

// sendRequest sends a request to the server and waits for a response
func (c *WebSocketClient) sendRequest(method string, params interface{}) ([]byte, error) {
	// Generate a request ID
	c.lock.Lock()
	id := fmt.Sprintf("req-%d", c.requestID)
	c.requestID++
	c.lock.Unlock()

	// Prepare the request
	request := struct {
		Type      string      `json:"type"`
		MessageId string      `json:"messageId"`
		Method    string      `json:"method"`
		Args      interface{} `json:"args,omitempty"`
	}{
		Type:      "request",
		MessageId: id,
		Method:    method,
		Args:      params,
	}

	// Create a channel to receive the response
	responseCh := make(chan []byte, 1)
	c.lock.Lock()
	c.responses[id] = responseCh
	c.lock.Unlock()

	// Send the request
	if err := c.conn.WriteJSON(request); err != nil {
		return nil, fmt.Errorf("error sending request: %w", err)
	}

	// Wait for the response with a timeout
	select {
	case responseData := <-responseCh:
		// Parse the response
		var response struct {
			Type      string          `json:"type"`
			MessageId string          `json:"messageId"`
			Success   bool            `json:"success"`
			Result    json.RawMessage `json:"result"`
			Error     struct {
				Code    string `json:"code"`
				Message string `json:"message"`
			} `json:"error"`
		}

		if err := json.Unmarshal(responseData, &response); err != nil {
			return nil, fmt.Errorf("error parsing response: %w", err)
		}

		if !response.Success {
			errorMsg := "unknown error"
			if response.Error.Message != "" {
				errorMsg = response.Error.Message
			}
			return nil, fmt.Errorf("request failed: %s", errorMsg)
		}

		return response.Result, nil

	case <-time.After(10 * time.Second):
		// Remove the response channel
		c.lock.Lock()
		delete(c.responses, id)
		c.lock.Unlock()

		return nil, fmt.Errorf("timeout waiting for response")
	}
}

// Initialize initializes the client connection
func (c *WebSocketClient) Initialize(ctx context.Context) error {
	// Send an initialize request
	_, err := c.sendRequest("initialize", map[string]interface{}{})
	if err != nil {
		return fmt.Errorf("initialization failed: %w", err)
	}

	c.initialized = true
	return nil
}

// ListPrompts lists all available prompts
func (c *WebSocketClient) ListPrompts(ctx context.Context) ([]types.Prompt, error) {
	// Send a listPrompts request
	data, err := c.sendRequest("listPrompts", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to list prompts: %w", err)
	}

	// Parse the response
	var prompts []types.Prompt
	if err := json.Unmarshal(data, &prompts); err != nil {
		return nil, fmt.Errorf("error parsing prompts: %w", err)
	}

	return prompts, nil
}

// ListPromptsPaginated lists prompts with pagination
func (c *WebSocketClient) ListPromptsPaginated(ctx context.Context, options types.PaginationOptions) (*types.PaginatedResult, error) {
	// Send a listPrompts request with pagination
	data, err := c.sendRequest("listPrompts", options)
	if err != nil {
		return nil, fmt.Errorf("failed to list prompts: %w", err)
	}

	// Parse the response
	var result types.PaginatedResult
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("error parsing paginated result: %w", err)
	}

	return &result, nil
}

// GetPrompt gets a prompt with the given name and arguments
func (c *WebSocketClient) GetPrompt(ctx context.Context, name string, arguments map[string]interface{}) (*types.GetPromptResult, error) {
	// Send a getPrompt request
	params := struct {
		Name      string                 `json:"name"`
		Arguments map[string]interface{} `json:"arguments"`
	}{
		Name:      name,
		Arguments: arguments,
	}

	data, err := c.sendRequest("getPrompt", params)
	if err != nil {
		return nil, fmt.Errorf("failed to get prompt: %w", err)
	}

	// Parse the response
	var result types.GetPromptResult
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("error parsing prompt result: %w", err)
	}

	return &result, nil
}

// ListTools lists all available tools
func (c *WebSocketClient) ListTools(ctx context.Context) ([]types.Tool, error) {
	// Send a listTools request
	data, err := c.sendRequest("listTools", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to list tools: %w", err)
	}

	// Parse the response
	var tools []types.Tool
	if err := json.Unmarshal(data, &tools); err != nil {
		return nil, fmt.Errorf("error parsing tools: %w", err)
	}

	return tools, nil
}

// ListToolsPaginated lists tools with pagination
func (c *WebSocketClient) ListToolsPaginated(ctx context.Context, options types.PaginationOptions) (*types.PaginatedResult, error) {
	// Send a listTools request with pagination
	data, err := c.sendRequest("listTools", options)
	if err != nil {
		return nil, fmt.Errorf("failed to list tools: %w", err)
	}

	// Parse the response
	var result types.PaginatedResult
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("error parsing paginated result: %w", err)
	}

	return &result, nil
}

// CallTool calls a tool with the given name and arguments
func (c *WebSocketClient) CallTool(ctx context.Context, name string, arguments map[string]interface{}) (*types.CallToolResult, error) {
	// Send a callTool request
	params := struct {
		Name      string                 `json:"name"`
		Arguments map[string]interface{} `json:"arguments"`
	}{
		Name:      name,
		Arguments: arguments,
	}

	data, err := c.sendRequest("callTool", params)
	if err != nil {
		return nil, fmt.Errorf("failed to call tool: %w", err)
	}

	// Parse the response
	var result types.CallToolResult
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("error parsing tool result: %w", err)
	}

	return &result, nil
}

// ListResources lists all available resources
func (c *WebSocketClient) ListResources(ctx context.Context) ([]types.Resource, error) {
	// Send a listResources request
	data, err := c.sendRequest("listResources", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to list resources: %w", err)
	}

	// Parse the response
	var resources []types.Resource
	if err := json.Unmarshal(data, &resources); err != nil {
		return nil, fmt.Errorf("error parsing resources: %w", err)
	}

	return resources, nil
}

// ListResourcesPaginated lists resources with pagination
func (c *WebSocketClient) ListResourcesPaginated(ctx context.Context, options types.PaginationOptions) (*types.PaginatedResult, error) {
	// Send a listResources request with pagination
	data, err := c.sendRequest("listResources", options)
	if err != nil {
		return nil, fmt.Errorf("failed to list resources: %w", err)
	}

	// Parse the response
	var result types.PaginatedResult
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("error parsing paginated result: %w", err)
	}

	return &result, nil
}

// ReadResource reads a resource with the given name
func (c *WebSocketClient) ReadResource(ctx context.Context, name string) ([]byte, string, error) {
	// Send a readResource request
	params := struct {
		Name string `json:"name"`
	}{
		Name: name,
	}

	data, err := c.sendRequest("readResource", params)
	if err != nil {
		return nil, "", fmt.Errorf("failed to read resource: %w", err)
	}

	// Parse the response
	var result struct {
		Content  []byte `json:"content"`
		MimeType string `json:"mimeType"`
	}
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, "", fmt.Errorf("error parsing resource result: %w", err)
	}

	return result.Content, result.MimeType, nil
}

// Close closes the WebSocket connection
func (c *WebSocketClient) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}
