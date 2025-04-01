package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/virgoC0der/go-mcp/server"
	"github.com/virgoC0der/go-mcp/transport"
	"github.com/virgoC0der/go-mcp/types"
)

// Caiyun Weather API address
const CaiyunAPI = "https://api.caiyunapp.com/v2.6/"

// WeatherServer 实现一个天气服务
type WeatherServer struct {
	*server.BaseServer
	APIKey string // Caiyun Weather API key
}

// WeatherRequest 天气请求参数
type WeatherRequest struct {
	Longitude float64 `json:"longitude" jsonschema:"required,description=经度"`
	Latitude  float64 `json:"latitude" jsonschema:"required,description=纬度"`
	Language  string  `json:"language" jsonschema:"description=语言，支持zh_CN, en_US, ja等"`
}

// NewWeatherServer 创建一个新的天气服务实例
func NewWeatherServer(apiKey string) *WeatherServer {
	s := &WeatherServer{
		BaseServer: server.NewBaseServer("caiyun-weather", "1.0.0"),
		APIKey:     apiKey,
	}

	// Register weather tool
	err := s.RegisterToolTyped("getWeather", "获取指定位置的天气信息", server.CreateToolHandler(s.GetWeather))
	if err != nil {
		log.Fatalf("Failed to register weather tool: %v", err)
	}

	return s
}

// GetWeather 获取天气信息
func (s *WeatherServer) GetWeather(request WeatherRequest) (*types.CallToolResult, error) {
	// Set default language
	language := request.Language
	if language == "" {
		language = "zh_CN"
	}

	// Build API URL
	url := fmt.Sprintf("%s%s/%f,%f/weather?alert=true&dailysteps=1&hourlysteps=24&lang=%s",
		CaiyunAPI, s.APIKey, request.Longitude, request.Latitude, language)

	// Send HTTP request
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("请求天气API失败: %w", err)
	}
	defer resp.Body.Close()

	// Read response content
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}

	// Parse JSON response
	var result map[string]any
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("解析JSON失败: %w", err)
	}

	// 检查API响应状态
	status, ok := result["status"].(string)
	if !ok || status != "ok" {
		return nil, fmt.Errorf("API响应错误: %v", result["status"])
	}

	// 返回结果
	return &types.CallToolResult{
		Content: result,
	}, nil
}

func main() {
	// 从环境变量获取API密钥
	apiKey := os.Getenv("CAIYUN_API_KEY")

	// 创建天气服务
	srv := NewWeatherServer(apiKey)

	// 初始化服务
	err := srv.Initialize(context.Background(), types.InitializationOptions{
		ServerName:    "caiyun-weather",
		ServerVersion: "1.0.0",
		Capabilities: types.ServerCapabilities{
			Tools: true,
		},
	})
	if err != nil {
		log.Fatalf("初始化服务失败: %v", err)
	}

	// 创建HTTP和WebSocket服务器
	httpServer := transport.NewHTTPServer(srv, ":8080")
	wsServer := transport.NewWSServer(srv, ":8081")

	// 使用WaitGroup等待所有服务器关闭
	var wg sync.WaitGroup
	wg.Add(2)

	// 启动HTTP服务器
	go func() {
		defer wg.Done()
		log.Printf("HTTP服务器启动在 :8080")
		if err := httpServer.Start(); err != nil {
			log.Printf("HTTP服务器错误: %v", err)
		}
	}()

	// 启动WebSocket服务器
	go func() {
		defer wg.Done()
		log.Printf("WebSocket服务器启动在 :8081")
		if err := wsServer.Start(); err != nil {
			log.Printf("WebSocket服务器错误: %v", err)
		}
	}()

	// 处理中断信号
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	log.Println("正在关闭服务器...")

	// 创建带超时的上下文用于优雅关闭
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 关闭服务器
	if err := httpServer.Shutdown(ctx); err != nil {
		log.Printf("关闭HTTP服务器出错: %v", err)
	}
	if err := wsServer.Shutdown(ctx); err != nil {
		log.Printf("关闭WebSocket服务器出错: %v", err)
	}

	// 等待服务器关闭
	wg.Wait()
	log.Println("所有服务器已停止")
}
