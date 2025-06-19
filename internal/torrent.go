// internal/torrent.go
package internal

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
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

func progressLog(msg string, td *TorrentDownload) {
	td.Message = msg

	saveTorrentDownload(td)
}

func StartMultipleDownloads(ctx context.Context, entries []TorrentEntry, outDir string) {
	var wg sync.WaitGroup
	for _, entry := range entries {
		wg.Add(1)
		go func(e TorrentEntry) {
			defer wg.Done()
			if err := StartTorrent(ctx, e, outDir); err != nil {
				log.Printf("❌ Error downloading %s: %v", e.Title, err)
			}
		}(entry)
	}
	wg.Wait()
}

func StartTorrent(ctx context.Context, entry TorrentEntry, outDir string) error {
	title := fmt.Sprintf("S%02d: %s (%s)", entry.DownloadKey, entry.SeasonName, entry.Quality)
	opcfg := LoadConfig()
	torrentURL := fmt.Sprintf("%s/download/%d.torrent", opcfg.TorrentAPIURL, entry.TorrentID)

	td := &TorrentDownload{
		Title:        title,
		TorrentID:    entry.TorrentID,
		Started:      time.Now(),
		OutDir:       outDir,
		ChapterRange: entry.ChapterRange,
		Message:      "🌱 Initializing...",
	}

	saveTorrentDownload(td)
	progressLog(fmt.Sprintf("🌍 Fetching torrent: %s", torrentURL), td)

	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, torrentURL, nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		progressLog("❌ HTTP request cancelled or failed", td)
		deleteTorrentDownload(td)
		return fmt.Errorf("❌ Could not fetch .torrent file: %w", err)
	}
	defer resp.Body.Close()

	meta, err := metainfo.Load(resp.Body)
	if err != nil {
		return fmt.Errorf("❌ Failed to parse .torrent file: %w", err)
	}

	tmpDir := filepath.Join(os.TempDir(), fmt.Sprintf("opfor-tmp-%d", entry.TorrentID))
	if err := os.MkdirAll(tmpDir, 0755); err != nil {
		return fmt.Errorf("❌ Failed to create temp dir: %w", err)
	}

	cfg := torrent.NewDefaultClientConfig()
	cfg.DataDir = tmpDir
	cfg.NoUpload = true
	cfg.ListenPort = 0

	client, err := torrent.NewClient(cfg)
	if err != nil {
		return fmt.Errorf("❌ Torrent client error: %w", err)
	}
	defer client.Close()

	t, err := client.AddTorrent(meta)
	if err != nil {
		return fmt.Errorf("❌ Could not add torrent: %w", err)
	}

	select {
	case <-t.GotInfo():
		progressLog("ℹ️ Torrent metadata loaded.", td)
	case <-time.After(20 * time.Second):
		td.Error = "❌ Timeout while waiting for torrent metadata"
		saveTorrentDownload(td)
		return fmt.Errorf("timeout while waiting for GotInfo for %s", title)
	case <-ctx.Done():
		progressLog("🛑 Cancelled before metadata loaded", td)
		deleteTorrentDownload(td)
		return ctx.Err()
	}

	td.TotalSize = t.Length()
	saveTorrentDownload(td)

	t.DownloadAll()

downloadLoop:
	for {
		select {
		case <-ctx.Done():
			progressLog("🛑 Cancelled by user", td)
			deleteTorrentDownload(td)
			return ctx.Err()
		case <-time.After(1 * time.Second):
			if t.BytesMissing() == 0 {
				break downloadLoop
			}
			td.Progress = t.BytesCompleted()
			td.Message = fmt.Sprintf("📥 Downloading... %.2f%%", (float64(td.Progress)/float64(td.TotalSize))*100)
			saveTorrentDownload(td)
		}
	}

	progressLog("✅ Download complete", td)

	var anyMatched bool
	for _, f := range t.Files() {
		ext := filepath.Ext(f.Path())
		if ext != ".mkv" && ext != ".mp4" {
			continue
		}
		src := filepath.Join(tmpDir, f.Path())
		if err := MatchAndPlaceVideo(src, outDir, td); err != nil {
			td.Error += fmt.Sprintf("❌ Error placing file %s: %v\n", f.Path(), err)
		} else {
			anyMatched = true
		}
	}

	if !anyMatched {
		td.Error += "⚠️ No videos were placed, no metadata matches found.\n"
	} else {
		td.Done = true
	}

	saveTorrentDownload(td)

	time.Sleep(2 * time.Second)
	deleteTorrentDownload(td)

	return nil
}

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

func deleteTorrentDownload(td *TorrentDownload) {
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
	if err := os.RemoveAll(tmpDir); err != nil {
		progressLog(fmt.Sprintf("⚠️ Could not remove tmp dir %s: %v", tmpDir, err), td)
	} else {
		progressLog(fmt.Sprintf("🧹 Removed tmp dir: %s", tmpDir), td)
	}

}

func loadActiveDownloadsFromFile() []*TorrentDownload {
	data, err := os.ReadFile(activeFile)
	if err != nil {
		return []*TorrentDownload{}
	}
	var list []*TorrentDownload
	json.Unmarshal(data, &list)
	return list
}

func GetActiveDownloads() []*TorrentDownload {
	return loadActiveDownloadsFromFile()
}

func CleanupStaleDownloads() {
	activeMu.Lock()
	defer activeMu.Unlock()

	list := loadActiveDownloadsFromFile()
	var stillActive []*TorrentDownload
	for _, td := range list {
		if !td.Done {
			stillActive = append(stillActive, td)
		}
	}

	data, _ := json.MarshalIndent(stillActive, "", "  ")
	_ = os.WriteFile(activeFile, data, 0644)
}
