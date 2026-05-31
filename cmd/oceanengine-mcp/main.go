// Command oceanengine-mcp is a Model Context Protocol server for the Ocean
// Engine (巨量引擎) advertising platform. It speaks MCP over stdio and exposes
// tools for querying advertisers, campaigns, ads and performance reports — and,
// when explicitly enabled, for mutating campaign status and budget.
//
// Configuration is via environment variables:
//
//	OCEANENGINE_ACCESS_TOKEN   (required) OAuth access token
//	OCEANENGINE_BASE_URL       (optional) API host override
//	OCEANENGINE_ENABLE_WRITES  (optional) set to "1"/"true" to register write tools
package main

import (
	"context"
	"log"
	"os"
	"strconv"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/virgoC0der/go-mcp/internal/mcpserver"
	"github.com/virgoC0der/go-mcp/internal/oceanengine"
)

// version is overridable at build time via -ldflags "-X main.version=...".
var version = "dev"

func main() {
	token := os.Getenv("OCEANENGINE_ACCESS_TOKEN")
	if token == "" {
		log.Fatal("OCEANENGINE_ACCESS_TOKEN is required")
	}

	var clientOpts []oceanengine.Option
	if base := os.Getenv("OCEANENGINE_BASE_URL"); base != "" {
		clientOpts = append(clientOpts, oceanengine.WithBaseURL(base))
	}
	client := oceanengine.NewClient(token, clientOpts...)

	srv := mcpserver.New(client, mcpserver.Config{
		Name:         "oceanengine-mcp",
		Version:      version,
		EnableWrites: envBool("OCEANENGINE_ENABLE_WRITES"),
	})

	if err := srv.Run(context.Background(), &mcp.StdioTransport{}); err != nil {
		log.Fatalf("oceanengine-mcp: %v", err)
	}
}

func envBool(key string) bool {
	b, _ := strconv.ParseBool(os.Getenv(key))
	return b
}
