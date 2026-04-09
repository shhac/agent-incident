package roles

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

func TestRolesList(t *testing.T) {
	var gotPath, gotMethod string

	shared.SetupMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		w.Write(testdata.Load("roles_list.json"))
	})

	root := newTestRoot()
	root.SetArgs([]string{"role", "list"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if gotPath != "/v2/incident_roles" {
		t.Errorf("expected path /v2/incident_roles, got %q", gotPath)
	}
	if gotMethod != http.MethodGet {
		t.Errorf("expected GET, got %s", gotMethod)
	}
}

func TestRolesGet(t *testing.T) {
	var gotPath, gotMethod string

	shared.SetupMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		json.NewEncoder(w).Encode(map[string]any{
			"incident_role": api.IncidentRoleFull{ID: "role-42", Name: "Scribe"},
		})
	})

	root := newTestRoot()
	root.SetArgs([]string{"role", "get", "role-42"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if gotPath != "/v2/incident_roles/role-42" {
		t.Errorf("expected path /v2/incident_roles/role-42, got %q", gotPath)
	}
	if gotMethod != http.MethodGet {
		t.Errorf("expected GET, got %s", gotMethod)
	}
}
