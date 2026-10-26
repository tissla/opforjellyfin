package ui

import (
	"os"
	"strings"
	"testing"
	"time"

	"opforjellyfin/internal/logger"
	"opforjellyfin/internal/shared"

	"github.com/charmbracelet/x/ansi"
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
	logger.Log(true, "Download timed out after 30 minutes: Water Seven")
	done <- struct{}{}
	<-done
	data, err := os.ReadFile(out.Name())
	if err != nil {
		t.Fatal(err)
	}
	output := string(data)
	for _, count := range []string{"0/3 episodes", "3 remaining", "1/3 episodes", "2 remaining"} {
		if !strings.Contains(output, count) {
			t.Fatalf("missing total count %q: %q", count, output)
		}
	}
	if !strings.Contains(output, "\033[3A") {
		t.Fatalf("redraw did not account for the fixed download, total and status rows: %q", output)
	}
	if !strings.Contains(output, "Latest: Download timed out") {
		t.Fatalf("missing status row: %q", output)
	}
	if !strings.Contains(output, "\033[?25l") || !strings.Contains(output, "\033[?25h") {
		t.Fatal("cursor was not hidden and restored")
	}
}

func TestDownloadLayoutFitsTerminal(t *testing.T) {
	td := &shared.TorrentDownload{
		Title: "176: Water Seven 08 (1080p)", Progress: 512 << 20, TotalSize: 1 << 30,
	}
	for _, width := range []int{40, 60, 80, 100} {
		line := renderDownload(td, width)
		if ansi.StringWidth(line) > width || strings.ContainsAny(line, "\r\n") {
			t.Fatalf("row wraps at width %d: %q", width, line)
		}
		if !strings.Contains(line, "50%") {
			t.Fatalf("percentage hidden at width %d: %q", width, line)
		}
		if width >= 80 && (!strings.Contains(line, td.Title) || !strings.Contains(line, "1.0 GiB")) {
			t.Fatalf("title or size missing at width %d: %q", width, line)
		}
		totalLine := renderTotalBar(1, 2, width)
		if ansi.StringWidth(totalLine) > width || ansi.StringWidth(line[:strings.Index(line, "━")]) != ansi.StringWidth(totalLine[:strings.Index(totalLine, "━")]) {
			t.Fatalf("total bar wraps or is misaligned at width %d: %q", width, totalLine)
		}
	}
	td.Progress = td.TotalSize
	if line := renderDownload(td, 80); !strings.Contains(line, "Verifying") {
		t.Fatalf("unverified download displayed as complete: %q", line)
	}
}
