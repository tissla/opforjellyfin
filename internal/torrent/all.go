package torrent

import (
	"context"
	"fmt"
	"path"
	"path/filepath"
	"sort"
	"strings"

	"opforjellyfin/internal/logger"
	"opforjellyfin/internal/matcher"
	"opforjellyfin/internal/metadata"
	"opforjellyfin/internal/shared"
)

// SelectAllDownloads assigns each missing episode to its most seeded source.
func SelectAllDownloads(ctx context.Context, entries []shared.TorrentEntry, cfg *shared.Config, index *shared.MetadataIndex, includeExtended, includeSpecials bool) ([]shared.TorrentEntry, error) {
	if index == nil || len(index.Entries) == 0 {
		return nil, fmt.Errorf("metadata index is missing; run 'opfor sync'")
	}
	wanted := make(map[string]bool)
	for _, entry := range index.Entries {
		if entry.EpisodeID == "" || (entry.Kind != "episode" && entry.Kind != "special") {
			return nil, fmt.Errorf("metadata index lacks episode information; run 'opfor sync'")
		}
		if entry.Kind == "special" && !includeSpecials || entry.Variant == "extended" && !includeExtended {
			continue
		}
		wanted[episodeKey(entry)] = true
	}
	for _, arc := range index.Arcs {
		found, err := metadata.ExistingVideos(filepath.Join(cfg.TargetDir, arc.Folder), index)
		if err != nil {
			return nil, fmt.Errorf("read existing videos: %w", err)
		}
		for id := range found {
			delete(wanted, episodeKey(index.Entries[id]))
		}
	}

	entries = append([]shared.TorrentEntry(nil), entries...)
	sort.Slice(entries, func(i, j int) bool {
		if entries[i].Seeders != entries[j].Seeders {
			return entries[i].Seeders > entries[j].Seeders
		}
		return entries[i].TorrentID < entries[j].TorrentID
	})
	seen := make(map[int]bool)
	var selected []shared.TorrentEntry
	var fetchError error
	for _, entry := range entries {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		if len(wanted) == 0 {
			break
		}
		if entry.Seeders <= 0 || entry.TorrentID <= 0 || seen[entry.TorrentID] {
			continue
		}
		seen[entry.TorrentID] = true
		meta, err := fetchTorrentMeta(ctx, cfg.Source.BaseURL, entry.TorrentID)
		if err != nil {
			if ctx.Err() != nil {
				return nil, ctx.Err()
			}
			fetchError = err
			logger.Log(true, "Skipping torrent %d: %v", entry.TorrentID, err)
			continue
		}
		info, err := meta.UnmarshalInfo()
		if err != nil {
			return nil, fmt.Errorf("torrent %d: %w", entry.TorrentID, err)
		}
		entry.SelectedFiles = nil
		for _, file := range info.UpvertedFiles() {
			filename := strings.Join(append([]string{info.BestName()}, file.BestPath()...), "/")
			if !shared.IsVideoFile(filename) {
				continue
			}
			stem := strings.TrimSuffix(path.Base(filename), path.Ext(filename))
			id, exists := index.ByFilename[stem]
			if !exists {
				continue
			}
			if _, err := matcher.MatchFile(filename, index); err != nil {
				return nil, err
			}
			key := episodeKey(index.Entries[id])
			if !wanted[key] {
				continue
			}
			entry.SelectedFiles = append(entry.SelectedFiles, filename)
			delete(wanted, key)
		}
		if len(entry.SelectedFiles) > 0 {
			selected = append(selected, entry)
		}
	}
	if len(selected) == 0 && fetchError != nil {
		return nil, fmt.Errorf("could not select downloads: %w", fetchError)
	}
	return selected, nil
}

func episodeKey(entry shared.MetadataFile) string {
	if entry.Variant == "extended" {
		return entry.EpisodeID + ":extended"
	}
	return entry.EpisodeID
}
