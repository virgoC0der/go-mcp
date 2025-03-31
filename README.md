# Go MCP SDK

Go implementation of the Model Context Protocol (MCP).

## Overview

Model Context Protocol (MCP) is a protocol for building AI applications that defines three core primitives:

- Prompts: User-controlled interactive templates
- Resources: Application-controlled contextual data
- Tools: Model-controlled executable functions

This SDK provides a Go implementation of the MCP protocol, including both server and client interfaces.

## Features

- Complete MCP protocol implementation
- Type-safe API with native Go struct support and automatic JSON Schema generation
- Multiple transport options:
  - HTTP: Stateless REST API
  - WebSocket: For bidirectional communication
  - Stdio: For integration with CLI tools and processes
- Pagination support for listing large sets of prompts, tools, and resources
- Change notifications for real-time updates
- Easy-to-use server and client interfaces
- Built-in base implementations and examples

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
    "fmt"
    "log"

    "github.com/virgoC0der/go-mcp/server"
    "github.com/virgoC0der/go-mcp/types"
)

// Define typed input structure
type GreetInput struct {
    Name    string `json:"name" jsonschema:"required,description=Name to greet"`
    Formal  bool   `json:"formal" jsonschema:"description=Whether to use formal greeting"`
}

func main() {
    // Create server
    srv := server.NewBaseServer("my-server", "1.0.0")

    // Register typed prompt
    err := srv.RegisterPromptTyped("greet", "Greet a person", 
        func(input GreetInput) (*types.GetPromptResult, error) {
            greeting := "Hi, " + input.Name
            if input.Formal {
                greeting = "Good day, " + input.Name
            }
            return &types.GetPromptResult{
                Description: "A greeting message",
                Message:     greeting,
            }, nil
        })
    if err != nil {
        log.Fatal(err)
    }

    // Initialize server
    err = srv.Initialize(context.Background(), types.InitializationOptions{
        ServerName:    "my-server",
        ServerVersion: "1.0.0",
        Capabilities: types.ServerCapabilities{
            Prompts: true,
        },
    })
    if err != nil {
        log.Fatal(err)
    }

    // Start server...
}
```

### Using the Client

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/virgoC0der/go-mcp/client"
)

func main() {
    // Create client (HTTP, WebSocket, or Stdio)
    c := client.NewHTTPClient("http://localhost:8080")
    // Alternatively: c, _ := client.NewWebSocketClient("ws://localhost:8081/ws")
    
    // Initialize connection
    err := c.Initialize(context.Background())
    if err != nil {
        log.Fatal(err)
    }

    // List available prompts
    prompts, err := c.ListPrompts(context.Background())
    if err != nil {
        log.Fatal(err)
    }

    for _, p := range prompts {
        fmt.Printf("Found prompt: %s (%s)\n", p.Name, p.Description)
    }
    
    // Call a prompt
    args := map[string]interface{}{
        "name": "World",
        "formal": true,
    }
    
    result, err := c.GetPrompt(context.Background(), "greet", args)
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Println(result.Message)
}
```

## Examples

Check out the `examples` directory for more examples:

- `examples/echo`: Simple echo server example
- `examples/advanced`: Advanced example with typed handlers and multiple transports
- More examples coming soon...

## Documentation

Full documentation coming soon.

## Contributing

Pull requests are welcome!

## License

MIT
