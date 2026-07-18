// shared/config.go
package shared

import (
	"encoding/json"
	"log"
	"os"
	"path/filepath"
	"sync"
)

var defaultCfg = Config{
	TargetDir:  "",
	GitHubRepo: "tissla/one-pace-jellyfin",
}

var (
	globalConfig    *Config
	configLoadOnce  sync.Once
	configLoadError error
)

// loads config-file file, creates it if it does not exist
func LoadConfig() (*Config, error) {

	configLoadOnce.Do(func() {
		path := EnsureConfigExists()

		data, err := os.ReadFile(path)
		if err != nil {
			configLoadError = err
		}

		var cfg Config
		if err := json.Unmarshal(data, &cfg); err != nil {
			configLoadError = err
		}
		globalConfig = &cfg
	})

	if configLoadError != nil {
		return nil, configLoadError
	}

	return globalConfig, nil
}

// writes static config file from config object, as json
func SaveConfig(cfg Config) error {
	path := getConfigPath()

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return err
	}

	globalConfig = &cfg

	return nil
}

// creates default config if no config-file is found
func EnsureConfigExists() string {
	path := getConfigPath()
	if _, err := os.Stat(path); os.IsNotExist(err) {

		SaveConfig(defaultCfg)
	}

	return path
}

// returns the config filepath from the OS's default config directory
func getConfigPath() string {
	return filepath.Join(ConfigDir(), "config.json")
}

// ConfigDir returns the app's config directory (creating it if needed), so other
// packages can place their own per-user files (e.g. the search cache) next to
// config.json instead of relying on the process's working directory.
func ConfigDir() string {
	dirname, err := os.UserConfigDir()
	if err != nil {
		log.Fatalf("could not determine config directory: %v", err)
	}
	path := filepath.Join(dirname, "opforjellyfin")
	if err := os.MkdirAll(path, 0755); err != nil {
		log.Fatalf("could not create config dir: %v", err)
	}
	return path
}
