// matcher/matcher.go
package matcher

import (
	"fmt"
	"opforjellyfin/internal/logger"
	"opforjellyfin/internal/shared"
	"os"
	"path/filepath"

	"github.com/charmbracelet/x/ansi"
)

func MatchAndPlaceVideo(videoPath, defaultDir string, index *shared.MetadataIndex) (string, error) {

	if _, err := os.Stat(videoPath); os.IsNotExist(err) {
		return "", nil
	}

	logger.DebugLog(false, "Checking if video file exists: %s", videoPath)
	fileName := filepath.Base(videoPath)
	logger.DebugLog(false, "Placing filename : %s", fileName)

	// strict
	dstPathNoSuffix := findMetadataMatch(fileName, index)

	logger.DebugLog(false, "dstPath for fileName %s will be %s", fileName, dstPathNoSuffix)

	ext := filepath.Ext(fileName)
	finalPath := dstPathNoSuffix + ext

	if err := SafeMoveFile(videoPath, finalPath); err != nil {
		logger.DebugLog(false, "sfm Error: %s", err)
		return "", err
	}

	//relative path for logs
	relPath, _ := filepath.Rel(defaultDir, finalPath)
	//debug
	logger.DebugLog(false, fmt.Sprintf("placed: %s → %s", fileName, relPath))

	// some formatting
	fileNameNoPrefix := fileName[10:]
	relPathNoPrefix := filepath.Base(relPath)[10:]
	outFileName := ansi.Truncate(fileNameNoPrefix, 36, "..")
	outRelPath := ansi.Truncate(".."+relPathNoPrefix, 36, "..")
	msg := fmt.Sprintf("🎞️  Placed: %s → %s", outFileName, outRelPath)

	return msg, nil
}

// returns directory to place file, without suffix
func findMetadataMatch(fileName string, index *shared.MetadataIndex) string {

	cfg := shared.LoadConfig()
	baseDir := cfg.TargetDir

	// moved this to findMetadataMatch to place file in strayfolder when no key can be found
	chapterKey, err := shared.ExtractChapterRangeFromTitle(fileName)
	if chapterKey == "" {
		chapterKey = "unknown"
	}
	strayfolder := filepath.Join(baseDir, "strayvideos", chapterKey)

	if err != nil {
		logger.DebugLog(false, "findMetaDataMatch: could not extract manga chapter: %s", err)
		return strayfolder
	}

	logger.DebugLog(false, "chapterkey extracted: %s from %s", chapterKey, fileName)

	seasonFolder, seasonIndex := findSeasonForChapter(chapterKey, index)
	if seasonFolder == "" {
		logger.DebugLog(false, "findMetaDataMatch: failed to find Season-folder")
		return strayfolder
	}

	fileName = findTitleForChapter(chapterKey, seasonIndex)
	if fileName == "" {
		logger.DebugLog(false, "findMetaDataMatch: failed to find episode-key")
		return strayfolder
	}

	logger.DebugLog(false, "Title match found: ChapterKey: %s - EpisodeKey%s", chapterKey, fileName)

	seasonDir := filepath.Join(baseDir, seasonFolder)

	fullPathNoSuffix := filepath.Join(seasonDir, fileName)

	logger.DebugLog(false, "findMetaDataMatch: returning %s", fullPathNoSuffix)
	return fullPathNoSuffix
}

// exact match, returns title
func findTitleForChapter(chapterKey string, sindex shared.SeasonIndex) string {
	normKey := shared.NormalizeDash(chapterKey)

	logger.DebugLog(false, "findEpisodeKeyForChapter: chapterKey: %s - normKey: %s ", chapterKey, normKey)

	for epRange, ep := range sindex.EpisodeRange {
		if shared.NormalizeDash(epRange) == normKey {
			return ep.Title
		}
	}
	return ""
}

// finds the season a ChapterKey belongs to. returns the season name as a string, also returns the whole SeasonIndex struct
func findSeasonForChapter(chapterKey string, index *shared.MetadataIndex) (string, shared.SeasonIndex) {
	chStart, chEnd := shared.ParseRange(chapterKey)

	for seasonName, season := range index.Seasons {
		seasonStart, seasonEnd := shared.ParseRange(season.Range)

		if chStart >= seasonStart && chEnd <= seasonEnd {
			return seasonName, season
		}
	}

	return "", shared.SeasonIndex{}

}
