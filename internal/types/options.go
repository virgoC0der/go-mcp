package types

// ServerOptions contains server configuration options
type ServerOptions struct {
	// Address is the address to listen on (e.g. ":8080")
	Address string `json:"address"`
}

// ClientOptions contains client configuration options
type ClientOptions struct {
	// ServerAddress is the server address to connect to (e.g. "localhost:8080")
	ServerAddress string `json:"server_address"`
	// Type is the client type ("http" or "sse")
	Type string `json:"type"`
}
