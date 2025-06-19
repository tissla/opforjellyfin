// internal/metadata.go
package internal

import (
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
)

var (
	metadataCache     *MetadataIndex
	metadataCacheOnce sync.Once
)

// MetadataIndex represents structured metadata for seasons and episodes.
type MetadataIndex struct {
	Seasons map[string]SeasonIndex `json:"seasons"`
}

// SeasonIndex represents episodes and their manga chapter ranges.
type SeasonIndex struct {
	Range    string            `json:"range"`
	Episodes map[string]string `json:"episodes"`
}

// FetchAllMetadata clones and indexes metadata from GitHub.
func FetchAllMetadata(baseDir string, cfg Config) error {
	return cloneAndCopyRepo(baseDir, cfg, false)
}

// SyncMetadata clones and syncs metadata updates from GitHub.
func SyncMetadata(baseDir string, cfg Config) error {
	return cloneAndCopyRepo(baseDir, cfg, true)
}

func cloneAndCopyRepo(baseDir string, cfg Config, syncOnly bool) error {
	tmpDir := filepath.Join(os.TempDir(), "repo-tmp")
	defer os.RemoveAll(tmpDir)

	fmt.Println("🌐 Fetching metadata from GitHub...")

	if err := exec.Command("git", "clone", "--depth=1", fmt.Sprintf("https://github.com/%s.git", cfg.GitHubRepo), tmpDir).Run(); err != nil {
		return fmt.Errorf("❌ git clone failed: %w", err)
	}

	srcDir := filepath.Join(tmpDir, "One Pace")
	var err error

	if syncOnly {
		err = syncDir(srcDir, baseDir)
	} else {
		err = copyDir(srcDir, baseDir)
	}

	if err != nil {
		return fmt.Errorf("❌ failed to copy metadata: %w", err)
	}

	if err := BuildMetadataIndex(baseDir); err != nil {
		return fmt.Errorf("❌ failed to build metadata index: %w", err)
	}

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

func buildIndexFromDir(baseDir string) (*MetadataIndex, error) {
	index := &MetadataIndex{Seasons: make(map[string]SeasonIndex)}

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
		epKey := fmt.Sprintf("S%02sE%02s", season, episode)

		if _, exists := index.Seasons[seasonKey]; !exists {
			index.Seasons[seasonKey] = SeasonIndex{Episodes: make(map[string]string)}
		}

		index.Seasons[seasonKey].Episodes[epKey] = chapterRange
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("❌ metadata indexing failed: %w", err)
	}

	calculateSeasonRanges(index)

	return index, nil
}

func saveMetadataIndex(index *MetadataIndex, baseDir string) error {
	path := filepath.Join(baseDir, "metadata-index.json")
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("❌ could not create index file: %w", err)
	}
	defer f.Close()

	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	if err := enc.Encode(index); err != nil {
		return fmt.Errorf("❌ could not encode metadata index: %w", err)
	}

	metadataCache = index // cache immediately after saving
	fmt.Println("✅ Metadata index written to", path)
	return nil
}

// HaveMetadata checks if metadata exists for given chapterRange.
func HaveMetadata(chapterRange string) bool {
	if chapterRange == "" {
		return false
	}

	loadMetadataCache()

	targetStart, targetEnd := parseRange(chapterRange)

	for _, season := range metadataCache.Seasons {
		for _, epRange := range season.Episodes {
			a, b := parseRange(epRange)
			if rangesOverlap(targetStart, targetEnd, a, b) {
				return true
			}
		}
	}

	return false
}

// HaveVideoFile checks if a video file exists for a given chapter range.
func HaveVideoFile(chapterRange string) bool {
	if chapterRange == "" {
		return false
	}

	loadMetadataCache()

	targetStart, targetEnd := parseRange(chapterRange)

	for seasonKey, season := range metadataCache.Seasons {
		for epKey, epRange := range season.Episodes {
			start, end := parseRange(epRange)
			if rangesOverlap(targetStart, targetEnd, start, end) {
				if videoExistsForEpisode(seasonKey, epKey) {
					return true
				}
			}
		}
	}

	return false
}

// helper for havevideofile
func videoExistsForEpisode(seasonKey, epKey string) bool {
	cfg := LoadConfig()
	baseDir := cfg.TargetDir

	seasonNum := extractSeasonNumberFromKey(epKey)
	episodeNum := extractEpisodeNumber(epKey)

	dir := filepath.Join(baseDir, seasonKey)

	mkvPath := filepath.Join(dir, fmt.Sprintf("One Pace - S%sE%s - *.mkv", seasonNum, episodeNum))
	mp4Path := filepath.Join(dir, fmt.Sprintf("One Pace - S%sE%s - *.mp4", seasonNum, episodeNum))

	mkvMatch, _ := filepath.Glob(mkvPath)
	mp4Match, _ := filepath.Glob(mp4Path)

	return len(mkvMatch) > 0 || len(mp4Match) > 0
}

// extract season from epkey e.g: S05E04 -> "05"
func extractSeasonNumberFromKey(episodeKey string) string {
	re := regexp.MustCompile(`S(\d+)E\d+`)
	matches := re.FindStringSubmatch(episodeKey)
	if len(matches) == 2 {
		return matches[1]
	}
	return "00"
}

// more helpers

func loadMetadataCache() {
	metadataCacheOnce.Do(func() {
		cfg := LoadConfig()
		data, err := os.ReadFile(filepath.Join(cfg.TargetDir, "metadata-index.json"))
		if err != nil {
			metadataCache = &MetadataIndex{}
			return
		}
		json.Unmarshal(data, &metadataCache)
	})
}

func calculateSeasonRanges(index *MetadataIndex) {
	for skey, sidx := range index.Seasons {
		min, max := 99999, -1
		for _, cr := range sidx.Episodes {
			start, end := parseRange(cr)
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

func extractEpisodeMetadata(data []byte) (string, string, string) {
	return extractXMLTag(data, "season"), extractXMLTag(data, "episode"), extractChapterRangeFromNFO(string(data))
}

func parseRange(r string) (int, int) {
	parts := strings.Split(r, "-")
	if len(parts) != 2 {
		return -1, -1
	}
	a, _ := strconv.Atoi(parts[0])
	b, _ := strconv.Atoi(parts[1])
	return a, b
}

func rangesOverlap(a1, a2, b1, b2 int) bool {
	return a1 <= b2 && b1 <= a2
}

func extractChapterRangeFromNFO(content string) string {
	re := regexp.MustCompile(`(?i)Manga\s*Chapter\(s\)?:\s*(\d+)(?:[\s,-]*(\d+))?`)
	match := re.FindStringSubmatch(content)
	if len(match) >= 2 {
		start := match[1]
		end := match[2]
		if end == "" {
			end = start
		}
		return fmt.Sprintf("%s-%s", start, end)
	}
	return ""
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

		return copyFile(path, destPath, info.Mode())
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

		if stat, err := os.Stat(destPath); err == nil && stat.Size() == info.Size() {
			return nil
		}

		return copyFile(path, destPath, info.Mode())
	})
}

// copyFile copies from src to dst with permissions
func copyFile(src, dst string, perm os.FileMode) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return err
	}

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	if _, err := io.Copy(out, in); err != nil {
		return err
	}

	return os.Chmod(dst, perm)
}
