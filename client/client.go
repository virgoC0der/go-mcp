package client

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/virgoC0der/go-mcp/types"
)

// Client defines the interface for an MCP client
type Client interface {
	// Initialize initializes the client connection
	Initialize(ctx context.Context) error

	// GetPrompt retrieves a prompt with the given name and arguments
	GetPrompt(ctx context.Context, name string, arguments map[string]interface{}) (*types.GetPromptResult, error)

	// ListPrompts lists all available prompts
	ListPrompts(ctx context.Context) ([]types.Prompt, error)

	// ListPromptsPaginated lists prompts with pagination
	ListPromptsPaginated(ctx context.Context, options types.PaginationOptions) (*types.PaginatedResult, error)

	// CallTool calls a tool with the given name and arguments
	CallTool(ctx context.Context, name string, arguments map[string]interface{}) (*types.CallToolResult, error)

	// ListTools lists all available tools
	ListTools(ctx context.Context) ([]types.Tool, error)

	// ListToolsPaginated lists tools with pagination
	ListToolsPaginated(ctx context.Context, options types.PaginationOptions) (*types.PaginatedResult, error)

	// ReadResource reads a resource with the given name
	ReadResource(ctx context.Context, name string) ([]byte, string, error)

	// ListResources lists all available resources
	ListResources(ctx context.Context) ([]types.Resource, error)

	// ListResourcesPaginated lists resources with pagination
	ListResourcesPaginated(ctx context.Context, options types.PaginationOptions) (*types.PaginatedResult, error)

	// Close closes the client connection
	Close() error
}

// HTTPClient is an HTTP-based MCP client
type HTTPClient struct {
	baseURL    string
	httpClient *http.Client
}

// NewHTTPClient creates a new HTTP client
func NewHTTPClient(baseURL string) *HTTPClient {
	// Ensure the base URL ends with a slash
	if !strings.HasSuffix(baseURL, "/") {
		baseURL += "/"
	}

	return &HTTPClient{
		baseURL:    baseURL,
		httpClient: &http.Client{},
	}
}

// Initialize initializes the client connection
func (c *HTTPClient) Initialize(ctx context.Context) error {
	// For HTTP clients, initialization is just checking that the server is reachable
	_, err := c.httpClient.Get(c.baseURL + "prompts")
	if err != nil {
		return fmt.Errorf("failed to connect to server: %w", err)
	}

	return nil
}

// GetPrompt retrieves a prompt with the given name and arguments
func (c *HTTPClient) GetPrompt(ctx context.Context, name string, arguments map[string]interface{}) (*types.GetPromptResult, error) {
	// Create the request payload
	payload := map[string]interface{}{
		"name":      name,
		"arguments": arguments,
	}

	// Encode the payload as JSON
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to encode payload: %w", err)
	}

	// Create the request
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		c.baseURL+"prompt",
		strings.NewReader(string(payloadBytes)),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")

	// Send the request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Check the response status
	if resp.StatusCode != http.StatusOK {
		var errorResp struct {
			Error string `json:"error"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&errorResp); err == nil && errorResp.Error != "" {
			return nil, fmt.Errorf("server error: %s", errorResp.Error)
		}
		return nil, fmt.Errorf("server returned non-OK status: %s", resp.Status)
	}

	// Decode the response
	var result types.GetPromptResult
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}

// ListPrompts lists all available prompts
func (c *HTTPClient) ListPrompts(ctx context.Context) ([]types.Prompt, error) {
	// Create the request
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		c.baseURL+"prompts",
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Send the request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Check the response status
	if resp.StatusCode != http.StatusOK {
		var errorResp struct {
			Error string `json:"error"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&errorResp); err == nil && errorResp.Error != "" {
			return nil, fmt.Errorf("server error: %s", errorResp.Error)
		}
		return nil, fmt.Errorf("server returned non-OK status: %s", resp.Status)
	}

	// Decode the response
	var prompts []types.Prompt
	if err := json.NewDecoder(resp.Body).Decode(&prompts); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return prompts, nil
}

