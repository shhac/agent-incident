package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
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
	if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
		return filepath.Join(xdg, "agent-incident")
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "agent-incident")
}

func configPath() string {
	return filepath.Join(ConfigDir(), "config.json")
}

func Read() *Config {
	cacheMu.Lock()
	defer cacheMu.Unlock()
	if cache != nil {
		return cache
	}
	data, err := os.ReadFile(configPath())
	if err != nil {
		return defaultConfig()
	}
	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return defaultConfig()
	}
	if cfg.Organizations == nil {
		cfg.Organizations = make(map[string]Organization)
	}
	cache = &cfg
	return cache
}

func Write(cfg *Config) error {
	cacheMu.Lock()
	cache = nil
	cacheMu.Unlock()

	dir := ConfigDir()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(configPath(), append(data, '\n'), 0o644)
}

func ClearCache() {
	cacheMu.Lock()
	defer cacheMu.Unlock()
	cache = nil
}

func defaultConfig() *Config {
	cfg := &Config{
		Organizations: make(map[string]Organization),
	}
	cache = cfg
	return cfg
}

func StoreOrganization(alias string) error {
	cfg := Read()
	cfg.Organizations[alias] = Organization{}
	if cfg.DefaultOrg == "" {
		cfg.DefaultOrg = alias
	}
	return Write(cfg)
}

func RemoveOrganization(alias string) error {
	cfg := Read()
	delete(cfg.Organizations, alias)
	if cfg.DefaultOrg == alias {
		cfg.DefaultOrg = ""
		for name := range cfg.Organizations {
			cfg.DefaultOrg = name
			break
		}
	}
	return Write(cfg)
}

func SetDefault(alias string) error {
	cfg := Read()
	if _, ok := cfg.Organizations[alias]; !ok {
		return nil
	}
	cfg.DefaultOrg = alias
	return Write(cfg)
}
