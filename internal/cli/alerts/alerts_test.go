package alerts

import (
	"encoding/json"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/spf13/cobra"

	"github.com/shhac/agent-incident/internal/api"
	"github.com/shhac/agent-incident/internal/cli/shared"
)

func newTestRoot() *cobra.Command {
	root := &cobra.Command{
		Use:           "agent-incident",
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	globals := func() *shared.GlobalFlags {
		return &shared.GlobalFlags{}
	}
	Register(root, globals)
	return root
}

func TestAlertsList(t *testing.T) {
	var gotPath, gotMethod string
	var gotQuery map[string][]string

	shared.SetupMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		gotQuery = r.URL.Query()
		json.NewEncoder(w).Encode(map[string]any{
			"alerts": []api.Alert{
				{
					ID:        "alert-1",
					Title:     "CPU High",
					Status:    "firing",
					CreatedAt: time.Now(),
				},
			},
			"pagination_meta": map[string]any{
				"after": "",
			},
		})
	})

	root := newTestRoot()
	root.SetArgs([]string{"alert", "list", "--status", "firing", "--limit", "5"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if gotPath != "/v2/alerts" {
		t.Errorf("expected path /v2/alerts, got %q", gotPath)
	}
	if gotMethod != http.MethodGet {
		t.Errorf("expected GET, got %s", gotMethod)
	}
	if vals := gotQuery["status[]"]; len(vals) == 0 || vals[0] != "firing" {
		t.Errorf("expected status[]=firing, got %v", gotQuery["status[]"])
	}
	if vals := gotQuery["page_size"]; len(vals) == 0 || vals[0] != "5" {
		t.Errorf("expected page_size=5, got %v", gotQuery["page_size"])
	}
}

func TestAlertsGet(t *testing.T) {
	var gotPath, gotMethod string

	shared.SetupMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		json.NewEncoder(w).Encode(map[string]any{
			"alert": api.Alert{
				ID:     "alert-99",
				Title:  "Disk Full",
				Status: "firing",
			},
		})
	})

	root := newTestRoot()
	root.SetArgs([]string{"alert", "get", "alert-99"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if gotPath != "/v2/alerts/alert-99" {
		t.Errorf("expected path /v2/alerts/alert-99, got %q", gotPath)
	}
	if gotMethod != http.MethodGet {
		t.Errorf("expected GET, got %s", gotMethod)
	}
}

func TestAlertsCreate(t *testing.T) {
	var gotPath, gotMethod string
	var gotBody map[string]any

	shared.SetupMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		body, _ := io.ReadAll(r.Body)
		json.Unmarshal(body, &gotBody)
		json.NewEncoder(w).Encode(map[string]any{
			"alert": api.Alert{
				ID:     "alert-new",
				Title:  "New Alert",
				Status: "firing",
			},
		})
	})

	root := newTestRoot()
	root.SetArgs([]string{"alert", "create", "--source-id", "src-abc", "--title", "New Alert", "--description", "Something happened"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if gotPath != "/v2/alert_events/http/src-abc" {
		t.Errorf("expected path /v2/alert_events/http/src-abc, got %q", gotPath)
	}
	if gotMethod != http.MethodPost {
		t.Errorf("expected POST, got %s", gotMethod)
	}
	if gotBody["title"] != "New Alert" {
		t.Errorf("expected title 'New Alert', got %v", gotBody["title"])
	}
	if gotBody["description"] != "Something happened" {
		t.Errorf("expected description 'Something happened', got %v", gotBody["description"])
	}
}
