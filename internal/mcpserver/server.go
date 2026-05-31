// Package mcpserver wires the Ocean Engine (巨量引擎) client into a Model
// Context Protocol server, built on the official modelcontextprotocol/go-sdk.
// It does not implement the MCP protocol itself; it only registers tools.
package mcpserver

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/virgoC0der/go-mcp/internal/oceanengine"
)

// Config controls which tools are registered.
type Config struct {
	// Name and Version are reported to MCP clients.
	Name    string
	Version string
	// EnableWrites registers the mutating tools (status/budget changes). When
	// false, the server is read-only — the safe default.
	EnableWrites bool
}

// New builds an MCP server exposing Ocean Engine tools backed by client.
func New(client *oceanengine.Client, cfg Config) *mcp.Server {
	if cfg.Name == "" {
		cfg.Name = "oceanengine-mcp"
	}
	if cfg.Version == "" {
		cfg.Version = "dev"
	}

	srv := mcp.NewServer(&mcp.Implementation{Name: cfg.Name, Version: cfg.Version}, nil)
	registerReadTools(srv, client)
	if cfg.EnableWrites {
		registerWriteTools(srv, client)
	}
	return srv
}

// ---------------------------------------------------------------------------
// Read tools
// ---------------------------------------------------------------------------

type advertiserInfoInput struct {
	AdvertiserIDs []int64  `json:"advertiser_ids" jsonschema:"Ocean Engine advertiser (account) IDs to look up"`
	Fields        []string `json:"fields,omitempty" jsonschema:"optional subset of fields to return; defaults to id,name,role,status,company"`
}

type advertiserInfoOutput struct {
	Advertisers []oceanengine.Advertiser `json:"advertisers"`
}

type listCampaignsInput struct {
	AdvertiserID int64 `json:"advertiser_id" jsonschema:"Ocean Engine advertiser (account) ID"`
	Page         int   `json:"page,omitempty" jsonschema:"1-based page number; defaults to 1"`
	PageSize     int   `json:"page_size,omitempty" jsonschema:"page size 1-100; defaults to 10"`
}

type listAdsInput struct {
	AdvertiserID int64 `json:"advertiser_id" jsonschema:"Ocean Engine advertiser (account) ID"`
	Page         int   `json:"page,omitempty" jsonschema:"1-based page number; defaults to 1"`
	PageSize     int   `json:"page_size,omitempty" jsonschema:"page size 1-100; defaults to 10"`
}

type getReportInput struct {
	AdvertiserID int64    `json:"advertiser_id" jsonschema:"Ocean Engine advertiser (account) ID"`
	StartDate    string   `json:"start_date" jsonschema:"report start date, YYYY-MM-DD"`
	EndDate      string   `json:"end_date" jsonschema:"report end date, YYYY-MM-DD"`
	GroupBy      []string `json:"group_by,omitempty" jsonschema:"dimensions to group by, e.g. [\"STAT_GROUP_BY_FIELD_ID\",\"STAT_GROUP_BY_FIELD_STAT_TIME\"]"`
	Fields       []string `json:"fields,omitempty" jsonschema:"metrics to return, e.g. [\"cost\",\"show\",\"click\",\"convert\"]"`
	Page         int      `json:"page,omitempty" jsonschema:"1-based page number; defaults to 1"`
	PageSize     int      `json:"page_size,omitempty" jsonschema:"page size 1-100; defaults to 10"`
}

