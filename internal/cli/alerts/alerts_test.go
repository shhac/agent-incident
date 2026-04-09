package alerts

import (
	"encoding/json"
	"io"
	"net/http"
	"testing"

	"github.com/shhac/agent-incident/internal/api"
	"github.com/shhac/agent-incident/internal/api/testdata"
	"github.com/shhac/agent-incident/internal/cli/shared"
)

func TestAlertsList(t *testing.T) {
	var gotPath, gotMethod string
	var gotQuery map[string][]string

	shared.SetupMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		gotQuery = r.URL.Query()
		w.Write(testdata.Load("alerts_list.json"))
	})

	root := shared.NewTestRoot(Register)
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

	root := shared.NewTestRoot(Register)
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

func TestAlertsIncidents(t *testing.T) {
	var gotPath, gotMethod string

	shared.SetupMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		json.NewEncoder(w).Encode(map[string]any{
			"incident_alerts": []api.IncidentAlert{
				{ID: "ia-1", AlertID: "alert-1", Title: "Alert in incident", Status: "firing"},
			},
		})
	})

	root := shared.NewTestRoot(Register)
	root.SetArgs([]string{"alert", "incidents"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if gotPath != "/v2/incident_alerts" {
		t.Errorf("expected path /v2/incident_alerts, got %q", gotPath)
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

	root := shared.NewTestRoot(Register)
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
