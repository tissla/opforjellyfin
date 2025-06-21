// internal/filters.go
package internal

import (
	"opforjellyfin/internal/shared"
	"strings"
)

func FilterByTitle(list []shared.TorrentEntry, keyword string) []shared.TorrentEntry {
	var filtered []shared.TorrentEntry
	for _, entry := range list {
		if strings.Contains(strings.ToLower(entry.SeasonName), strings.ToLower(keyword)) {
			filtered = append(filtered, entry)
		}
	}
	return filtered
}

func FilterBySeasonRange(list []shared.TorrentEntry, from, to int) []shared.TorrentEntry {
	var filtered []shared.TorrentEntry
	for _, entry := range list {
		if entry.DownloadKey >= from && entry.DownloadKey <= to {
			filtered = append(filtered, entry)
		}
	}
	return filtered
}
