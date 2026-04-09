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

func TestParseKeyValue(t *testing.T) {
	tests := []struct {
		input     string
		wantKey   string
		wantValue string
		wantErr   bool
	}{
		{"Name=Value", "Name", "Value", false},
		{"Key=", "Key", "", false},
		{"=Value", "", "Value", false},
		{"Key=Val=ue", "Key", "Val=ue", false},
		{"NoEquals", "", "", true},
		{"", "", "", true},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			key, value, err := parseKeyValue(tt.input)
			if (err != nil) != tt.wantErr {
				t.Fatalf("parseKeyValue(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
			if key != tt.wantKey {
				t.Errorf("key = %q, want %q", key, tt.wantKey)
			}
			if value != tt.wantValue {
				t.Errorf("value = %q, want %q", value, tt.wantValue)
			}
		})
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

func TestIncidentsEditWithNumericField(t *testing.T) {
	var gotBody map[string]any

	shared.SetupMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/v2/custom_fields":
			json.NewEncoder(w).Encode(map[string]any{
				"custom_fields": []map[string]any{
					{"id": "cf-count", "name": "Affected Count", "field_type": "numeric"},
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
	root.SetArgs([]string{"incident", "edit", "inc-1", "--field", "Affected Count=42"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	incident := gotBody["incident"].(map[string]any)
	entries := incident["custom_field_entries"].([]any)
	entry := entries[0].(map[string]any)
	vals := entry["values"].([]any)
	val := vals[0].(map[string]any)
	if val["value_numeric"] != "42" {
		t.Errorf("expected value_numeric '42', got %v", val["value_numeric"])
	}
}

func TestIncidentsEditWithLinkField(t *testing.T) {
	var gotBody map[string]any

	shared.SetupMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/v2/custom_fields":
			json.NewEncoder(w).Encode(map[string]any{
				"custom_fields": []map[string]any{
					{"id": "cf-link", "name": "Runbook", "field_type": "link"},
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
	root.SetArgs([]string{"incident", "edit", "inc-1", "--field", "Runbook=https://example.com/runbook"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	incident := gotBody["incident"].(map[string]any)
	entries := incident["custom_field_entries"].([]any)
	entry := entries[0].(map[string]any)
	vals := entry["values"].([]any)
	val := vals[0].(map[string]any)
	if val["value_link"] != "https://example.com/runbook" {
		t.Errorf("expected value_link, got %v", val["value_link"])
	}
}

func TestIncidentsEditClearField(t *testing.T) {
	var gotBody map[string]any

	shared.SetupMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/v2/custom_fields":
			json.NewEncoder(w).Encode(map[string]any{
				"custom_fields": []map[string]any{
					{"id": "cf-cause", "name": "Root Cause", "field_type": "text"},
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
	root.SetArgs([]string{"incident", "edit", "inc-1", "--field", "Root Cause="})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	incident := gotBody["incident"].(map[string]any)
	entries := incident["custom_field_entries"].([]any)
	entry := entries[0].(map[string]any)
	vals := entry["values"].([]any)
	if len(vals) != 0 {
		t.Errorf("expected empty values array to clear field, got %v", vals)
	}
}

func TestIncidentsEditClearTimestamp(t *testing.T) {
	var gotBody map[string]any

	shared.SetupMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/v2/incident_timestamps":
			json.NewEncoder(w).Encode(map[string]any{
				"incident_timestamps": []map[string]any{
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
	root.SetArgs([]string{"incident", "edit", "inc-1", "--timestamp", "Resolved at="})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	incident := gotBody["incident"].(map[string]any)
	tsValues := incident["incident_timestamp_values"].([]any)
	tsVal := tsValues[0].(map[string]any)
	if tsVal["value"] != nil {
		t.Errorf("expected null value to clear timestamp, got %v", tsVal["value"])
	}
}

func TestIncidentsEditTimestampParseError(t *testing.T) {
	shared.SetupMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/v2/incident_timestamps":
			json.NewEncoder(w).Encode(map[string]any{
				"incident_timestamps": []map[string]any{
					{"id": "ts-resolved", "name": "Resolved at", "rank": 5},
				},
			})
		default:
			json.NewEncoder(w).Encode(map[string]any{
				"incident": api.Incident{ID: "inc-1", Name: "Test"},
			})
		}
	})

	root := newTestRoot()
	root.SetArgs([]string{"incident", "edit", "inc-1", "--timestamp", "Resolved at=not-a-date"})

	err := root.Execute()
	if err == nil {
		t.Fatal("expected error for invalid timestamp")
	}
	if !strings.Contains(err.Error(), "Resolved at") {
		t.Errorf("expected error to mention timestamp name, got %q", err.Error())
	}
}

func TestIncidentsEditFieldNotFound(t *testing.T) {
	shared.SetupMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/v2/custom_fields":
			json.NewEncoder(w).Encode(map[string]any{
				"custom_fields": []map[string]any{
					{"id": "cf-cause", "name": "Root Cause", "field_type": "text"},
				},
			})
		default:
			json.NewEncoder(w).Encode(map[string]any{
				"incident": api.Incident{ID: "inc-1", Name: "Test"},
			})
		}
	})

	root := newTestRoot()
	root.SetArgs([]string{"incident", "edit", "inc-1", "--field", "Nonexistent=value"})

	err := root.Execute()
	if err == nil {
		t.Fatal("expected error for unknown custom field")
	}
	if !strings.Contains(err.Error(), "Nonexistent") {
		t.Errorf("expected error to mention field name, got %q", err.Error())
	}
}
