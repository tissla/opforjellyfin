package metadata

import (
	"opforjellyfin/internal/logger"
	"opforjellyfin/internal/shared"
	"os"
	"path/filepath"
	"strings"
)

// checks a range in metadata. 0 = does not have, 1 = have some, 2 = have all
func HaveVideoStatus(chapterRange string) int {
	if chapterRange == "" {
		return 0
	}

	index := LoadMetadataCache()
	cfg := shared.LoadConfig()
	baseDir := cfg.TargetDir

	matches := shared.FindMatchingEpisodes(chapterRange, index)
	if len(matches) == 0 {
		return 0
	}

	// Season-level match (bundle torrent or exact season)
	if matches[0].EpisodeRange == "" {
		seasonDir := filepath.Join(baseDir, matches[0].SeasonKey)
		v, n := CountVideosAndTotal(seasonDir)
		logger.Log(false, "HaveVideoStatus: counted %d videos and %d nfos for seasonKey: %s", v, n, matches[0].SeasonKey)
		if v == 0 {
			return 0
		}
		if v < n {
			return 1
		}
		return 2
	}

	// Episode-level matches: check if video files exist for each matched episode
	haveCount := 0
	for _, m := range matches {
		seasonDir := filepath.Join(baseDir, m.SeasonKey)
		videoPathMP4 := filepath.Join(seasonDir, m.EpisodeTitle+".mp4")
		videoPathMKV := filepath.Join(seasonDir, m.EpisodeTitle+".mkv")

		if shared.FileExists(videoPathMP4) || shared.FileExists(videoPathMKV) {
			haveCount++
		}
	}

	if haveCount == 0 {
		return 0
	}
	if haveCount < len(matches) {
		return 1
	}
	return 2
}

// HaveMetadata checks if metadata exists for given chapterRange.
func HaveMetadata(chapterRange string) bool {
	if chapterRange == "" {
		return false
	}

	LoadMetadataCache()
	matches := shared.FindMatchingEpisodes(chapterRange, metadataCache)
	return len(matches) > 0
}

// video and .nfo file counter. Returns: number of videos matched with episode .nfo file, number of episode .nfo files
func CountVideosAndTotal(dir string) (matched int, totalNFO int) {
	videoFiles := map[string]bool{}

	err := filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			return nil
		}

		lower := strings.ToLower(d.Name())
		base := strings.TrimSuffix(d.Name(), filepath.Ext(d.Name()))

		if strings.HasSuffix(lower, ".mkv") || strings.HasSuffix(lower, ".mp4") {
			videoFiles[base] = false
		}

		if shared.IsEpisodeNFO(lower) {
			totalNFO++
			if _, exists := videoFiles[base]; exists {
				videoFiles[base] = true //
			}
		}

		return nil
	})

	if err != nil {
		return 0, 0
	}

	// count matched
	for _, matchedFlag := range videoFiles {
		if matchedFlag {
			matched++
		}
	}

	return matched, totalNFO
}
