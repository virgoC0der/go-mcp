package oceanengine

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetAdvertiserInfo(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("Access-Token"); got != "tok" {
			t.Errorf("Access-Token header = %q, want %q", got, "tok")
		}
		if r.URL.Path != "/open_api/2/advertiser/info/" {
			t.Errorf("path = %q", r.URL.Path)
		}
		if got := r.URL.Query().Get("advertiser_ids"); got != "[123,456]" {
			t.Errorf("advertiser_ids = %q", got)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"code":0,"message":"OK","request_id":"req-1",
			"data":[{"id":123,"name":"acct-a","role":"ROLE_ADVERTISER","status":"STATUS_ENABLE","company":"co"}]}`))
	}))
	defer ts.Close()

	c := NewClient("tok", WithBaseURL(ts.URL))
	ads, err := c.GetAdvertiserInfo(context.Background(), []int64{123, 456}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(ads) != 1 || ads[0].ID != 123 || ads[0].Name != "acct-a" {
		t.Fatalf("unexpected advertisers: %+v", ads)
	}
}

func TestListCampaignsPagination(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		if q.Get("advertiser_id") != "777" || q.Get("page") != "2" || q.Get("page_size") != "5" {
			t.Errorf("unexpected query: %v", q)
		}
		_, _ = w.Write([]byte(`{"code":0,"data":{"list":[{"id":1,"name":"c1"}],
			"page_info":{"page":2,"page_size":5,"total_number":6,"total_page":2}}}`))
	}))
	defer ts.Close()

	c := NewClient("tok", WithBaseURL(ts.URL))
	res, err := c.ListCampaigns(context.Background(), 777, 2, 5)
	if err != nil {
		t.Fatal(err)
	}
	if len(res.List) != 1 || res.PageInfo.TotalNumber != 6 {
		t.Fatalf("unexpected result: %+v", res)
	}
}

func TestAPIErrorPropagated(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"code":40001,"message":"invalid token","request_id":"req-x"}`))
	}))
	defer ts.Close()

	c := NewClient("tok", WithBaseURL(ts.URL))
	_, err := c.ListAds(context.Background(), 1, 1, 10)
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected *APIError, got %v", err)
	}
	if apiErr.Code != 40001 || apiErr.RequestID != "req-x" {
		t.Fatalf("unexpected api error: %+v", apiErr)
	}
}

func TestMissingTokenFailsFast(t *testing.T) {
	c := NewClient("")
	_, err := c.GetAdvertiserInfo(context.Background(), []int64{1}, nil)
	if err == nil {
		t.Fatal("expected error for missing token")
	}
}

func TestNormPageSize(t *testing.T) {
	cases := []struct{ in, want int }{{0, 10}, {-3, 10}, {50, 50}, {500, 100}}
	for _, tc := range cases {
		if got := normPageSize(tc.in); got != tc.want {
			t.Errorf("normPageSize(%d) = %d, want %d", tc.in, got, tc.want)
		}
	}
}
