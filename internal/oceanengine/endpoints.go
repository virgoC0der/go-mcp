package oceanengine

import (
	"context"
	"net/url"
	"strconv"
)

// PageInfo is the pagination block returned by list endpoints.
type PageInfo struct {
	Page        int `json:"page"`
	PageSize    int `json:"page_size"`
	TotalNumber int `json:"total_number"`
	TotalPage   int `json:"total_page"`
}

// ---------------------------------------------------------------------------
// Advertiser
// ---------------------------------------------------------------------------

// Advertiser is a subset of the fields returned by /2/advertiser/info/.
type Advertiser struct {
	ID      int64  `json:"id"`
	Name    string `json:"name"`
	Role    string `json:"role"`
	Status  string `json:"status"`
	Company string `json:"company"`
}

// GetAdvertiserInfo returns account information for the given advertiser IDs.
// If fields is empty a sensible default set is requested.
//
// GET /open_api/2/advertiser/info/
func (c *Client) GetAdvertiserInfo(ctx context.Context, advertiserIDs []int64, fields []string) ([]Advertiser, error) {
	if len(fields) == 0 {
		fields = []string{"id", "name", "role", "status", "company"}
	}
	q := url.Values{}
	q.Set("advertiser_ids", jsonParam(advertiserIDs))
	q.Set("fields", jsonParam(fields))

	var out []Advertiser
	if err := c.get(ctx, "/open_api/2/advertiser/info/", q, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// ---------------------------------------------------------------------------
// Campaigns (广告组)
// ---------------------------------------------------------------------------

// Campaign is a subset of the fields returned by /2/campaign/get/.
type Campaign struct {
	ID           int64   `json:"id"`
	Name         string  `json:"name"`
	AdvertiserID int64   `json:"advertiser_id"`
	Budget       float64 `json:"budget"`
	BudgetMode   string  `json:"budget_mode"`
	LandingType  string  `json:"landing_type"`
	Status       string  `json:"status"`
	OptStatus    string  `json:"opt_status"`
}

// CampaignList is the data payload of /2/campaign/get/.
type CampaignList struct {
	List     []Campaign `json:"list"`
	PageInfo PageInfo   `json:"page_info"`
}

// ListCampaigns returns campaigns for an advertiser, paginated.
//
// GET /open_api/2/campaign/get/
func (c *Client) ListCampaigns(ctx context.Context, advertiserID int64, page, pageSize int) (*CampaignList, error) {
	q := url.Values{}
	q.Set("advertiser_id", strconv.FormatInt(advertiserID, 10))
	q.Set("page", strconv.Itoa(normPage(page)))
	q.Set("page_size", strconv.Itoa(normPageSize(pageSize)))

	var out CampaignList
	if err := c.get(ctx, "/open_api/2/campaign/get/", q, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// ---------------------------------------------------------------------------
// Ads (广告计划)
// ---------------------------------------------------------------------------

// Ad is a subset of the fields returned by /2/ad/get/.
type Ad struct {
	ID           int64   `json:"id"`
	Name         string  `json:"name"`
	CampaignID   int64   `json:"campaign_id"`
	AdvertiserID int64   `json:"advertiser_id"`
	Budget       float64 `json:"budget"`
	BudgetMode   string  `json:"budget_mode"`
	Status       string  `json:"status"`
	OptStatus    string  `json:"opt_status"`
}

// AdList is the data payload of /2/ad/get/.
type AdList struct {
	List     []Ad     `json:"list"`
	PageInfo PageInfo `json:"page_info"`
}

// ListAds returns ads for an advertiser, paginated.
//
// GET /open_api/2/ad/get/
func (c *Client) ListAds(ctx context.Context, advertiserID int64, page, pageSize int) (*AdList, error) {
	q := url.Values{}
	q.Set("advertiser_id", strconv.FormatInt(advertiserID, 10))
	q.Set("page", strconv.Itoa(normPage(page)))
	q.Set("page_size", strconv.Itoa(normPageSize(pageSize)))

	var out AdList
	if err := c.get(ctx, "/open_api/2/ad/get/", q, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// ---------------------------------------------------------------------------
// Reporting
// ---------------------------------------------------------------------------

// ReportRequest describes a performance report query.
type ReportRequest struct {
	AdvertiserID int64
	StartDate    string // YYYY-MM-DD
	EndDate      string // YYYY-MM-DD
	GroupBy      []string
	Fields       []string
	Page         int
	PageSize     int
}

// ReportResult is the data payload of the report endpoint. Rows are left
// untyped because the available metrics depend on the requested fields.
type ReportResult struct {
	List     []map[string]any `json:"list"`
	PageInfo PageInfo         `json:"page_info"`
}

// GetReport returns an ad-level performance report.
//
// GET /open_api/2/report/ad/get/
func (c *Client) GetReport(ctx context.Context, req ReportRequest) (*ReportResult, error) {
	q := url.Values{}
	q.Set("advertiser_id", strconv.FormatInt(req.AdvertiserID, 10))
	q.Set("start_date", req.StartDate)
	q.Set("end_date", req.EndDate)
	if len(req.GroupBy) > 0 {
		q.Set("group_by", jsonParam(req.GroupBy))
	}
	if len(req.Fields) > 0 {
		q.Set("fields", jsonParam(req.Fields))
	}
	q.Set("page", strconv.Itoa(normPage(req.Page)))
	q.Set("page_size", strconv.Itoa(normPageSize(req.PageSize)))

	var out ReportResult
	if err := c.get(ctx, "/open_api/2/report/ad/get/", q, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// ---------------------------------------------------------------------------
// Writes (gated by the server's write flag)
// ---------------------------------------------------------------------------

// UpdateCampaignStatus enables, disables or deletes campaigns. optStatus is one
// of "enable", "disable" or "delete".
//
// POST /open_api/2/campaign/update/status/
func (c *Client) UpdateCampaignStatus(ctx context.Context, advertiserID int64, campaignIDs []int64, optStatus string) error {
	body := map[string]any{
		"advertiser_id": advertiserID,
		"campaign_ids":  campaignIDs,
		"opt_status":    optStatus,
	}
	return c.post(ctx, "/open_api/2/campaign/update/status/", body, nil)
}

// UpdateCampaignBudget sets a new daily budget for a campaign. budgetMode is
// typically "BUDGET_MODE_DAY".
//
// POST /open_api/2/campaign/update/budget/
func (c *Client) UpdateCampaignBudget(ctx context.Context, advertiserID, campaignID int64, budget float64, budgetMode string) error {
	body := map[string]any{
		"advertiser_id": advertiserID,
		"data": []map[string]any{{
			"campaign_id": campaignID,
			"budget":      budget,
			"budget_mode": budgetMode,
		}},
	}
	return c.post(ctx, "/open_api/2/campaign/update/budget/", body, nil)
}

func normPage(p int) int {
	if p <= 0 {
		return 1
	}
	return p
}

func normPageSize(s int) int {
	if s <= 0 {
		return 10
	}
	if s > 100 {
		return 100
	}
	return s
}
