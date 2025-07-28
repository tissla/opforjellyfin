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
	metadataCacheOnce sync.Once
)

// we read metadata-index.json once when we need it, and if multiple checks are needed we read the cache.
// returns a reference to the index-variable
func LoadMetadataCache() *shared.MetadataIndex {
	metadataCacheOnce.Do(func() {
		cfg := shared.LoadConfig()
		data, err := os.ReadFile(filepath.Join(cfg.TargetDir, "metadata-index.json"))
		if err != nil {
			metadataCache = &shared.MetadataIndex{}
			return
		}
		json.Unmarshal(data, &metadataCache)
	})

	return metadataCache
}

// saves file and creates cache
func saveMetadataIndex(index *shared.MetadataIndex, baseDir string) error {
	path := filepath.Join(baseDir, "metadata-index.json")
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("could not create index file: %w", err)
	}
	defer f.Close()

	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	if err := enc.Encode(index); err != nil {
		return fmt.Errorf("could not encode metadata index: %w", err)
	}
	if err := f.Sync(); err != nil {
		return fmt.Errorf("could not flush metadata index to disk: %w", err)
	}

	metadataCache = index // cache immediately after saving

	return nil
}
