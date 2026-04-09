package users

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/shhac/agent-incident/internal/api"
	"github.com/shhac/agent-incident/internal/cli/shared"
)

func TestUsersList(t *testing.T) {
	var gotPath, gotMethod string

	shared.SetupMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		json.NewEncoder(w).Encode(map[string]any{
			"users": []api.User{
				{ID: "u-1", Name: "Alice", Email: "alice@example.com"},
				{ID: "u-2", Name: "Bob", Email: "bob@example.com"},
			},
			"pagination_meta": map[string]any{},
		})
	})

	root := shared.NewTestRoot(Register)
	root.SetArgs([]string{"user", "list"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if gotPath != "/v2/users" {
		t.Errorf("expected path /v2/users, got %q", gotPath)
	}
	if gotMethod != http.MethodGet {
		t.Errorf("expected GET, got %s", gotMethod)
	}
}

func TestUsersGet(t *testing.T) {
	var gotPath, gotMethod string

	shared.SetupMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		json.NewEncoder(w).Encode(map[string]any{
			"user": api.User{ID: "u-42", Name: "Charlie", Email: "charlie@example.com"},
		})
	})

	root := shared.NewTestRoot(Register)
	root.SetArgs([]string{"user", "get", "u-42"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if gotPath != "/v2/users/u-42" {
		t.Errorf("expected path /v2/users/u-42, got %q", gotPath)
	}
	if gotMethod != http.MethodGet {
		t.Errorf("expected GET, got %s", gotMethod)
	}
}
