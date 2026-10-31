package torrent

import (
	"context"
	"fmt"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"opforjellyfin/internal/logger"
	"opforjellyfin/internal/matcher"
	"opforjellyfin/internal/metadata"
	"opforjellyfin/internal/shared"
)

var resolutionTag = regexp.MustCompile(`(?i)\[(\d{3,4})p\]`)

// SelectAllDownloads prefers 1080p or higher, falling back to 720p per missing episode.
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
	highQuality := make(map[string]bool)
	var candidates []shared.TorrentEntry
	var fetchError error
	for _, entry := range entries {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		if len(highQuality) == len(wanted) {
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
			resolution := videoResolution(filename)
			if !shared.IsVideoFile(filename) || (resolution != 720 && resolution < 1080) {
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
			if resolution >= 1080 {
				highQuality[key] = true
			}
		}
		if len(entry.SelectedFiles) > 0 {
			candidates = append(candidates, entry)
		}
	}

	var selected []shared.TorrentEntry
	for _, entry := range candidates {
		files := entry.SelectedFiles
		entry.SelectedFiles = nil
		for _, filename := range files {
			stem := strings.TrimSuffix(path.Base(filename), path.Ext(filename))
			key := episodeKey(index.Entries[index.ByFilename[stem]])
			if !wanted[key] || (highQuality[key] && videoResolution(filename) < 1080) {
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

func videoResolution(filename string) int {
	match := resolutionTag.FindStringSubmatch(path.Base(filename))
	if len(match) != 2 {
		return 0
	}
	height, _ := strconv.Atoi(match[1])
	return height
}

func episodeKey(entry shared.MetadataFile) string {
	if entry.Variant == "extended" {
		return entry.EpisodeID + ":extended"
	}
	return entry.EpisodeID
}
