package metadata

import (
	"encoding/json"
	"fmt"
	"opforjellyfin/internal/shared"
	"os"
	"path/filepath"
	"sync"
)

var (
	metadataCache     *shared.MetadataIndex
	metadataCacheErr  error
	metadataCacheOnce sync.Once
)

// LoadMetadataCache reads the generated index once per command.
func LoadMetadataCache() (*shared.MetadataIndex, error) {
	metadataCacheOnce.Do(func() {
		cfg, err := shared.LoadConfig()
		if err != nil {
			metadataCacheErr = err
			return
		}
		metadataCache, metadataCacheErr = readIndex(cfg.TargetDir)
	})
	return metadataCache, metadataCacheErr
}

func readIndex(baseDir string) (*shared.MetadataIndex, error) {
	if baseDir == "" {
		return nil, fmt.Errorf("set a library directory with 'opfor setDir <path>' first")
	}
	data, err := os.ReadFile(filepath.Join(baseDir, "metadata-index.json"))
	if err != nil {
		return nil, fmt.Errorf("read metadata index (run 'opfor sync'): %w", err)
	}
	var index shared.MetadataIndex
	if err := json.Unmarshal(data, &index); err != nil {
		return nil, fmt.Errorf("read metadata index: %w", err)
	}
	if len(index.ByFilename) == 0 || len(index.Entries) == 0 {
		return nil, fmt.Errorf("metadata index has no filename mappings; run 'opfor sync'")
	}
	return &index, nil
}
