package metadata

import (
	"opforjellyfin/internal/shared"
	"os"
	"path"
	"path/filepath"
	"strings"
)

// HaveVideoStatus reports none (0), some (1), or all (2) for an episode or arc title.
func HaveVideoStatus(title string) int {
	index, err := LoadMetadataCache()
	if err != nil {
		return 0
	}
	cfg, err := shared.LoadConfig()
	if err != nil {
		return 0
	}
	if shared.IsVideoFile(title) {
		title = strings.TrimSuffix(title, filepath.Ext(title))
	}
	if id, ok := index.ByFilename[title]; ok {
		entry, exists := index.Entries[id]
		if !exists {
			return 0
		}
		dir := filepath.Join(cfg.TargetDir, filepath.FromSlash(path.Dir(entry.NFOPath)))
		if videoEntries(dir, index)[id] {
			return 2
		}
		return 0
	}
	for _, arc := range index.Arcs {
		if arc.Folder == title {
			matched, total := CountVideosAndTotal(filepath.Join(cfg.TargetDir, arc.Folder))
			if matched == 0 {
				return 0
			}
			if matched < total {
				return 1
			}
			return 2
		}
	}
	return 0
}

func HaveMetadata(title string) bool {
	index, err := LoadMetadataCache()
	if err != nil {
		return false
	}
	if shared.IsVideoFile(title) {
		title = strings.TrimSuffix(title, filepath.Ext(title))
	}
	if _, exists := index.ByFilename[title]; exists {
		return true
	}
	for _, arc := range index.Arcs {
		if arc.Folder == title {
			return true
		}
	}
	return false
}

// CountVideosAndTotal counts indexed episodes, so alias sidecars are not counted twice.
func CountVideosAndTotal(dir string) (matched int, total int) {
	index, err := LoadMetadataCache()
	if err != nil {
		return 0, 0
	}
	for _, entry := range index.Entries {
		if path.Dir(entry.NFOPath) == filepath.Base(dir) {
			total++
		}
	}
	return len(videoEntries(dir, index)), total
}

func videoEntries(dir string, index *shared.MetadataIndex) map[string]bool {
	found := make(map[string]bool)
	files, _ := os.ReadDir(dir)
	for _, file := range files {
		if !file.Type().IsRegular() || !shared.IsVideoFile(file.Name()) {
			continue
		}
		stem := strings.TrimSuffix(file.Name(), filepath.Ext(file.Name()))
		id, exists := index.ByFilename[stem]
		if exists && path.Dir(index.Entries[id].NFOPath) == filepath.Base(dir) {
			found[id] = true
		}
	}
	return found
}
