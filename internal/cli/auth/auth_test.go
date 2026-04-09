package auth

import (
	"bytes"
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"github.com/spf13/cobra"

	"github.com/shhac/agent-incident/internal/api"
	"github.com/shhac/agent-incident/internal/cli/shared"
	"github.com/shhac/agent-incident/internal/config"
	"github.com/shhac/agent-incident/internal/credential"
)

func newTestRoot() *cobra.Command {
	root := &cobra.Command{
		Use:           "agent-incident",
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	Register(root)
	return root
}

func TestAuthAdd(t *testing.T) {
	config.SetConfigDir(t.TempDir())

	root := newTestRoot()
	buf := &bytes.Buffer{}
	root.SetOut(buf)
	root.SetArgs([]string{"auth", "add", "myorg", "--api-key", "test-key-123"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify credential was stored
	cred, err := credential.Get("myorg")
	if err != nil {
		t.Fatalf("credential not found after add: %v", err)
	}
	if cred.APIKey != "test-key-123" && !cred.KeychainManaged {
		t.Errorf("expected api key test-key-123 or keychain managed, got key=%q keychain=%v", cred.APIKey, cred.KeychainManaged)
	}

	// Verify config was updated
	cfg := config.Read()
	if _, ok := cfg.Organizations["myorg"]; !ok {
		t.Error("organization not found in config after add")
	}
	if cfg.DefaultOrg != "myorg" {
		t.Errorf("expected default org 'myorg', got %q", cfg.DefaultOrg)
	}
}

func TestAuthCheck(t *testing.T) {
	tmpDir := t.TempDir()
	config.SetConfigDir(tmpDir)

	// Store a credential so auth check can resolve it
	_, _ = credential.Store("testorg", credential.Credential{APIKey: "test-key"})
	_ = config.StoreOrganization("testorg")

	var gotPath string
	shared.SetupMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		json.NewEncoder(w).Encode(map[string]any{
			"identity": api.Identity{
				Name:  "Test Org",
				Roles: []string{"owner"},
			},
		})
	})

	root := newTestRoot()
	root.SetArgs([]string{"auth", "check"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if gotPath != "/v1/identity" {
		t.Errorf("expected path /v1/identity, got %q", gotPath)
	}
}

func TestAuthList(t *testing.T) {
	tmpDir := t.TempDir()
	config.SetConfigDir(tmpDir)

	// Add two orgs to config
	_ = config.StoreOrganization("org-a")
	_ = config.StoreOrganization("org-b")

	root := newTestRoot()
	buf := &bytes.Buffer{}
	root.SetOut(buf)
	root.SetArgs([]string{"auth", "list"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	cfg := config.Read()
	if len(cfg.Organizations) != 2 {
		t.Errorf("expected 2 organizations, got %d", len(cfg.Organizations))
	}
	if _, ok := cfg.Organizations["org-a"]; !ok {
		t.Error("org-a not found")
	}
	if _, ok := cfg.Organizations["org-b"]; !ok {
		t.Error("org-b not found")
	}
}

func TestAuthRemove(t *testing.T) {
	tmpDir := t.TempDir()
	config.SetConfigDir(tmpDir)

	// First store a credential and org
	_, err := credential.Store("removeme", credential.Credential{APIKey: "key-to-remove"})
	if err != nil {
		t.Fatalf("setup store failed: %v", err)
	}
	_ = config.StoreOrganization("removeme")

	root := newTestRoot()
	root.SetArgs([]string{"auth", "remove", "removeme"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify credential is gone
	_, err = credential.Get("removeme")
	if err == nil {
		t.Error("expected credential to be removed")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("unexpected error: %v", err)
	}

	// Verify org removed from config
	cfg := config.Read()
	if _, ok := cfg.Organizations["removeme"]; ok {
		t.Error("organization still present in config after remove")
	}
}
