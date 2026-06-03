package oceanengine

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"
)

func TestStaticTokenEmpty(t *testing.T) {
	if _, err := StaticToken("").Token(context.Background()); err == nil {
		t.Fatal("expected error for empty static token")
	}
	got, err := StaticToken("abc").Token(context.Background())
	if err != nil || got != "abc" {
		t.Fatalf("Token() = %q, %v", got, err)
	}
}

// refreshServer returns an httptest server that answers refresh_token calls,
// echoing a sequence of access/refresh tokens and recording request bodies.
func refreshServer(t *testing.T, calls *int32, bodies *[]refreshRequest) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != oauthRefreshPath {
			t.Errorf("path = %q, want %q", r.URL.Path, oauthRefreshPath)
		}
		raw, _ := io.ReadAll(r.Body)
		var req refreshRequest
		_ = json.Unmarshal(raw, &req)
		*bodies = append(*bodies, req)
		n := atomic.AddInt32(calls, 1)

		resp := map[string]any{"code": 0, "data": map[string]any{
			"access_token":             "access-" + itoa(n),
			"refresh_token":            "refresh-" + itoa(n),
			"expires_in":               86400,
			"refresh_token_expires_in": 2592000,
		}}
		_ = json.NewEncoder(w).Encode(resp)
	}))
}

func itoa(n int32) string { return string(rune('0' + n)) }

func TestRefreshWhenExpired(t *testing.T) {
	var calls int32
	var bodies []refreshRequest
	ts := refreshServer(t, &calls, &bodies)
	defer ts.Close()

	clk := time.Now()
	src := NewRefreshingTokenSource(123, "sec", "old-access", "seed-refresh", 0,
		WithRefreshBaseURL(ts.URL))
	src.now = func() time.Time { return clk }

	tok, err := src.Token(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if tok != "access-1" {
		t.Fatalf("token = %q, want access-1", tok)
	}
	if calls != 1 {
		t.Fatalf("calls = %d, want 1", calls)
	}
	if bodies[0].AppID != 123 || bodies[0].Secret != "sec" ||
		bodies[0].GrantType != "refresh_token" || bodies[0].RefreshToken != "seed-refresh" {
		t.Fatalf("unexpected refresh body: %+v", bodies[0])
	}
}

func TestCachesValidToken(t *testing.T) {
	var calls int32
	var bodies []refreshRequest
	ts := refreshServer(t, &calls, &bodies)
	defer ts.Close()

	clk := time.Now()
	// Fresh token valid for 1h; should be served from cache without refreshing.
	src := NewRefreshingTokenSource(1, "s", "fresh", "r", time.Hour, WithRefreshBaseURL(ts.URL))
	src.now = func() time.Time { return clk }

	for i := 0; i < 3; i++ {
		tok, err := src.Token(context.Background())
		if err != nil || tok != "fresh" {
			t.Fatalf("Token() = %q, %v", tok, err)
		}
	}
	if calls != 0 {
		t.Fatalf("expected no refresh calls, got %d", calls)
	}
}

func TestRefreshTokenRotation(t *testing.T) {
	var calls int32
	var bodies []refreshRequest
	ts := refreshServer(t, &calls, &bodies)
	defer ts.Close()

	clk := time.Now()
	src := NewRefreshingTokenSource(1, "s", "", "seed-refresh", 0, WithRefreshBaseURL(ts.URL))
	src.now = func() time.Time { return clk }

	// First call refreshes using the seed refresh token.
	if _, err := src.Token(context.Background()); err != nil {
		t.Fatal(err)
	}
	// Advance past expiry to force a second refresh; it must use the rotated token.
	clk = clk.Add(48 * time.Hour)
	if _, err := src.Token(context.Background()); err != nil {
		t.Fatal(err)
	}
	if calls != 2 {
		t.Fatalf("calls = %d, want 2", calls)
	}
	if bodies[1].RefreshToken != "refresh-1" {
		t.Fatalf("second refresh used %q, want rotated refresh-1", bodies[1].RefreshToken)
	}
}

func TestOnRefreshCallback(t *testing.T) {
	var calls int32
	var bodies []refreshRequest
	ts := refreshServer(t, &calls, &bodies)
	defer ts.Close()

	var gotAccess, gotRefresh string
	src := NewRefreshingTokenSource(1, "s", "", "seed", 0,
		WithRefreshBaseURL(ts.URL),
		WithOnRefresh(func(a, r string, _, _ time.Time) { gotAccess, gotRefresh = a, r }))
	if _, err := src.Token(context.Background()); err != nil {
		t.Fatal(err)
	}
	if gotAccess != "access-1" || gotRefresh != "refresh-1" {
		t.Fatalf("callback got %q/%q", gotAccess, gotRefresh)
	}
}

func TestNoRefreshToken(t *testing.T) {
	src := NewRefreshingTokenSource(1, "s", "", "", 0)
	if _, err := src.Token(context.Background()); err == nil {
		t.Fatal("expected error when no refresh token is available")
	}
}

func TestClientUsesRefreshingSource(t *testing.T) {
	var calls int32
	var bodies []refreshRequest
	authTS := refreshServer(t, &calls, &bodies)
	defer authTS.Close()

	// API server asserts the Access-Token header carries the refreshed token.
	apiTS := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("Access-Token"); got != "access-1" {
			t.Errorf("Access-Token = %q, want access-1", got)
		}
		_, _ = w.Write([]byte(`{"code":0,"data":[{"id":1,"name":"a"}]}`))
	}))
	defer apiTS.Close()

	src := NewRefreshingTokenSource(1, "s", "", "seed", 0, WithRefreshBaseURL(authTS.URL))
	c := NewClient("", WithBaseURL(apiTS.URL), WithTokenProvider(src))

	if _, err := c.GetAdvertiserInfo(context.Background(), []int64{1}, nil); err != nil {
		t.Fatal(err)
	}
	if calls != 1 {
		t.Fatalf("expected 1 refresh, got %d", calls)
	}
}
