package credential

import (
	"fmt"
	"path/filepath"

	"github.com/shhac/agent-incident/internal/config"
	"github.com/shhac/lib-agent-cli/creds"
)

const keychainSentinel = "__KEYCHAIN__"

// keychainService is the reverse-domain service name under which this CLI's
// credentials live in the macOS keychain. The library is service-agnostic; the
// CLI owns this identifier.
const keychainService = "app.paulie.agent-incident"

// MCPKeychainService is the Keychain service for the MCP server's local-OAuth
// secrets — the CLI's service plus a ".mcp" namespace, separate from the API creds.
func MCPKeychainService() string { return keychainService + ".mcp" }

var keychain = creds.NewKeychain(keychainService)

type Credential struct {
	APIKey          string `json:"api_key"`
	KeychainManaged bool   `json:"keychain_managed,omitempty"`
}

type credentialEntry struct {
	APIKey          string `json:"api_key"`
	KeychainManaged bool   `json:"keychain_managed,omitempty"`
}

type NotFoundError struct {
	Name string
}

func (e *NotFoundError) Error() string {
	return fmt.Sprintf("organization credential %q not found", e.Name)
}

func credentialsPath() string {
	return filepath.Join(config.ConfigDir(), "credentials.json")
}

// store is the credential index's file: 0600 writes into a 0700 parent, atomic
// replacement, and Update for a locked read-modify-write. This used to be
// hand-rolled with os.ReadFile/os.WriteFile, which carried a lost-update race —
// two concurrent writers could each build their write from a stale snapshot,
// and the loser's entry vanished while its secret stayed in the keychain,
// unreferenced and un-removable (auth list can't show it, auth remove can't
// look it up).
func store() creds.Store {
	return creds.Store{Path: credentialsPath()}
}

func readIndex() (map[string]credentialEntry, error) {
	index := make(map[string]credentialEntry)
	if err := store().Load(&index); err != nil {
		return nil, err
	}
	if index == nil {
		index = make(map[string]credentialEntry)
	}
	return index, nil
}

// updateIndex applies mutate to the index under an exclusive lock, so two
// concurrent `auth add`/`auth remove` invocations serialize instead of
// clobbering each other.
func updateIndex(mutate func(index map[string]credentialEntry) error) error {
	index := make(map[string]credentialEntry)
	return store().Update(&index, func() error {
		if index == nil {
			index = make(map[string]credentialEntry)
		}
		return mutate(index)
	})
}

func Store(name string, cred Credential) (string, error) {
	storage := "file"
	entry := credentialEntry{
		APIKey: cred.APIKey,
	}

	if err := keychain.Set(name, cred.APIKey); err == nil {
		entry.APIKey = keychainSentinel
		entry.KeychainManaged = true
		storage = "keychain"
	}

	// The index write is the step that must not race: the keychain already
	// holds the secret by now, so an entry lost to a concurrent writer leaves
	// that secret referenced by nothing.
	if err := updateIndex(func(index map[string]credentialEntry) error {
		index[name] = entry
		return nil
	}); err != nil {
		return "", err
	}
	return storage, nil
}

func Get(name string) (*Credential, error) {
	index, err := readIndex()
	if err != nil {
		return nil, err
	}
	entry, ok := index[name]
	if !ok {
		return nil, &NotFoundError{Name: name}
	}

	cred := &Credential{
		APIKey:          entry.APIKey,
		KeychainManaged: entry.KeychainManaged,
	}

	if entry.KeychainManaged {
		if apiKey, ok := keychain.Get(name); ok {
			cred.APIKey = apiKey
		}
	}

	return cred, nil
}

func Remove(name string) error {
	return updateIndex(func(index map[string]credentialEntry) error {
		entry, ok := index[name]
		if !ok {
			return &NotFoundError{Name: name}
		}

		if entry.KeychainManaged {
			_ = keychain.Delete(name)
		}

		delete(index, name)
		return nil
	})
}

func List() ([]string, error) {
	index, err := readIndex()
	if err != nil {
		return nil, err
	}
	names := make([]string, 0, len(index))
	for name := range index {
		names = append(names, name)
	}
	return names, nil
}
