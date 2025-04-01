package mcp

import (
	"github.com/virgoC0der/go-mcp/internal/transport"
	"github.com/virgoC0der/go-mcp/internal/types"
)

// ServerOption represents an option for configuring the server
type ServerOption func(*types.ServerOptions)

// WithAddress sets the server address
func WithAddress(addr string) ServerOption {
	return func(o *types.ServerOptions) {
		o.Address = addr
	}
}

// NewServer creates a new MCP server instance
func NewServer(service types.MCPService, options *types.ServerOptions) (types.Server, error) {
	return transport.NewServer(service, options)
}
