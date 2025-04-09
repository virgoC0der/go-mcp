package types

import "context"

// MCPService 定义了 MCP 服务的核心功能
type MCPService interface {
	// ListPrompts returns a list of available prompts
	ListPrompts(ctx context.Context, cursor string) (*PromptListResult, error)

	// GetPrompt retrieves a specific prompt by name with optional arguments
	GetPrompt(ctx context.Context, name string, args map[string]any) (*PromptResult, error)

	// ListTools returns a list of available tools
	ListTools(ctx context.Context, cursor string) (*ToolListResult, error)

	// CallTool invokes a specific tool by name with arguments
	CallTool(ctx context.Context, name string, args map[string]any) (*CallToolResult, error)

	// ListResources returns a list of available resources
	ListResources(ctx context.Context, cursor string) (*ResourceListResult, error)

	// ReadResource reads the content of a specific resource
	ReadResource(ctx context.Context, uri string) (*ResourceContent, error)

	// ListResourceTemplates returns a list of available resource templates
	ListResourceTemplates(ctx context.Context) ([]ResourceTemplate, error)

	// SubscribeToResource subscribes to changes on a specific resource
	SubscribeToResource(ctx context.Context, uri string) error
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
	Type     string       `json:"type"`               // "text", "image", "audio", or "resource"
	Text     string       `json:"text,omitempty"`     // For text content
	Data     string       `json:"data,omitempty"`     // For base64-encoded image/audio data
	MimeType string       `json:"mimeType,omitempty"` // MIME type for image/audio content
	Resource *ResourceRef `json:"resource,omitempty"` // For embedded resources
}

// ResourceRef represents a reference to a resource
type ResourceRef struct {
	URI      string `json:"uri"`
	MimeType string `json:"mimeType"`
	Text     string `json:"text,omitempty"`
	Blob     string `json:"blob,omitempty"` // base64-encoded blob data
}

// Message represents a message in the MCP protocol
type Message struct {
	Role    string  `json:"role"` // "user" or "assistant"
	Content Content `json:"content"`
}

// Prompt represents a prompt template
type Prompt struct {
	Name        string           `json:"name"`
	Description string           `json:"description,omitempty"`
	Arguments   []PromptArgument `json:"arguments,omitempty"`
}

// PromptArgument represents an argument for a prompt
type PromptArgument struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Required    bool   `json:"required"`
}

// PromptListResult represents the result of listing prompts with pagination
type PromptListResult struct {
	Prompts    []Prompt `json:"prompts"`
	NextCursor string   `json:"nextCursor,omitempty"`
}

// PromptResult represents the result of getting a prompt
type PromptResult struct {
	Description string    `json:"description,omitempty"`
	Messages    []Message `json:"messages"`
}

// Tool represents a tool that can be called
type Tool struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	InputSchema map[string]interface{} `json:"inputSchema,omitempty"`
	Annotations map[string]interface{} `json:"annotations,omitempty"`
}

// ToolListResult represents the result of listing tools with pagination
type ToolListResult struct {
	Tools      []Tool `json:"tools"`
	NextCursor string `json:"nextCursor,omitempty"`
}

// Resource represents a resource that can be accessed
type Resource struct {
	URI         string `json:"uri"`
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
	MimeType    string `json:"mimeType,omitempty"`
	Size        int64  `json:"size,omitempty"`
}

// ResourceTemplate represents a template for parameterized resources
type ResourceTemplate struct {
	URITemplate string `json:"uriTemplate"`
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
	MimeType    string `json:"mimeType,omitempty"`
}

// ResourceListResult represents the result of listing resources with pagination
type ResourceListResult struct {
	Resources  []Resource `json:"resources"`
	NextCursor string     `json:"nextCursor,omitempty"`
}

// ResourceContent represents the content of a resource
type ResourceContent struct {
	URI      string `json:"uri"`
	MimeType string `json:"mimeType"`
	Text     string `json:"text,omitempty"` // For text content
	Blob     string `json:"blob,omitempty"` // For base64-encoded binary content
}

// ServerCapabilities represents the capabilities of an MCP server
type ServerCapabilities struct {
	Prompts    *PromptCapabilities   `json:"prompts,omitempty"`
	Resources  *ResourceCapabilities `json:"resources,omitempty"`
	Tools      *ToolCapabilities     `json:"tools,omitempty"`
	Logging    bool                  `json:"logging,omitempty"`
	Completion bool                  `json:"completion,omitempty"`
}

// PromptCapabilities represents the capabilities of the prompts feature
type PromptCapabilities struct {
	ListChanged bool `json:"listChanged"`
}

// ResourceCapabilities represents the capabilities of the resources feature
type ResourceCapabilities struct {
	ListChanged bool `json:"listChanged"`
	Subscribe   bool `json:"subscribe,omitempty"`
	Templates   bool `json:"templates,omitempty"`
}

// ToolCapabilities represents the capabilities of the tools feature
type ToolCapabilities struct {
	ListChanged bool `json:"listChanged"`
	Streaming   bool `json:"streaming,omitempty"`
}

// InitializationOptions represents the options for initializing an MCP server
type InitializationOptions struct {
	ServerName    string             `json:"serverName"`
	ServerVersion string             `json:"serverVersion"`
	Capabilities  ServerCapabilities `json:"capabilities"`
}

// ToolContent represents a content item in a tool result
type ToolContent struct {
	Type     string       `json:"type"`               // "text", "image", "audio", or "resource"
	Text     string       `json:"text,omitempty"`     // For text content
	Data     string       `json:"data,omitempty"`     // For base64-encoded image/audio data
	MimeType string       `json:"mimeType,omitempty"` // MIME type for image/audio content
	Resource *ResourceRef `json:"resource,omitempty"` // For embedded resources
}

// CallToolResult represents the result of calling a tool
type CallToolResult struct {
	Content []ToolContent `json:"content"`
	IsError bool          `json:"isError,omitempty"`
}
