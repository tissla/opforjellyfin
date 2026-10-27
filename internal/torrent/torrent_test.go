package torrent

import (
	"context"
	"crypto/sha1"
	"os"
	"path/filepath"
	"testing"
	"time"

	"opforjellyfin/internal/shared"

	anacrolix "github.com/anacrolix/torrent"
	"github.com/anacrolix/torrent/bencode"
	"github.com/anacrolix/torrent/metainfo"
)

func TestSelectedFilesCompleteWithoutRestOfBatch(t *testing.T) {
	cfg := anacrolix.NewDefaultClientConfig()
	cfg.DataDir = t.TempDir()
	cfg.DisableTCP, cfg.DisableUTP, cfg.NoDHT = true, true, true
	cfg.DisableTrackers, cfg.DisableWebtorrent, cfg.NoDefaultPortForwarding = true, true, true
	if err := os.Mkdir(filepath.Join(cfg.DataDir, "Batch"), 0755); err != nil {
		t.Fatal(err)
	}
	selectedPath := filepath.Join(cfg.DataDir, "Batch", "A.mkv")
	if err := os.WriteFile(selectedPath, []byte("abcd"), 0644); err != nil {
		t.Fatal(err)
	}
	client, err := anacrolix.NewClient(cfg)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { client.Close() })
	hashA, hashB := sha1.Sum([]byte("abcd")), sha1.Sum([]byte("efgh"))
	infoBytes, err := bencode.Marshal(metainfo.Info{
		Name: "Batch", PieceLength: 4, Pieces: append(append(hashA[:], hashB[:]...), hashB[:]...),
		Files: []metainfo.FileInfo{
			{Path: []string{"A.mkv"}, Length: 4},
			{Path: []string{"B.MP4"}, Length: 4},
			{Path: []string{"episode.mkv.exe"}, Length: 4},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	tr, err := client.AddTorrent(&metainfo.MetaInfo{InfoBytes: infoBytes})
	if err != nil {
		t.Fatal(err)
	}
	tr.VerifyData()
	selected, err := filesToDownload(tr, []string{"Batch/A.mkv"})
	if err != nil || len(selected) != 1 {
		t.Fatalf("selected %v, error %v", selected, err)
	}
	if progress, complete := fileProgress(selected); !complete || progress != 4 {
		t.Fatalf("selected file: %d bytes, complete=%v", progress, complete)
	}
	all, err := filesToDownload(tr, nil)
	if err != nil || len(all) != 2 {
		t.Fatalf("manual download: %v, error %v", all, err)
	}
	if all[0].Path() != "Batch/A.mkv" || all[1].Path() != "Batch/B.MP4" {
		t.Fatalf("unexpected video selection: %v", all)
	}
	if _, complete := fileProgress(all); complete {
		t.Fatal("whole torrent should still be incomplete")
	}
	for _, paths := range [][]string{{}, {"Batch/absent.mkv"}, {"Batch/episode.mkv.exe"}} {
		if _, err := filesToDownload(tr, paths); err == nil {
			t.Fatalf("invalid selection accepted: %v", paths)
		}
	}
	if err := os.WriteFile(selectedPath, []byte("xxxx"), 0644); err != nil {
		t.Fatal(err)
	}
	tr.VerifyData()
	if _, complete := fileProgress(selected); complete {
		t.Fatal("corrupted file must not count as complete")
	}

	t.Cleanup(shared.ClearActiveDownloads)
	td := &shared.TorrentDownload{TorrentID: 1}
	t.Run("stalled download times out", func(t *testing.T) {
		if err := waitForDownload(context.Background(), td, all, 40*time.Millisecond); err != context.DeadlineExceeded {
			t.Fatalf("stalled download: %v", err)
		}
	})
	t.Run("cancelled download stops", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		if err := waitForDownload(ctx, td, all, time.Minute); err != context.Canceled {
			t.Fatalf("cancelled download: %v", err)
		}
	})
	t.Run("progress extends timeout", func(t *testing.T) {
		written := make(chan error, 1)
		go func() {
			time.Sleep(500 * time.Millisecond)
			if err := os.WriteFile(selectedPath, []byte("abcd"), 0644); err != nil {
				written <- err
				return
			}
			tr.VerifyData()
			time.Sleep(1300 * time.Millisecond)
			if err := os.WriteFile(filepath.Join(cfg.DataDir, "Batch", "B.MP4"), []byte("efgh"), 0644); err != nil {
				written <- err
				return
			}
			tr.VerifyData()
			written <- nil
		}()
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		err := waitForDownload(ctx, td, all, 1500*time.Millisecond)
		if writeErr := <-written; writeErr != nil {
			t.Fatal(writeErr)
		}
		if err != nil {
			t.Fatalf("download made progress but stopped: %v", err)
		}
		if td.Progress != 8 {
			t.Fatalf("finished with %d bytes, want 8", td.Progress)
		}
	})
}
