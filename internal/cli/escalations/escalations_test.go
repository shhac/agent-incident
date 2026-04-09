package escalations

import (
	"encoding/json"
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

	root := newTestRoot()
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

	root := newTestRoot()
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
				{ID: "path-1", Name: "Primary Path"},
				{ID: "path-2", Name: "Secondary Path"},
			},
		})
	})

	root := newTestRoot()
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
			"escalation_path": api.EscalationPath{ID: "path-42", Name: "Critical Path"},
		})
	})

	root := newTestRoot()
	root.SetArgs([]string{"escalation", "path", "get", "path-42"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if gotPath != "/v2/escalation_paths/path-42" {
		t.Errorf("expected path /v2/escalation_paths/path-42, got %q", gotPath)
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

	root := newTestRoot()
	root.SetArgs([]string{
		"escalation", "create",
		"--incident", "inc-1",
		"--path", "path-1",
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
	if gotBody["escalation_path_id"] != "path-1" {
		t.Errorf("expected escalation_path_id path-1, got %v", gotBody["escalation_path_id"])
	}
}
