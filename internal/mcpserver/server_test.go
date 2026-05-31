package mcpserver

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/virgoC0der/go-mcp/internal/oceanengine"
)

// connect wires an in-memory MCP client to a server backed by the given
// Ocean Engine HTTP test server.
func connect(t *testing.T, apiURL string, cfg Config) *mcp.ClientSession {
	t.Helper()
	ctx := context.Background()

	client := oceanengine.NewClient("tok", oceanengine.WithBaseURL(apiURL))
	srv := New(client, cfg)

	c := mcp.NewClient(&mcp.Implementation{Name: "test", Version: "v0"}, nil)
	st, ct := mcp.NewInMemoryTransports()
	if _, err := srv.Connect(ctx, st, nil); err != nil {
		t.Fatalf("server connect: %v", err)
	}
	cs, err := c.Connect(ctx, ct, nil)
	if err != nil {
		t.Fatalf("client connect: %v", err)
	}
	t.Cleanup(func() { _ = cs.Close() })
	return cs
}

func toolNames(t *testing.T, cs *mcp.ClientSession) map[string]bool {
	t.Helper()
	names := map[string]bool{}
	for tool, err := range cs.Tools(context.Background(), nil) {
		if err != nil {
			t.Fatal(err)
		}
		names[tool.Name] = true
	}
	return names
}

func TestReadOnlyByDefault(t *testing.T) {
	cs := connect(t, "http://unused", Config{})
	names := toolNames(t, cs)

	for _, want := range []string{
		"oceanengine_get_advertiser_info",
		"oceanengine_list_campaigns",
		"oceanengine_list_ads",
		"oceanengine_get_report",
	} {
		if !names[want] {
			t.Errorf("missing read tool %q", want)
		}
	}
	if names["oceanengine_update_campaign_status"] || names["oceanengine_update_campaign_budget"] {
		t.Error("write tools must not be registered when EnableWrites is false")
	}
}

func TestWriteToolsGated(t *testing.T) {
	cs := connect(t, "http://unused", Config{EnableWrites: true})
	names := toolNames(t, cs)
	if !names["oceanengine_update_campaign_status"] || !names["oceanengine_update_campaign_budget"] {
		t.Error("write tools should be registered when EnableWrites is true")
	}
}

func TestCallToolRoundTrip(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"code":0,"data":[{"id":123,"name":"acct-a"}]}`))
	}))
	defer ts.Close()

	cs := connect(t, ts.URL, Config{})
	res, err := cs.CallTool(context.Background(), &mcp.CallToolParams{
		Name:      "oceanengine_get_advertiser_info",
		Arguments: map[string]any{"advertiser_ids": []int64{123}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if res.IsError {
		t.Fatalf("tool returned error result: %+v", res.Content)
	}
	if len(res.Content) == 0 {
		t.Fatal("expected content in result")
	}
}

func TestCallToolValidationError(t *testing.T) {
	cs := connect(t, "http://unused", Config{})
	res, err := cs.CallTool(context.Background(), &mcp.CallToolParams{
		Name:      "oceanengine_get_advertiser_info",
		Arguments: map[string]any{"advertiser_ids": []int64{}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if !res.IsError {
		t.Fatal("expected IsError for empty advertiser_ids")
	}
}
