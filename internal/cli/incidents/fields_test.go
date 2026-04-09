package incidents

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/shhac/agent-incident/internal/api"
	"github.com/shhac/agent-incident/internal/cli/shared"
)

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
