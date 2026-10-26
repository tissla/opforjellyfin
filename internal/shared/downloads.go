// shared/downloads.go

package shared

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strings"
	"sync"
)

var (
	activeDownloads = make(map[int]TorrentDownload)
	mu              sync.RWMutex
)

func SaveTorrentDownload(td *TorrentDownload) {
	mu.Lock()
	defer mu.Unlock()
	snapshot := *td
	snapshot.PlacementFull = slices.Clone(td.PlacementFull)
	snapshot.VideoPaths = slices.Clone(td.VideoPaths)
	activeDownloads[td.TorrentID] = snapshot
}

func RemoveActiveDownload(id int) {
	mu.Lock()
	defer mu.Unlock()
	delete(activeDownloads, id)
}

func GetActiveDownloads() []*TorrentDownload {
	mu.RLock()
	defer mu.RUnlock()

	var list []*TorrentDownload
	for _, td := range activeDownloads {
		snapshot := td
		snapshot.PlacementFull = slices.Clone(td.PlacementFull)
		snapshot.VideoPaths = slices.Clone(td.VideoPaths)
		list = append(list, &snapshot)
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

// Clear in-memory download state; unfinished files remain available for retry.
func ClearActiveDownloads() {
	mu.Lock()
	defer mu.Unlock()
	activeDownloads = make(map[int]TorrentDownload)
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
