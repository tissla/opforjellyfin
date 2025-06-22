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

// TODO: sort this file, and add cool animations where possible

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

	fmt.Println("✅ Metadata index written to", path)

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
	index := &shared.MetadataIndex{Seasons: make(map[string]shared.SeasonIndex)}

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
		epKey := fmt.Sprintf("S%02sE%02s", season, episode)

		if _, exists := index.Seasons[seasonKey]; !exists {
			index.Seasons[seasonKey] = shared.SeasonIndex{Episodes: make(map[string]string)}
		}

		index.Seasons[seasonKey].Episodes[epKey] = chapterRange
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("metadata indexing failed: %w", err)
	}

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
// TODO: make more rigid
func HaveMetadata(chapterRange string) bool {
	if chapterRange == "" {
		return false
	}

	LoadMetadataCache()

	for _, season := range metadataCache.Seasons {

		if season.Range == chapterRange {
			return true
		}
		for _, epRange := range season.Episodes {
			if epRange == chapterRange {
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

	LoadMetadataCache()
	cfg := shared.LoadConfig()
	baseDir := cfg.TargetDir

	targetStart, targetEnd := shared.ParseRange(chapterRange)

	totalRelevant := 0
	totalFound := 0

	for seasonKey, season := range metadataCache.Seasons {
		seasonDir := filepath.Join(baseDir, seasonKey)

		for epKey, epRange := range season.Episodes {
			start, end := shared.ParseRange(epRange)

			if start >= targetStart && end <= targetEnd {
				totalRelevant++

				seasonNum := shared.ExtractSeasonNumberFromKey(epKey)
				episodeNum := shared.ExtractEpisodeNumberFromKey(epKey)
				expectedPrefix := fmt.Sprintf("One Pace - S%sE%s -", seasonNum, episodeNum)

				files, err := os.ReadDir(seasonDir)
				if err != nil {
					continue
				}

				for _, file := range files {
					if file.IsDir() {
						continue
					}
					name := file.Name()
					if strings.HasPrefix(name, expectedPrefix) &&
						(strings.HasSuffix(name, ".mkv") || strings.HasSuffix(name, ".mp4")) {
						totalFound++
						break
					}
				}
			}
		}
	}

	if totalRelevant == 0 {
		return 0 // no metadata episodes found
	}
	if totalFound == 0 {
		return 0 // metadata found but no files
	}
	if totalFound < totalRelevant {
		return 1 // partial match
	}
	return 2 // all matching episodes found
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
		for _, cr := range sidx.Episodes {
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

func rangesOverlap(a1, a2, b1, b2 int) bool {
	return a1 <= b2 && b1 <= a2
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
