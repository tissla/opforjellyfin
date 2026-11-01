package matcher

import (
	"os"
	"path/filepath"
	"testing"

	"opforjellyfin/internal/shared"
)

func TestHardlinkPlacementKeepsSource(t *testing.T) {
	source, target, destination, index := placementFixture(t)
	message, err := MatchAndPlaceVideo(source, target, index, true)
	if err != nil || message == "" {
		t.Fatalf("place: message=%q, err=%v", message, err)
	}
	sourceInfo, err := os.Stat(source)
	if err != nil {
		t.Fatalf("source must remain available for seeding: %v", err)
	}
	destInfo, err := os.Stat(destination)
	if err != nil {
		t.Fatal(err)
	}
	if !os.SameFile(sourceInfo, destInfo) {
		t.Fatal("source and destination must share the same file")
	}
	message, err = MatchAndPlaceVideo(source, target, index, true)
	if err != nil || message != "" {
		t.Fatalf("repeat placement: message=%q, err=%v", message, err)
	}
}

func TestMovePlacementRemovesSource(t *testing.T) {
	source, target, destination, index := placementFixture(t)
	if _, err := MatchAndPlaceVideo(source, target, index, false); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(source); !os.IsNotExist(err) {
		t.Fatalf("source should be removed after a move: %v", err)
	}
	contents, err := os.ReadFile(destination)
	if err != nil || string(contents) != "video" {
		t.Fatalf("destination: contents=%q, err=%v", contents, err)
	}
}

func TestHardlinkPlacementPreservesConflictingFiles(t *testing.T) {
	source, target, destination, index := placementFixture(t)
	if err := os.WriteFile(destination, []byte("existing video"), 0644); err != nil {
		t.Fatal(err)
	}
	if _, err := MatchAndPlaceVideo(source, target, index, true); err == nil {
		t.Fatal("expected a destination conflict")
	}
	for path, want := range map[string]string{source: "video", destination: "existing video"} {
		contents, err := os.ReadFile(path)
		if err != nil || string(contents) != want {
			t.Fatalf("%s changed: contents=%q, err=%v", path, contents, err)
		}
	}
}

func placementFixture(t *testing.T) (string, string, string, *shared.MetadataIndex) {
	t.Helper()
	root := t.TempDir()
	source := filepath.Join(root, "Episode.mkv")
	target := filepath.Join(root, "library")
	arc := filepath.Join(target, "Arc")
	if err := os.MkdirAll(arc, 0755); err != nil {
		t.Fatal(err)
	}
	for path, contents := range map[string]string{
		source:                            "video",
		filepath.Join(arc, "Episode.nfo"): "metadata",
	} {
		if err := os.WriteFile(path, []byte(contents), 0644); err != nil {
			t.Fatal(err)
		}
	}
	index := &shared.MetadataIndex{
		ByFilename: map[string]string{"Episode": "episode"},
		Entries: map[string]shared.MetadataFile{
			"episode": {NFOPath: "Arc/Episode.nfo"},
		},
	}
	return source, target, filepath.Join(arc, "Episode.mkv"), index
}
