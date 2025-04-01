package mcp

import (
	"github.com/virgoC0der/go-mcp/internal/transport"
	"github.com/virgoC0der/go-mcp/internal/types"
)

// ClientOption represents an option for configuring the client
type ClientOption func(*types.ClientOptions)

// WithServerAddress sets the server address for the client
func WithServerAddress(addr string) ClientOption {
	return func(o *types.ClientOptions) {
		o.ServerAddress = addr
	}
}

// NewClient creates a new MCP client instance
func NewClient(options *types.ClientOptions) (types.Client, error) {
	return transport.NewClient(options)
}
