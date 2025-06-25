package metadata

import (
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

	for seasonKey, season := range index.Seasons {
		seasonDir := filepath.Join(baseDir, seasonKey)

		if seasonKey == chapterRange {
			v, n := CountVideosAndTotal(seasonDir)
			if v < n {
				return 0
			}
			return 1
		}

		for epRange, epData := range season.EpisodeRange {
			if epRange == chapterRange {
				videoPathMP4 := filepath.Join(seasonDir, epData.Title+".mp4")
				videoPathMKV := filepath.Join(seasonDir, epData.Title+".mkv")

				if shared.FileExists(videoPathMP4) || shared.FileExists(videoPathMKV) {
					return 2
				}
			}

		}
	}

	return 0
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
