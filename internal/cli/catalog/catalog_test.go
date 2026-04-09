package catalog

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

func TestCatalogTypesList(t *testing.T) {
	var gotPath, gotMethod string

	shared.SetupMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		json.NewEncoder(w).Encode(map[string]any{
			"catalog_types": []api.CatalogType{
				{ID: "ct-1", Name: "Service"},
				{ID: "ct-2", Name: "Team"},
			},
		})
	})

	root := newTestRoot()
	root.SetArgs([]string{"catalog", "types", "list"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if gotPath != "/v2/catalog_types" {
		t.Errorf("expected path /v2/catalog_types, got %q", gotPath)
	}
	if gotMethod != http.MethodGet {
		t.Errorf("expected GET, got %s", gotMethod)
	}
}

func TestCatalogTypesGet(t *testing.T) {
	var gotPath, gotMethod string

	shared.SetupMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		json.NewEncoder(w).Encode(map[string]any{
			"catalog_type": api.CatalogType{ID: "ct-42", Name: "Service"},
		})
	})

	root := newTestRoot()
	root.SetArgs([]string{"catalog", "types", "get", "ct-42"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if gotPath != "/v2/catalog_types/ct-42" {
		t.Errorf("expected path /v2/catalog_types/ct-42, got %q", gotPath)
	}
	if gotMethod != http.MethodGet {
		t.Errorf("expected GET, got %s", gotMethod)
	}
}

func TestCatalogEntriesList(t *testing.T) {
	var gotPath, gotMethod string

	shared.SetupMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		json.NewEncoder(w).Encode(map[string]any{
			"catalog_entries": []api.CatalogEntry{
				{ID: "ce-1", Name: "API Gateway"},
				{ID: "ce-2", Name: "Auth Service"},
			},
		})
	})

	root := newTestRoot()
	root.SetArgs([]string{"catalog", "entries", "list"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if gotPath != "/v2/catalog_entries" {
		t.Errorf("expected path /v2/catalog_entries, got %q", gotPath)
	}
	if gotMethod != http.MethodGet {
		t.Errorf("expected GET, got %s", gotMethod)
	}
}

func TestCatalogEntriesGet(t *testing.T) {
	var gotPath, gotMethod string

	shared.SetupMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		json.NewEncoder(w).Encode(map[string]any{
			"catalog_entry": api.CatalogEntry{ID: "ce-42", Name: "Payment Service"},
		})
	})

	root := newTestRoot()
	root.SetArgs([]string{"catalog", "entries", "get", "ce-42"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if gotPath != "/v2/catalog_entries/ce-42" {
		t.Errorf("expected path /v2/catalog_entries/ce-42, got %q", gotPath)
	}
	if gotMethod != http.MethodGet {
		t.Errorf("expected GET, got %s", gotMethod)
	}
}
