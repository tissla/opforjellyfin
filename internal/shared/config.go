// shared/config.go
package shared

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
)

var defaultCfg = Config{
	TargetDir:     "",
	GitHubRepo:    "tissla/one-pace-jellyfin",
	TorrentAPIURL: "https://nyaa.si",
}

// loads config-file file, creates it if it does not exist
func LoadConfig() Config {

	path := EnsureConfigExists()

	data, err := os.ReadFile(path)
	if err != nil {
		log.Fatalf("❌ Could not read config: %v", err)
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		log.Fatalf("❌ Invalid config format: %v", err)
	}

	return cfg
}

// writes static config file from config object, as json
func SaveConfig(cfg Config) {
	path := getConfigPath()

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		log.Fatalf("❌ Failed to serialize config: %v", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		log.Fatalf("❌ Failed to save config: %v", err)
	}
}

// creates default config if no config-file is found
func EnsureConfigExists() string {
	path := getConfigPath()
	if _, err := os.Stat(path); os.IsNotExist(err) {

		SaveConfig(defaultCfg)
		fmt.Printf("📁 Created default config at: %s\n", path)
	}

	return path
}

// returns the config filepath from the OS's default config directory
func getConfigPath() string {
	dirname, err := os.UserConfigDir()
	if err != nil {
		log.Fatalf("❌ Could not determine config directory: %v", err)
	}
	path := filepath.Join(dirname, "opforjellyfin")
	err = os.MkdirAll(path, 0755)
	if err != nil {
		log.Fatalf("❌ Could not create config dir: %v", err)
	}
	return filepath.Join(path, "config.json")
}
