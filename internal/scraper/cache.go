package scraper

import (
	"encoding/json"
	"opforjellyfin/internal/shared"
	"os"
	"path/filepath"
)

type SearchCache struct {
	Results []shared.TorrentEntry `json:"results"`
}

const cacheFileName = "search_cache.json"

// cacheFilePath returns the absolute path to the search cache, inside the app's
// config directory.
func cacheFilePath() string {
	return filepath.Join(shared.ConfigDir(), cacheFileName)
}

// SaveSearchCache saves the current search results to cache, returns error if failed
func SaveSearchCache(results []shared.TorrentEntry) error {
	cache := SearchCache{
		Results: results,
	}

	data, err := json.MarshalIndent(cache, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(cacheFilePath(), data, 0644)
}

// LoadSearchCache loads the search cache, returns the adress to the cache and error
func LoadSearchCache() (*SearchCache, error) {
	data, err := os.ReadFile(cacheFilePath())
	if err != nil {
		return nil, err
	}

	var cache SearchCache
	err = json.Unmarshal(data, &cache)
	if err != nil {
		return nil, err
	}

	return &cache, nil
}

// GetTorrentByKey tries to find the torrent by key, returns result and error
func GetTorrentByKey(key int) (*shared.TorrentEntry, error) {
	cache, err := LoadSearchCache()
	if err != nil {
		return nil, err
	}

	for _, result := range cache.Results {
		if result.DownloadKey == key {
			return &result, nil
		}
	}

	return nil, nil
}
