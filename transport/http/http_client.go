package http

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"

	"github.com/virgoC0der/go-mcp/internal/types"
)

// HTTPClient implements the Client interface using HTTP
type HTTPClient struct {
	serverURL string
	client    *http.Client
}

// NewHTTPClient creates a new HTTP client instance
func NewHTTPClient(serverAddress string) *HTTPClient {
	return &HTTPClient{
		serverURL: serverAddress,
		client:    &http.Client{},
	}
}

// Connect implements the Client interface
func (c *HTTPClient) Connect(ctx context.Context) error {
	// Validate server URL
	_, err := url.Parse(c.serverURL)
	return err
}

// Close implements the Client interface
func (c *HTTPClient) Close() error {
	// HTTP client doesn't need explicit cleanup
	return nil
}

// Service implements the Client interface
func (c *HTTPClient) Service() types.MCPService {
	return c
}

// ListPrompts implements the MCPService interface
func (c *HTTPClient) ListPrompts(ctx context.Context) ([]types.Prompt, error) {
	var prompts []types.Prompt
	err := c.get(ctx, "/prompts", &prompts)
	return prompts, err
}

// GetPrompt implements the MCPService interface
func (c *HTTPClient) GetPrompt(ctx context.Context, name string, args map[string]any) (*types.GetPromptResult, error) {
	var result types.GetPromptResult
	err := c.post(ctx, fmt.Sprintf("/prompts/%s", name), args, &result)
	return &result, err
}

// ListTools implements the MCPService interface
func (c *HTTPClient) ListTools(ctx context.Context) ([]types.Tool, error) {
	var tools []types.Tool
	err := c.get(ctx, "/tools", &tools)
	return tools, err
}

// CallTool implements the MCPService interface
func (c *HTTPClient) CallTool(ctx context.Context, name string, args map[string]any) (*types.CallToolResult, error) {
	var result types.CallToolResult
	err := c.post(ctx, fmt.Sprintf("/tools/%s", name), args, &result)
	return &result, err
}

// ListResources implements the MCPService interface
func (c *HTTPClient) ListResources(ctx context.Context) ([]types.Resource, error) {
	var resources []types.Resource
	err := c.get(ctx, "/resources", &resources)
	return resources, err
}

// ReadResource implements the MCPService interface
func (c *HTTPClient) ReadResource(ctx context.Context, name string) ([]byte, string, error) {
	url := c.buildURL(fmt.Sprintf("/resources/%s", name))
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, "", fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, "", fmt.Errorf("server returned error: %s", resp.Status)
	}

	content, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, "", fmt.Errorf("failed to read response: %w", err)
	}

	return content, resp.Header.Get("Content-Type"), nil
}

// Helper methods for HTTP requests
func (c *HTTPClient) get(ctx context.Context, endpoint string, result interface{}) error {
	url := c.buildURL(endpoint)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("server returned error: %s", resp.Status)
	}

	if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	return nil
}

func (c *HTTPClient) post(ctx context.Context, endpoint string, data interface{}, result interface{}) error {
	body, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to encode request body: %w", err)
	}

	url := c.buildURL(endpoint)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("server returned error: %s", resp.Status)
	}

	if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	return nil
}

func (c *HTTPClient) buildURL(endpoint string) string {
	return c.serverURL + path.Clean("/"+endpoint)
}
