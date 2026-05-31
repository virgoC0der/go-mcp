# oceanengine-mcp

[![CI](https://github.com/virgoC0der/go-mcp/actions/workflows/go.yml/badge.svg)](https://github.com/virgoC0der/go-mcp/actions/workflows/go.yml)

A [Model Context Protocol](https://modelcontextprotocol.io) (MCP) server for
**Ocean Engine (巨量引擎)** — ByteDance's domestic advertising platform (巨量广告 /
巨量千川 / 本地推 / 星图). It lets AI agents query advertiser accounts, campaigns,
ads and performance reports — and, when explicitly enabled, adjust campaign
status and budget — over the standard MCP interface.

> **Why this exists.** The "general-purpose Go MCP SDK" question is settled: the
> [official `modelcontextprotocol/go-sdk`](https://github.com/modelcontextprotocol/go-sdk)
> (maintained with Google) and [`mark3labs/mcp-go`](https://github.com/mark3labs/mcp-go)
> own that space. The open niche is **vertical** MCP servers. Google Ads and Meta
> Ads already have MCP servers (and TikTok Ads has several), but **Ocean Engine —
> the domestic Chinese platform — has none.** This project fills that gap, and is
> built *on top of* the official SDK rather than reimplementing the protocol.

## Status

Early but functional. The MCP layer is fully working (verified end-to-end via the
in-memory transport and a stdio smoke test); the Ocean Engine client implements
the documented Marketing API endpoints. Hitting the live API requires a valid
`access_token` from the [Ocean Engine open platform](https://open.oceanengine.com/).

## Install

```bash
go install github.com/virgoC0der/go-mcp/cmd/oceanengine-mcp@latest
```

Or build from source:

```bash
make build   # produces ./bin/oceanengine-mcp
```

## Configuration

The server is configured via environment variables:

| Variable | Required | Description |
|---|---|---|
| `OCEANENGINE_ACCESS_TOKEN` | yes | OAuth access token from the Ocean Engine open platform |
| `OCEANENGINE_BASE_URL` | no | API host override (defaults to `https://api.oceanengine.com`) |
| `OCEANENGINE_ENABLE_WRITES` | no | set to `1`/`true` to register the mutating tools (off by default) |

### Use with an MCP client

```json
{
  "mcpServers": {
    "oceanengine": {
      "command": "oceanengine-mcp",
      "env": { "OCEANENGINE_ACCESS_TOKEN": "your-token" }
    }
  }
}
```

## Tools

Read tools (always available):

| Tool | Ocean Engine endpoint | Purpose |
|---|---|---|
| `oceanengine_get_advertiser_info` | `GET /2/advertiser/info/` | account info by advertiser ID |
| `oceanengine_list_campaigns` | `GET /2/campaign/get/` | list campaigns (广告组), paginated |
| `oceanengine_list_ads` | `GET /2/ad/get/` | list ads (广告计划), paginated |
| `oceanengine_get_report` | `GET /2/report/ad/get/` | performance report by date range/dimensions |

Write tools (only when `OCEANENGINE_ENABLE_WRITES` is set — they mutate the live
account):

| Tool | Ocean Engine endpoint | Purpose |
|---|---|---|
| `oceanengine_update_campaign_status` | `POST /2/campaign/update/status/` | enable / disable / delete campaigns |
| `oceanengine_update_campaign_budget` | `POST /2/campaign/update/budget/` | set a campaign budget |

## Architecture

```
cmd/oceanengine-mcp     entrypoint: reads env, runs MCP over stdio
internal/mcpserver      registers tools on the official go-sdk; no protocol code
internal/oceanengine    thin Marketing API client (auth, envelope, endpoints)
```

The Ocean Engine client is deliberately thin and dependency-light. To broaden
endpoint coverage later, its methods can be backed by the much larger community
SDK [`bububa/oceanengine`](https://github.com/bububa/oceanengine) without touching
the MCP tool layer.

## Roadmap

- OAuth token refresh helper (the open platform issues short-lived tokens)
- 千川 (Qianchuan) e-commerce ad endpoints
- Broader report dimensions/metrics and async report export
- Optional `bububa/oceanengine` backend for full endpoint coverage

## Note on the repository name

This project currently lives in the `go-mcp` repository for historical reasons.
The intended home is a dedicated `oceanengine-mcp` repository — renaming the repo
(which preserves stars and history) updates the import path accordingly.

## Development

```bash
make test    # go test ./...
make vet     # go vet ./...
make build   # build the binary
```

## License

MIT — see [LICENSE](./LICENSE).
