package mcp

import (
	"github.com/virgoC0der/go-mcp/internal/types"
)

// Server defines the interface for MCP server implementations
type Server = types.Server

// Client defines the interface for MCP client implementations
type Client = types.Client

// Prompt represents a prompt template
type Prompt = types.Prompt

// Tool represents a tool that can be called
type Tool = types.Tool

// Resource represents a resource that can be accessed
type Resource = types.Resource

// PromptResult represents the result of getting a prompt
type PromptResult = types.PromptResult

// CallToolResult represents the result of calling a tool
type CallToolResult = types.CallToolResult

// ServerOptions contains server configuration options
type ServerOptions = types.ServerOptions

// ClientOptions contains client configuration options
type ClientOptions = types.ClientOptions
