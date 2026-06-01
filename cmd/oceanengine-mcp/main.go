// Command oceanengine-mcp is a Model Context Protocol server for the Ocean
// Engine (巨量引擎) advertising platform. It speaks MCP over stdio and exposes
// tools for querying advertisers, campaigns, ads and performance reports — and,
// when explicitly enabled, for mutating campaign status and budget.
//
// Configuration is via environment variables.
//
// Authentication — either supply a static token:
//
//	OCEANENGINE_ACCESS_TOKEN   OAuth access token
//
// or enable automatic refresh by supplying app credentials and a refresh token:
//
//	OCEANENGINE_APP_ID                 developer app ID (numeric)
//	OCEANENGINE_APP_SECRET             developer app secret
//	OCEANENGINE_REFRESH_TOKEN          OAuth refresh token
//	OCEANENGINE_ACCESS_TOKEN           (optional) current access token
//	OCEANENGINE_ACCESS_TOKEN_EXPIRES_IN (optional) remaining lifetime, seconds
//
// Other options:
//
//	OCEANENGINE_BASE_URL       (optional) API host override
//	OCEANENGINE_ENABLE_WRITES  (optional) set to "1"/"true" to register write tools
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/virgoC0der/go-mcp/internal/mcpserver"
	"github.com/virgoC0der/go-mcp/internal/oceanengine"
)

// version is overridable at build time via -ldflags "-X main.version=...".
var version = "dev"

func main() {
	baseURL := os.Getenv("OCEANENGINE_BASE_URL")

	var clientOpts []oceanengine.Option
	if baseURL != "" {
		clientOpts = append(clientOpts, oceanengine.WithBaseURL(baseURL))
	}

	tokens, err := tokenProvider(baseURL)
	if err != nil {
		log.Fatal(err)
	}
	clientOpts = append(clientOpts, oceanengine.WithTokenProvider(tokens))

	client := oceanengine.NewClient("", clientOpts...)

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

// tokenProvider builds a token source from the environment, preferring the
// self-refreshing OAuth source when app credentials and a refresh token are
// present, and otherwise falling back to a static access token.
func tokenProvider(baseURL string) (oceanengine.TokenProvider, error) {
	accessToken := os.Getenv("OCEANENGINE_ACCESS_TOKEN")
	refreshToken := os.Getenv("OCEANENGINE_REFRESH_TOKEN")
	appIDStr := os.Getenv("OCEANENGINE_APP_ID")
	secret := os.Getenv("OCEANENGINE_APP_SECRET")

	if refreshToken != "" && appIDStr != "" && secret != "" {
		appID, err := strconv.ParseUint(appIDStr, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("OCEANENGINE_APP_ID must be numeric: %w", err)
		}
		var expiresIn time.Duration
		if s := os.Getenv("OCEANENGINE_ACCESS_TOKEN_EXPIRES_IN"); s != "" {
			secs, err := strconv.ParseInt(s, 10, 64)
			if err != nil {
				return nil, fmt.Errorf("OCEANENGINE_ACCESS_TOKEN_EXPIRES_IN must be numeric seconds: %w", err)
			}
			expiresIn = time.Duration(secs) * time.Second
		}
		var opts []oceanengine.RefreshOption
		if baseURL != "" {
			opts = append(opts, oceanengine.WithRefreshBaseURL(baseURL))
		}
		return oceanengine.NewRefreshingTokenSource(appID, secret, accessToken, refreshToken, expiresIn, opts...), nil
	}

	if accessToken == "" {
		return nil, fmt.Errorf("authentication required: set OCEANENGINE_ACCESS_TOKEN, or OCEANENGINE_APP_ID + OCEANENGINE_APP_SECRET + OCEANENGINE_REFRESH_TOKEN for auto-refresh")
	}
	return oceanengine.StaticToken(accessToken), nil
}
