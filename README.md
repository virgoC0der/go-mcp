# Go-MCP

[![Go Test and Lint](https://github.com/virgoC0der/go-mcp/actions/workflows/go.yml/badge.svg)](https://github.com/virgoC0der/go-mcp/actions/workflows/go.yml)
[![codecov](https://codecov.io/gh/virgoC0der/go-mcp/branch/main/graph/badge.svg)](https://codecov.io/gh/virgoC0der/go-mcp)
[![GoDoc](https://godoc.org/github.com/virgoC0der/go-mcp?status.svg)](https://godoc.org/github.com/virgoC0der/go-mcp)

Go-MCP is a Go implementation of the Model Context Protocol (MCP). MCP is a protocol for building AI services, defining three core primitives: Prompts, Tools, and Resources.

## Features

- Complete MCP protocol implementation
- Type-safe API
- Multiple transport options (HTTP, SSE)
- Unified response structure
- Pagination support
- Change notifications support

## Installation

```bash
go get github.com/virgoC0der/go-mcp
```

## Quick Start

### Creating a Server

```go
package main

import (
    "context"
    "log"
    "github.com/virgoC0der/go-mcp"
    "github.com/virgoC0der/go-mcp/internal/types"
)

// Implement MCPService interface
type MyService struct {
    // ... your service implementation
}

func main() {
    // Create service instance
    service := &MyService{}

    // Create server
    server, err := mcp.NewServer(service, &types.ServerOptions{
        Address: ":8080",
        Type:    "sse", // or "http"
    })
    if err != nil {
        log.Fatal(err)
    }

    // Initialize server
    ctx := context.Background()
    if err := server.Initialize(ctx, nil); err != nil {
        log.Fatal(err)
    }

    // Start server
    if err := server.Start(); err != nil {
        log.Fatal(err)
    }
}
```

### Creating a Client

```go
package main

import (
    "context"
    "log"
    "github.com/virgoC0der/go-mcp"
    "github.com/virgoC0der/go-mcp/internal/types"
)

func main() {
    // Create client
    client, err := mcp.NewClient(&types.ClientOptions{
        ServerAddress: "localhost:8080",
        Type:         "http", // or "sse", "websocket"
        UseJSONRPC: true,
        SubscribeToNotifications: true,
    })
    if err != nil {
        log.Fatal(err)
    }

    // Connect to server
    ctx := context.Background()
    if err := client.Connect(ctx); err != nil {
        log.Fatal(err)
    }
    defer client.Close()

    // Get service interface
    service := client.Service()

    // Use service
    result, err := service.ListPrompts(ctx, "")
    if err != nil {
        log.Fatal(err)
    }
    log.Printf("Available prompts: %v", result.Prompts)

    // Get next page if available
    if result.NextCursor != "" {
        nextPage, err := service.ListPrompts(ctx, result.NextCursor)
        if err != nil {
            log.Fatal(err)
        }
        log.Printf("Next page prompts: %v", nextPage.Prompts)
    }
}
```

## Response Structure

### JSON-RPC Responses

For JSON-RPC endpoints, responses follow the JSON-RPC 2.0 specification with a unified response handling system across all transport layers:

```go
type JSONRPCResponse struct {
    JSONRPC string        `json:"jsonrpc"`
    ID      interface{}   `json:"id"`
    Result  interface{}   `json:"result,omitempty"`
    Error   *JSONRPCError `json:"error,omitempty"`
}

type JSONRPCError struct {
    Code    int         `json:"code"`
    Message string      `json:"message"`
    Data    interface{} `json:"data,omitempty"`
}
```

JSON-RPC Success Response Example:
```json
{
    "jsonrpc": "2.0",
    "id": 1,
    "result": {
        "prompts": [
            {
                "name": "example_prompt",
                "description": "An example prompt"
            }
        ],
        "nextCursor": ""
    }
}
```

JSON-RPC Error Response Example:
```json
{
    "jsonrpc": "2.0",
    "id": 1,
    "error": {
        "code": -32602,
        "message": "Invalid params",
        "data": "Missing required parameter: name"
    }
}
```

The library provides a unified response handling system that works across HTTP, WebSocket, and stdio transport layers, ensuring consistent error handling and response formatting.

## Examples

- [Echo Server](examples/echo/main.go) - A simple echo server example
- [Weather Service](examples/weather/main.go) - A weather service example using OpenWeatherMap API
- [App Launcher](examples/app-launcher/main.go) - A macOS application launcher example with stdio server support

## API Documentation

### Server Interface

Servers must implement the `types.Server` interface:

```go
type Server interface {
    MCPService
    Initialize(ctx context.Context, options any) error
    Start() error
    Shutdown(ctx context.Context) error
}
```

The `MCPService` interface defines the core functionality:

```go
type MCPService interface {
    // ListPrompts returns a list of available prompts with pagination support
    ListPrompts(ctx context.Context, cursor string) (*PromptListResult, error)

    // GetPrompt retrieves a specific prompt by name with optional arguments
    GetPrompt(ctx context.Context, name string, args map[string]any) (*PromptResult, error)

    // ListTools returns a list of available tools with pagination support
    ListTools(ctx context.Context, cursor string) (*ToolListResult, error)

    // CallTool invokes a specific tool by name with arguments
    CallTool(ctx context.Context, name string, args map[string]any) (*CallToolResult, error)

    // ListResources returns a list of available resources with pagination support
    ListResources(ctx context.Context, cursor string) (*ResourceListResult, error)

    // ReadResource reads the content of a specific resource
    ReadResource(ctx context.Context, uri string) (*ResourceContent, error)

    // ListResourceTemplates returns a list of available resource templates
    ListResourceTemplates(ctx context.Context) ([]ResourceTemplate, error)

    // SubscribeToResource subscribes to changes on a specific resource
    SubscribeToResource(ctx context.Context, uri string) error
}
```

### Client Interface

Clients access services through the `types.Client` interface:

```go
type Client interface {
    // Connect establishes a connection to the server
    Connect(ctx context.Context) error

    // Close terminates the connection
    Close() error

    // Service returns the underlying MCPService interface
    Service() MCPService
}
```

## Contributing

Contributions are welcome! Please follow these steps:

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details

## Updated to MCP Specification 2025-03-26

This library has been updated to support the Model Context Protocol (MCP) 2025-03-26 specification. Major updates include:

### New Features

- **Complete JSON-RPC Support**: Implemented JSON-RPC 2.0 API endpoints compliant with the latest MCP specification
- **Enhanced Multimodal Content**: Support for text, image, audio, and embedded resource content transmission
- **Pagination Support**: All list APIs now support cursor-based pagination
- **Resource Templates**: Support for parameterized resource URI templates
- **Resource Subscriptions**: Support for client subscriptions to resource change notifications
- **Rich Server Capabilities**: More granular server capability declarations
- **Unified Response Handling**: Standardized response handling across different transport layers

### Backward Compatibility

- Preserved original REST API endpoints to ensure compatibility with older clients
- Notification system supports both new and old notification formats

### Example Usage

```go
// Create server
server, err := mcp.NewServer(service, &types.ServerOptions{
    Address: ":8080",
    Capabilities: &types.ServerCapabilities{
        Prompts: &types.PromptCapabilities{
            ListChanged: true,
        },
        Resources: &types.ResourceCapabilities{
            ListChanged: true,
            Subscribe: true,
            Templates: true,
        },
    },
})

// Create client
client, err := mcp.NewClient(&types.ClientOptions{
    ServerAddress: "localhost:8080",
    Type: "http",
    UseJSONRPC: true,
    SubscribeToNotifications: true,
})

// Get prompts list (with pagination support)
result, err := client.Service().ListPrompts(ctx, "")
// Get next page
nextPage, err := client.Service().ListPrompts(ctx, result.NextCursor)

// Subscribe to resource changes
err := client.Service().SubscribeToResource(ctx, "file:///example.txt")
```
