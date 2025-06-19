// internal/filters.go
package internal

import "strings"

func FilterByTitle(list []TorrentEntry, keyword string) []TorrentEntry {
	var filtered []TorrentEntry
	for _, entry := range list {
		if strings.Contains(strings.ToLower(entry.SeasonName), strings.ToLower(keyword)) {
			filtered = append(filtered, entry)
		}
	}
	return filtered
}

func FilterBySeasonRange(list []TorrentEntry, from, to int) []TorrentEntry {
	var filtered []TorrentEntry
	for _, entry := range list {
		if entry.DownloadKey >= from && entry.DownloadKey <= to {
			filtered = append(filtered, entry)
		}
	}
	return filtered
}
