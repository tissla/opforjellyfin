// shared/downloads.go

package shared

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
)

var (
	activeDownloads = make(map[int]*TorrentDownload)
	mu              sync.RWMutex
)

func SaveTorrentDownload(td *TorrentDownload) {
	mu.Lock()
	defer mu.Unlock()
	existing, ok := activeDownloads[td.TorrentID]
	if ok {
		*existing = *td // overwrite
	} else {
		activeDownloads[td.TorrentID] = td
	}
}

func GetActiveDownloads() []*TorrentDownload {
	mu.RLock()
	defer mu.RUnlock()

	var list []*TorrentDownload
	for _, td := range activeDownloads {
		list = append(list, td)
	}

	// Sort by ChapterRange, then by TorrentID as a deterministic tiebreak.
	sort.SliceStable(list, func(i, j int) bool {
		if list[i].ChapterRange != list[j].ChapterRange {
			return list[i].ChapterRange < list[j].ChapterRange
		}
		return list[i].TorrentID < list[j].TorrentID
	})

	return list
}

// clear cache and remove temp download directories
func ClearActiveDownloads() {
	mu.Lock()
	defer mu.Unlock()
	activeDownloads = make(map[int]*TorrentDownload)
	CleanupTempDirs()
}

func CleanupTempDirs() error {
	tmpDir, _ := GetTempDir()
	files, err := os.ReadDir(tmpDir)
	if err != nil {
		return err
	}
	for _, f := range files {
		if strings.HasPrefix(f.Name(), "opfor-tmp-") {
			path := filepath.Join(tmpDir, f.Name())
			if err := os.RemoveAll(path); err != nil {
				fmt.Printf("⚠️  Failed to remove %s: %v\n", path, err)
			}
		}
	}
	return nil
}

// helper
func (td *TorrentDownload) MarkPlaced(msg string) {
	td.PlacementProgress = msg
	td.Placed = true
	SaveTorrentDownload(td)
}
