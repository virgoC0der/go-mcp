package transport

import (
	"context"

	tclient "github.com/virgoC0der/go-mcp/internal/transport/client"
	tserver "github.com/virgoC0der/go-mcp/internal/transport/server"
	"github.com/virgoC0der/go-mcp/internal/types"
)

// NewServer creates a new server instance based on the provided options
func NewServer(service types.MCPService, options *types.ServerOptions) (types.Server, error) {
	if options == nil {
		options = &types.ServerOptions{
			Address: ":8080",
		}
	}
	return tserver.NewHTTPServer(service, options.Address), nil
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
		return tclient.NewHTTPClient(options.ServerAddress), nil
	case "websocket":
		client := tclient.NewWebSocketClient(options.ServerAddress)
		if err := client.Connect(context.Background()); err != nil {
			return nil, err
		}
		return client, nil
	default:
		// Default to WebSocket client
		client := tclient.NewWebSocketClient(options.ServerAddress)
		if err := client.Connect(context.Background()); err != nil {
			return nil, err
		}
		return client, nil
	}
}
