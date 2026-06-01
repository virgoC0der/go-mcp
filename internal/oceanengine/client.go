// Package oceanengine is a thin client for the Ocean Engine (巨量引擎) Marketing
// API. It deliberately covers only the handful of endpoints the MCP server
// exposes as tools, keeping the dependency surface and the maintenance burden
// small. To expand coverage later, methods here can be backed by the much
// larger community SDK github.com/bububa/oceanengine without changing the MCP
// tool layer.
package oceanengine

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

// DefaultBaseURL is the production Ocean Engine open API host.
const DefaultBaseURL = "https://api.oceanengine.com"

// Client talks to the Ocean Engine Marketing API. It obtains the access token
// for each request from a TokenProvider, so callers can supply either a static
// token or a self-refreshing OAuth token source.
type Client struct {
	baseURL    string
	tokens     TokenProvider
	httpClient *http.Client
}

// Option customizes a Client.
type Option func(*Client)

// WithBaseURL overrides the API host (useful for the sandbox or for tests).
func WithBaseURL(u string) Option {
	return func(c *Client) {
		if u != "" {
			c.baseURL = u
		}
	}
}

// WithHTTPClient injects a custom *http.Client.
func WithHTTPClient(h *http.Client) Option {
	return func(c *Client) {
		if h != nil {
			c.httpClient = h
		}
	}
}

// WithTokenProvider supplies the token source used to authenticate requests.
// When set, it overrides the access token passed to NewClient. Use this with a
// *RefreshingTokenSource to enable automatic OAuth token refresh.
func WithTokenProvider(tp TokenProvider) Option {
	return func(c *Client) {
		if tp != nil {
			c.tokens = tp
		}
	}
}

// NewClient builds a Client. The accessToken is the OAuth access_token issued
// by the Ocean Engine open platform; pass WithTokenProvider to manage refresh
// automatically instead.
func NewClient(accessToken string, opts ...Option) *Client {
	c := &Client{
		baseURL:    DefaultBaseURL,
		tokens:     StaticToken(accessToken),
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

// envelope is the standard Ocean Engine response wrapper. Every endpoint
// returns a non-zero Code on failure; Data carries the endpoint-specific
// payload on success.
type envelope struct {
	Code      int             `json:"code"`
	Message   string          `json:"message"`
	RequestID string          `json:"request_id"`
	Data      json.RawMessage `json:"data"`
}

// APIError represents a non-zero Code returned by the Ocean Engine API.
type APIError struct {
	Code      int
	Message   string
	RequestID string
}

func (e *APIError) Error() string {
	return fmt.Sprintf("oceanengine api error: code=%d message=%q request_id=%s", e.Code, e.Message, e.RequestID)
}

// get performs an authenticated GET request and unmarshals the data field of
// the response envelope into out.
func (c *Client) get(ctx context.Context, path string, query url.Values, out any) error {
	u := c.baseURL + path
	if len(query) > 0 {
		u += "?" + query.Encode()
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return err
	}
	return c.authedDo(req, out)
}

// post performs an authenticated POST request with a JSON body and unmarshals
// the data field of the response envelope into out.
func (c *Client) post(ctx context.Context, path string, body, out any) error {
	buf, err := json.Marshal(body)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+path, bytes.NewReader(buf))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	return c.authedDo(req, out)
}

// authedDo attaches the access token from the provider and executes the request.
func (c *Client) authedDo(req *http.Request, out any) error {
	token, err := c.tokens.Token(req.Context())
	if err != nil {
		return err
	}
	// Ocean Engine authenticates via the Access-Token header.
	req.Header.Set("Access-Token", token)
	return doRequest(c.httpClient, req, out)
}

// doRequest executes req, parses the standard Ocean Engine response envelope and
// unmarshals the data field into out. A non-zero envelope code becomes an
// *APIError. It performs no authentication, so it is also used for the OAuth
// endpoints, which authenticate with app credentials in the body.
func doRequest(httpClient *http.Client, req *http.Request, out any) error {
	resp, err := httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var env envelope
	if err := json.Unmarshal(raw, &env); err != nil {
		return fmt.Errorf("oceanengine: decode response (http %d): %w", resp.StatusCode, err)
	}
	if env.Code != 0 {
		return &APIError{Code: env.Code, Message: env.Message, RequestID: env.RequestID}
	}
	if out == nil || len(env.Data) == 0 {
		return nil
	}
	if err := json.Unmarshal(env.Data, out); err != nil {
		return fmt.Errorf("oceanengine: decode data: %w", err)
	}
	return nil
}

// postJSON sends an unauthenticated JSON POST to url and decodes the response
// envelope's data into out. Used by the OAuth token endpoints.
func postJSON(ctx context.Context, httpClient *http.Client, url string, body, out any) error {
	buf, err := json.Marshal(body)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(buf))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	return doRequest(httpClient, req, out)
}

// jsonParam encodes v as a compact JSON string for use as a query parameter,
// which is how Ocean Engine expects array/object parameters on GET endpoints
// (e.g. advertiser_ids=[123], fields=["id","name"]).
func jsonParam(v any) string {
	b, _ := json.Marshal(v)
	return string(b)
}
