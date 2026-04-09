package severities

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

func TestSeveritiesList(t *testing.T) {
	var gotPath, gotMethod string

	shared.SetupMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		w.Write(testdata.Load("severities_list.json"))
	})

	root := newTestRoot()
	root.SetArgs([]string{"severity", "list"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if gotPath != "/v1/severities" {
		t.Errorf("expected path /v1/severities, got %q", gotPath)
	}
	if gotMethod != http.MethodGet {
		t.Errorf("expected GET, got %s", gotMethod)
	}
}

func TestSeveritiesGet(t *testing.T) {
	var gotPath, gotMethod string

	shared.SetupMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		json.NewEncoder(w).Encode(map[string]any{
			"severity": api.Severity{
				ID:   "sev-42",
				Name: "Minor",
				Rank: 3,
			},
		})
	})

	root := newTestRoot()
	root.SetArgs([]string{"severity", "get", "sev-42"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if gotPath != "/v1/severities/sev-42" {
		t.Errorf("expected path /v1/severities/sev-42, got %q", gotPath)
	}
	if gotMethod != http.MethodGet {
		t.Errorf("expected GET, got %s", gotMethod)
	}
}
