package metadata

import (
	"testing"

	"opforjellyfin/internal/shared"
)

func TestFindArcMatchesNameCaseInsensitively(t *testing.T) {
	index := &shared.MetadataIndex{
		Seasons: map[string]shared.SeasonIndex{
			"Season 4": {
				Name:  "Skypiea",
				Range: "237-302",
				EpisodeRange: map[string]shared.EpisodeData{
					"237-238": {Title: "S04E01 - The Knock Up Stream"},
				},
			},
		},
	}

	arc, err := FindArc(index, "skypiea")
	if err != nil {
		t.Fatalf("FindArc() error = %v", err)
	}
	if arc.Name != "Skypiea" {
		t.Fatalf("arc.Name = %q, want Skypiea", arc.Name)
	}
}

func TestFindArcAllowsCloseMisspelling(t *testing.T) {
	index := &shared.MetadataIndex{
		Seasons: map[string]shared.SeasonIndex{
			"Season 4": {
				Name:  "Skypiea",
				Range: "237-302",
				EpisodeRange: map[string]shared.EpisodeData{
					"237-238": {Title: "S04E01 - The Knock Up Stream"},
				},
			},
		},
	}

	arc, err := FindArc(index, "Skypia")
	if err != nil {
		t.Fatalf("FindArc() error = %v", err)
	}
	if arc.Name != "Skypiea" {
		t.Fatalf("arc.Name = %q, want Skypiea", arc.Name)
	}
}
