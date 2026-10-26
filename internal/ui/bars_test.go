package ui

import (
	"os"
	"strings"
	"testing"
	"time"

	"opforjellyfin/internal/shared"
)

func TestTotalBarSurvivesActiveDownloadRemoval(t *testing.T) {
	shared.ClearActiveDownloads()
	t.Cleanup(shared.ClearActiveDownloads)
	shared.SetDownloadTotal(3)
	shared.SaveTorrentDownload(&shared.TorrentDownload{TorrentID: 1, Title: "active", TotalSize: 100})
	out, err := os.CreateTemp(t.TempDir(), "output")
	if err != nil {
		t.Fatal(err)
	}
	previous := os.Stdout
	os.Stdout = out
	defer func() { os.Stdout = previous; out.Close() }()
	done := make(chan struct{})
	go FollowProgress(done)
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		data, _ := os.ReadFile(out.Name())
		if strings.Contains(string(data), "0/3 episodes") {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	shared.RecordPlacedVideo()
	shared.RemoveActiveDownload(1)
	done <- struct{}{}
	<-done
	data, err := os.ReadFile(out.Name())
	if err != nil {
		t.Fatal(err)
	}
	output := string(data)
	if !strings.Contains(output, "0/3 episodes | 3 remaining") || !strings.Contains(output, "1/3 episodes | 2 remaining") {
		t.Fatalf("missing total counts: %q", output)
	}
	if !strings.Contains(output, "\033[2A") {
		t.Fatalf("redraw did not account for both download and total rows: %q", output)
	}
}
