package statuspages

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

func TestStatusPagesList(t *testing.T) {
	var gotPath, gotMethod string

	shared.SetupMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		json.NewEncoder(w).Encode(map[string]any{
			"status_pages": []api.StatusPage{
				{ID: "sp-1", Name: "Public Status"},
				{ID: "sp-2", Name: "Internal Status"},
			},
		})
	})

	root := newTestRoot()
	root.SetArgs([]string{"status-page", "list"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if gotPath != "/v2/status_pages" {
		t.Errorf("expected path /v2/status_pages, got %q", gotPath)
	}
	if gotMethod != http.MethodGet {
		t.Errorf("expected GET, got %s", gotMethod)
	}
}

func TestStatusPagesIncidentsList(t *testing.T) {
	var gotPath, gotMethod string

	shared.SetupMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		json.NewEncoder(w).Encode(map[string]any{
			"status_page_incidents": []api.StatusPageIncident{
				{ID: "spi-1", Name: "Database Outage", Status: "investigating"},
			},
		})
	})

	root := newTestRoot()
	root.SetArgs([]string{"status-page", "update", "list"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if gotPath != "/v2/status_page_incidents" {
		t.Errorf("expected path /v2/status_page_incidents, got %q", gotPath)
	}
	if gotMethod != http.MethodGet {
		t.Errorf("expected GET, got %s", gotMethod)
	}
}

func TestStatusPagesIncidentsCreate(t *testing.T) {
	var gotPath, gotMethod string
	var gotBody map[string]any

	shared.SetupMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		if r.Method == http.MethodPost {
			json.NewDecoder(r.Body).Decode(&gotBody)
		}
		json.NewEncoder(w).Encode(map[string]any{
			"status_page_incident": api.StatusPageIncident{
				ID:     "spi-new",
				Name:   "API Degradation",
				Status: "investigating",
			},
		})
	})

	root := newTestRoot()
	root.SetArgs([]string{
		"status-page", "update", "create",
		"--page", "sp-1",
		"--name", "API Degradation",
	})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if gotPath != "/v2/status_page_incidents" {
		t.Errorf("expected path /v2/status_page_incidents, got %q", gotPath)
	}
	if gotMethod != http.MethodPost {
		t.Errorf("expected POST, got %s", gotMethod)
	}
	if gotBody["status_page_id"] != "sp-1" {
		t.Errorf("expected status_page_id sp-1, got %v", gotBody["status_page_id"])
	}
	if gotBody["name"] != "API Degradation" {
		t.Errorf("expected name 'API Degradation', got %v", gotBody["name"])
	}
}

func TestStatusPagesIncidentsUpdate(t *testing.T) {
	var gotPath, gotMethod string
	var gotBody map[string]any

	shared.SetupMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		if r.Method == http.MethodPut {
			json.NewDecoder(r.Body).Decode(&gotBody)
		}
		json.NewEncoder(w).Encode(map[string]any{
			"status_page_incident": api.StatusPageIncident{
				ID:     "spi-1",
				Name:   "API Degradation",
				Status: "resolved",
			},
		})
	})

	root := newTestRoot()
	root.SetArgs([]string{
		"status-page", "update", "update", "spi-1",
		"--status", "resolved",
	})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if gotPath != "/v2/status_page_incidents/spi-1" {
		t.Errorf("expected path /v2/status_page_incidents/spi-1, got %q", gotPath)
	}
	if gotMethod != http.MethodPut {
		t.Errorf("expected PUT, got %s", gotMethod)
	}
	if gotBody["status"] != "resolved" {
		t.Errorf("expected status resolved, got %v", gotBody["status"])
	}
}
