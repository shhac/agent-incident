package config

import (
	"errors"
	"path/filepath"
	"sync"

	"github.com/shhac/lib-agent-cli/creds"
	"github.com/shhac/lib-agent-cli/xdg"
)

type Config struct {
	DefaultOrg    string                  `json:"default_org,omitempty"`
	Organizations map[string]Organization `json:"organizations"`
}

type Organization struct {
	// incident.io uses a single API host, so no Site field needed (unlike agent-dd).
}

var (
	cache       *Config
	cacheMu     sync.Mutex
	overrideDir string
)

func SetConfigDir(dir string) {
	cacheMu.Lock()
	defer cacheMu.Unlock()
	overrideDir = dir
	cache = nil
}

func ConfigDir() string {
	if overrideDir != "" {
		return overrideDir
	}
	return xdg.ConfigDir("agent-incident")
}

func configPath() string {
	return filepath.Join(ConfigDir(), "config.json")
}

// store is config.json's file: 0600 writes into a 0700 parent, atomic
// replacement, and Update for a locked read-modify-write. This used to be
// hand-rolled with os.ReadFile/os.WriteFile, which carried a lost-update race —
// two concurrent CLI invocations (e.g. `auth add` racing `auth default`) could
// each build their write from a snapshot taken before the other landed,
// silently erasing one of them.
func store() creds.Store {
	return creds.Store{Path: configPath()}
}

func Read() *Config {
	cacheMu.Lock()
	defer cacheMu.Unlock()
	if cache != nil {
		return cache
	}
	cache = loadConfig()
	return cache
}

// loadConfig reads config.json fresh from disk, bypassing the package cache.
// It is the single definition of "what a from-scratch read looks like", shared
// by Read (cached) and updateConfig (which must never hand a mutate callback
// the stale in-memory cache while holding the store's lock).
func loadConfig() *Config {
	var cfg Config
	if err := store().Load(&cfg); err != nil {
		return defaultConfig()
	}
	if cfg.Organizations == nil {
		cfg.Organizations = make(map[string]Organization)
	}
	return &cfg
}

func Write(cfg *Config) error {
	err := store().Save(cfg)
	cacheMu.Lock()
	cache = nil
	cacheMu.Unlock()
	return err
}

func ClearCache() {
	cacheMu.Lock()
	defer cacheMu.Unlock()
	cache = nil
}

func defaultConfig() *Config {
	return &Config{
		Organizations: make(map[string]Organization),
	}
}

// errSkipWrite lets a mutate callback decline to persist anything (e.g.
// SetDefault on an unknown alias) without updateConfig treating it as a real
// failure.
var errSkipWrite = errors.New("config: skip write")

// updateConfig applies mutate to a freshly loaded config under ONE exclusive
// lock spanning read, mutate, and write, so two concurrent invocations
// serialize instead of each building its write from a stale snapshot. The
// package-level cache is bypassed entirely while the lock is held — mutate
// always sees what store().Update just loaded from disk, never the cache — and
// is invalidated afterward so a later Read() cannot hand back the pre-write
// value.
func updateConfig(mutate func(cfg *Config) error) error {
	var cfg Config
	err := store().Update(&cfg, func() error {
		if cfg.Organizations == nil {
			cfg.Organizations = make(map[string]Organization)
		}
		return mutate(&cfg)
	})

	cacheMu.Lock()
	cache = nil
	cacheMu.Unlock()

	if errors.Is(err, errSkipWrite) {
		return nil
	}
	return err
}

func StoreOrganization(alias string) error {
	return updateConfig(func(cfg *Config) error {
		cfg.Organizations[alias] = Organization{}
		if cfg.DefaultOrg == "" {
			cfg.DefaultOrg = alias
		}
		return nil
	})
}

func RemoveOrganization(alias string) error {
	return updateConfig(func(cfg *Config) error {
		delete(cfg.Organizations, alias)
		if cfg.DefaultOrg == alias {
			cfg.DefaultOrg = ""
			for name := range cfg.Organizations {
				cfg.DefaultOrg = name
				break
			}
		}
		return nil
	})
}

func SetDefault(alias string) error {
	return updateConfig(func(cfg *Config) error {
		if _, ok := cfg.Organizations[alias]; !ok {
			return errSkipWrite
		}
		cfg.DefaultOrg = alias
		return nil
	})
}
