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
	activeDownloads[td.TorrentID] = td
}

func GetActiveDownloads() []*TorrentDownload {
	mu.RLock()
	defer mu.RUnlock()

	var list []*TorrentDownload
	for _, td := range activeDownloads {
		list = append(list, td)
	}

	//make sure they come in the right order
	sort.SliceStable(list, func(i, j int) bool {
		return list[i].ChapterRange < list[j].ChapterRange
	})

	return list
}

func ClearActiveDownloads() {
	mu.Lock()
	defer mu.Unlock()
	activeDownloads = make(map[int]*TorrentDownload)
	CleanupTempDirs()
}

func CleanupTempDirs() error {
	files, err := os.ReadDir(os.TempDir())
	if err != nil {
		return err
	}
	for _, f := range files {
		if strings.HasPrefix(f.Name(), "opfor-tmp-") {
			path := filepath.Join(os.TempDir(), f.Name())
			if err := os.RemoveAll(path); err != nil {
				fmt.Printf("⚠️  Failed to remove %s: %v\n", path, err)
			}
		}
	}
	return nil
}

// helper
func (td *TorrentDownload) MarkPlaced(msg string) {
	td.ProgressMessage = msg
	td.Placed = true
	SaveTorrentDownload(td)
}
