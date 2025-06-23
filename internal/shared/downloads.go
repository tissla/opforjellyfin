// shared/active.go

package shared

import (
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
	return list
}

func ClearActiveDownloads() {
	mu.Lock()
	defer mu.Unlock()
	activeDownloads = make(map[int]*TorrentDownload)
}