func registerReadTools(srv *mcp.Server, client *oceanengine.Client) {
	mcp.AddTool(srv, &mcp.Tool{
		Name:        "oceanengine_get_advertiser_info",
		Description: "Get Ocean Engine (巨量引擎) advertiser account information by advertiser ID.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, in advertiserInfoInput) (*mcp.CallToolResult, advertiserInfoOutput, error) {
		if len(in.AdvertiserIDs) == 0 {
			return nil, advertiserInfoOutput{}, fmt.Errorf("advertiser_ids must not be empty")
		}
		ads, err := client.GetAdvertiserInfo(ctx, in.AdvertiserIDs, in.Fields)
		if err != nil {
			return nil, advertiserInfoOutput{}, err
		}
		return nil, advertiserInfoOutput{Advertisers: ads}, nil
	})

	mcp.AddTool(srv, &mcp.Tool{
		Name:        "oceanengine_list_campaigns",
		Description: "List Ocean Engine (巨量引擎) campaigns (广告组) for an advertiser, with pagination.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, in listCampaignsInput) (*mcp.CallToolResult, *oceanengine.CampaignList, error) {
		if in.AdvertiserID == 0 {
			return nil, nil, fmt.Errorf("advertiser_id is required")
		}
		res, err := client.ListCampaigns(ctx, in.AdvertiserID, in.Page, in.PageSize)
		if err != nil {
			return nil, nil, err
		}
		return nil, res, nil
	})

	mcp.AddTool(srv, &mcp.Tool{
		Name:        "oceanengine_list_ads",
		Description: "List Ocean Engine (巨量引擎) ads (广告计划) for an advertiser, with pagination.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, in listAdsInput) (*mcp.CallToolResult, *oceanengine.AdList, error) {
		if in.AdvertiserID == 0 {
			return nil, nil, fmt.Errorf("advertiser_id is required")
		}
		res, err := client.ListAds(ctx, in.AdvertiserID, in.Page, in.PageSize)
		if err != nil {
			return nil, nil, err
		}
		return nil, res, nil
	})

	mcp.AddTool(srv, &mcp.Tool{
		Name:        "oceanengine_get_report",
		Description: "Get an Ocean Engine (巨量引擎) ad performance report for a date range, grouped by the given dimensions.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, in getReportInput) (*mcp.CallToolResult, *oceanengine.ReportResult, error) {
		if in.AdvertiserID == 0 {
			return nil, nil, fmt.Errorf("advertiser_id is required")
		}
		if in.StartDate == "" || in.EndDate == "" {
			return nil, nil, fmt.Errorf("start_date and end_date are required")
		}
		res, err := client.GetReport(ctx, oceanengine.ReportRequest{
			AdvertiserID: in.AdvertiserID,
			StartDate:    in.StartDate,
			EndDate:      in.EndDate,
			GroupBy:      in.GroupBy,
			Fields:       in.Fields,
			Page:         in.Page,
			PageSize:     in.PageSize,
		})
		if err != nil {
			return nil, nil, err
		}
		return nil, res, nil
	})
}

// ---------------------------------------------------------------------------
// Write tools (only registered when EnableWrites is true)
// ---------------------------------------------------------------------------

type updateStatusInput struct {
	AdvertiserID int64   `json:"advertiser_id" jsonschema:"Ocean Engine advertiser (account) ID"`
	CampaignIDs  []int64 `json:"campaign_ids" jsonschema:"campaign IDs to update"`
	OptStatus    string  `json:"opt_status" jsonschema:"one of: enable, disable, delete"`
}

type updateBudgetInput struct {
	AdvertiserID int64   `json:"advertiser_id" jsonschema:"Ocean Engine advertiser (account) ID"`
	CampaignID   int64   `json:"campaign_id" jsonschema:"campaign ID to update"`
	Budget       float64 `json:"budget" jsonschema:"new budget amount"`
	BudgetMode   string  `json:"budget_mode,omitempty" jsonschema:"budget mode; defaults to BUDGET_MODE_DAY"`
}

type okOutput struct {
	OK bool `json:"ok"`
}

func registerWriteTools(srv *mcp.Server, client *oceanengine.Client) {
	mcp.AddTool(srv, &mcp.Tool{
		Name:        "oceanengine_update_campaign_status",
		Description: "WRITE: enable, disable or delete Ocean Engine (巨量引擎) campaigns. This mutates the live account.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, in updateStatusInput) (*mcp.CallToolResult, okOutput, error) {
		switch in.OptStatus {
		case "enable", "disable", "delete":
		default:
			return nil, okOutput{}, fmt.Errorf("opt_status must be one of enable, disable, delete")
		}
		if in.AdvertiserID == 0 || len(in.CampaignIDs) == 0 {
			return nil, okOutput{}, fmt.Errorf("advertiser_id and campaign_ids are required")
		}
		if err := client.UpdateCampaignStatus(ctx, in.AdvertiserID, in.CampaignIDs, in.OptStatus); err != nil {
			return nil, okOutput{}, err
		}
		return nil, okOutput{OK: true}, nil
	})

	mcp.AddTool(srv, &mcp.Tool{
		Name:        "oceanengine_update_campaign_budget",
		Description: "WRITE: set a new budget for an Ocean Engine (巨量引擎) campaign. This mutates the live account.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, in updateBudgetInput) (*mcp.CallToolResult, okOutput, error) {
		if in.AdvertiserID == 0 || in.CampaignID == 0 {
			return nil, okOutput{}, fmt.Errorf("advertiser_id and campaign_id are required")
		}
		mode := in.BudgetMode
		if mode == "" {
			mode = "BUDGET_MODE_DAY"
		}
		if err := client.UpdateCampaignBudget(ctx, in.AdvertiserID, in.CampaignID, in.Budget, mode); err != nil {
			return nil, okOutput{}, err
		}
		return nil, okOutput{OK: true}, nil
	})
}
