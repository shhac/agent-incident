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

func TestIncidentsEditWithStatus(t *testing.T) {
	var gotBody map[string]any

	shared.SetupMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/v1/incident_statuses":
			json.NewEncoder(w).Encode(map[string]any{
				"incident_statuses": []map[string]any{
					{"id": "stat-closed", "name": "Closed", "category": "closed"},
					{"id": "stat-active", "name": "Active", "category": "active"},
				},
			})
		case strings.HasSuffix(r.URL.Path, "/actions/edit"):
			body, _ := io.ReadAll(r.Body)
			json.Unmarshal(body, &gotBody)
			json.NewEncoder(w).Encode(map[string]any{
				"incident": api.Incident{ID: "inc-1", Name: "Test", Status: api.IncidentStatusRef{Name: "Closed"}},
			})
		default:
			json.NewEncoder(w).Encode(map[string]any{
				"incident": api.Incident{ID: "inc-1", Name: "Test"},
			})
		}
	})

	root := newTestRoot()
	root.SetArgs([]string{"incident", "edit", "inc-1", "--status", "Closed"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	incident := gotBody["incident"].(map[string]any)
	if incident["incident_status_id"] != "stat-closed" {
		t.Errorf("expected incident_status_id 'stat-closed', got %v", incident["incident_status_id"])
	}
}

func TestIncidentsEditWithSeverityName(t *testing.T) {
	var gotBody map[string]any

	shared.SetupMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/v1/severities":
			json.NewEncoder(w).Encode(map[string]any{
				"severities": []map[string]any{
					{"id": "sev-crit", "name": "Critical", "rank": 1},
					{"id": "sev-major", "name": "Major", "rank": 2},
				},
			})
		case strings.HasSuffix(r.URL.Path, "/actions/edit"):
			body, _ := io.ReadAll(r.Body)
			json.Unmarshal(body, &gotBody)
			json.NewEncoder(w).Encode(map[string]any{
				"incident": api.Incident{ID: "inc-1", Name: "Test"},
			})
		default:
			json.NewEncoder(w).Encode(map[string]any{
				"incident": api.Incident{ID: "inc-1", Name: "Test"},
			})
		}
	})

	root := newTestRoot()
	root.SetArgs([]string{"incident", "edit", "inc-1", "--severity", "Critical"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	incident := gotBody["incident"].(map[string]any)
	if incident["severity_id"] != "sev-crit" {
		t.Errorf("expected severity_id 'sev-crit', got %v", incident["severity_id"])
	}
}

func TestIncidentsEditWithCustomField(t *testing.T) {
	var gotBody map[string]any

	shared.SetupMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/v2/custom_fields":
			json.NewEncoder(w).Encode(map[string]any{
				"custom_fields": []map[string]any{
					{"id": "cf-team", "name": "Affected Team", "field_type": "single_select"},
					{"id": "cf-cause", "name": "Root Cause", "field_type": "text"},
				},
			})
		case r.URL.Path == "/v1/custom_field_options":
			json.NewEncoder(w).Encode(map[string]any{
				"custom_field_options": []map[string]any{
					{"id": "opt-platform", "value": "Platform", "custom_field_id": "cf-team"},
					{"id": "opt-mobile", "value": "Mobile", "custom_field_id": "cf-team"},
				},
			})
		case strings.HasSuffix(r.URL.Path, "/actions/edit"):
			body, _ := io.ReadAll(r.Body)
			json.Unmarshal(body, &gotBody)
			json.NewEncoder(w).Encode(map[string]any{
				"incident": api.Incident{ID: "inc-1", Name: "Test"},
			})
		default:
			json.NewEncoder(w).Encode(map[string]any{
				"incident": api.Incident{ID: "inc-1", Name: "Test"},
			})
		}
	})

	root := newTestRoot()
	root.SetArgs([]string{"incident", "edit", "inc-1", "--field", "Affected Team=Platform", "--field", "Root Cause=DNS misconfiguration"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	incident := gotBody["incident"].(map[string]any)
	entries, ok := incident["custom_field_entries"].([]any)
	if !ok {
		t.Fatal("expected custom_field_entries in body")
	}
	if len(entries) != 2 {
		t.Fatalf("expected 2 custom field entries, got %d", len(entries))
	}

	entry0 := entries[0].(map[string]any)
	if entry0["custom_field_id"] != "cf-team" {
		t.Errorf("expected custom_field_id 'cf-team', got %v", entry0["custom_field_id"])
	}
	vals0 := entry0["values"].([]any)
	val0 := vals0[0].(map[string]any)
	if val0["value_option_id"] != "opt-platform" {
		t.Errorf("expected value_option_id 'opt-platform', got %v", val0["value_option_id"])
	}

	entry1 := entries[1].(map[string]any)
	if entry1["custom_field_id"] != "cf-cause" {
		t.Errorf("expected custom_field_id 'cf-cause', got %v", entry1["custom_field_id"])
	}
	vals1 := entry1["values"].([]any)
	val1 := vals1[0].(map[string]any)
	if val1["value_text"] != "DNS misconfiguration" {
		t.Errorf("expected value_text 'DNS misconfiguration', got %v", val1["value_text"])
	}
}

func TestIncidentsEditWithCatalogField(t *testing.T) {
	var gotBody map[string]any

	shared.SetupMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/v2/custom_fields":
			json.NewEncoder(w).Encode(map[string]any{
				"custom_fields": []map[string]any{
					{
						"id": "cf-team", "name": "Team", "field_type": "multi_select",
						"catalog_type_id": "cat-type-teams",
					},
				},
			})
		case r.URL.Path == "/v2/catalog_entries":
			json.NewEncoder(w).Encode(map[string]any{
				"catalog_entries": []map[string]any{
					{"id": "cat-entry-platform", "name": "Platform", "catalog_type_id": "cat-type-teams"},
				},
			})
		case strings.HasSuffix(r.URL.Path, "/actions/edit"):
			body, _ := io.ReadAll(r.Body)
			json.Unmarshal(body, &gotBody)
			json.NewEncoder(w).Encode(map[string]any{
				"incident": api.Incident{ID: "inc-1", Name: "Test"},
			})
		default:
			json.NewEncoder(w).Encode(map[string]any{
				"incident": api.Incident{ID: "inc-1", Name: "Test"},
			})
		}
	})

	root := newTestRoot()
	root.SetArgs([]string{"incident", "edit", "inc-1", "--field", "Team=Platform"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	incident := gotBody["incident"].(map[string]any)
	entries := incident["custom_field_entries"].([]any)
	entry := entries[0].(map[string]any)
	if entry["custom_field_id"] != "cf-team" {
		t.Errorf("expected custom_field_id 'cf-team', got %v", entry["custom_field_id"])
	}
	vals := entry["values"].([]any)
	val := vals[0].(map[string]any)
	if val["value_catalog_entry_id"] != "cat-entry-platform" {
		t.Errorf("expected value_catalog_entry_id 'cat-entry-platform', got %v", val["value_catalog_entry_id"])
	}
}

func TestIncidentsEditWithTimestamp(t *testing.T) {
	var gotBody map[string]any

	shared.SetupMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/v2/incident_timestamps":
			json.NewEncoder(w).Encode(map[string]any{
				"incident_timestamps": []map[string]any{
					{"id": "ts-reported", "name": "Reported at", "rank": 0},
					{"id": "ts-resolved", "name": "Resolved at", "rank": 5},
				},
			})
		case strings.HasSuffix(r.URL.Path, "/actions/edit"):
			body, _ := io.ReadAll(r.Body)
			json.Unmarshal(body, &gotBody)
			json.NewEncoder(w).Encode(map[string]any{
				"incident": api.Incident{ID: "inc-1", Name: "Test"},
			})
		default:
			json.NewEncoder(w).Encode(map[string]any{
				"incident": api.Incident{ID: "inc-1", Name: "Test"},
			})
		}
	})

	root := newTestRoot()
	root.SetArgs([]string{"incident", "edit", "inc-1", "--timestamp", "Resolved at=2026-04-09T15:30:00Z"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	incident := gotBody["incident"].(map[string]any)
	tsValues, ok := incident["incident_timestamp_values"].([]any)
	if !ok {
		t.Fatal("expected incident_timestamp_values in body")
	}
	if len(tsValues) != 1 {
		t.Fatalf("expected 1 timestamp value, got %d", len(tsValues))
	}

	tsVal := tsValues[0].(map[string]any)
	if tsVal["incident_timestamp_id"] != "ts-resolved" {
		t.Errorf("expected incident_timestamp_id 'ts-resolved', got %v", tsVal["incident_timestamp_id"])
	}
	if tsVal["value"] != "2026-04-09T15:30:00Z" {
		t.Errorf("expected value '2026-04-09T15:30:00Z', got %v", tsVal["value"])
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
