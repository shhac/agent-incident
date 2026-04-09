package incidents

import (
	"encoding/json"
	"io"
	"net/http"
	"testing"

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

func TestIncidentsList(t *testing.T) {
	var gotPath, gotMethod string
	var gotQuery map[string][]string

	shared.SetupMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		gotQuery = r.URL.Query()
		json.NewEncoder(w).Encode(map[string]any{
			"incidents": []api.Incident{
				{
					ID:   "inc-1",
					Name: "Test Incident",
					Status: api.IncidentStatusRef{
						Category: "active",
						Name:     "Investigating",
					},
					CreatedAt: "2024-01-01T00:00:00Z",
				},
			},
			"pagination_meta": map[string]any{
				"after": "cursor-abc",
			},
		})
	})

	root := newTestRoot()
	root.SetArgs([]string{"incidents", "list", "--status", "active", "--limit", "10"})

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
		json.NewEncoder(w).Encode(map[string]any{
			"incident": api.Incident{
				ID:   "01ABC123DEF456",
				Name: "Specific Incident",
				Status: api.IncidentStatusRef{
					Category: "active",
					Name:     "Investigating",
				},
			},
		})
	})

	root := newTestRoot()
	root.SetArgs([]string{"incidents", "get", "01ABC123DEF456"})

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
	root.SetArgs([]string{"incidents", "create", "--name", "New Incident", "--summary", "Something broke"})

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
	root.SetArgs([]string{"incidents", "edit", "inc-edit", "--summary", "Updated summary"})

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

func TestIncidentsUpdates(t *testing.T) {
	var gotPath, gotMethod string
	var gotQuery map[string][]string

	shared.SetupMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		gotQuery = r.URL.Query()
		json.NewEncoder(w).Encode(map[string]any{
			"incident_updates": []api.IncidentUpdate{
				{
					ID:         "upd-1",
					IncidentID: "inc-42",
					Message:    "Status update",
					CreatedAt:  "2024-01-01T00:00:00Z",
				},
			},
		})
	})

	root := newTestRoot()
	root.SetArgs([]string{"incidents", "updates", "inc-42"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if gotPath != "/v2/incident_updates" {
		t.Errorf("expected path /v2/incident_updates, got %q", gotPath)
	}
	if gotMethod != http.MethodGet {
		t.Errorf("expected GET, got %s", gotMethod)
	}
	if vals := gotQuery["incident_id"]; len(vals) == 0 || vals[0] != "inc-42" {
		t.Errorf("expected incident_id=inc-42, got %v", gotQuery["incident_id"])
	}
}
