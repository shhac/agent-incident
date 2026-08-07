package config

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
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

// Concurrent StoreOrganization calls must not lose each other's entries.
//
// Before updateConfig routed through creds.Store.Update, StoreOrganization did
// Read() (from the shared in-memory cache) -> mutate -> Write(). Two concurrent
// CLI invocations — in-process sharing the package cache, or across processes
// sharing config.json — each built their write from a snapshot taken before the
// other's landed, so all but the last writer's organization were silently
// erased.
func TestConcurrentStoreOrganizationDoesNotLoseEntries(t *testing.T) {
	setupTestDir(t)

	const writers = 20
	var wg sync.WaitGroup
	for i := range writers {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			if err := StoreOrganization(fmt.Sprintf("org-%02d", i)); err != nil {
				t.Errorf("StoreOrganization: %v", err)
			}
		}(i)
	}
	wg.Wait()

	ClearCache()
	cfg := Read()
	if len(cfg.Organizations) != writers {
		t.Fatalf("%d of %d concurrent StoreOrganization calls survived — updates were lost", len(cfg.Organizations), writers)
	}
	for i := range writers {
		name := fmt.Sprintf("org-%02d", i)
		if _, ok := cfg.Organizations[name]; !ok {
			t.Errorf("%s was lost from config.json", name)
		}
	}
}

// config.json now goes through creds.Store (see updateConfig), which writes
// every file 0600 regardless of content — one audited place to get file
// permissions right rather than a per-file policy. That's a tightening from the
// previous 0644 default, not a regression: this directory is shared with
// credentials.json, which holds a real API key whenever the keychain is
// unavailable.
func TestConfigFilePerms(t *testing.T) {
	setupTestDir(t)

	if err := Write(defaultConfig()); err != nil {
		t.Fatalf("Write: %v", err)
	}
	info, err := os.Stat(filepath.Join(ConfigDir(), "config.json"))
	if err != nil {
		t.Fatalf("config file not written: %v", err)
	}
	if perm := info.Mode().Perm(); perm != 0o600 {
		t.Errorf("config file perms = %o, want 600", perm)
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
