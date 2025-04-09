package types

// ServerOptions contains server configuration options
type ServerOptions struct {
	// Address is the address to listen on (e.g. ":8080")
	Address string `json:"address"`

	// Type is the server type ("http" or "websocket")
	Type string `json:"type,omitempty"`

	// Capabilities defines the server capabilities
	Capabilities *ServerCapabilities `json:"capabilities,omitempty"`

	// ServerName is the name of the server
	ServerName string `json:"serverName,omitempty"`

	// ServerVersion is the version of the server
	ServerVersion string `json:"serverVersion,omitempty"`
}

// ClientOptions contains client configuration options
type ClientOptions struct {
	// ServerAddress is the server address to connect to (e.g. "localhost:8080")
	ServerAddress string `json:"server_address"`

	// Type is the client type ("http", "websocket" or "sse")
	Type string `json:"type"`

	// UseJSONRPC indicates whether to use JSON-RPC protocol
	UseJSONRPC bool `json:"use_jsonrpc"`

	// SubscribeToNotifications indicates whether to subscribe to server notifications
	SubscribeToNotifications bool `json:"subscribe_to_notifications"`
}
