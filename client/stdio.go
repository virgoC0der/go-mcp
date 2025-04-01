package client

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"sync"

	"github.com/virgoC0der/go-mcp/types"
)

// StdioClient implements a client that communicates with a server over stdio
type StdioClient struct {
	reader      io.Reader
	writer      io.Writer
	requestID   int
	responses   map[string]chan []byte
	lock        sync.Mutex
	initialized bool
}

// NewStdioClient creates a new stdio client
func NewStdioClient(reader io.Reader, writer io.Writer) *StdioClient {
	c := &StdioClient{
		reader:    reader,
		writer:    writer,
		requestID: 0,
		responses: make(map[string]chan []byte),
	}

	// Start reading responses
	go c.readResponses()

	return c
}

// readResponses reads and processes responses from the reader
func (c *StdioClient) readResponses() {
	decoder := json.NewDecoder(c.reader)
	for {
		var response struct {
			Type    string          `json:"type"`
			ID      string          `json:"id"`
			Success bool            `json:"success"`
			Content json.RawMessage `json:"content"`
			Error   string          `json:"error"`
		}

		if err := decoder.Decode(&response); err != nil {
			if err == io.EOF {
				return
			}
			fmt.Printf("Error decoding response: %v\n", err)
			continue
		}

		// Get the response channel for this ID
		c.lock.Lock()
		ch, ok := c.responses[response.ID]
		c.lock.Unlock()

		if ok {
			// Send the response back to the waiting goroutine
			ch <- []byte(fmt.Sprintf(`{"success":%t,"content":%s,"error":%q}`,
				response.Success,
				string(response.Content),
				response.Error))

			// Delete the channel since we don't need it anymore
			c.lock.Lock()
			delete(c.responses, response.ID)
			c.lock.Unlock()
		}
	}
}

// sendRequest sends a request to the server and waits for a response
func (c *StdioClient) sendRequest(method string, params any) ([]byte, error) {
	// Generate a request ID
	c.lock.Lock()
	id := fmt.Sprintf("req-%d", c.requestID)
	c.requestID++
	c.lock.Unlock()

	// Prepare the request
	request := struct {
		Type   string `json:"type"`
		ID     string `json:"id"`
		Method string `json:"method"`
		Params any    `json:"params,omitempty"`
	}{
		Type:   "request",
		ID:     id,
		Method: method,
		Params: params,
	}

	// Create a channel to receive the response
	responseCh := make(chan []byte, 1)
	c.lock.Lock()
	c.responses[id] = responseCh
	c.lock.Unlock()

	// Send the request
	data, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("error marshaling request: %w", err)
	}

	if _, err := fmt.Fprintln(c.writer, string(data)); err != nil {
		return nil, fmt.Errorf("error writing request: %w", err)
	}

	// Wait for the response
	responseData := <-responseCh

	// Parse the response
	var response struct {
		Success bool            `json:"success"`
		Content json.RawMessage `json:"content"`
		Error   string          `json:"error"`
	}
	if err := json.Unmarshal(responseData, &response); err != nil {
		return nil, fmt.Errorf("error unmarshaling response: %w", err)
	}

	if !response.Success {
		return nil, fmt.Errorf("request failed: %s", response.Error)
	}

	return response.Content, nil
}

// Initialize initializes the client
func (c *StdioClient) Initialize(ctx context.Context) error {
	// Send an initialize request
	_, err := c.sendRequest("initialize", map[string]any{})
	if err != nil {
		return fmt.Errorf("initialization failed: %w", err)
	}

	c.initialized = true
	return nil
}

// ListPrompts lists all available prompts
func (c *StdioClient) ListPrompts(ctx context.Context) ([]types.Prompt, error) {
	// Send a listPrompts request
	data, err := c.sendRequest("listPrompts", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to list prompts: %w", err)
	}

	// Parse the response
	var prompts []types.Prompt
	if err := json.Unmarshal(data, &prompts); err != nil {
		return nil, fmt.Errorf("error unmarshaling prompts: %w", err)
	}

	return prompts, nil
}

// GetPrompt gets a prompt with the given name and arguments
func (c *StdioClient) GetPrompt(ctx context.Context, name string, arguments map[string]any) (*types.GetPromptResult, error) {
	// Send a getPrompt request
	params := struct {
		Name      string         `json:"name"`
		Arguments map[string]any `json:"arguments"`
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
		return nil, fmt.Errorf("error unmarshaling prompt result: %w", err)
	}

	return &result, nil
}

// ListTools lists all available tools
func (c *StdioClient) ListTools(ctx context.Context) ([]types.Tool, error) {
	// Send a listTools request
	data, err := c.sendRequest("listTools", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to list tools: %w", err)
	}

	// Parse the response
	var tools []types.Tool
	if err := json.Unmarshal(data, &tools); err != nil {
		return nil, fmt.Errorf("error unmarshaling tools: %w", err)
	}

	return tools, nil
}

// CallTool calls a tool with the given name and arguments
func (c *StdioClient) CallTool(ctx context.Context, name string, arguments map[string]any) (*types.CallToolResult, error) {
	// Send a callTool request
	params := struct {
		Name      string         `json:"name"`
		Arguments map[string]any `json:"arguments"`
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
		return nil, fmt.Errorf("error unmarshaling tool result: %w", err)
	}

	return &result, nil
}

// ListResources lists all available resources
func (c *StdioClient) ListResources(ctx context.Context) ([]types.Resource, error) {
	// Send a listResources request
	data, err := c.sendRequest("listResources", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to list resources: %w", err)
	}

	// Parse the response
	var resources []types.Resource
	if err := json.Unmarshal(data, &resources); err != nil {
		return nil, fmt.Errorf("error unmarshaling resources: %w", err)
	}

	return resources, nil
}

// ReadResource reads a resource with the given name
func (c *StdioClient) ReadResource(ctx context.Context, name string) ([]byte, string, error) {
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
		return nil, "", fmt.Errorf("error unmarshaling resource result: %w", err)
	}

	return result.Content, result.MimeType, nil
}

// Close closes the client
func (c *StdioClient) Close() error {
	// Nothing to do since we don't own the reader/writer
	return nil
}
