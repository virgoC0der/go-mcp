# macOS App Launcher Example

This example demonstrates how to create a macOS application launcher using the Go-MCP framework.

## Features

- Launch macOS applications using the MCP protocol
- List common macOS applications
- Simple client for testing

## Prerequisites

Before running this example, you need:

1. macOS operating system (this example is designed specifically for macOS)
2. Go 1.20 or later

## Running the Example

Start the server:

```bash
go run main.go
```

This will start a stdio server that can be used with JSON-RPC requests. The HTTP server is commented out in the code but can be enabled by uncommenting the relevant section in main.go.

## Using the Client

Run the client to open a specific application:

```bash
# Open Calculator (default)
go run client/main.go

# Open a specific application
go run client/main.go "Safari"
go run client/main.go "Terminal"
go run client/main.go "System Settings"
```

## API Usage

### Prompts

The app launcher service provides an `openApp` prompt:

```json
{
  "name": "openApp",
  "args": {
    "appName": "Calculator"
  }
}
```

Response:
```json
{
  "description": "Application launcher result",
  "messages": [
    {
      "role": "assistant",
      "content": {
        "type": "text",
        "text": "Successfully opened application: Calculator"
      }
    }
  ]
}
```

### Tools

The service provides an `openApp` tool:

```json
{
  "name": "openApp",
  "args": {
    "appName": "Safari"
  }
}
```

Response:
```json
{
  "content": [
    {
      "type": "text",
      "text": "Successfully opened application: Safari"
    }
  ],
  "isError": false
}
```

### Resources

The service provides an `apps` resource that lists common macOS applications:

```json
GET /resources/apps
```

Response:
```json
{
  "uri": "apps",
  "mimeType": "application/json",
  "text": "[\"Safari\",\"Mail\",\"Calendar\",\"Notes\",\"Maps\",\"Photos\",\"Messages\",\"FaceTime\",\"Music\",\"App Store\",\"System Settings\",\"Terminal\",\"Calculator\",\"TextEdit\",\"Preview\",\"GoLand\",\"Edge\",\"Cursor\",\"Warp\",\"iTerm 2\"]"
}
```

### JSON-RPC Support

The service also supports JSON-RPC 2.0 calls:

```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "method": "tools/call",
  "params": {
    "name": "openApp",
    "arguments": {
      "appName": "Calculator"
    }
  }
}
```

Response:
```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "result": {
    "content": [
      {
        "type": "text",
        "text": "Successfully opened application: Calculator"
      }
    ],
    "isError": false
  }
}
```

## Error Handling

The service handles various error cases:
- Invalid application names
- Applications that don't exist
- Running on non-macOS platforms

## Implementation Details

The example demonstrates:
- Implementing the MCP Server interface
- Using the `exec.Command` to execute the macOS `open` command
- Handling prompts, tools, and resources
- Error handling and validation
- Graceful shutdown
- JSON-RPC support
- Stdio server implementation
- Server capabilities configuration
