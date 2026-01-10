package scraper

import (
	"encoding/json"
	"opforjellyfin/internal/shared"
	"os"
)

type SearchCache struct {
	Results []shared.TorrentEntry `json:"results"`
}

const CacheFile = ".search_cache.json"

// saves the current search results to cache, returns error if failed
func SaveSearchCache(results []shared.TorrentEntry) error {
	cache := SearchCache{
		Results: results,
	}

	data, err := json.MarshalIndent(cache, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(CacheFile, data, 0644)
}

// loads the search cache, returns the adress to the cache and error
func LoadSearchCache() (*SearchCache, error) {
	data, err := os.ReadFile(CacheFile)
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

// tries to find the torrent by key, returns result and error
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
