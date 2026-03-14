package metadata

import (
	"opforjellyfin/internal/logger"
	"opforjellyfin/internal/shared"
	"os"
	"path/filepath"
	"strings"
)

// checks a range in metadata. 0 = does not have, 1 = have some, 2 = have all
// Uses numeric comparison for range matching.
func HaveVideoStatus(chapterRange string) int {
	if chapterRange == "" {
		return 0
	}

	index := LoadMetadataCache()
	cfg := shared.LoadConfig()
	baseDir := cfg.TargetDir
	crStart, crEnd := shared.ParseRange(shared.NormalizeDash(chapterRange))

	for seasonKey, season := range index.Seasons {
		seasonDir := filepath.Join(baseDir, seasonKey)

		sStart, sEnd := shared.ParseRange(shared.NormalizeDash(season.Range))
		if sStart >= 0 && sStart == crStart && sEnd == crEnd {
			v, n := CountVideosAndTotal(seasonDir)
			logger.Log(false, "HaveVideoStatus: counted %d videos and %d nfos for seasonKey: %s", v, n, seasonKey)
			if v == 0 {
				return 0
			}

			if v < n {
				return 1
			}

			return 2
		}

		for epRange, epData := range season.EpisodeRange {
			epStart, epEnd := shared.ParseRange(shared.NormalizeDash(epRange))
			if epStart >= 0 && epStart == crStart && epEnd == crEnd {
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
// Uses numeric comparison so "023-041" matches "23-41", and overlap matching
// so "85-88" matches metadata keyed as "85-87".
func HaveMetadata(chapterRange string) bool {
	if chapterRange == "" {
		return false
	}

	LoadMetadataCache()
	crStart, crEnd := shared.ParseRange(shared.NormalizeDash(chapterRange))

	for _, season := range metadataCache.Seasons {
		// season range match
		sStart, sEnd := shared.ParseRange(shared.NormalizeDash(season.Range))
		if sStart >= 0 && sStart == crStart && sEnd == crEnd {
			return true
		}

		// match individual episodes (exact numeric, then overlap)
		for epRange := range season.EpisodeRange {
			epStart, epEnd := shared.ParseRange(shared.NormalizeDash(epRange))
			if epStart >= 0 && epStart == crStart && epEnd == crEnd {
				return true
			}
		}

		// overlap fallback
		if crStart >= 0 && crEnd >= 0 {
			for epRange := range season.EpisodeRange {
				epStart, epEnd := shared.ParseRange(shared.NormalizeDash(epRange))
				if epStart < 0 || epEnd < 0 {
					continue
				}
				overlapStart := max(crStart, epStart)
				overlapEnd := min(crEnd, epEnd)
				if overlapEnd-overlapStart+1 > 0 {
					return true
				}
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