// ListPromptsPaginated lists prompts with pagination
func (c *HTTPClient) ListPromptsPaginated(ctx context.Context, options types.PaginationOptions) (*types.PaginatedResult, error) {
	// Create the request
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		fmt.Sprintf("%sprompts?page=%d&pageSize=%d", c.baseURL, options.Page, options.PageSize),
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Send the request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Check the response status
	if resp.StatusCode != http.StatusOK {
		var errorResp struct {
			Error string `json:"error"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&errorResp); err == nil && errorResp.Error != "" {
			return nil, fmt.Errorf("server error: %s", errorResp.Error)
		}
		return nil, fmt.Errorf("server returned non-OK status: %s", resp.Status)
	}

	// Decode the response
	var result types.PaginatedResult
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}

// CallTool calls a tool with the given name and arguments
func (c *HTTPClient) CallTool(ctx context.Context, name string, arguments map[string]interface{}) (*types.CallToolResult, error) {
	// Create the request payload
	payload := map[string]interface{}{
		"name":      name,
		"arguments": arguments,
	}

	// Encode the payload as JSON
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to encode payload: %w", err)
	}

	// Create the request
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		c.baseURL+"tool",
		strings.NewReader(string(payloadBytes)),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")

	// Send the request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Check the response status
	if resp.StatusCode != http.StatusOK {
		var errorResp struct {
			Error string `json:"error"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&errorResp); err == nil && errorResp.Error != "" {
			return nil, fmt.Errorf("server error: %s", errorResp.Error)
		}
		return nil, fmt.Errorf("server returned non-OK status: %s", resp.Status)
	}

	// Decode the response
	var result types.CallToolResult
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}

// ListTools lists all available tools
func (c *HTTPClient) ListTools(ctx context.Context) ([]types.Tool, error) {
	// Create the request
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		c.baseURL+"tools",
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Send the request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Check the response status
	if resp.StatusCode != http.StatusOK {
		var errorResp struct {
			Error string `json:"error"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&errorResp); err == nil && errorResp.Error != "" {
			return nil, fmt.Errorf("server error: %s", errorResp.Error)
		}
		return nil, fmt.Errorf("server returned non-OK status: %s", resp.Status)
	}

	// Decode the response
	var tools []types.Tool
	if err := json.NewDecoder(resp.Body).Decode(&tools); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return tools, nil
}

// ListToolsPaginated lists tools with pagination
func (c *HTTPClient) ListToolsPaginated(ctx context.Context, options types.PaginationOptions) (*types.PaginatedResult, error) {
	// Create the request
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		fmt.Sprintf("%stools?page=%d&pageSize=%d", c.baseURL, options.Page, options.PageSize),
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Send the request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Check the response status
	if resp.StatusCode != http.StatusOK {
		var errorResp struct {
			Error string `json:"error"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&errorResp); err == nil && errorResp.Error != "" {
			return nil, fmt.Errorf("server error: %s", errorResp.Error)
		}
		return nil, fmt.Errorf("server returned non-OK status: %s", resp.Status)
	}

	// Decode the response
	var result types.PaginatedResult
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}

