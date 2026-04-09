package actions

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/shhac/agent-incident/internal/api"
	"github.com/shhac/agent-incident/internal/cli/shared"
)

func TestActionsList(t *testing.T) {
	var gotPath, gotMethod string

	shared.SetupMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		json.NewEncoder(w).Encode(map[string]any{
			"actions": []api.Action{
				{ID: "act-1", Status: "outstanding"},
				{ID: "act-2", Status: "completed"},
			},
		})
	})

	root := shared.NewTestRoot(Register)
	root.SetArgs([]string{"action", "list"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if gotPath != "/v2/actions" {
		t.Errorf("expected path /v2/actions, got %q", gotPath)
	}
	if gotMethod != http.MethodGet {
		t.Errorf("expected GET, got %s", gotMethod)
	}
}

func TestActionsGet(t *testing.T) {
	var gotPath, gotMethod string

	shared.SetupMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		json.NewEncoder(w).Encode(map[string]any{
			"action": api.Action{ID: "act-42", Status: "outstanding"},
		})
	})

	root := shared.NewTestRoot(Register)
	root.SetArgs([]string{"action", "get", "act-42"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if gotPath != "/v2/actions/act-42" {
		t.Errorf("expected path /v2/actions/act-42, got %q", gotPath)
	}
	if gotMethod != http.MethodGet {
		t.Errorf("expected GET, got %s", gotMethod)
	}
}
