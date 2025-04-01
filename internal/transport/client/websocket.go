package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/gorilla/websocket"
	"github.com/virgoC0der/go-mcp/internal/types"
)

// WebSocketClient implements the Client interface using WebSocket
type WebSocketClient struct {
	serverAddr string
	conn       *websocket.Conn
	service    types.MCPService
}

// NewWebSocketClient creates a new WebSocket client
func NewWebSocketClient(serverAddr string) *WebSocketClient {
	return &WebSocketClient{
		serverAddr: serverAddr,
	}
}

// Connect implements the Client interface
func (c *WebSocketClient) Connect(ctx context.Context) error {
	u := url.URL{Scheme: "ws", Host: c.serverAddr, Path: "/ws"}

	var err error
	c.conn, _, err = websocket.DefaultDialer.DialContext(ctx, u.String(), nil)
	if err != nil {
		return fmt.Errorf("failed to connect to WebSocket server: %w", err)
	}

	// Create a service proxy that forwards requests through the WebSocket connection
	c.service = &wsServiceProxy{conn: c.conn}
	return nil
}

// Close implements the Client interface
func (c *WebSocketClient) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

// Service implements the Client interface
func (c *WebSocketClient) Service() types.MCPService {
	return c.service
}

// wsServiceProxy implements the MCPService interface by forwarding requests through WebSocket
type wsServiceProxy struct {
	conn *websocket.Conn
}

// ListPrompts implements the MCPService interface
func (p *wsServiceProxy) ListPrompts(ctx context.Context) ([]types.Prompt, error) {
	var result []types.Prompt
	err := p.sendRequest(ctx, "ListPrompts", nil, &result)
	return result, err
}

// GetPrompt implements the MCPService interface
func (p *wsServiceProxy) GetPrompt(ctx context.Context, name string, args map[string]any) (*types.GetPromptResult, error) {
	var result types.GetPromptResult
	err := p.sendRequest(ctx, "GetPrompt", map[string]any{"name": name, "args": args}, &result)
	return &result, err
}

// ListTools implements the MCPService interface
func (p *wsServiceProxy) ListTools(ctx context.Context) ([]types.Tool, error) {
	var result []types.Tool
	err := p.sendRequest(ctx, "ListTools", nil, &result)
	return result, err
}

// CallTool implements the MCPService interface
func (p *wsServiceProxy) CallTool(ctx context.Context, name string, args map[string]any) (*types.CallToolResult, error) {
	var result types.CallToolResult
	err := p.sendRequest(ctx, "CallTool", map[string]any{"name": name, "args": args}, &result)
	return &result, err
}

// ListResources implements the MCPService interface
func (p *wsServiceProxy) ListResources(ctx context.Context) ([]types.Resource, error) {
	var result []types.Resource
	err := p.sendRequest(ctx, "ListResources", nil, &result)
	return result, err
}

// ReadResource implements the MCPService interface
func (p *wsServiceProxy) ReadResource(ctx context.Context, name string) ([]byte, string, error) {
	var result struct {
		Content  []byte `json:"content"`
		MimeType string `json:"mimeType"`
	}
	err := p.sendRequest(ctx, "ReadResource", map[string]any{"name": name}, &result)
	return result.Content, result.MimeType, err
}

// sendRequest sends a request through the WebSocket connection and waits for the response
func (p *wsServiceProxy) sendRequest(ctx context.Context, method string, params any, result ...any) error {
	// Send request
	req := map[string]any{
		"method": method,
		"params": params,
	}
	if err := p.conn.WriteJSON(req); err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}

	// Read response
	var resp struct {
		Error  string          `json:"error,omitempty"`
		Result json.RawMessage `json:"result,omitempty"`
	}
	if err := p.conn.ReadJSON(&resp); err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	if resp.Error != "" {
		return fmt.Errorf("server error: %s", resp.Error)
	}

	if len(result) > 0 && result[0] != nil {
		if err := json.Unmarshal(resp.Result, result[0]); err != nil {
			return fmt.Errorf("failed to unmarshal response: %w", err)
		}
	}

	return nil
}