// ReadResource reads a resource with the given name
func (c *HTTPClient) ReadResource(ctx context.Context, name string) ([]byte, string, error) {
	// Create the request
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		c.baseURL+"resource/"+name,
		nil,
	)
	if err != nil {
		return nil, "", fmt.Errorf("failed to create request: %w", err)
	}

	// Send the request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Check the response status
	if resp.StatusCode != http.StatusOK {
		var errorResp struct {
			Error string `json:"error"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&errorResp); err == nil && errorResp.Error != "" {
			return nil, "", fmt.Errorf("server error: %s", errorResp.Error)
		}
		return nil, "", fmt.Errorf("server returned non-OK status: %s", resp.Status)
	}

	// Get the content type
	contentType := resp.Header.Get("Content-Type")

	// Read the response body
	var content []byte
	if err := json.NewDecoder(resp.Body).Decode(&content); err != nil {
		return nil, "", fmt.Errorf("failed to decode response: %w", err)
	}

	return content, contentType, nil
}

// ListResources lists all available resources
func (c *HTTPClient) ListResources(ctx context.Context) ([]types.Resource, error) {
	// Create the request
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		c.baseURL+"resources",
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Send the request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Check the response status
	if resp.StatusCode != http.StatusOK {
		var errorResp struct {
			Error string `json:"error"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&errorResp); err == nil && errorResp.Error != "" {
			return nil, fmt.Errorf("server error: %s", errorResp.Error)
		}
		return nil, fmt.Errorf("server returned non-OK status: %s", resp.Status)
	}

	// Decode the response
	var resources []types.Resource
	if err := json.NewDecoder(resp.Body).Decode(&resources); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return resources, nil
}

