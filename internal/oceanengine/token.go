package oceanengine

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"
)

// oauthRefreshPath is the Ocean Engine OAuth2 refresh endpoint.
const oauthRefreshPath = "/open_api/oauth2/refresh_token/"

// TokenProvider supplies a valid access token for API requests, refreshing it
// transparently when necessary. Implementations must be safe for concurrent use.
type TokenProvider interface {
	Token(ctx context.Context) (string, error)
}

// StaticToken is a TokenProvider that always returns the same token. Use it
// when the access token's lifecycle is managed outside this package.
type StaticToken string

// Token implements TokenProvider.
func (s StaticToken) Token(context.Context) (string, error) {
	if s == "" {
		return "", fmt.Errorf("oceanengine: missing access token")
	}
	return string(s), nil
}

// refreshRequest is the body of the oauth2/refresh_token endpoint.
type refreshRequest struct {
	AppID        uint64 `json:"app_id"`
	Secret       string `json:"secret"`
	GrantType    string `json:"grant_type"`
	RefreshToken string `json:"refresh_token"`
}

// tokenData is the data payload returned by the OAuth token endpoints.
type tokenData struct {
	AccessToken           string `json:"access_token"`
	RefreshToken          string `json:"refresh_token"`
	ExpiresIn             int64  `json:"expires_in"`               // seconds
	RefreshTokenExpiresIn int64  `json:"refresh_token_expires_in"` // seconds
}

// RefreshingTokenSource holds an Ocean Engine OAuth token pair and refreshes the
// access token via the oauth2/refresh_token endpoint shortly before it expires.
//
// Ocean Engine rotates the refresh token on every refresh (the previous one is
// invalidated), so callers that need to persist the refresh token across process
// restarts should register an OnRefresh callback and store the latest value.
//
// It is safe for concurrent use.
type RefreshingTokenSource struct {
	appID  uint64
	secret string

	baseURL    string
	httpClient *http.Client
	skew       time.Duration // refresh this long before expiry
	now        func() time.Time
	onRefresh  func(accessToken, refreshToken string, accessExpiry, refreshExpiry time.Time)

	mu            sync.Mutex
	accessToken   string
	refreshToken  string
	accessExpiry  time.Time
	refreshExpiry time.Time
}

// RefreshOption customizes a RefreshingTokenSource.
type RefreshOption func(*RefreshingTokenSource)

// WithRefreshBaseURL overrides the API host used for token refresh.
func WithRefreshBaseURL(u string) RefreshOption {
	return func(s *RefreshingTokenSource) {
		if u != "" {
			s.baseURL = u
		}
	}
}

// WithRefreshHTTPClient injects a custom *http.Client for token refresh.
func WithRefreshHTTPClient(h *http.Client) RefreshOption {
	return func(s *RefreshingTokenSource) {
		if h != nil {
			s.httpClient = h
		}
	}
}

// WithRefreshSkew refreshes the access token this long before it actually
// expires (default 5 minutes), guarding against clock skew and in-flight calls.
func WithRefreshSkew(d time.Duration) RefreshOption {
	return func(s *RefreshingTokenSource) {
		if d >= 0 {
			s.skew = d
		}
	}
}

// WithOnRefresh registers a callback invoked after every successful refresh with
// the new token pair and their expiry times. Use it to persist the rotated
// refresh token.
func WithOnRefresh(fn func(accessToken, refreshToken string, accessExpiry, refreshExpiry time.Time)) RefreshOption {
	return func(s *RefreshingTokenSource) { s.onRefresh = fn }
}

// NewRefreshingTokenSource builds a self-refreshing token source from an OAuth
// token pair. accessExpiresIn is the remaining lifetime of accessToken; pass 0
// (or an empty accessToken) to force a refresh on first use.
func NewRefreshingTokenSource(appID uint64, secret, accessToken, refreshToken string, accessExpiresIn time.Duration, opts ...RefreshOption) *RefreshingTokenSource {
	s := &RefreshingTokenSource{
		appID:        appID,
		secret:       secret,
		accessToken:  accessToken,
		refreshToken: refreshToken,
		baseURL:      DefaultBaseURL,
		httpClient:   &http.Client{Timeout: 30 * time.Second},
		skew:         5 * time.Minute,
		now:          time.Now,
	}
	if accessExpiresIn > 0 {
		s.accessExpiry = s.now().Add(accessExpiresIn)
	}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

// Token implements TokenProvider, refreshing the access token if it is missing
// or within the skew window of expiry.
func (s *RefreshingTokenSource) Token(ctx context.Context) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.accessToken != "" && !s.accessExpiry.IsZero() && s.now().Before(s.accessExpiry.Add(-s.skew)) {
		return s.accessToken, nil
	}
	if err := s.refreshLocked(ctx); err != nil {
		return "", err
	}
	return s.accessToken, nil
}

func (s *RefreshingTokenSource) refreshLocked(ctx context.Context) error {
	if s.refreshToken == "" {
		return fmt.Errorf("oceanengine: no refresh token available")
	}
	body := refreshRequest{
		AppID:        s.appID,
		Secret:       s.secret,
		GrantType:    "refresh_token",
		RefreshToken: s.refreshToken,
	}
	var data tokenData
	if err := postJSON(ctx, s.httpClient, s.baseURL+oauthRefreshPath, body, &data); err != nil {
		return fmt.Errorf("oceanengine: refresh token: %w", err)
	}
	if data.AccessToken == "" {
		return fmt.Errorf("oceanengine: refresh returned empty access token")
	}

	s.accessToken = data.AccessToken
	if data.RefreshToken != "" {
		s.refreshToken = data.RefreshToken // rotated; old one is now invalid
	}
	now := s.now()
	s.accessExpiry = now.Add(time.Duration(data.ExpiresIn) * time.Second)
	if data.RefreshTokenExpiresIn > 0 {
		s.refreshExpiry = now.Add(time.Duration(data.RefreshTokenExpiresIn) * time.Second)
	}
	if s.onRefresh != nil {
		s.onRefresh(s.accessToken, s.refreshToken, s.accessExpiry, s.refreshExpiry)
	}
	return nil
}
