package escalations

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/shhac/agent-incident/internal/api"
	"github.com/shhac/agent-incident/internal/cli/shared"
)

func TestEscalationsList(t *testing.T) {
	var gotPath, gotMethod string

	shared.SetupMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		json.NewEncoder(w).Encode(map[string]any{
			"escalations": []api.Escalation{
				{ID: "esc-1", Title: "P1 Escalation"},
				{ID: "esc-2", Title: "P2 Escalation"},
			},
		})
	})

	root := shared.NewTestRoot(Register)
	root.SetArgs([]string{"escalation", "list"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if gotPath != "/v2/escalations" {
		t.Errorf("expected path /v2/escalations, got %q", gotPath)
	}
	if gotMethod != http.MethodGet {
		t.Errorf("expected GET, got %s", gotMethod)
	}
}

func TestEscalationsGet(t *testing.T) {
	var gotPath, gotMethod string

	shared.SetupMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		json.NewEncoder(w).Encode(map[string]any{
			"escalation": api.Escalation{ID: "esc-42", Title: "Critical"},
		})
	})

	root := shared.NewTestRoot(Register)
	root.SetArgs([]string{"escalation", "get", "esc-42"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if gotPath != "/v2/escalations/esc-42" {
		t.Errorf("expected path /v2/escalations/esc-42, got %q", gotPath)
	}
	if gotMethod != http.MethodGet {
		t.Errorf("expected GET, got %s", gotMethod)
	}
}

func TestEscalationPathsList(t *testing.T) {
	var gotPath, gotMethod string

	shared.SetupMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		json.NewEncoder(w).Encode(map[string]any{
			"escalation_paths": []api.EscalationPath{
				{ID: "01HRDKWWNGX330JQ4J1PATH001", Name: "Primary Path"},
				{ID: "01HRDKWWNGX330JQ4J1PATH002", Name: "Secondary Path"},
			},
		})
	})

	root := shared.NewTestRoot(Register)
	root.SetArgs([]string{"escalation", "path", "list"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if gotPath != "/v2/escalation_paths" {
		t.Errorf("expected path /v2/escalation_paths, got %q", gotPath)
	}
	if gotMethod != http.MethodGet {
		t.Errorf("expected GET, got %s", gotMethod)
	}
}

func TestEscalationPathsGet(t *testing.T) {
	var gotPath, gotMethod string

	shared.SetupMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		json.NewEncoder(w).Encode(map[string]any{
			"escalation_path": api.EscalationPath{ID: "01HRDKWWNGX330JQ4JPATH0042", Name: "Critical Path"},
		})
	})

	root := shared.NewTestRoot(Register)
	root.SetArgs([]string{"escalation", "path", "get", "01HRDKWWNGX330JQ4JPATH0042"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if gotPath != "/v2/escalation_paths/01HRDKWWNGX330JQ4JPATH0042" {
		t.Errorf("expected path /v2/escalation_paths/01HRDKWWNGX330JQ4JPATH0042, got %q", gotPath)
	}
	if gotMethod != http.MethodGet {
		t.Errorf("expected GET, got %s", gotMethod)
	}
}

func TestEscalationsCreate(t *testing.T) {
	var gotPath, gotMethod string
	var gotBody map[string]any

	shared.SetupMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		if r.Method == http.MethodPost {
			json.NewDecoder(r.Body).Decode(&gotBody)
		}
		json.NewEncoder(w).Encode(map[string]any{
			"escalation": api.Escalation{ID: "esc-new", Title: "New Escalation"},
		})
	})

	root := shared.NewTestRoot(Register)
	root.SetArgs([]string{
		"escalation", "create",
		"--incident", "inc-1",
		"--path", "01HRDKWWNGX330JQ4J1PATH001",
	})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if gotPath != "/v2/escalations" {
		t.Errorf("expected path /v2/escalations, got %q", gotPath)
	}
	if gotMethod != http.MethodPost {
		t.Errorf("expected POST, got %s", gotMethod)
	}
	if gotBody["incident_id"] != "inc-1" {
		t.Errorf("expected incident_id inc-1, got %v", gotBody["incident_id"])
	}
	if gotBody["escalation_path_id"] != "01HRDKWWNGX330JQ4J1PATH001" {
		t.Errorf("expected escalation_path_id 01HRDKWWNGX330JQ4J1PATH001, got %v", gotBody["escalation_path_id"])
	}
}
