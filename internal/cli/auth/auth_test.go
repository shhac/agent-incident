package auth

import (
	"bytes"
	"net/http"
	"strings"
	"testing"

	"github.com/spf13/cobra"

	"github.com/shhac/agent-incident/internal/api/testdata"
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

// TestAuthAddStdinFills verifies that a key piped on stdin is used when no
// --api-key flag is given. The keychain opt-out forces the deterministic file
// path so the stored value can be read back and compared.
func TestAuthAddStdinFills(t *testing.T) {
	t.Setenv("AGENT_INCIDENT_NO_KEYCHAIN", "1")
	config.SetConfigDir(t.TempDir())

	root := newTestRoot()
	root.SetOut(&bytes.Buffer{})
	root.SetIn(strings.NewReader("stdin-key-123\n"))
	root.SetArgs([]string{"auth", "add", "myorg"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	cred, err := credential.Get("myorg")
	if err != nil {
		t.Fatalf("credential not found after add: %v", err)
	}
	if cred.APIKey != "stdin-key-123" {
		t.Errorf("expected api key from stdin, got %q", cred.APIKey)
	}
}

// TestAuthAddFlagWinsOverStdin verifies precedence: an explicit --api-key flag
// takes priority over a value piped on stdin.
func TestAuthAddFlagWinsOverStdin(t *testing.T) {
	t.Setenv("AGENT_INCIDENT_NO_KEYCHAIN", "1")
	config.SetConfigDir(t.TempDir())

	root := newTestRoot()
	root.SetOut(&bytes.Buffer{})
	root.SetIn(strings.NewReader("stdin-key\n"))
	root.SetArgs([]string{"auth", "add", "myorg", "--api-key", "flag-key"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	cred, err := credential.Get("myorg")
	if err != nil {
		t.Fatalf("credential not found after add: %v", err)
	}
	if cred.APIKey != "flag-key" {
		t.Errorf("expected flag to win over stdin, got %q", cred.APIKey)
	}
}

// TestAuthAddNoKeyErrors verifies that with no flag, no piped stdin, and no
// --form, the command surfaces the RequireFlag error.
func TestAuthAddNoKeyErrors(t *testing.T) {
	config.SetConfigDir(t.TempDir())

	root := newTestRoot()
	root.SetOut(&bytes.Buffer{})
	root.SetIn(&bytes.Buffer{})
	root.SetArgs([]string{"auth", "add", "myorg"})

	err := root.Execute()
	if err == nil {
		t.Fatal("expected an error when no api key is supplied")
	}
	if !strings.Contains(err.Error(), "--api-key is required") {
		t.Errorf("expected --api-key required error, got %v", err)
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
		w.Write(testdata.Load("identity.json"))
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
