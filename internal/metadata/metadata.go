// metadata/metadata.go
package metadata

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"

	"opforjellyfin/internal/shared"
	"opforjellyfin/internal/ui"
)

// TODO: sort this file, and mutex for reading/writing to cache

var (
	metadataCache     *shared.MetadataIndex
	metadataCacheOnce sync.Once
)

// FetchAllMetadata clones and indexes metadata from GitHub.
func FetchAllMetadata(baseDir string, cfg shared.Config) error {
	return cloneAndCopyRepo(baseDir, cfg, false)
}

// SyncMetadata clones and syncs metadata updates from GitHub.
func SyncMetadata(baseDir string, cfg shared.Config) error {
	return cloneAndCopyRepo(baseDir, cfg, true)
}

// Main dataobtainer, builds or rebuilds index when complete.
func cloneAndCopyRepo(baseDir string, cfg shared.Config, syncOnly bool) error {
	tmpDir := filepath.Join(os.TempDir(), "repo-tmp")
	defer os.RemoveAll(tmpDir)

	repo := fmt.Sprintf("https://github.com/%s.git", cfg.GitHubRepo)

	fmt.Printf("🌐 Fetching metadata from " + repo + "\n")

	spinner := ui.NewSpinner("🗃️ Downloading.. ", ui.Animations["MetaFetcher"])

	if err := exec.Command("git", "clone", "--depth=1", repo, tmpDir).Run(); err != nil {
		spinner.Stop()
		return fmt.Errorf("git clone failed: %w", err)
	}

	srcDir := filepath.Join(tmpDir, "One Pace")
	var err error

	if syncOnly {
		err = syncDir(srcDir, baseDir)
	} else {
		err = copyDir(srcDir, baseDir)
	}

	if err != nil {
		spinner.Stop()
		return fmt.Errorf("failed to copy metadata: %w", err)
	}

	if err := BuildMetadataIndex(baseDir); err != nil {
		spinner.Stop()
		return fmt.Errorf("failed to build metadata index: %w", err)
	}

	spinner.Stop()

	path := filepath.Join(baseDir, "metadata-index.json")
	fmt.Println("\n✅ Saved metadata index to", path)

	fmt.Println("✅ Metadata fetch and indexing complete.")
	return nil
}

// BuildMetadataIndex constructs and caches metadata index.
func BuildMetadataIndex(baseDir string) error {
	index, err := buildIndexFromDir(baseDir)
	if err != nil {
		return err
	}

	return saveMetadataIndex(index, baseDir)
}

func buildIndexFromDir(baseDir string) (*shared.MetadataIndex, error) {
	index := &shared.MetadataIndex{
		Seasons: make(map[string]shared.SeasonIndex),
	}

	err := filepath.WalkDir(baseDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || !isEpisodeNFO(d.Name()) {
			return nil
		}

		data, err := os.ReadFile(path)
		if err != nil {
			return nil
		}

		season, episode, chapterRange := extractEpisodeMetadata(data)
		if season == "" || episode == "" || chapterRange == "" {
			return nil
		}

		seasonKey := fmt.Sprintf("Season %s", season)
		if season == "00" || season == "0" {
			seasonKey = "Specials"
		}

		// chapterRange used by index
		normalized := shared.NormalizeDash(chapterRange)
		// filename withouth .nfo for the index
		epTitle := strings.TrimSuffix(d.Name(), ".nfo")

		// check if SeasonIndex is there
		if _, exists := index.Seasons[seasonKey]; !exists {
			index.Seasons[seasonKey] = shared.SeasonIndex{
				EpisodeRange: make(map[string]shared.EpisodeData),
			}
		}
		// just store title, use baseDir+seasonKey+epTitle+mp4/mkv for storing
		index.Seasons[seasonKey].EpisodeRange[normalized] = shared.EpisodeData{
			Title: epTitle,
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("metadata indexing failed: %w", err)
	}

	// put range on season
	calculateSeasonRanges(index)

	return index, nil
}

// saves file and creates cache
func saveMetadataIndex(index *shared.MetadataIndex, baseDir string) error {
	path := filepath.Join(baseDir, "metadata-index.json")
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("could not create index file: %w", err)
	}
	defer f.Close()

	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	if err := enc.Encode(index); err != nil {
		return fmt.Errorf("could not encode metadata index: %w", err)
	}

	metadataCache = index // cache immediately after saving

	return nil
}

