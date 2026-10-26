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
		"by_filename":{"A":"a","A fixed":"a","A old":"old","B":"b","B Extended":"ext","C":"c","Special 199-201":"sp"},
		"entries":{
			"a":{"nfo_path":"Arc/A.nfo","episode_id":"ep1","kind":"episode","variant":"standard"},
			"old":{"nfo_path":"Arc/A old.nfo","episode_id":"ep1","kind":"episode","variant":"alternate"},
			"b":{"nfo_path":"Arc/B.nfo","episode_id":"ep2","kind":"episode","variant":"standard"},
			"ext":{"nfo_path":"Arc/B Extended.nfo","episode_id":"ep2","kind":"episode","variant":"extended"},
			"c":{"nfo_path":"Arc/C.nfo","episode_id":"ep3","kind":"episode","variant":"standard"},
			"sp":{"nfo_path":"Specials/Special 199-201.nfo","episode_id":"special1","kind":"special","variant":"standard"}
		}
	}`), &index)
	if err != nil {
		t.Fatal(err)
	}
	installManifests(t, map[int]metainfo.Info{
		1: {Name: "Batch", Files: []metainfo.FileInfo{
			{Path: []string{"A.mkv"}, Length: 10}, {Path: []string{"B.mkv"}, Length: 10},
			{Path: []string{"B Extended.mkv"}, Length: 10}, {Path: []string{"C.mkv"}, Length: 10},
			{Path: []string{"Special 199-201.mkv"}, Length: 10}, {Path: []string{"Unknown.mkv"}, Length: 10},
			{Path: []string{"readme.txt"}, Length: 10},
		}},
		2: {Name: "B.mkv", Length: 10},
		3: {Name: "C.mkv", Length: 10},
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
		{"default", false, false, "A fixed", map[int][]string{2: {"B.mkv"}, 1: {"Batch/C.mkv"}}},
		{"extended", true, false, "A old", map[int][]string{2: {"B.mkv"}, 1: {"Batch/B Extended.mkv", "Batch/C.mkv"}}},
		{"specials", false, true, "A", map[int][]string{2: {"B.mkv"}, 1: {"Batch/C.mkv", "Batch/Special 199-201.mkv"}}},
		{"both", true, true, "A", map[int][]string{2: {"B.mkv"}, 1: {"Batch/B Extended.mkv", "Batch/C.mkv", "Batch/Special 199-201.mkv"}}},
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
	if err := json.Unmarshal([]byte(`{"arcs":{"arc":{"folder":"Arc"}},"by_filename":{"A":"a"},"entries":{"a":{"nfo_path":"Arc/A.nfo","episode_id":"ep1","kind":"episode","variant":"standard"}}}`), &index); err != nil {
		t.Fatal(err)
	}
	cfg := &shared.Config{TargetDir: t.TempDir(), Source: shared.ScraperConfig{BaseURL: "https://tracker.test"}}
	if err := os.Mkdir(filepath.Join(cfg.TargetDir, "Arc"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(cfg.TargetDir, "Arc", "A.mkv"), []byte("owned"), 0644); err != nil {
		t.Fatal(err)
	}
	installManifests(t, nil) // No manifest needs fetching when everything is owned.
	selected, err := SelectAllDownloads(context.Background(), []shared.TorrentEntry{{TorrentID: 1, Seeders: 10}}, cfg, &index, false, false)
	if err != nil || len(selected) != 0 {
		t.Fatalf("selected %v, error %v", selected, err)
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
