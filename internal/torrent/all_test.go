package torrent

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"opforjellyfin/internal/shared"

	"github.com/anacrolix/torrent/bencode"
	"github.com/anacrolix/torrent/metainfo"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func TestSelectAllDownloads(t *testing.T) {
	var index shared.MetadataIndex
	err := json.Unmarshal([]byte(`{
		"arcs":{"arc":{"folder":"Arc"},"specials":{"folder":"Specials"}},
		"by_filename":{"A [1080p]":"a","A fixed [1080p]":"a","A old [1080p]":"old","B [1080p]":"b","B Extended [720p]":"ext","C [1080p]":"c","Special 199-201 [720p]":"sp"},
		"entries":{
			"a":{"nfo_path":"Arc/A [1080p].nfo","episode_id":"ep1","kind":"episode","variant":"standard"},
			"old":{"nfo_path":"Arc/A old [1080p].nfo","episode_id":"ep1","kind":"episode","variant":"alternate"},
			"b":{"nfo_path":"Arc/B [1080p].nfo","episode_id":"ep2","kind":"episode","variant":"standard"},
			"ext":{"nfo_path":"Arc/B Extended [720p].nfo","episode_id":"ep2","kind":"episode","variant":"extended"},
			"c":{"nfo_path":"Arc/C [1080p].nfo","episode_id":"ep3","kind":"episode","variant":"standard"},
			"sp":{"nfo_path":"Specials/Special 199-201 [720p].nfo","episode_id":"special1","kind":"special","variant":"standard"}
		}
	}`), &index)
	if err != nil {
		t.Fatal(err)
	}
	installManifests(t, map[int]metainfo.Info{
		1: {Name: "Batch", Files: []metainfo.FileInfo{
			{Path: []string{"A [1080p].mkv"}, Length: 10}, {Path: []string{"B [1080p].mkv"}, Length: 10},
			{Path: []string{"B Extended [720p].mkv"}, Length: 10}, {Path: []string{"C [1080p].mkv"}, Length: 10},
			{Path: []string{"Special 199-201 [720p].mkv"}, Length: 10}, {Path: []string{"Unknown.mkv"}, Length: 10},
			{Path: []string{"readme.txt"}, Length: 10},
		}},
		2: {Name: "B [1080p].mkv", Length: 10},
		3: {Name: "C [1080p].mkv", Length: 10},
	})
	entries := []shared.TorrentEntry{
		{TorrentID: 3, Seeders: 2}, {TorrentID: 1, Seeders: 20, HaveIt: 1},
		{TorrentID: 2, Seeders: 30}, {TorrentID: 1, Seeders: 20},
	}
	for _, tc := range []struct {
		name               string
		extended, specials bool
		owned              string
		want               map[int][]string
	}{
		{"default", false, false, "A fixed [1080p]", map[int][]string{2: {"B [1080p].mkv"}, 1: {"Batch/C [1080p].mkv"}}},
		{"extended", true, false, "A old [1080p]", map[int][]string{2: {"B [1080p].mkv"}, 1: {"Batch/B Extended [720p].mkv", "Batch/C [1080p].mkv"}}},
		{"specials", false, true, "A [1080p]", map[int][]string{2: {"B [1080p].mkv"}, 1: {"Batch/C [1080p].mkv", "Batch/Special 199-201 [720p].mkv"}}},
		{"both", true, true, "A [1080p]", map[int][]string{2: {"B [1080p].mkv"}, 1: {"Batch/B Extended [720p].mkv", "Batch/C [1080p].mkv", "Batch/Special 199-201 [720p].mkv"}}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			cfg := &shared.Config{TargetDir: t.TempDir(), Source: shared.ScraperConfig{BaseURL: "https://tracker.test"}}
			if err := os.Mkdir(filepath.Join(cfg.TargetDir, "Arc"), 0755); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(cfg.TargetDir, "Arc", tc.owned+".mkv"), []byte("owned"), 0644); err != nil {
				t.Fatal(err)
			}
			selected, err := SelectAllDownloads(context.Background(), entries, cfg, &index, tc.extended, tc.specials)
			if err != nil {
				t.Fatal(err)
			}
			got := map[int][]string{}
			for _, entry := range selected {
				got[entry.TorrentID] = entry.SelectedFiles
			}
			if !reflect.DeepEqual(got, tc.want) {
				t.Fatalf("selected %v; want %v", got, tc.want)
			}
		})
	}
}

