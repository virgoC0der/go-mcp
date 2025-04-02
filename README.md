# Go-MCP

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
        Type:         "sse", // or "http"
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
    prompts, err := service.ListPrompts(ctx)
    if err != nil {
        log.Fatal(err)
    }
    log.Printf("Available prompts: %v", prompts)
}
```

## Response Structure

All API responses follow a unified structure:

```go
type Response struct {
    Success bool        `json:"success"`
    Result  interface{} `json:"result,omitempty"`
    Error   *ErrorInfo  `json:"error,omitempty"`
}

type ErrorInfo struct {
    Code    string `json:"code"`
    Message string `json:"message"`
}
```

Success Response Example:
```json
{
    "success": true,
    "result": {
        "name": "example_prompt",
        "description": "An example prompt"
    }
}
```

Error Response Example:
```json
{
    "success": false,
    "error": {
        "code": "invalid_request",
        "message": "Invalid request parameters"
    }
}
```

## Examples

- [Echo Server](examples/echo/main.go) - A simple echo server example
- [Weather Service](examples/weather/main.go) - A weather service example using OpenWeatherMap API
- [Advanced Usage](examples/advanced/main.go) - Example showcasing advanced features

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
    ListPrompts(ctx context.Context) ([]Prompt, error)
    GetPrompt(ctx context.Context, name string, args map[string]any) (*GetPromptResult, error)
    ListTools(ctx context.Context) ([]Tool, error)
    CallTool(ctx context.Context, name string, args map[string]any) (*CallToolResult, error)
    ListResources(ctx context.Context) ([]Resource, error)
    ReadResource(ctx context.Context, name string) ([]byte, string, error)
}
```

### Client Interface

Clients access services through the `types.Client` interface:

```go
type Client interface {
    Connect(ctx context.Context) error
    Close() error
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
