package matcher

import (
	"os"
	"path/filepath"
	"testing"

	"opforjellyfin/internal/shared"
)

func TestTotalProgressCountsOnlyPlacedVideos(t *testing.T) {
	shared.ClearActiveDownloads()
	t.Cleanup(shared.ClearActiveDownloads)
	shared.SetDownloadTotal(2)
	source, target := t.TempDir(), t.TempDir()
	if err := os.Mkdir(filepath.Join(target, "Arc"), 0755); err != nil {
		t.Fatal(err)
	}
	for filename, contents := range map[string]string{
		filepath.Join(source, "A.mkv"):        "video A",
		filepath.Join(source, "B.mkv"):        "video B",
		filepath.Join(target, "Arc", "A.nfo"): "metadata A",
	} {
		if err := os.WriteFile(filename, []byte(contents), 0644); err != nil {
			t.Fatal(err)
		}
	}
	index := &shared.MetadataIndex{
		ByFilename: map[string]string{"A": "a", "B": "b"},
		Entries: map[string]shared.MetadataFile{
			"a": {NFOPath: "Arc/A.nfo"}, "b": {NFOPath: "Arc/B.nfo"},
		},
	}
	td := &shared.TorrentDownload{TorrentID: 1, VideoPaths: []string{filepath.Join(source, "A.mkv"), filepath.Join(source, "B.mkv")}}
	if err := ProcessTorrentFiles(target, td, index); err == nil {
		t.Fatal("expected placement failure for B's missing NFO")
	}
	shared.RemoveActiveDownload(td.TorrentID)
	if placed, total := shared.GetDownloadProgress(); placed != 1 || total != 2 {
		t.Fatalf("got %d/%d placed, want 1/2 after batch leaves active downloads", placed, total)
	}
	shared.ClearActiveDownloads()
	if placed, total := shared.GetDownloadProgress(); placed != 0 || total != 0 {
		t.Fatalf("progress not reset: %d/%d", placed, total)
	}
}
