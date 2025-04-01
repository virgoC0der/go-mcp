package types

import "context"

// MCPService 定义了 MCP 服务的核心功能
type MCPService interface {
	// ListPrompts returns a list of available prompts
	ListPrompts(ctx context.Context) ([]Prompt, error)

	// GetPrompt retrieves a specific prompt by name with optional arguments
	GetPrompt(ctx context.Context, name string, args map[string]any) (*GetPromptResult, error)

	// ListTools returns a list of available tools
	ListTools(ctx context.Context) ([]Tool, error)

	// CallTool invokes a specific tool by name with arguments
	CallTool(ctx context.Context, name string, args map[string]any) (*CallToolResult, error)

	// ListResources returns a list of available resources
	ListResources(ctx context.Context) ([]Resource, error)

	// ReadResource reads the content of a specific resource
	ReadResource(ctx context.Context, name string) ([]byte, string, error)
}

// Server defines the interface for MCP server implementations
type Server interface {
	MCPService
	// Initialize initializes the server with given options
	Initialize(ctx context.Context, options any) error
	// Start starts the server
	Start() error
	// Shutdown gracefully shuts down the server
	Shutdown(ctx context.Context) error
}

// Client defines the interface for MCP client implementations
type Client interface {
	// Connect establishes a connection to the server
	Connect(ctx context.Context) error

	// Close terminates the connection
	Close() error

	// Service returns the underlying MCPService interface
	Service() MCPService
}

// Content represents the content of a message
type Content struct {
	Type string `json:"type"`
	Text string `json:"text,omitempty"`
}

// Message represents a message in the MCP protocol
type Message struct {
	Role    string  `json:"role"`
	Content Content `json:"content"`
}

// Prompt represents a prompt template
type Prompt struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Template    string                 `json:"template"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// PromptArgument represents an argument for a prompt
type PromptArgument struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Required    bool   `json:"required"`
}

// GetPromptResult represents the result of getting a prompt
type GetPromptResult struct {
	Content string `json:"content"`
}

// Tool represents a tool that can be called
type Tool struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Parameters  map[string]interface{} `json:"parameters,omitempty"`
}

// Resource represents a resource that can be accessed
type Resource struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	Description string `json:"description,omitempty"`
}

// ServerCapabilities represents the capabilities of an MCP server
type ServerCapabilities struct {
	Prompts    bool `json:"prompts"`
	Resources  bool `json:"resources"`
	Tools      bool `json:"tools"`
	Logging    bool `json:"logging"`
	Completion bool `json:"completion"`
}

// InitializationOptions represents the options for initializing an MCP server
type InitializationOptions struct {
	ServerName    string             `json:"serverName"`
	ServerVersion string             `json:"serverVersion"`
	Capabilities  ServerCapabilities `json:"capabilities"`
}

// CallToolResult represents the result of calling a tool
type CallToolResult struct {
	Output interface{} `json:"output"`
	Error  string      `json:"error,omitempty"`
}
