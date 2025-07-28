package metadata

import (
	"fmt"
	"io/fs"
	"opforjellyfin/internal/logger"
	"opforjellyfin/internal/shared"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

func buildIndexFromDir(baseDir string) (*shared.MetadataIndex, error) {
	index := &shared.MetadataIndex{
		Seasons: make(map[string]shared.SeasonIndex),
	}

	err := filepath.WalkDir(baseDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || !shared.IsEpisodeNFO(d.Name()) {
			return nil
		}

		data, err := os.ReadFile(path)
		if err != nil {
			return nil
		}

		season, episode, chapterRange := extractEpisodeMetadata(data)
		if season == "" || episode == "" || chapterRange == "" {
			logger.Log(false, "indexbuilder: missed param for %s - season: %s - episode %s - chapterRange %s", d.Name(), season, episode, chapterRange)
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
	nameSeasons(index, baseDir)

	return index, nil
}

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

func nameSeasons(index *shared.MetadataIndex, baseDir string) {
	tvshowPath := filepath.Join(baseDir, "tvshow.nfo")

	data, err := os.ReadFile(tvshowPath)
	if err != nil {
		logger.Log(false, "indexbuilder: Could not read tvshow.nfo: %v", err)
		return
	}

	// Parse all namedseason tags
	// Pattern: <namedseason number="X">Y. Season Name</namedseason>
	re := regexp.MustCompile(`<namedseason\s+number="(\d+)">([^<]+)</namedseason>`)
	matches := re.FindAllSubmatch(data, -1)

	seasonNames := make(map[string]string)
	for _, match := range matches {
		if len(match) >= 3 {
			seasonNum := string(match[1])
			seasonName := string(match[2])

			// Create the season key (e.g., "Season 01", "Season 10")
			snum, _ := strconv.Atoi(seasonNum)
			seasonKey := fmt.Sprintf("Season %d", snum)

			// Remove the number prefix from the name (e.g., "1. Romance Dawn" -> "Romance Dawn")
			cleanName := strings.TrimSpace(seasonName)
			if idx := strings.Index(cleanName, ". "); idx != -1 {
				cleanName = cleanName[idx+2:]
			}

			seasonNames[seasonKey] = cleanName
		}
	}

	// Apply names to the index
	for seasonKey, seasonData := range index.Seasons {
		if name, exists := seasonNames[seasonKey]; exists {
			seasonData.Name = name
			index.Seasons[seasonKey] = seasonData
			logger.Log(false, "Named %s as '%s'", seasonKey, name)
		}
	}
}

// important
func extractEpisodeMetadata(data []byte) (string, string, string) {
	return shared.ExtractXMLTag(data, "season"), shared.ExtractXMLTag(data, "episode"), shared.ExtractChapterRangeFromNFO(string(data))
}
