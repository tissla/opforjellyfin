package torrent

import (
	"crypto/sha1"
	"os"
	"path/filepath"
	"testing"

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
		Name: "Batch", PieceLength: 4, Pieces: append(hashA[:], hashB[:]...),
		Files: []metainfo.FileInfo{{Path: []string{"A.mkv"}, Length: 4}, {Path: []string{"B.mkv"}, Length: 4}},
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
	if _, complete := fileProgress(all); complete {
		t.Fatal("whole torrent should still be incomplete")
	}
	for _, paths := range [][]string{{}, {"Batch/absent.mkv"}} {
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
}
