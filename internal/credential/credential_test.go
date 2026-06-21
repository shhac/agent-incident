package credential

import (
	"os"
	"path/filepath"
	"sort"
	"testing"

	"github.com/shhac/agent-incident/internal/config"
)

func setupTestDir(t *testing.T) {
	t.Helper()
	dir := t.TempDir()
	config.SetConfigDir(dir)
	t.Cleanup(func() { config.SetConfigDir("") })
}

func TestStoreAndGet(t *testing.T) {
	setupTestDir(t)

	cred := Credential{APIKey: "test-api-key-123"}
	storage, err := Store("myorg", cred)
	if err != nil {
		t.Fatalf("Store failed: %v", err)
	}
	// storage is either "file" or "keychain" depending on platform
	if storage != "file" && storage != "keychain" {
		t.Fatalf("unexpected storage type: %q", storage)
	}

	got, err := Get("myorg")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if got.APIKey != "test-api-key-123" {
		// On macOS with keychain, the sentinel is stored in file but Get retrieves from keychain
		// If keychain retrieval fails, we get the sentinel; that's acceptable in CI
		if got.APIKey == keychainSentinel {
			t.Skip("keychain sentinel returned — keychain access may be blocked in test env")
		}
		t.Fatalf("expected API key %q, got %q", "test-api-key-123", got.APIKey)
	}
}

// TestStore_Headless_FileFallback exercises the real credential-WRITE path
// non-interactively. Setting the per-CLI keychain opt-out (derived by
// lib-agent-cli from the "app.paulie.agent-incident" service) makes the keychain
// backend report unavailable, so Store deterministically takes the 0600 file
// fallback on every platform — including darwin, where it would otherwise reach
// the `security` CLI and its GUI prompt. The per-CLI env var also proves the
// lib's prefix derivation.
func TestStore_Headless_FileFallback(t *testing.T) {
	t.Setenv("AGENT_INCIDENT_NO_KEYCHAIN", "1")
	dir := t.TempDir()
	config.SetConfigDir(dir)
	t.Cleanup(func() { config.SetConfigDir("") })

	storage, err := Store("headless", Credential{APIKey: "api-headless"})
	if err != nil {
		t.Fatalf("Store: %v", err)
	}
	if storage != "file" {
		t.Fatalf("storage = %q, want \"file\" (keychain opt-out should force the file path)", storage)
	}

	credsPath := filepath.Join(dir, "credentials.json")
	info, err := os.Stat(credsPath)
	if err != nil {
		t.Fatalf("credentials file not written: %v", err)
	}
	if mode := info.Mode().Perm(); mode != 0o600 {
		t.Errorf("credentials mode = %o, want 0600", mode)
	}

	got, err := Get("headless")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if got.KeychainManaged {
		t.Error("KeychainManaged = true, want false (keychain should not have been used)")
	}
	if got.APIKey != "api-headless" {
		t.Errorf("APIKey = %q, want %q (file fallback stores the key directly, not the sentinel)", got.APIKey, "api-headless")
	}

	if err := Remove("headless"); err != nil {
		t.Fatalf("Remove: %v", err)
	}
	_, err = Get("headless")
	var notFound *NotFoundError
	if !isNotFoundError(err, &notFound) {
		t.Fatalf("after Remove, Get should return *NotFoundError, got %T: %v", err, err)
	}
}

func TestGetMissing(t *testing.T) {
	setupTestDir(t)

	_, err := Get("nonexistent")
	if err == nil {
		t.Fatal("expected error for missing credential")
	}
	var notFound *NotFoundError
	if !isNotFoundError(err, &notFound) {
		t.Fatalf("expected NotFoundError, got %T: %v", err, err)
	}
	if notFound.Name != "nonexistent" {
		t.Fatalf("expected name %q, got %q", "nonexistent", notFound.Name)
	}
}

func isNotFoundError(err error, target **NotFoundError) bool {
	nfe, ok := err.(*NotFoundError)
	if ok {
		*target = nfe
	}
	return ok
}

func TestRemove(t *testing.T) {
	setupTestDir(t)

	Store("toremove", Credential{APIKey: "key"})

	if err := Remove("toremove"); err != nil {
		t.Fatalf("Remove failed: %v", err)
	}

	_, err := Get("toremove")
	if err == nil {
		t.Fatal("expected error after removal")
	}
	var notFound *NotFoundError
	if !isNotFoundError(err, &notFound) {
		t.Fatalf("expected NotFoundError, got %T", err)
	}
}

func TestRemoveMissing(t *testing.T) {
	setupTestDir(t)

	err := Remove("ghost")
	if err == nil {
		t.Fatal("expected error removing nonexistent credential")
	}
	var notFound *NotFoundError
	if !isNotFoundError(err, &notFound) {
		t.Fatalf("expected NotFoundError, got %T", err)
	}
}

func TestList(t *testing.T) {
	setupTestDir(t)

	Store("org-a", Credential{APIKey: "a"})
	Store("org-b", Credential{APIKey: "b"})
	Store("org-c", Credential{APIKey: "c"})

	names, err := List()
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	sort.Strings(names)
	if len(names) != 3 {
		t.Fatalf("expected 3 names, got %d", len(names))
	}
	expected := []string{"org-a", "org-b", "org-c"}
	for i, name := range names {
		if name != expected[i] {
			t.Fatalf("expected %q at index %d, got %q", expected[i], i, name)
		}
	}
}

func TestListEmpty(t *testing.T) {
	setupTestDir(t)

	names, err := List()
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(names) != 0 {
		t.Fatalf("expected 0 names, got %d", len(names))
	}
}

func TestStoreOverwrite(t *testing.T) {
	setupTestDir(t)

	Store("myorg", Credential{APIKey: "old-key"})
	Store("myorg", Credential{APIKey: "new-key"})

	got, err := Get("myorg")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if got.APIKey != "new-key" {
		if got.APIKey == keychainSentinel {
			t.Skip("keychain sentinel — skipping overwrite verification")
		}
		t.Fatalf("expected %q, got %q", "new-key", got.APIKey)
	}
}
