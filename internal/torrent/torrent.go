// torrent/torrent.go
package torrent

import (
	"context"
	"fmt"
	"net/http"
	"opforjellyfin/internal/logger"
	"opforjellyfin/internal/shared"
	"path/filepath"
	"strings"
	"time"

	"github.com/anacrolix/torrent"
	"github.com/anacrolix/torrent/metainfo"
)

// main torrent download and tracker. When seed is true, the client keeps
// uploading after the download finishes until ctx is cancelled (Ctrl+C) -
// the caller is responsible for not treating that as a failure.
func StartTorrent(ctx context.Context, td *shared.TorrentDownload, seed bool, selectedFiles []string) error {
	config, err := shared.LoadConfig()
	if err != nil {
		return err
	}
	meta, err := fetchTorrentMeta(ctx, config.Source.BaseURL, td.TorrentID)
	if err != nil {
		return err
	}

	// create tempdir using safe function
	tmpDir, err := shared.CreateTempTorrentDir(td.TorrentID)
	if err != nil {
		return fmt.Errorf("failed to create temp dir: %w", err)
	}

	// start the torrent-client with optimized settings for concurrent downloads
	cfg := torrent.NewDefaultClientConfig()
	cfg.DataDir = tmpDir
	cfg.NoUpload = !seed
	cfg.Seed = seed // upload even once we have nothing left to gain ourselves
	cfg.ListenPort = 0

	client, err := torrent.NewClient(cfg)
	if err != nil {
		return err
	}
	defer closeWithLogs(client)

	// add torrent
	t, err := client.AddTorrent(meta)
	if err != nil {
		return err
	}

	// get torrent metadata
	select {
	case <-t.GotInfo():
		logger.Log(false, "Torrent metadata loaded: %s", td.Title)
	case <-time.After(20 * time.Second):
		return fmt.Errorf("timeout waiting for torrent metadata")
	case <-ctx.Done():
		return ctx.Err()
	}

	files, err := filesToDownload(t, selectedFiles)
	if err != nil {
		return err
	}
	td.VideoPaths = nil
	td.TotalSize = 0
	for _, file := range files {
		td.TotalSize += file.Length()
		if shared.IsVideoFile(file.Path()) {
			td.VideoPaths = append(td.VideoPaths, filepath.Join(tmpDir, filepath.FromSlash(file.Path())))
		}
		file.Download()
	}

	shared.SaveTorrentDownload(td)

	// watch progress, save to activefile. Bounded independently of the
	// caller's ctx so a stalled download times out even during a --seed run,
	// without also capping how long seeding is allowed to continue below.
	downloadCtx, downloadCancel := context.WithTimeout(ctx, 30*time.Minute)
	defer downloadCancel()

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		progress, complete := fileProgress(files)
		td.Progress = progress
		shared.SaveTorrentDownload(td)
		if complete {
			break
		}
		select {
		case <-downloadCtx.Done():
			return downloadCtx.Err()
		case <-ticker.C:
		}
	}

	// close
	td.Progress = td.TotalSize
	logger.Log(false, "Torrent contains %d files", len(t.Files()))

	td.Done = true
	td.PlacementProgress = "⏳ Waiting to place.."
	shared.SaveTorrentDownload(td)
	logger.Log(false, "Download complete: %s", td.Title)

	if seed {
		td.PlacementProgress = "🌱 Seeding until stopped (Ctrl+C)..."
		shared.SaveTorrentDownload(td)
		<-ctx.Done()
		logger.Log(false, "Stopped seeding: %s", td.Title)
		return ctx.Err()
	}

	return nil
}

func fetchTorrentMeta(ctx context.Context, baseURL string, id int) (*metainfo.MetaInfo, error) {
	ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()
	url := fmt.Sprintf("%s/download/%d.torrent", strings.TrimRight(baseURL, "/"), id)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("fetch torrent %d: HTTP %d", id, resp.StatusCode)
	}
	return metainfo.Load(resp.Body)
}

func filesToDownload(t *torrent.Torrent, selected []string) ([]*torrent.File, error) {
	if selected == nil {
		return t.Files(), nil
	}
	wanted := make(map[string]bool, len(selected))
	for _, filename := range selected {
		wanted[filename] = true
	}
	var files []*torrent.File
	for _, file := range t.Files() {
		if wanted[file.Path()] {
			files = append(files, file)
			delete(wanted, file.Path())
		}
	}
	if len(wanted) > 0 || len(files) == 0 {
		return nil, fmt.Errorf("selected files are missing from torrent")
	}
	return files, nil
}

func fileProgress(files []*torrent.File) (bytes int64, complete bool) {
	complete = true
	for _, file := range files {
		bytes += file.BytesCompleted()
		for _, piece := range file.State() {
			if !piece.Complete {
				complete = false
			}
		}
	}
	return
}

// loghelper
func closeWithLogs(client *torrent.Client) {
	if client != nil {
		logger.Log(false, "Closing torrent client")
		client.Close()
	}
}
