package shared

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

var (
	activeFile = filepath.Join(os.TempDir(), "opfor-active.json")
	SilentLogs bool
	activeMu   sync.Mutex
)

// updates the active.json with current status of download
func SaveTorrentDownload(td *TorrentDownload) {
	activeMu.Lock()
	defer activeMu.Unlock()

	list := loadActiveDownloadsFromFile()
	updated := false
	for i, existing := range list {
		if existing.TorrentID == td.TorrentID {
			list[i] = td
			updated = true
			break
		}
	}
	if !updated {
		list = append(list, td)
	}
	data, _ := json.MarshalIndent(list, "", "  ")
	os.WriteFile(activeFile, data, 0644)
}

// returns list of torrents in active.json
func loadActiveDownloadsFromFile() []*TorrentDownload {
	data, err := os.ReadFile(activeFile)
	if err != nil {
		return []*TorrentDownload{}
	}
	var list []*TorrentDownload
	json.Unmarshal(data, &list)
	return list
}

// public func
func GetActiveDownloads() []*TorrentDownload {
	return loadActiveDownloadsFromFile()
}

// clear the opfor-active.json and any tmp-files
func ClearActiveDownloads() error {
	activeMu.Lock()
	defer activeMu.Unlock()

	data, err := json.MarshalIndent([]*TorrentDownload{}, "", "  ")
	if err != nil {
		return fmt.Errorf("could not marshal empty active list: %w", err)
	}
	if err := os.WriteFile(activeFile, data, 0644); err != nil {
		return fmt.Errorf("could not clear active file: %w", err)
	}

	files, err := os.ReadDir(os.TempDir())
	if err == nil {
		for _, fi := range files {
			if strings.HasPrefix(fi.Name(), "opfor-tmp-") {
				os.RemoveAll(filepath.Join(os.TempDir(), fi.Name()))
			}
		}
	}
	return nil
}
