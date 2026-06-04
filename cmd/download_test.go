package cmd

import (
	"testing"

	"opforjellyfin/internal/metadata"
	"opforjellyfin/internal/shared"
)

func TestSelectArcDownloadsPicksHighestSeededNonOverlappingRanges(t *testing.T) {
	arc := &metadata.ArcMatch{
		Name:   "Long Ring Long Land",
		Season: "Season 5",
		Range:  "303-321",
	}

	torrents := []shared.TorrentEntry{
		{DownloadKey: 1, ChapterRange: "303-303", Seeders: 3, TorrentName: "low"},
		{DownloadKey: 2, ChapterRange: "303-303", Seeders: 10, TorrentName: "high"},
		{DownloadKey: 3, ChapterRange: "304-305", Seeders: 4, TorrentName: "bundle"},
		{DownloadKey: 4, ChapterRange: "304-304", Seeders: 100, TorrentName: "overlap"},
		{DownloadKey: 5, ChapterRange: "400-401", Seeders: 50, TorrentName: "outside"},
	}

	selected := selectArcDownloads(torrents, arc)
	if len(selected) != 2 {
		t.Fatalf("len(selected) = %d, want 2", len(selected))
	}
	if selected[0].DownloadKey != 2 {
		t.Fatalf("selected[0].DownloadKey = %d, want 2", selected[0].DownloadKey)
	}
	if selected[1].DownloadKey != 3 {
		t.Fatalf("selected[1].DownloadKey = %d, want 3", selected[1].DownloadKey)
	}
}
