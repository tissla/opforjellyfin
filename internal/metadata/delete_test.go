package metadata

import (
	"os"
	"path/filepath"
	"testing"

	"opforjellyfin/internal/shared"
)

func TestFindEpisodeVideosForDelete(t *testing.T) {
	baseDir := t.TempDir()
	seasonDir := filepath.Join(baseDir, "Season 1")
	if err := os.MkdirAll(seasonDir, 0755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}

	videoPath := filepath.Join(seasonDir, "S01E01 - Romance Dawn.mkv")
	nfoPath := filepath.Join(seasonDir, "S01E01 - Romance Dawn.nfo")
	for _, path := range []string{videoPath, nfoPath} {
		if err := os.WriteFile(path, []byte("test"), 0644); err != nil {
			t.Fatalf("WriteFile() error = %v", err)
		}
	}

	index := &shared.MetadataIndex{
		Seasons: map[string]shared.SeasonIndex{
			"Season 1": {
				Range: "1-7",
				EpisodeRange: map[string]shared.EpisodeData{
					"1-1": {Title: "S01E01 - Romance Dawn"},
				},
			},
		},
	}

	candidates, missing, err := FindEpisodeVideosForDelete(baseDir, index, []string{"1"})
	if err != nil {
		t.Fatalf("FindEpisodeVideosForDelete() error = %v", err)
	}
	if len(missing) != 0 {
		t.Fatalf("missing = %v, want none", missing)
	}
	if len(candidates) != 1 {
		t.Fatalf("len(candidates) = %d, want 1", len(candidates))
	}
	if candidates[0].Path != videoPath {
		t.Fatalf("candidate path = %q, want %q", candidates[0].Path, videoPath)
	}
}

func TestDeleteEpisodeVideosLeavesMetadata(t *testing.T) {
	baseDir := t.TempDir()
	seasonDir := filepath.Join(baseDir, "Season 1")
	if err := os.MkdirAll(seasonDir, 0755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}

	videoPath := filepath.Join(seasonDir, "S01E01 - Romance Dawn.mp4")
	nfoPath := filepath.Join(seasonDir, "S01E01 - Romance Dawn.nfo")
	for _, path := range []string{videoPath, nfoPath} {
		if err := os.WriteFile(path, []byte("test"), 0644); err != nil {
			t.Fatalf("WriteFile() error = %v", err)
		}
	}

	err := DeleteEpisodeVideos([]DeleteCandidate{{Path: videoPath}})
	if err != nil {
		t.Fatalf("DeleteEpisodeVideos() error = %v", err)
	}
	if shared.FileExists(videoPath) {
		t.Fatalf("video file still exists after delete")
	}
	if !shared.FileExists(nfoPath) {
		t.Fatalf("metadata file was deleted")
	}
}

func TestFindEpisodeVideosForDeleteReportsMissingMetadata(t *testing.T) {
	index := &shared.MetadataIndex{
		Seasons: map[string]shared.SeasonIndex{
			"Season 1": {
				Range:        "1-7",
				EpisodeRange: map[string]shared.EpisodeData{},
			},
		},
	}

	candidates, missing, err := FindEpisodeVideosForDelete(t.TempDir(), index, []string{"99"})
	if err != nil {
		t.Fatalf("FindEpisodeVideosForDelete() error = %v", err)
	}
	if len(candidates) != 0 {
		t.Fatalf("len(candidates) = %d, want 0", len(candidates))
	}
	if len(missing) != 1 || missing[0] != "99-99" {
		t.Fatalf("missing = %v, want [99-99]", missing)
	}
}

func TestFindEpisodeVideosForDeleteRejectsInvalidRange(t *testing.T) {
	index := &shared.MetadataIndex{Seasons: map[string]shared.SeasonIndex{}}

	_, _, err := FindEpisodeVideosForDelete(t.TempDir(), index, []string{"1-x"})
	if err == nil {
		t.Fatalf("FindEpisodeVideosForDelete() error = nil, want error")
	}
}
