package incidents

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/spf13/cobra"

	"github.com/shhac/agent-incident/internal/api"
	"github.com/shhac/agent-incident/internal/api/testdata"
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

func TestIncidentsList(t *testing.T) {
	var gotPath, gotMethod string
	var gotQuery map[string][]string

	shared.SetupMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		gotQuery = r.URL.Query()
		w.Write(testdata.Load("incidents_list.json"))
	})

	root := newTestRoot()
	root.SetArgs([]string{"incident", "list", "--status", "active", "--limit", "10"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if gotPath != "/v2/incidents" {
		t.Errorf("expected path /v2/incidents, got %q", gotPath)
	}
	if gotMethod != http.MethodGet {
		t.Errorf("expected GET, got %s", gotMethod)
	}
	if vals := gotQuery["status_category[one_of][]"]; len(vals) == 0 || vals[0] != "active" {
		t.Errorf("expected status_category[one_of][]=active, got %v", gotQuery["status_category[one_of][]"])
	}
	if vals := gotQuery["page_size"]; len(vals) == 0 || vals[0] != "10" {
		t.Errorf("expected page_size=10, got %v", gotQuery["page_size"])
	}
}

func TestIncidentsGet(t *testing.T) {
	var gotPath, gotMethod string

	shared.SetupMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		w.Write(testdata.Load("incident_get.json"))
	})

	root := newTestRoot()
	root.SetArgs([]string{"incident", "get", "01ABC123DEF456"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if gotPath != "/v2/incidents/01ABC123DEF456" {
		t.Errorf("expected path /v2/incidents/01ABC123DEF456, got %q", gotPath)
	}
	if gotMethod != http.MethodGet {
		t.Errorf("expected GET, got %s", gotMethod)
	}
}

func TestIncidentsCreate(t *testing.T) {
	var gotPath, gotMethod string
	var gotBody map[string]any

	shared.SetupMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		body, _ := io.ReadAll(r.Body)
		json.Unmarshal(body, &gotBody)
		json.NewEncoder(w).Encode(map[string]any{
			"incident": api.Incident{
				ID:   "inc-new",
				Name: "New Incident",
				Status: api.IncidentStatusRef{
					Category: "active",
					Name:     "Investigating",
				},
			},
		})
	})

	root := newTestRoot()
	root.SetArgs([]string{"incident", "create", "--name", "New Incident", "--summary", "Something broke"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if gotPath != "/v2/incidents" {
		t.Errorf("expected path /v2/incidents, got %q", gotPath)
	}
	if gotMethod != http.MethodPost {
		t.Errorf("expected POST, got %s", gotMethod)
	}
	if gotBody["name"] != "New Incident" {
		t.Errorf("expected name 'New Incident', got %v", gotBody["name"])
	}
	if gotBody["summary"] != "Something broke" {
		t.Errorf("expected summary 'Something broke', got %v", gotBody["summary"])
	}
	if _, ok := gotBody["idempotency_key"]; !ok {
		t.Error("expected idempotency_key in body")
	}
}

func TestIncidentsEdit(t *testing.T) {
	var gotPath, gotMethod string
	var gotBody map[string]any

	shared.SetupMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		body, _ := io.ReadAll(r.Body)
		json.Unmarshal(body, &gotBody)
		json.NewEncoder(w).Encode(map[string]any{
			"incident": api.Incident{
				ID:   "inc-edit",
				Name: "Edited Incident",
				Status: api.IncidentStatusRef{
					Category: "active",
					Name:     "Investigating",
				},
			},
		})
	})

	root := newTestRoot()
	root.SetArgs([]string{"incident", "edit", "inc-edit", "--summary", "Updated summary"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if gotPath != "/v2/incidents/inc-edit/actions/edit" {
		t.Errorf("expected path /v2/incidents/inc-edit/actions/edit, got %q", gotPath)
	}
	if gotMethod != http.MethodPost {
		t.Errorf("expected POST, got %s", gotMethod)
	}
	incident, ok := gotBody["incident"].(map[string]any)
	if !ok {
		t.Fatal("expected incident field in body")
	}
	if incident["summary"] != "Updated summary" {
		t.Errorf("expected summary 'Updated summary', got %v", incident["summary"])
	}
}

func TestNormalizeIncidentRef(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"2000", "2000"},
		{"INC-2000", "2000"},
		{"inc-2000", "2000"},
		{"Inc-2000", "2000"},
		{"01ABC123DEF456", "01ABC123DEF456"},
		{"INC-abc", "INC-abc"},
		{"", ""},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := normalizeIncidentRef(tt.input)
			if got != tt.want {
				t.Errorf("normalizeIncidentRef(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestNormalizeIncidentRefViaGet(t *testing.T) {
	var gotPath string

	shared.SetupMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		json.NewEncoder(w).Encode(map[string]any{
			"incident": api.Incident{
				ID:   "inc-2000",
				Name: "Test",
				Status: api.IncidentStatusRef{
					Category: "active",
					Name:     "Investigating",
				},
			},
		})
	})

	root := newTestRoot()
	root.SetArgs([]string{"incident", "get", "INC-2000"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if gotPath != "/v2/incidents/2000" {
		t.Errorf("expected path /v2/incidents/2000, got %q", gotPath)
	}
}

func TestIncidentsUpdates(t *testing.T) {
	var gotUpdatesPath, gotUpdatesMethod string
	var gotQuery map[string][]string

	shared.SetupMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/v2/incidents/") && !strings.Contains(r.URL.Path, "incident_updates") {
			// Resolve incident ID
			json.NewEncoder(w).Encode(map[string]any{
				"incident": api.Incident{
					ID:   "01ABC-RESOLVED-UUID",
					Name: "Test",
				},
			})
			return
		}
		gotUpdatesPath = r.URL.Path
		gotUpdatesMethod = r.Method
		gotQuery = r.URL.Query()
		json.NewEncoder(w).Encode(map[string]any{
			"incident_updates": []api.IncidentUpdate{
				{
					ID:         "upd-1",
					IncidentID: "01ABC-RESOLVED-UUID",
					Message:    "Status update",
					CreatedAt:  "2024-01-01T00:00:00Z",
				},
			},
		})
	})

	root := newTestRoot()
	root.SetArgs([]string{"incident", "updates", "inc-42"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if gotUpdatesPath != "/v2/incident_updates" {
		t.Errorf("expected path /v2/incident_updates, got %q", gotUpdatesPath)
	}
	if gotUpdatesMethod != http.MethodGet {
		t.Errorf("expected GET, got %s", gotUpdatesMethod)
	}
	if vals := gotQuery["incident_id"]; len(vals) == 0 || vals[0] != "01ABC-RESOLVED-UUID" {
		t.Errorf("expected incident_id=inc-42, got %v", gotQuery["incident_id"])
	}
}
