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

	"github.com/virgoC0der/go-mcp/internal/response"
	"github.com/virgoC0der/go-mcp/internal/types"
)

// HTTPClient implements the Client interface using HTTP
type HTTPClient struct {
	serverURL string
	client    *http.Client
	jsonRPC   bool // Whether to use JSON-RPC
	requestID int  // JSON-RPC request ID counter
}

// Use common JSON-RPC request/response structures
type JSONRPCRequest = response.JSONRPCRequest
type JSONRPCResponse = response.JSONRPCResponse

// NewHTTPClient creates a new HTTP client instance
func NewHTTPClient(serverAddress string) *HTTPClient {
	return &HTTPClient{
		serverURL: serverAddress,
		client:    &http.Client{},
		jsonRPC:   true, // Use JSON-RPC by default
		requestID: 0,    // Initialize request ID
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

// nextRequestID generates the next request ID
func (c *HTTPClient) nextRequestID() interface{} {
	c.requestID++
	return c.requestID
}

// ListPrompts implements the MCPService interface
func (c *HTTPClient) ListPrompts(ctx context.Context, cursor string) (*types.PromptListResult, error) {
	if c.jsonRPC {
		var result types.PromptListResult
		err := c.jsonRPCCall(ctx, "prompts/list", map[string]interface{}{
			"cursor": cursor,
		}, &result)
		return &result, err
	}

	// 传统REST API回退
	var promptList types.PromptListResult
	endpoint := "/prompts"
	if cursor != "" {
		endpoint = fmt.Sprintf("%s?cursor=%s", endpoint, url.QueryEscape(cursor))
	}
	err := c.get(ctx, endpoint, &promptList)
	return &promptList, err
}

// GetPrompt implements the MCPService interface
func (c *HTTPClient) GetPrompt(ctx context.Context, name string, args map[string]any) (*types.PromptResult, error) {
	if c.jsonRPC {
		var result types.PromptResult
		err := c.jsonRPCCall(ctx, "prompts/get", map[string]interface{}{
			"name":      name,
			"arguments": args,
		}, &result)
		return &result, err
	}

	// 传统REST API回退
	var result types.PromptResult
	err := c.post(ctx, fmt.Sprintf("/prompts/%s", name), map[string]interface{}{
		"arguments": args,
	}, &result)
	return &result, err
}

// ListTools implements the MCPService interface
func (c *HTTPClient) ListTools(ctx context.Context, cursor string) (*types.ToolListResult, error) {
	if c.jsonRPC {
		var result types.ToolListResult
		err := c.jsonRPCCall(ctx, "tools/list", map[string]interface{}{
			"cursor": cursor,
		}, &result)
		return &result, err
	}

	// 传统REST API回退
	var toolList types.ToolListResult
	endpoint := "/tools"
	if cursor != "" {
		endpoint = fmt.Sprintf("%s?cursor=%s", endpoint, url.QueryEscape(cursor))
	}
	err := c.get(ctx, endpoint, &toolList)
	return &toolList, err
}

// CallTool implements the MCPService interface
func (c *HTTPClient) CallTool(ctx context.Context, name string, args map[string]any) (*types.CallToolResult, error) {
	if c.jsonRPC {
		// 对于JSON-RPC，直接获取CallToolResult
		var result types.CallToolResult
		err := c.jsonRPCCall(ctx, "tools/call", map[string]interface{}{
			"name":      name,
			"arguments": args,
		}, &result)

		if err != nil {
			return nil, err
		}

		return &result, nil
	}

	// 传统REST API回退
	var result types.CallToolResult
	err := c.post(ctx, fmt.Sprintf("/tools/%s", name), map[string]interface{}{
		"arguments": args,
	}, &result)
	return &result, err
}

// ListResources implements the MCPService interface
func (c *HTTPClient) ListResources(ctx context.Context, cursor string) (*types.ResourceListResult, error) {
	if c.jsonRPC {
		var result types.ResourceListResult
		err := c.jsonRPCCall(ctx, "resources/list", map[string]interface{}{
			"cursor": cursor,
		}, &result)
		return &result, err
	}

	// 传统REST API回退
	var resourceList types.ResourceListResult
	endpoint := "/resources"
	if cursor != "" {
		endpoint = fmt.Sprintf("%s?cursor=%s", endpoint, url.QueryEscape(cursor))
	}
	err := c.get(ctx, endpoint, &resourceList)
	return &resourceList, err
}

// ReadResource implements the MCPService interface
func (c *HTTPClient) ReadResource(ctx context.Context, uri string) (*types.ResourceContent, error) {
	if c.jsonRPC {
		var result types.ResourceContent
		err := c.jsonRPCCall(ctx, "resources/read", map[string]interface{}{
			"uri": uri,
		}, &result)
		return &result, err
	}

	// 传统REST API回退
	url := c.buildURL(fmt.Sprintf("/resources/%s", uri))
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("server returned error: %s", resp.Status)
	}

	content, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	mimeType := resp.Header.Get("Content-Type")
	result := &types.ResourceContent{
		URI:      uri,
		MimeType: mimeType,
	}

	// 根据MIME类型决定如何处理内容
	if isTextMIME(mimeType) {
		result.Text = string(content)
	} else {
		// 对于二进制内容，不再需要Base64编码，因为已经处理过了
		result.Blob = string(content)
	}

	return result, nil
}

// ListResourceTemplates implements the MCPService interface
func (c *HTTPClient) ListResourceTemplates(ctx context.Context) ([]types.ResourceTemplate, error) {
	if c.jsonRPC {
		var response struct {
			ResourceTemplates []types.ResourceTemplate `json:"resourceTemplates"`
		}
		err := c.jsonRPCCall(ctx, "resources/templates/list", nil, &response)
		return response.ResourceTemplates, err
	}

	// 对于没有实现此功能的传统API，返回空列表和错误
	return nil, types.NewError("not_implemented", "resource templates not supported in legacy API mode")
}

// SubscribeToResource implements the MCPService interface
func (c *HTTPClient) SubscribeToResource(ctx context.Context, uri string) error {
	if c.jsonRPC {
		var result bool
		err := c.jsonRPCCall(ctx, "resources/subscribe", map[string]interface{}{
			"uri": uri,
		}, &result)
		return err
	}

	// 对于没有实现此功能的传统API，返回错误
	return types.NewError("not_implemented", "resource subscription not supported in legacy API mode")
}

// Helper methods for HTTP requests

// jsonRPCCall executes a JSON-RPC call
func (c *HTTPClient) jsonRPCCall(ctx context.Context, method string, params interface{}, result interface{}) error {
	// Create request
	reqID := c.nextRequestID()
	request := response.NewJSONRPCRequest(reqID, method, params)

	// Serialize request
	reqBody, err := json.Marshal(request)
	if err != nil {
		return fmt.Errorf("failed to encode JSON-RPC request: %w", err)
	}

	// Create HTTP request
	url := c.buildURL("/jsonrpc")
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(reqBody))
	if err != nil {
		return fmt.Errorf("failed to create JSON-RPC request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	// Send request
	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send JSON-RPC request: %w", err)
	}
	defer resp.Body.Close()

	// Check HTTP status code
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("server returned HTTP error: %s", resp.Status)
	}

	// Parse response
	var response JSONRPCResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return fmt.Errorf("failed to decode JSON-RPC response: %w", err)
	}

	// Check JSON-RPC error
	if response.Error != nil {
		return fmt.Errorf("JSON-RPC error (code %d): %s", response.Error.Code, response.Error.Message)
	}

	// 将结果解析到目标结构体
	if result != nil && response.Result != nil {
		resultData, err := json.Marshal(response.Result)
		if err != nil {
			return fmt.Errorf("failed to re-encode JSON-RPC result: %w", err)
		}

		if err := json.Unmarshal(resultData, result); err != nil {
			return fmt.Errorf("failed to decode JSON-RPC result into target structure: %w", err)
		}
	}

	return nil
}

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

	var response types.Response
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	if !response.Success {
		if response.Error != nil {
			return types.NewError(response.Error.Code, response.Error.Message)
		}
		return fmt.Errorf("unknown error in response")
	}

	// 将结果解析到目标结构体
	if result != nil && response.Result != nil {
		resultData, err := json.Marshal(response.Result)
		if err != nil {
			return fmt.Errorf("failed to re-encode result: %w", err)
		}

		if err := json.Unmarshal(resultData, result); err != nil {
			return fmt.Errorf("failed to decode result into target structure: %w", err)
		}
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

	var response types.Response
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	if !response.Success {
		if response.Error != nil {
			return types.NewError(response.Error.Code, response.Error.Message)
		}
		return fmt.Errorf("unknown error in response")
	}

	// 将结果解析到目标结构体
	if result != nil && response.Result != nil {
		resultData, err := json.Marshal(response.Result)
		if err != nil {
			return fmt.Errorf("failed to re-encode result: %w", err)
		}

		if err := json.Unmarshal(resultData, result); err != nil {
			return fmt.Errorf("failed to decode result into target structure: %w", err)
		}
	}

	return nil
}

func (c *HTTPClient) buildURL(endpoint string) string {
	return c.serverURL + path.Clean("/"+endpoint)
}

// isTextMIME 判断MIME类型是否为文本类型
func isTextMIME(mimeType string) bool {
	switch {
	case mimeType == "application/json":
		return true
	case mimeType == "application/xml":
		return true
	case mimeType == "application/javascript":
		return true
	case mimeType == "text/plain":
		return true
	case mimeType == "text/html":
		return true
	case mimeType == "text/css":
		return true
	case mimeType == "text/xml":
		return true
	case mimeType == "text/markdown":
		return true
	default:
		// 检查前缀
		if len(mimeType) > 5 && mimeType[0:5] == "text/" {
			return true
		}
		return false
	}
}
