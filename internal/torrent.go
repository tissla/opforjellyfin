// internal/torrent.go
package internal

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/anacrolix/torrent"
	"github.com/anacrolix/torrent/metainfo"
)

type TorrentDownload struct {
	Title        string
	TorrentID    int
	Started      time.Time
	OutDir       string
	Progress     int64
	TotalSize    int64
	Message      string
	Done         bool
	Error        string
	ChapterRange string
}

var (
	activeFile = filepath.Join(os.TempDir(), "opfor-active.json")
	SilentLogs bool
	activeMu   sync.Mutex
)

// used for writing to json-file for live tracking of concurrent background downloads. currently unused
func progressLog(msg string, td *TorrentDownload) {

	td.Message = msg
	DebugLog(false, msg, td.TorrentID)
	saveTorrentDownload(td)
}

// sync waitgroup lets every download start at its own pace.
func StartMultipleDownloads(ctx context.Context, entries []TorrentEntry, outDir string) {
	var wg sync.WaitGroup
	for _, entry := range entries {
		wg.Add(1)
		go func(e TorrentEntry) {
			defer wg.Done()
			if err := StartTorrent(ctx, e, outDir); err != nil {
				DebugLog(true, "❌ Error downloading %s: %v", e.Title, err)
			}
		}(entry)
	}
	wg.Wait()
}

// main torrent download and tracker
func StartTorrent(ctx context.Context, entry TorrentEntry, outDir string) error {
	// init download obj
	td := &TorrentDownload{
		Title:        fmt.Sprintf("S%02d: %s (%s)", entry.DownloadKey, entry.SeasonName, entry.Quality),
		TorrentID:    entry.TorrentID,
		Started:      time.Now(),
		OutDir:       outDir,
		ChapterRange: entry.ChapterRange,
		Message:      "🌱 Initializing...",
	}
	saveTorrentDownload(td)

	// 2) get torrent meta-info
	torrentURL := fmt.Sprintf("%s/download/%d.torrent", LoadConfig().TorrentAPIURL, entry.TorrentID)
	progressLog(fmt.Sprintf("🌍 Fetching torrent: %s", torrentURL), td)

	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, torrentURL, nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		progressLog("❌ HTTP request failed", td)
		return cleanupWithError(td, err)
	}
	defer resp.Body.Close()

	meta, err := metainfo.Load(resp.Body)
	if err != nil {
		return cleanupWithError(td, err)
	}

	// create tempdir
	tmpDir := filepath.Join(os.TempDir(), fmt.Sprintf("opfor-tmp-%d", entry.TorrentID))
	if err := os.MkdirAll(tmpDir, 0755); err != nil {
		return cleanupWithError(td, err)
	}

	// start the torrent-client
	cfg := torrent.NewDefaultClientConfig()
	cfg.DataDir = tmpDir
	cfg.NoUpload = true
	cfg.ListenPort = 0

	client, err := torrent.NewClient(cfg)
	if err != nil {
		return cleanupWithError(td, err)
	}
	defer client.Close()

	t, err := client.AddTorrent(meta)
	if err != nil {
		return cleanupWithError(td, err)
	}

	// get torrent metadata
	select {
	case <-t.GotInfo():
		progressLog("ℹ️ Torrent metadata loaded", td)
	case <-time.After(20 * time.Second):
		return cleanupWithError(td, fmt.Errorf("timeout waiting for torrent info"))
	case <-ctx.Done():
		return cleanupWithError(td, ctx.Err())
	}

	// start download
	td.TotalSize = t.Length()
	saveTorrentDownload(td)
	t.DownloadAll()

	// watch progress
	for t.BytesMissing() > 0 {
		select {
		case <-ctx.Done():
			return cleanupWithError(td, ctx.Err())
		case <-time.After(1 * time.Second):
			td.Progress = t.BytesCompleted()
			td.Message = fmt.Sprintf("📥 Downloading... %.2f%%", 100*float64(td.Progress)/float64(td.TotalSize))
			saveTorrentDownload(td)
		}
	}
	td.Progress = td.TotalSize
	progressLog("✅ Download complete, placing files...", td)

	for _, f := range t.Files() {
		src := filepath.Join(tmpDir, f.Path())
		msg, err := MatchAndPlaceVideo(src, outDir, td)
		if err != nil {
			td.Error += fmt.Sprintf("❌ Error placing file %s: %v\n", f.Path(), err)
		} else if msg != "" {
			progressLog(msg, td)
		}
	}

	td.Done = true
	progressLog("🚀 All done", td)
	return nil
}

// removes and cleans up the torrent when an error is cast
func cleanupWithError(td *TorrentDownload, err error) error {
	td.Error = err.Error()
	progressLog(td.Error, td)
	DeleteTorrentDownload(td)
	return err
}

// updates the active.json with current status of download
func saveTorrentDownload(td *TorrentDownload) {
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

// deletes torrent from active.json
func DeleteTorrentDownload(td *TorrentDownload) {
	activeMu.Lock()
	defer activeMu.Unlock()

	list := loadActiveDownloadsFromFile()
	updated := []*TorrentDownload{}
	for _, d := range list {
		if d.TorrentID != td.TorrentID {
			updated = append(updated, d)
		}
	}
	data, _ := json.MarshalIndent(updated, "", "  ")
	os.WriteFile(activeFile, data, 0644)
	tmpDir := filepath.Join(os.TempDir(), fmt.Sprintf("opfor-tmp-%d", td.TorrentID))

	os.RemoveAll(tmpDir)

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
