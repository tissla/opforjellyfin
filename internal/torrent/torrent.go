// torrent/torrent.go
package torrent

import (
	"context"

	"fmt"
	"net/http"
	"opforjellyfin/internal/logger"
	"opforjellyfin/internal/shared"
	"os"
	"path/filepath"

	"time"

	"github.com/anacrolix/torrent"
	"github.com/anacrolix/torrent/metainfo"
)

// used for writing to json-file for live tracking of concurrent background downloads. currently just saves progress

// main torrent download and tracker
func StartTorrent(ctx context.Context, td *shared.TorrentDownload) error {

	// get torrent meta-info
	torrentURL := fmt.Sprintf("%s/download/%d.torrent", shared.LoadConfig().TorrentAPIURL, td.TorrentID)
	logger.DebugLog(false, "Fetching torrent: %s, ID: %s", torrentURL, td)

	// get metadata
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, torrentURL, nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		logger.DebugLog(false, "HTTP request for metadata failed %s", td)
		return err
	}
	defer resp.Body.Close()

	//build meta
	meta, err := metainfo.Load(resp.Body)
	if err != nil {
		return err
	}

	// create tempdir
	tmpDir := filepath.Join(os.TempDir(), fmt.Sprintf("opfor-tmp-%d", td.TorrentID))
	if err := os.MkdirAll(tmpDir, 0755); err != nil {
		return err
	}

	// start the torrent-client
	cfg := torrent.NewDefaultClientConfig()
	cfg.DataDir = tmpDir
	cfg.NoUpload = true
	cfg.ListenPort = 0

	client, err := torrent.NewClient(cfg)
	if err != nil {
		return err
	}

	// add torrent
	t, err := client.AddTorrent(meta)
	if err != nil {
		return err
	}

	// get torrent metadata
	select {
	case <-t.GotInfo():
		td.TotalSize = t.Length()
		logger.DebugLog(false, "Torrent metadata loaded: %s", td)
	case <-time.After(20 * time.Second):
		return err
	case <-ctx.Done():
		return ctx.Err()
	}

	// start download
	td.TotalSize = t.Length()

	t.DownloadAll()
	shared.SaveTorrentDownload(td)

	// watch progress, save to activefile
	for t.BytesMissing() > 0 {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(1 * time.Second):
			td.Progress = t.BytesCompleted()
			shared.SaveTorrentDownload(td)
		}
	}

	// close
	td.Progress = td.TotalSize
	logger.DebugLog(false, "Torrent contains %d files", len(t.Files()))

	closeWithLogs(client)
	td.Done = true
	td.PlacementProgress = "⏳ Waiting to place.."
	shared.SaveTorrentDownload(td)
	logger.DebugLog(false, "Download complete: %s", td)

	return nil
}

// loghelper
func closeWithLogs(client *torrent.Client) {
	logger.DebugLog(false, "Torrentclient closed for: %s", client)
	client.Close()
}
