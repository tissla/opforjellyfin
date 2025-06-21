// torrent/torrent.go
package torrent

import (
	"context"

	"fmt"
	"net/http"
	"opforjellyfin/internal/logger"
	"opforjellyfin/internal/matcher"
	"opforjellyfin/internal/shared"
	"opforjellyfin/internal/ui"
	"os"
	"path/filepath"

	"time"

	"github.com/anacrolix/torrent"
	"github.com/anacrolix/torrent/metainfo"
)

// used for writing to json-file for live tracking of concurrent background downloads. currently just saves progress
func progressLog(msg string, td *shared.TorrentDownload) {

	logger.DebugLog(false, msg, td.TorrentID)
	shared.SaveTorrentDownload(td)
}

// main torrent download and tracker
func StartTorrent(ctx context.Context, entry shared.TorrentEntry, outDir string) error {
	// init download obj

	dKey := ui.StyleFactory(fmt.Sprintf("%4d", entry.DownloadKey), ui.Style.Pink)
	title := ui.StyleFactory(entry.SeasonName, ui.Style.LBlue)

	td := &shared.TorrentDownload{
		Title:        fmt.Sprintf("%s: %s (%s)", dKey, title, entry.Quality),
		TorrentID:    entry.TorrentID,
		Started:      time.Now(),
		OutDir:       outDir,
		ChapterRange: entry.ChapterRange,
	}
	shared.SaveTorrentDownload(td)

	// get torrent meta-info
	torrentURL := fmt.Sprintf("%s/download/%d.torrent", shared.LoadConfig().TorrentAPIURL, entry.TorrentID)
	progressLog(fmt.Sprintf("🌍 Fetching torrent: %s", torrentURL), td)

	// get metadata
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, torrentURL, nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		progressLog("HTTP request for metadata failed", td)
		return cleanupWithError(td, err)
	}
	defer resp.Body.Close()

	//build meta
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

	// add torrent
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
	shared.SaveTorrentDownload(td)
	t.DownloadAll()

	// watch progress
	for t.BytesMissing() > 0 {
		select {
		case <-ctx.Done():
			return cleanupWithError(td, ctx.Err())
		case <-time.After(1 * time.Second):
			td.Progress = t.BytesCompleted()
			shared.SaveTorrentDownload(td)
		}
	}

	td.Progress = td.TotalSize
	logger.DebugLog(false, "Torrent contains %d files", len(t.Files()))
	td.Done = true

	CloseWithLogs(client)

	progressLog("✅ Download complete, placing files...", td)

	logger.DebugLog(false, "Waiting for OS to release files..")
	time.Sleep(2 * time.Second)

	matcher.ProcessTorrentFiles(tmpDir, outDir, td)

	return nil
}

// loghelper
func CloseWithLogs(client *torrent.Client) {
	logger.DebugLog(false, "Torrentclient closed for: %s", client)
	client.Close()
}

// removes and cleans up the torrent when an error is cast
func cleanupWithError(td *shared.TorrentDownload, err error) error {
	logger.DebugLog(false, "cleanupWithError called:", err)
	td.Error = err.Error()
	progressLog(td.Error, td)

	// we clear all downloads when this happens
	shared.ClearActiveDownloads()
	return err
}
