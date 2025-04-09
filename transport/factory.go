package transport

import (
	"github.com/virgoC0der/go-mcp/internal/types"
	"github.com/virgoC0der/go-mcp/transport/http"
)

// NewServer creates a new server instance based on the provided options
func NewServer(service types.MCPService, options *types.ServerOptions) (types.Server, error) {
	if options == nil {
		options = &types.ServerOptions{
			Address: ":8080",
		}
	}
	// 默认使用 HTTP 服务器
	return http.NewHTTPServer(service, options), nil
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
	default:
		return http.NewHTTPClient(options.ServerAddress), nil
	}
}
