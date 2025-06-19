// internal/config.go
package internal

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
)



type Config struct {
    TargetDir       string `json:"target_dir"`
    GitHubRepo   string `json:"github_base_url"`
    TorrentAPIURL   string `json:"torrent_api_url"`
}

func LoadConfig() Config {
	path := getConfigPath()

	if _, err := os.Stat(path); os.IsNotExist(err) {
		defaultCfg := Config{
			TargetDir:      "",
			GitHubRepo:     "tissla/one-pace-jellyfin",
			TorrentAPIURL:  "https://nyaa.si",
		}
		SaveConfig(defaultCfg)
		fmt.Printf("📁 Created default config at: %s\n", path)
		return defaultCfg
	}

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



func SetDir(dir string, force bool) {
	abs, err := filepath.Abs(dir)
	if err != nil {
		log.Fatalf("❌ Invalid directory: %v", err)
	}

	cfg := LoadConfig()
	cfg.TargetDir = abs
	SaveConfig(cfg)

	fmt.Println("✅ Default target directory set to:", abs)

    if (force) {
        FetchAllMetadata(abs, cfg)
    } else {
	    SyncMetadata(abs, cfg)
    }
}


func EnsureConfigExists() {
	path := getConfigPath()
	if _, err := os.Stat(path); os.IsNotExist(err) {
		defaultCfg := Config{
			TargetDir:      "",
			GitHubRepo:  "tissla/one-pace-jellyfin",
			TorrentAPIURL:  "https://nyaa.si",
		}
		SaveConfig(defaultCfg)
		fmt.Printf("📁 Created default config at: %s\n", path)
	}
}


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
