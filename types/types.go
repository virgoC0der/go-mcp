package types

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
	Name        string           `json:"name"`
	Description string           `json:"description"`
	Arguments   []PromptArgument `json:"arguments,omitempty"`
}

// PromptArgument represents an argument for a prompt
type PromptArgument struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Required    bool   `json:"required"`
}

// GetPromptResult represents the result of getting a prompt
type GetPromptResult struct {
	Description string    `json:"description"`
	Messages    []Message `json:"messages"`
	Message     string    `json:"message,omitempty"` // Single text message for simple responses
}

// Tool represents a tool that can be called by the model
type Tool struct {
	Name        string           `json:"name"`
	Description string           `json:"description"`
	Arguments   []PromptArgument `json:"arguments,omitempty"`
	Schema      interface{}      `json:"schema,omitempty"`
}

// Resource represents a resource that can be accessed
type Resource struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	MimeType    string `json:"mimeType"`
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

// CallToolResult represents the result of a CallTool operation
type CallToolResult struct {
	Content interface{} `json:"content"`
}
