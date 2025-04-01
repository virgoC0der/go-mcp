package client

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/virgoC0der/go-mcp/types"
)

// Client defines the interface for an MCP client
type Client interface {
	// Initialize initializes the client connection
	Initialize(ctx context.Context) error

	// GetPrompt retrieves a prompt with the given name and arguments
	GetPrompt(ctx context.Context, name string, arguments map[string]any) (*types.GetPromptResult, error)

	// ListPrompts lists all available prompts
	ListPrompts(ctx context.Context) ([]types.Prompt, error)

	// ListPromptsPaginated lists prompts with pagination
	ListPromptsPaginated(ctx context.Context, options types.PaginationOptions) (*types.PaginatedResult, error)

	// CallTool calls a tool with the given name and arguments
	CallTool(ctx context.Context, name string, arguments map[string]any) (*types.CallToolResult, error)

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
func (c *HTTPClient) GetPrompt(ctx context.Context, name string, arguments map[string]any) (*types.GetPromptResult, error) {
	// Create the request payload
	payload := map[string]any{
		"args": arguments,
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
		c.baseURL+"prompts/"+name,
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

	// First try to decode as wrapped response format
	var wrappedResponse struct {
		Success bool                  `json:"success"`
		Result  types.GetPromptResult `json:"result"`
	}

	// Decode response
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Try to decode as wrapped response
	if err := json.Unmarshal(respBody, &wrappedResponse); err == nil && wrappedResponse.Success {
		return &wrappedResponse.Result, nil
	}

	// If decoding as wrapped response fails, try to decode directly as result
	var result types.GetPromptResult
	if err := json.Unmarshal(respBody, &result); err != nil {
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
		http.NoBody,
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

	// First try to decode as wrapped response format
	var wrappedResponse struct {
		Success bool           `json:"success"`
		Result  []types.Prompt `json:"result"`
	}

	// Decode response
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Try to decode as wrapped response
	if err := json.Unmarshal(respBody, &wrappedResponse); err == nil && wrappedResponse.Success {
		return wrappedResponse.Result, nil
	}

	// If decoding as wrapped response fails, try to decode directly as prompt list
	var prompts []types.Prompt
	if err := json.Unmarshal(respBody, &prompts); err != nil {
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
		http.NoBody,
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

	// First try to decode as wrapped response format
	var wrappedResponse struct {
		Success bool                  `json:"success"`
		Result  types.PaginatedResult `json:"result"`
	}

	// Decode response
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Try to decode as wrapped response
	if err := json.Unmarshal(respBody, &wrappedResponse); err == nil && wrappedResponse.Success {
		return &wrappedResponse.Result, nil
	}

	// If decoding as wrapped response fails, try to decode directly as result
	var result types.PaginatedResult
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}

// CallTool calls a tool with the given name and arguments
func (c *HTTPClient) CallTool(ctx context.Context, name string, arguments map[string]any) (*types.CallToolResult, error) {
	// Create the request payload
	payload := map[string]any{
		"args": arguments,
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
		c.baseURL+"tools/"+name,
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

	// First try to decode as wrapped response format
	var wrappedResponse struct {
		Success bool                 `json:"success"`
		Result  types.CallToolResult `json:"result"`
	}

	// Decode response
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Try to decode as wrapped response
	if err := json.Unmarshal(respBody, &wrappedResponse); err == nil && wrappedResponse.Success {
		return &wrappedResponse.Result, nil
	}

	// If decoding as wrapped response fails, try to decode directly as result
	var result types.CallToolResult
	if err := json.Unmarshal(respBody, &result); err != nil {
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
		http.NoBody,
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

	// First try to decode as wrapped response format
	var wrappedResponse struct {
		Success bool         `json:"success"`
		Result  []types.Tool `json:"result"`
	}

	// Decode response
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Try to decode as wrapped response
	if err := json.Unmarshal(respBody, &wrappedResponse); err == nil && wrappedResponse.Success {
		return wrappedResponse.Result, nil
	}

	// If decoding as wrapped response fails, try to decode directly as tool list
	var tools []types.Tool
	if err := json.Unmarshal(respBody, &tools); err != nil {
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
		http.NoBody,
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

	// First try to decode as wrapped response format
	var wrappedResponse struct {
		Success bool                  `json:"success"`
		Result  types.PaginatedResult `json:"result"`
	}

	// Decode response
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Try to decode as wrapped response
	if err := json.Unmarshal(respBody, &wrappedResponse); err == nil && wrappedResponse.Success {
		return &wrappedResponse.Result, nil
	}

	// If decoding as wrapped response fails, try to decode directly as result
	var result types.PaginatedResult
	if err := json.Unmarshal(respBody, &result); err != nil {
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
		c.baseURL+"resources/"+name,
		http.NoBody,
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

	// Check Content-Type header, if not application/json, it means the server returned resource content directly
	contentType := resp.Header.Get("Content-Type")
	if contentType != "application/json" {
		// Read response body directly
		content, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, "", fmt.Errorf("failed to read response body: %w", err)
		}
		return content, contentType, nil
	}

	// Read response body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, "", fmt.Errorf("failed to read response body: %w", err)
	}

	// First try to parse as wrapped response format
	var wrappedResponse struct {
		Success bool `json:"success"`
		Result  struct {
			Content  string `json:"content"`
			MimeType string `json:"mimeType"`
		} `json:"result"`
	}

	if err := json.Unmarshal(respBody, &wrappedResponse); err == nil && wrappedResponse.Success {
		// Decode Base64 content
		content, err := base64.StdEncoding.DecodeString(wrappedResponse.Result.Content)
		if err != nil {
			return nil, "", fmt.Errorf("failed to decode base64 content: %w", err)
		}
		return content, wrappedResponse.Result.MimeType, nil
	}

	// If not a wrapped response, try to decode directly
	var rawResponse struct {
		Content  string `json:"content"`
		MimeType string `json:"mimeType"`
	}

	if err := json.Unmarshal(respBody, &rawResponse); err != nil {
		return nil, "", fmt.Errorf("failed to decode response: %w", err)
	}

	// Decode Base64 content
	content, err := base64.StdEncoding.DecodeString(rawResponse.Content)
	if err != nil {
		return nil, "", fmt.Errorf("failed to decode base64 content: %w", err)
	}

	return content, rawResponse.MimeType, nil
}

// ListResources lists all available resources
func (c *HTTPClient) ListResources(ctx context.Context) ([]types.Resource, error) {
	// Create the request
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		c.baseURL+"resources",
		http.NoBody,
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

	// First try to decode as wrapped response format
	var wrappedResponse struct {
		Success bool             `json:"success"`
		Result  []types.Resource `json:"result"`
	}

	// Decode response
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Try to decode as wrapped response
	if err := json.Unmarshal(respBody, &wrappedResponse); err == nil && wrappedResponse.Success {
		return wrappedResponse.Result, nil
	}

	// If decoding as wrapped response fails, try to decode directly as resource list
	var resources []types.Resource
	if err := json.Unmarshal(respBody, &resources); err != nil {
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
		http.NoBody,
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

	// First try to decode as wrapped response format
	var wrappedResponse struct {
		Success bool                  `json:"success"`
		Result  types.PaginatedResult `json:"result"`
	}

	// Decode response
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Try to decode as wrapped response
	if err := json.Unmarshal(respBody, &wrappedResponse); err == nil && wrappedResponse.Success {
		return &wrappedResponse.Result, nil
	}

	// If decoding as wrapped response fails, try to decode directly as result
	var result types.PaginatedResult
	if err := json.Unmarshal(respBody, &result); err != nil {
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
	_, err := c.transport.SendRequest(ctx, "initialize", map[string]any{
		"serverName":    "client",
		"serverVersion": "1.0.0",
	})
	return err
}

// GetPrompt retrieves a prompt with the given name and arguments
func (c *BaseClient) GetPrompt(ctx context.Context, name string, arguments map[string]any) (*types.GetPromptResult, error) {
	resp, err := c.transport.SendRequest(ctx, "getPrompt", map[string]any{
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
	resp, err := c.transport.SendRequest(ctx, "listPrompts", map[string]any{
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
func (c *BaseClient) CallTool(ctx context.Context, name string, arguments map[string]any) (*types.CallToolResult, error) {
	resp, err := c.transport.SendRequest(ctx, "callTool", map[string]any{
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
	resp, err := c.transport.SendRequest(ctx, "listTools", map[string]any{
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
	resp, err := c.transport.SendRequest(ctx, "readResource", map[string]any{
		"name": name,
	})
	if err != nil {
		return nil, "", err
	}

	// Convert response to resource data
	resourceResp, ok := resp.(map[string]any)
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
	resp, err := c.transport.SendRequest(ctx, "listResources", map[string]any{
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