func TestSelectAllDownloadsAlreadyOwned(t *testing.T) {
	var index shared.MetadataIndex
	if err := json.Unmarshal([]byte(`{"arcs":{"arc":{"folder":"Arc"}},"by_filename":{"A [1080p]":"a"},"entries":{"a":{"nfo_path":"Arc/A [1080p].nfo","episode_id":"ep1","kind":"episode","variant":"standard"}}}`), &index); err != nil {
		t.Fatal(err)
	}
	cfg := &shared.Config{TargetDir: t.TempDir(), Source: shared.ScraperConfig{BaseURL: "https://tracker.test"}}
	if err := os.Mkdir(filepath.Join(cfg.TargetDir, "Arc"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(cfg.TargetDir, "Arc", "A [1080p].mkv"), []byte("owned"), 0644); err != nil {
		t.Fatal(err)
	}
	installManifests(t, nil) // No manifest needs fetching when everything is owned.
	selected, err := SelectAllDownloads(context.Background(), []shared.TorrentEntry{{TorrentID: 1, Seeders: 10}}, cfg, &index, false, false)
	if err != nil || len(selected) != 0 {
		t.Fatalf("selected %v, error %v", selected, err)
	}
}

func TestSelectAllDownloadsPrefersHDWith720pFallback(t *testing.T) {
	index := &shared.MetadataIndex{
		Arcs:       map[string]shared.MetadataArc{"arc": {Folder: "Arc"}},
		ByFilename: make(map[string]string),
		Entries:    make(map[string]shared.MetadataFile),
	}
	for stem, episode := range map[string]string{
		"A [720p]": "a", "A [1080p]": "a", "A [2160p]": "a",
		"B [2160P]": "b", "B [1080p]": "b", "C [720p]": "c",
		"C [1080p]": "c", "D": "d", "E [1440p]": "e", "F [480p]": "f",
	} {
		index.ByFilename[stem] = stem
		index.Entries[stem] = shared.MetadataFile{NFOPath: "Arc/" + stem + ".nfo", EpisodeID: episode, Kind: "episode", Variant: "standard"}
	}
	installManifests(t, map[int]metainfo.Info{
		1: {Name: "Batch [1080p]", Files: []metainfo.FileInfo{
			{Path: []string{"A [720p].mkv"}, Length: 10},
			{Path: []string{"B [2160P].mkv"}, Length: 10},
			{Path: []string{"C [720p].mkv"}, Length: 10},
			{Path: []string{"D.mkv"}, Length: 10},
			{Path: []string{"E [1440p].mp4"}, Length: 10},
			{Path: []string{"F [480p].mkv"}, Length: 10},
		}},
		2: {Name: "A [1080p].mkv", Length: 10},
		3: {Name: "A [2160p].mkv", Length: 10},
		4: {Name: "B [1080p].mkv", Length: 10},
		5: {Name: "C [720p].mkv", Length: 10},
	})
	entries := []shared.TorrentEntry{
		{TorrentID: 1, Seeders: 100, Quality: "1080p"},
		{TorrentID: 2, Seeders: 30, Quality: "1080p"},
		{TorrentID: 3, Seeders: 20, Quality: "n/a"},
		{TorrentID: 4, Seeders: 10, Quality: "1080p"},
		{TorrentID: 5, Seeders: 50, Quality: "720p"},
		{TorrentID: 6, Seeders: 0, Quality: "1080p"},
	}
	cfg := &shared.Config{TargetDir: t.TempDir(), Source: shared.ScraperConfig{BaseURL: "https://tracker.test"}}
	selected, err := SelectAllDownloads(context.Background(), entries, cfg, index, false, false)
	if err != nil {
		t.Fatal(err)
	}
	got := make(map[int][]string)
	for _, entry := range selected {
		if _, exists := got[entry.TorrentID]; exists {
			t.Fatalf("torrent %d selected more than once", entry.TorrentID)
		}
		got[entry.TorrentID] = entry.SelectedFiles
	}
	want := map[int][]string{
		1: {"Batch [1080p]/B [2160P].mkv", "Batch [1080p]/C [720p].mkv", "Batch [1080p]/E [1440p].mp4"},
		2: {"A [1080p].mkv"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("selected %v; want %v", got, want)
	}
}

func installManifests(t *testing.T, manifests map[int]metainfo.Info) {
	t.Helper()
	previous := http.DefaultClient
	t.Cleanup(func() { http.DefaultClient = previous })
	http.DefaultClient = &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		var id int
		if _, err := fmt.Sscanf(req.URL.Path, "/download/%d.torrent", &id); err != nil {
			return nil, err
		}
		info, ok := manifests[id]
		if !ok {
			t.Errorf("unexpected manifest request: %s", req.URL.Path)
			return nil, fmt.Errorf("missing fixture")
		}
		infoBytes, err := bencode.Marshal(info)
		if err != nil {
			return nil, err
		}
		var body bytes.Buffer
		meta := metainfo.MetaInfo{InfoBytes: infoBytes}
		if err := meta.Write(&body); err != nil {
			return nil, err
		}
		return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(&body), Header: make(http.Header)}, nil
	})}
}
