package transport

import (
	"context"

	"github.com/virgoC0der/go-mcp/internal/types"
	"github.com/virgoC0der/go-mcp/transport/http"
	"github.com/virgoC0der/go-mcp/transport/websocket"
)

// NewServer creates a new server instance based on the provided options
func NewServer(service types.MCPService, options *types.ServerOptions) (types.Server, error) {
	if options == nil {
		options = &types.ServerOptions{
			Address: ":8080",
		}
	}
	return http.NewHTTPServer(service, options.Address), nil
}

// NewClient creates a new client instance based on the provided options
func NewClient(options *types.ClientOptions) (types.Client, error) {
	if options == nil {
		options = &types.ClientOptions{
			ServerAddress: "localhost:8080",
			Type:          "websocket",
		}
	}

	switch options.Type {
	case "http":
		return http.NewHTTPClient(options.ServerAddress), nil
	//case "sse":
	//	return tclient.NewSSEClient(options.ServerAddress), nil
	case "websocket":
		client := websocket.NewWebSocketClient(options.ServerAddress)
		if err := client.Connect(context.Background()); err != nil {
			return nil, err
		}
		return client, nil
	default:
		// Default to WebSocket client
		client := websocket.NewWebSocketClient(options.ServerAddress)
		if err := client.Connect(context.Background()); err != nil {
			return nil, err
		}
		return client, nil
	}
}
