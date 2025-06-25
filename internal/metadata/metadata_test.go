package metadata

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"opforjellyfin/internal/shared"
)

// testable HaveVideoStatus
func HaveVideoStatusWithIndex(index *shared.MetadataIndex, baseDir, chapterRange string) int {
	if chapterRange == "" {
		return 0
	}

	norm := shared.NormalizeDash(chapterRange)
	targetStart, targetEnd := shared.ParseRange(norm)

	totalRelevant := 0
	totalFound := 0

	for seasonKey, season := range index.Seasons {
		seasonDir := filepath.Join(baseDir, seasonKey)

		for epRange, epData := range season.EpisodeRange {
			epStart, epEnd := shared.ParseRange(epRange)

			if epStart >= targetStart && epEnd <= targetEnd {
				totalRelevant++
				mp4 := filepath.Join(seasonDir, epData.Title+".mp4")
				mkv := filepath.Join(seasonDir, epData.Title+".mkv")
				if shared.FileExists(mp4) || shared.FileExists(mkv) {
					totalFound++
				}
			}
		}
	}

	switch {
	case totalRelevant == 0:
		return 0
	case totalFound == 0:
		return 0
	case totalFound < totalRelevant:
		return 1
	default:
		return 2
	}
}

func TestHaveVideoStatus(t *testing.T) {
	tmpDir := t.TempDir()

	season := shared.SeasonIndex{
		Range: "1-2",
		EpisodeRange: map[string]shared.EpisodeData{
			"1-1": {Title: "Episode_1"},
			"2-2": {Title: "Episode_2"},
		},
	}

	index := shared.MetadataIndex{
		Seasons: map[string]shared.SeasonIndex{
			"Season 01": season,
		},
	}

	indexPath := filepath.Join(tmpDir, "metadata-index.json")
	f, err := os.Create(indexPath)
	if err != nil {
		t.Fatalf("Failed to create index file: %v", err)
	}
	if err := json.NewEncoder(f).Encode(index); err != nil {
		t.Fatalf("Failed to write index: %v", err)
	}
	f.Close()

	seasonDir := filepath.Join(tmpDir, "Season 01")
	err = os.MkdirAll(seasonDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create season dir: %v", err)
	}

	videoFile := filepath.Join(seasonDir, "Episode_1.mp4")
	err = os.WriteFile(videoFile, []byte("dummy"), 0644)
	if err != nil {
		t.Fatalf("Failed to write dummy video: %v", err)
	}

	tests := []struct {
		rangeStr string
		want     int
	}{
		{"1-1", 2},
		{"2-2", 0},
		{"1-2", 1},
		{"3-4", 0},
		{"", 0},
	}

	for _, tt := range tests {
		got := HaveVideoStatusWithIndex(&index, tmpDir, tt.rangeStr)
		if got != tt.want {
			t.Errorf("HaveVideoStatus(%q) = %d, want %d", tt.rangeStr, got, tt.want)
		}
	}
}

func TestCalculateSeasonRanges(t *testing.T) {
	index := shared.MetadataIndex{
		Seasons: map[string]shared.SeasonIndex{
			"Season 1": {
				EpisodeRange: map[string]shared.EpisodeData{
					"1-1":   {Title: "ep1"},
					"2-2":   {Title: "ep2"},
					"5-5":   {Title: "ep5"},
					"10-10": {Title: "ep10"},
				},
			},
			"Season 2": {
				EpisodeRange: map[string]shared.EpisodeData{
					"42-22": {Title: "gaimon?"},
				},
			},
			"Specials": {
				EpisodeRange: map[string]shared.EpisodeData{
					"900-999": {Title: "special"},
				},
			},
		},
	}

	calculateSeasonRanges(&index)

	got := index.Seasons["Season 1"].Range
	if got != "1-10" {
		t.Errorf("expected Season 1 range to be 1-10, got %s", got)
	}

	got = index.Seasons["Season 2"].Range
	if got != "22-42" && got != "42-42" {
		t.Errorf("expected Season 2 range to be 42-42 or 22-42, got %s", got)
	}

	got = index.Seasons["Specials"].Range
	if got != "00-00" {
		t.Errorf("expected Specials range to be 00-00, got %s", got)
	}
}