// HaveMetadata checks if metadata exists for given chapterRange.
func HaveMetadata(chapterRange string) bool {
	if chapterRange == "" {
		return false
	}

	LoadMetadataCache()
	norm := shared.NormalizeDash(chapterRange)

	for _, season := range metadataCache.Seasons {
		// season range match instantly
		if shared.NormalizeDash(season.Range) == norm {
			return true
		}

		// match individual episodes
		for epRange := range season.EpisodeRange {
			if shared.NormalizeDash(epRange) == norm {
				return true
			}
		}
	}

	return false
}

// checks a range in metadata. 0 = does not have, 1 = have some, 2 = have all
func HaveVideoStatus(chapterRange string) int {
	if chapterRange == "" {
		return 0
	}

	index := LoadMetadataCache()
	cfg := shared.LoadConfig()
	baseDir := cfg.TargetDir

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

				videoPathMP4 := filepath.Join(seasonDir, epData.Title+".mp4")
				videoPathMKV := filepath.Join(seasonDir, epData.Title+".mkv")

				if shared.FileExists(videoPathMP4) || shared.FileExists(videoPathMKV) {
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

// more helpers

// we read metadata-index.json once when we need it, and if multiple checks are needed we read the cache.
// returns a reference to the index-variable
func LoadMetadataCache() *shared.MetadataIndex {
	metadataCacheOnce.Do(func() {
		cfg := shared.LoadConfig()
		data, err := os.ReadFile(filepath.Join(cfg.TargetDir, "metadata-index.json"))
		if err != nil {
			metadataCache = &shared.MetadataIndex{}
			return
		}
		json.Unmarshal(data, &metadataCache)
	})

	return metadataCache
}

// helpers
func calculateSeasonRanges(index *shared.MetadataIndex) {

	for skey, sidx := range index.Seasons {

		if skey == "Specials" {
			sidx.Range = "00-00"
			index.Seasons[skey] = sidx
			continue
		}

		min, max := 99999, -1
		for cr := range sidx.EpisodeRange {
			start, end := shared.ParseRange(cr)
			if start < min {
				min = start
			}
			if end > max {
				max = end
			}
		}
		sidx.Range = fmt.Sprintf("%d-%d", min, max)
		index.Seasons[skey] = sidx
	}
}

func isEpisodeNFO(filename string) bool {
	return strings.HasSuffix(filename, ".nfo") && !strings.Contains(filename, "season") && !strings.Contains(filename, "tvshow")
}

// important
func extractEpisodeMetadata(data []byte) (string, string, string) {
	return shared.ExtractXMLTag(data, "season"), shared.ExtractXMLTag(data, "episode"), shared.ExtractChapterRangeFromNFO(string(data))
}

// copyDir copies all files from src to dst
func copyDir(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}

		destPath := filepath.Join(dst, relPath)

		if info.IsDir() {
			return os.MkdirAll(destPath, 0755)
		}

		return shared.CopyFile(path, destPath, info.Mode())
	})
}

// syncDir copies new or changed files from src to dst
func syncDir(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}

		destPath := filepath.Join(dst, relPath)

		if info.IsDir() {
			return os.MkdirAll(destPath, 0755)
		}

		if _, err := os.Stat(destPath); err == nil {
			existingData, err1 := os.ReadFile(destPath)
			newData, err2 := os.ReadFile(path)
			if err1 == nil && err2 == nil && string(existingData) == string(newData) {
				return nil // identical, skip
			}
		}

		return shared.CopyFile(path, destPath, info.Mode())
	})
}
