package shared

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/shhac/agent-incident/internal/config"
	"github.com/shhac/agent-incident/internal/credential"
	agenterrors "github.com/shhac/agent-incident/internal/errors"
)

func setupIsolatedConfig(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	config.SetConfigDir(dir)
	t.Cleanup(func() { config.SetConfigDir("") })
	return dir
}

func writeCredentials(t *testing.T, dir string, creds map[string]credential.Credential) {
	t.Helper()
	data, err := json.MarshalIndent(creds, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "credentials.json"), data, 0o600); err != nil {
		t.Fatal(err)
	}
}

func TestNewClientFromFlags_APIURLEnv(t *testing.T) {
	dir := setupIsolatedConfig(t)
	_ = dir

	t.Setenv("INCIDENT_API_URL", "http://localhost:9999")
	t.Setenv("INCIDENT_API_KEY", "env-key-123")

	client, err := NewClientFromFlags("", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if client == nil {
		t.Fatal("expected non-nil client")
	}
}

func TestNewClientFromFlags_APIURLEnvWithFlag(t *testing.T) {
	setupIsolatedConfig(t)

	t.Setenv("INCIDENT_API_URL", "http://localhost:9999")
	t.Setenv("INCIDENT_API_KEY", "")

	client, err := NewClientFromFlags("flag-key", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if client == nil {
		t.Fatal("expected non-nil client")
	}
}

func TestNewClientFromFlags_APIKeyFlag(t *testing.T) {
	setupIsolatedConfig(t)

	client, err := NewClientFromFlags("my-api-key", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if client == nil {
		t.Fatal("expected non-nil client")
	}
}

func TestNewClientFromFlags_EnvKey(t *testing.T) {
	setupIsolatedConfig(t)
	t.Setenv("INCIDENT_API_KEY", "env-key-456")

	client, err := NewClientFromFlags("", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if client == nil {
		t.Fatal("expected non-nil client")
	}
}

func TestNewClientFromFlags_CredentialStore(t *testing.T) {
	dir := setupIsolatedConfig(t)

	// Write a config with a default org
	cfg := &config.Config{
		DefaultOrg:    "myorg",
		Organizations: map[string]config.Organization{"myorg": {}},
	}
	if err := config.Write(cfg); err != nil {
		t.Fatal(err)
	}

	// Write credentials directly (bypasses keychain)
	creds := map[string]credential.Credential{
		"myorg": {APIKey: "stored-key-789"},
	}
	writeCredentials(t, dir, creds)

	client, err := NewClientFromFlags("", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if client == nil {
		t.Fatal("expected non-nil client")
	}
}

func TestNewClientFromFlags_CredentialNotFound(t *testing.T) {
	setupIsolatedConfig(t)

	// Config has a default org but no credentials file
	cfg := &config.Config{
		DefaultOrg:    "missing-org",
		Organizations: map[string]config.Organization{"missing-org": {}},
	}
	if err := config.Write(cfg); err != nil {
		t.Fatal(err)
	}

	_, err := NewClientFromFlags("", "")
	if err == nil {
		t.Fatal("expected error for missing credentials")
	}
	var apiErr *agenterrors.APIError
	if !agenterrors.As(err, &apiErr) {
		t.Fatalf("expected APIError, got %T: %v", err, err)
	}
}

func TestNewClientFromFlags_EmptyAPIKey(t *testing.T) {
	dir := setupIsolatedConfig(t)

	cfg := &config.Config{
		DefaultOrg:    "myorg",
		Organizations: map[string]config.Organization{"myorg": {}},
	}
	if err := config.Write(cfg); err != nil {
		t.Fatal(err)
	}

	// Credential exists but has empty API key
	creds := map[string]credential.Credential{
		"myorg": {APIKey: ""},
	}
	writeCredentials(t, dir, creds)

	_, err := NewClientFromFlags("", "")
	if err == nil {
		t.Fatal("expected error for empty API key")
	}
	var apiErr *agenterrors.APIError
	if !agenterrors.As(err, &apiErr) {
		t.Fatalf("expected APIError, got %T: %v", err, err)
	}
}

func TestResolveOrg_ExplicitAlias(t *testing.T) {
	setupIsolatedConfig(t)

	org, err := ResolveOrg("explicit-org")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if org != "explicit-org" {
		t.Errorf("expected explicit-org, got %q", org)
	}
}

func TestResolveOrg_EnvShortcircuit(t *testing.T) {
	setupIsolatedConfig(t)
	t.Setenv("INCIDENT_API_KEY", "some-key")

	org, err := ResolveOrg("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if org != "" {
		t.Errorf("expected empty string for env shortcircuit, got %q", org)
	}
}

func TestResolveOrg_ConfigDefault(t *testing.T) {
	setupIsolatedConfig(t)

	cfg := &config.Config{
		DefaultOrg:    "default-org",
		Organizations: map[string]config.Organization{"default-org": {}},
	}
	if err := config.Write(cfg); err != nil {
		t.Fatal(err)
	}

	org, err := ResolveOrg("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if org != "default-org" {
		t.Errorf("expected default-org, got %q", org)
	}
}

func TestResolveOrg_ErrorWithHint(t *testing.T) {
	setupIsolatedConfig(t)

	// No default org, no env key, no orgs configured
	_, err := ResolveOrg("")
	if err == nil {
		t.Fatal("expected error when no org configured")
	}
	var apiErr *agenterrors.APIError
	if !agenterrors.As(err, &apiErr) {
		t.Fatalf("expected APIError, got %T: %v", err, err)
	}
	if apiErr.Hint == "" {
		t.Error("expected non-empty hint")
	}
}

func TestResolveOrg_ErrorWithAvailableOrgs(t *testing.T) {
	setupIsolatedConfig(t)

	// Orgs exist but no default set
	cfg := &config.Config{
		Organizations: map[string]config.Organization{
			"org-a": {},
		},
	}
	if err := config.Write(cfg); err != nil {
		t.Fatal(err)
	}

	_, err := ResolveOrg("")
	if err == nil {
		t.Fatal("expected error when no default org set")
	}
	var apiErr *agenterrors.APIError
	if !agenterrors.As(err, &apiErr) {
		t.Fatalf("expected APIError, got %T: %v", err, err)
	}
	if apiErr.Hint == "" {
		t.Error("expected hint listing available orgs")
	}
}
