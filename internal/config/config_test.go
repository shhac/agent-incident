package config

import (
	"os"
	"path/filepath"
	"testing"
)

func setupTestDir(t *testing.T) {
	t.Helper()
	dir := t.TempDir()
	SetConfigDir(dir)
	t.Cleanup(func() { SetConfigDir("") })
}

func TestReadEmptyDir(t *testing.T) {
	setupTestDir(t)
	cfg := Read()
	if cfg == nil {
		t.Fatal("expected non-nil config")
	}
	if cfg.DefaultOrg != "" {
		t.Fatalf("expected empty default org, got %q", cfg.DefaultOrg)
	}
	if len(cfg.Organizations) != 0 {
		t.Fatalf("expected 0 organizations, got %d", len(cfg.Organizations))
	}
}

func TestWriteReadRoundtrip(t *testing.T) {
	setupTestDir(t)
	cfg := &Config{
		DefaultOrg: "acme",
		Organizations: map[string]Organization{
			"acme": {},
			"beta": {},
		},
	}
	if err := Write(cfg); err != nil {
		t.Fatalf("write failed: %v", err)
	}

	ClearCache()
	got := Read()
	if got.DefaultOrg != "acme" {
		t.Fatalf("expected default org %q, got %q", "acme", got.DefaultOrg)
	}
	if len(got.Organizations) != 2 {
		t.Fatalf("expected 2 orgs, got %d", len(got.Organizations))
	}
}

func TestStoreOrganization(t *testing.T) {
	setupTestDir(t)

	if err := StoreOrganization("first"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	ClearCache()
	cfg := Read()
	if cfg.DefaultOrg != "first" {
		t.Fatalf("first org should become default, got %q", cfg.DefaultOrg)
	}

	if err := StoreOrganization("second"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	ClearCache()
	cfg = Read()
	if cfg.DefaultOrg != "first" {
		t.Fatalf("default should remain %q, got %q", "first", cfg.DefaultOrg)
	}
	if len(cfg.Organizations) != 2 {
		t.Fatalf("expected 2 orgs, got %d", len(cfg.Organizations))
	}
}

func TestRemoveOrganization(t *testing.T) {
	setupTestDir(t)

	StoreOrganization("alpha")
	StoreOrganization("beta")
	ClearCache()

	if err := RemoveOrganization("alpha"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	ClearCache()
	cfg := Read()
	if _, ok := cfg.Organizations["alpha"]; ok {
		t.Fatal("alpha should be removed")
	}
	// default should shift to beta (the only remaining org)
	if cfg.DefaultOrg != "beta" {
		t.Fatalf("expected default to shift to beta, got %q", cfg.DefaultOrg)
	}
}

func TestRemoveOrganizationLastOrg(t *testing.T) {
	setupTestDir(t)

	StoreOrganization("only")
	ClearCache()

	if err := RemoveOrganization("only"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	ClearCache()
	cfg := Read()
	if cfg.DefaultOrg != "" {
		t.Fatalf("expected empty default, got %q", cfg.DefaultOrg)
	}
	if len(cfg.Organizations) != 0 {
		t.Fatalf("expected 0 orgs, got %d", len(cfg.Organizations))
	}
}

func TestSetDefault(t *testing.T) {
	setupTestDir(t)

	StoreOrganization("a")
	StoreOrganization("b")
	ClearCache()

	if err := SetDefault("b"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	ClearCache()
	cfg := Read()
	if cfg.DefaultOrg != "b" {
		t.Fatalf("expected default %q, got %q", "b", cfg.DefaultOrg)
	}
}

func TestSetDefaultNonExistent(t *testing.T) {
	setupTestDir(t)

	StoreOrganization("a")
	ClearCache()

	if err := SetDefault("nonexistent"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	ClearCache()
	cfg := Read()
	if cfg.DefaultOrg != "a" {
		t.Fatalf("default should remain %q, got %q", "a", cfg.DefaultOrg)
	}
}

func TestClearCache(t *testing.T) {
	setupTestDir(t)

	_ = Read() // populate cache
	ClearCache()

	// Write a config file directly
	dir := ConfigDir()
	data := []byte(`{"default_org":"direct","organizations":{"direct":{}}}`)
	os.WriteFile(filepath.Join(dir, "config.json"), data, 0o644)

	cfg := Read()
	if cfg.DefaultOrg != "direct" {
		t.Fatalf("after cache clear, expected to read fresh config, got default %q", cfg.DefaultOrg)
	}
}
