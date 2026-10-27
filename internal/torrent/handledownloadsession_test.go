package torrent

import (
	"os"
	"strings"
	"testing"
	"time"

	"opforjellyfin/internal/shared"
)

func TestDownloadAllSummaryKeepsFailuresWithoutPlacementList(t *testing.T) {
	shared.SetDownloadTotal(3)
	t.Cleanup(shared.ClearActiveDownloads)
	shared.RecordPlacedVideo()
	shared.RecordPlacedVideo()
	results := []*shared.TorrentDownload{
		{Title: "Batch", PlacementFull: []string{"Placed: A.mkv", "Placed: B.mkv", "Left in C.mkv: destination already exists"}},
	}
	out, err := os.CreateTemp(t.TempDir(), "output")
	if err != nil {
		t.Fatal(err)
	}
	previous := os.Stdout
	os.Stdout = out
	defer func() { os.Stdout = previous; out.Close() }()
	printDownloadResults(results, true, false, 65*time.Second)
	data, err := os.ReadFile(out.Name())
	if err != nil {
		t.Fatal(err)
	}
	output := string(data)
	if strings.Contains(output, "Placed:") || strings.Contains(output, "🎞️") {
		t.Fatalf("download all printed the placement list: %q", output)
	}
	for _, want := range []string{"Left in C.mkv: destination already exists", "2/3 files", "1m5s", "1 remaining"} {
		if !strings.Contains(output, want) {
			t.Fatalf("missing %q: %q", want, output)
		}
	}
}
