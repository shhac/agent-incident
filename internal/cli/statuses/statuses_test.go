package statuses

import (
	"encoding/json"
	"net/http"
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

func TestStatusesList(t *testing.T) {
	var gotPath, gotMethod string

	shared.SetupMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		w.Write(testdata.Load("statuses_list.json"))
	})

	root := newTestRoot()
	root.SetArgs([]string{"status", "list"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if gotPath != "/v1/incident_statuses" {
		t.Errorf("expected path /v1/incident_statuses, got %q", gotPath)
	}
	if gotMethod != http.MethodGet {
		t.Errorf("expected GET, got %s", gotMethod)
	}
}

func TestStatusesGet(t *testing.T) {
	var gotPath, gotMethod string

	shared.SetupMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		json.NewEncoder(w).Encode(map[string]any{
			"incident_status": api.IncidentStatusResource{
				ID:       "st-99",
				Name:     "Monitoring",
				Category: "active",
				Rank:     3,
			},
		})
	})

	root := newTestRoot()
	root.SetArgs([]string{"status", "get", "st-99"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if gotPath != "/v1/incident_statuses/st-99" {
		t.Errorf("expected path /v1/incident_statuses/st-99, got %q", gotPath)
	}
	if gotMethod != http.MethodGet {
		t.Errorf("expected GET, got %s", gotMethod)
	}
}
