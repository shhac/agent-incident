package credential

import (
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