// ListResourcesPaginated lists resources with pagination
func (c *HTTPClient) ListResourcesPaginated(ctx context.Context, options types.PaginationOptions) (*types.PaginatedResult, error) {
	// Create the request
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		fmt.Sprintf("%sresources?page=%d&pageSize=%d", c.baseURL, options.Page, options.PageSize),
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Send the request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Check the response status
	if resp.StatusCode != http.StatusOK {
		var errorResp struct {
			Error string `json:"error"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&errorResp); err == nil && errorResp.Error != "" {
			return nil, fmt.Errorf("server error: %s", errorResp.Error)
		}
		return nil, fmt.Errorf("server returned non-OK status: %s", resp.Status)
	}

	// Decode the response
	var result types.PaginatedResult
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}

// Close closes the client connection
func (c *HTTPClient) Close() error {
	// HTTP clients don't need to be closed
	return nil
}

// BaseClient is the base implementation of the Client interface
type BaseClient struct {
	transport Transport
}

// NewClient creates a new client with the specified transport
func NewClient(transport Transport) Client {
	return &BaseClient{
		transport: transport,
	}
}

// Initialize initializes the client connection
func (c *BaseClient) Initialize(ctx context.Context) error {
	_, err := c.transport.SendRequest(ctx, "initialize", map[string]interface{}{
		"serverName":    "client",
		"serverVersion": "1.0.0",
	})
	return err
}

// GetPrompt retrieves a prompt with the given name and arguments
func (c *BaseClient) GetPrompt(ctx context.Context, name string, arguments map[string]interface{}) (*types.GetPromptResult, error) {
	resp, err := c.transport.SendRequest(ctx, "getPrompt", map[string]interface{}{
		"name": name,
		"args": arguments,
	})
	if err != nil {
		return nil, err
	}

	// Convert response to GetPromptResult
	result, ok := resp.(*types.GetPromptResult)
	if !ok {
		return nil, fmt.Errorf("unexpected response type: %T", resp)
	}

	return result, nil
}

// ListPrompts lists all available prompts
func (c *BaseClient) ListPrompts(ctx context.Context) ([]types.Prompt, error) {
	resp, err := c.transport.SendRequest(ctx, "listPrompts", nil)
	if err != nil {
		return nil, err
	}

	// Convert response to []types.Prompt
	prompts, ok := resp.([]types.Prompt)
	if !ok {
		return nil, fmt.Errorf("unexpected response type: %T", resp)
	}

	return prompts, nil
}

// ListPromptsPaginated lists prompts with pagination
func (c *BaseClient) ListPromptsPaginated(ctx context.Context, options types.PaginationOptions) (*types.PaginatedResult, error) {
	resp, err := c.transport.SendRequest(ctx, "listPrompts", map[string]interface{}{
		"page":     options.Page,
		"pageSize": options.PageSize,
	})
	if err != nil {
		return nil, err
	}

	// Convert response to PaginatedResult
	result, ok := resp.(*types.PaginatedResult)
	if !ok {
		return nil, fmt.Errorf("unexpected response type: %T", resp)
	}

	return result, nil
}

// CallTool calls a tool with the given name and arguments
func (c *BaseClient) CallTool(ctx context.Context, name string, arguments map[string]interface{}) (*types.CallToolResult, error) {
	resp, err := c.transport.SendRequest(ctx, "callTool", map[string]interface{}{
		"name": name,
		"args": arguments,
	})
	if err != nil {
		return nil, err
	}

	// Convert response to CallToolResult
	result, ok := resp.(*types.CallToolResult)
	if !ok {
		return nil, fmt.Errorf("unexpected response type: %T", resp)
	}

	return result, nil
}

// ListTools lists all available tools
func (c *BaseClient) ListTools(ctx context.Context) ([]types.Tool, error) {
	resp, err := c.transport.SendRequest(ctx, "listTools", nil)
	if err != nil {
		return nil, err
	}

	// Convert response to []types.Tool
	tools, ok := resp.([]types.Tool)
	if !ok {
		return nil, fmt.Errorf("unexpected response type: %T", resp)
	}

	return tools, nil
}

// ListToolsPaginated lists tools with pagination
func (c *BaseClient) ListToolsPaginated(ctx context.Context, options types.PaginationOptions) (*types.PaginatedResult, error) {
	resp, err := c.transport.SendRequest(ctx, "listTools", map[string]interface{}{
		"page":     options.Page,
		"pageSize": options.PageSize,
	})
	if err != nil {
		return nil, err
	}

	// Convert response to PaginatedResult
	result, ok := resp.(*types.PaginatedResult)
	if !ok {
		return nil, fmt.Errorf("unexpected response type: %T", resp)
	}

	return result, nil
}

// ReadResource reads a resource with the given name
func (c *BaseClient) ReadResource(ctx context.Context, name string) ([]byte, string, error) {
	resp, err := c.transport.SendRequest(ctx, "readResource", map[string]interface{}{
		"name": name,
	})
	if err != nil {
		return nil, "", err
	}

	// Convert response to resource data
	resourceResp, ok := resp.(map[string]interface{})
	if !ok {
		return nil, "", fmt.Errorf("unexpected response type: %T", resp)
	}

	contentStr, ok := resourceResp["content"].(string)
	if !ok {
		return nil, "", fmt.Errorf("unexpected content type: %T", resourceResp["content"])
	}

	// Decode base64 content
	content, err := base64.StdEncoding.DecodeString(contentStr)
	if err != nil {
		return nil, "", fmt.Errorf("failed to decode base64 content: %w", err)
	}

	mimeType, ok := resourceResp["mimeType"].(string)
	if !ok {
		return nil, "", fmt.Errorf("unexpected mimeType type: %T", resourceResp["mimeType"])
	}

	return content, mimeType, nil
}

// ListResources lists all available resources
func (c *BaseClient) ListResources(ctx context.Context) ([]types.Resource, error) {
	resp, err := c.transport.SendRequest(ctx, "listResources", nil)
	if err != nil {
		return nil, err
	}

	// Convert response to []types.Resource
	resources, ok := resp.([]types.Resource)
	if !ok {
		return nil, fmt.Errorf("unexpected response type: %T", resp)
	}

	return resources, nil
}

// ListResourcesPaginated lists resources with pagination
func (c *BaseClient) ListResourcesPaginated(ctx context.Context, options types.PaginationOptions) (*types.PaginatedResult, error) {
	resp, err := c.transport.SendRequest(ctx, "listResources", map[string]interface{}{
		"page":     options.Page,
		"pageSize": options.PageSize,
	})
	if err != nil {
		return nil, err
	}

	// Convert response to PaginatedResult
	result, ok := resp.(*types.PaginatedResult)
	if !ok {
		return nil, fmt.Errorf("unexpected response type: %T", resp)
	}

	return result, nil
}

// Close closes the client connection
func (c *BaseClient) Close() error {
	return c.transport.Close()
}
