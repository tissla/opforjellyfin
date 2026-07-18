package metadata

import (
	"encoding/json"
	"fmt"
	"opforjellyfin/internal/logger"
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
		cfg, err := shared.LoadConfig()
		if err != nil {
			logger.Log(false, "metadata: could not load config: %v", err)
			metadataCache = &shared.MetadataIndex{}
			return
		}

		path := filepath.Join(cfg.TargetDir, "metadata-index.json")
		data, err := os.ReadFile(path)
		if err != nil {
			// Not existing yet is the expected state before the first setDir/sync -
			// anything else (permissions, etc.) is worth a log entry.
			if !os.IsNotExist(err) {
				logger.Log(false, "metadata: could not read %s: %v", path, err)
			}
			metadataCache = &shared.MetadataIndex{}
			return
		}

		if err := json.Unmarshal(data, &metadataCache); err != nil {
			logger.Log(false, "metadata: could not parse %s: %v", path, err)
			metadataCache = &shared.MetadataIndex{}
		}
	})

	return metadataCache
}

// loadIndexIntoCache reads an existing metadata-index.json (e.g. one shipped
// pre-built by the metadata repo, copied in as-is) and puts it straight into
// the in-memory cache, without re-deriving it from the .nfo files on disk.
func loadIndexIntoCache(baseDir string) error {
	path := filepath.Join(baseDir, "metadata-index.json")

	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("could not read metadata index: %w", err)
	}

	var index shared.MetadataIndex
	if err := json.Unmarshal(data, &index); err != nil {
		return fmt.Errorf("could not parse metadata index: %w", err)
	}

	metadataCache = &index
	return nil
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
