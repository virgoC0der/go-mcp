package sse

//
//import (
//	"bufio"
//	"context"
//	"encoding/json"
//	"fmt"
//	"net/http"
//	"strings"
//	"sync"
//
//	"github.com/virgoC0der/go-mcp/internal/types"
//)
//
//// SSEClient 实现了基于 Server-Sent Events 的客户端
//type SSEClient struct {
//	baseURL    string
//	httpClient *http.Client
//	eventChan  chan types.Response
//	done       chan struct{}
//	mu         sync.RWMutex
//	connected  bool
//}
//
//// NewSSEClient 创建一个新的 SSE 客户端
//func NewSSEClient(baseURL string) *SSEClient {
//	return &SSEClient{
//		baseURL:    baseURL,
//		httpClient: &http.Client{},
//		eventChan:  make(chan types.Response, 100),
//		done:       make(chan struct{}),
//	}
//}
//
//// Connect 连接到 SSE 服务器并开始接收事件
//func (c *SSEClient) Connect(ctx context.Context) error {
//	c.mu.Lock()
//	if c.connected {
//		c.mu.Unlock()
//		return fmt.Errorf("client already connected")
//	}
//	c.connected = true
//	c.mu.Unlock()
//
//	go func() {
//		defer func() {
//			c.mu.Lock()
//			c.connected = false
//			c.mu.Unlock()
//		}()
//
//		for {
//			select {
//			case <-ctx.Done():
//				return
//			case <-c.done:
//				return
//			default:
//				if err := c.connect(ctx); err != nil {
//					// 如果连接失败，等待一段时间后重试
//					select {
//					case <-ctx.Done():
//						return
//					case <-c.done:
//						return
//					default:
//						continue
//					}
//				}
//			}
//		}
//	}()
//
//	return nil
//}
//
//// connect 建立 SSE 连接并处理事件流
//func (c *SSEClient) connect(ctx context.Context) error {
//	req, err := http.NewRequestWithContext(ctx, "GET", c.baseURL+"/events", nil)
//	if err != nil {
//		return fmt.Errorf("create request failed: %w", err)
//	}
//
//	resp, err := c.httpClient.Do(req)
//	if err != nil {
//		return fmt.Errorf("request failed: %w", err)
//	}
//	defer resp.Body.Close()
//
//	if resp.StatusCode != http.StatusOK {
//		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
//	}
//
//	reader := bufio.NewReader(resp.Body)
//	errChan := make(chan error, 1)
//
//	go func() {
//		for {
//			line, err := reader.ReadString('\n')
//			if err != nil {
//				errChan <- fmt.Errorf("read line failed: %w", err)
//				return
//			}
//
//			line = strings.TrimSpace(line)
//			if line == "" {
//				continue
//			}
//
//			if strings.HasPrefix(line, "data: ") {
//				data := strings.TrimPrefix(line, "data: ")
//				var response types.Response
//				if err := json.Unmarshal([]byte(data), &response); err != nil {
//					continue
//				}
//
//				select {
//				case c.eventChan <- response:
//				case <-ctx.Done():
//					return
//				case <-c.done:
//					return
//				default:
//					// 如果通道已满，丢弃事件
//				}
//			}
//		}
//	}()
//
//	select {
//	case <-ctx.Done():
//		return ctx.Err()
//	case <-c.done:
//		return nil
//	case err := <-errChan:
//		return err
//	}
//}
//
//// Close 关闭 SSE 客户端连接
//func (c *SSEClient) Close() error {
//	c.mu.Lock()
//	defer c.mu.Unlock()
//
//	if !c.connected {
//		return nil
//	}
//
//	close(c.done)
//	c.connected = false
//	return nil
//}
//
//func (c *SSEClient) Service() types.MCPService {
//	return c
//}
//
//// Events 返回用于接收事件的通道
//func (c *SSEClient) Events() <-chan types.Response {
//	return c.eventChan
//}
//
//// ListPrompts 获取可用的提示列表
//func (c *SSEClient) ListPrompts(ctx context.Context) ([]types.Prompt, error) {
//	resp, err := c.doRequest(ctx, "GET", "/api/prompts", nil)
//	if err != nil {
//		return nil, err
//	}
//
//	resultBytes, err := json.Marshal(resp.Result)
//	if err != nil {
//		return nil, fmt.Errorf("marshal result failed: %w", err)
//	}
//
//	var prompts []types.Prompt
//	if err := json.Unmarshal(resultBytes, &prompts); err != nil {
//		return nil, fmt.Errorf("unmarshal prompts failed: %w", err)
//	}
//	return prompts, nil
//}
//
//// GetPrompt 获取指定提示的内容
//func (c *SSEClient) GetPrompt(ctx context.Context, name string, args map[string]any) (*types.GetPromptResult, error) {
//	resp, err := c.doRequest(ctx, "POST", fmt.Sprintf("/api/prompts/%s", name), args)
//	if err != nil {
//		return nil, err
//	}
//
//	resultBytes, err := json.Marshal(resp.Result)
//	if err != nil {
//		return nil, fmt.Errorf("marshal result failed: %w", err)
//	}
//
//	var result types.GetPromptResult
//	if err := json.Unmarshal(resultBytes, &result); err != nil {
//		return nil, fmt.Errorf("unmarshal prompt result failed: %w", err)
//	}
//	return &result, nil
//}
//
//// ListTools 获取可用的工具列表
//func (c *SSEClient) ListTools(ctx context.Context) ([]types.Tool, error) {
//	resp, err := c.doRequest(ctx, "GET", "/api/tools", nil)
//	if err != nil {
//		return nil, err
//	}
//
//	resultBytes, err := json.Marshal(resp.Result)
//	if err != nil {
//		return nil, fmt.Errorf("marshal result failed: %w", err)
//	}
//
//	var tools []types.Tool
//	if err := json.Unmarshal(resultBytes, &tools); err != nil {
//		return nil, fmt.Errorf("unmarshal tools failed: %w", err)
//	}
//	return tools, nil
//}
//
//// CallTool 调用指定的工具
//func (c *SSEClient) CallTool(ctx context.Context, name string, args map[string]any) (*types.CallToolResult, error) {
//	resp, err := c.doRequest(ctx, "POST", fmt.Sprintf("/api/tools/%s", name), args)
//	if err != nil {
//		return nil, err
//	}
//
//	resultBytes, err := json.Marshal(resp.Result)
//	if err != nil {
//		return nil, fmt.Errorf("marshal result failed: %w", err)
//	}
//
//	var result types.CallToolResult
//	if err := json.Unmarshal(resultBytes, &result); err != nil {
//		return nil, fmt.Errorf("unmarshal tool result failed: %w", err)
//	}
//	return &result, nil
//}
//
//// ListResources 获取可用的资源列表
//func (c *SSEClient) ListResources(ctx context.Context) ([]types.Resource, error) {
//	resp, err := c.doRequest(ctx, "GET", "/api/resources", nil)
//	if err != nil {
//		return nil, err
//	}
//
//	resultBytes, err := json.Marshal(resp.Result)
//	if err != nil {
//		return nil, fmt.Errorf("marshal result failed: %w", err)
//	}
//
//	var resources []types.Resource
//	if err := json.Unmarshal(resultBytes, &resources); err != nil {
//		return nil, fmt.Errorf("unmarshal resources failed: %w", err)
//	}
//	return resources, nil
//}
//
//// ReadResource 读取指定资源的内容
//func (c *SSEClient) ReadResource(ctx context.Context, name string) ([]byte, string, error) {
//	resp, err := c.doRequest(ctx, "GET", fmt.Sprintf("/api/resources/%s", name), nil)
//	if err != nil {
//		return nil, "", err
//	}
//
//	result, ok := resp.Result.(map[string]interface{})
//	if !ok {
//		return nil, "", fmt.Errorf("unexpected result type")
//	}
//
//	content, ok := result["content"].(string)
//	if !ok {
//		return nil, "", fmt.Errorf("content not found or invalid type")
//	}
//
//	contentType, ok := result["content_type"].(string)
//	if !ok {
//		return nil, "", fmt.Errorf("content_type not found or invalid type")
//	}
//
//	return []byte(content), contentType, nil
//}
//
//// doRequest 执行 HTTP 请求
//func (c *SSEClient) doRequest(ctx context.Context, method, path string, body interface{}) (*types.Response, error) {
//	var reqBody []byte
//	var err error
//	if body != nil {
//		reqBody, err = json.Marshal(body)
//		if err != nil {
//			return nil, fmt.Errorf("marshal request body failed: %w", err)
//		}
//	}
//
//	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, strings.NewReader(string(reqBody)))
//	if err != nil {
//		return nil, fmt.Errorf("create request failed: %w", err)
//	}
//
//	if body != nil {
//		req.Header.Set("Content-Type", "application/json")
//	}
//
//	resp, err := c.httpClient.Do(req)
//	if err != nil {
//		return nil, fmt.Errorf("request failed: %w", err)
//	}
//	defer resp.Body.Close()
//
//	var response types.Response
//	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
//		return nil, fmt.Errorf("decode response failed: %w", err)
//	}
//
//	if !response.Success {
//		return nil, fmt.Errorf("request failed: %s - %s", response.Error.Code, response.Error.Message)
//	}
//
//	return &response, nil
//}
