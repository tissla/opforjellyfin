// matcher/matcher.go
package matcher

import (
	"fmt"
	"opforjellyfin/internal/logger"
	"opforjellyfin/internal/shared"
	"opforjellyfin/internal/ui"
	"os"
	"path/filepath"
)

// Matches video-file to metadata, then places it
func MatchAndPlaceVideo(videoPath, defaultDir string, index *shared.MetadataIndex, ogcr string) (string, error) {

	if _, err := os.Stat(videoPath); os.IsNotExist(err) {
		return "", nil
	}

	logger.DebugLog(false, "Checking if video file exists: %s", videoPath)
	fileName := filepath.Base(videoPath)
	logger.DebugLog(false, "Placing filename : %s", fileName)

	// strict
	dstPathNoSuffix := findMetadataMatch(fileName, index, ogcr)

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
	outFileName := ui.AnsiPadRight(fileNameNoPrefix, 26, "..")
	outRelPath := ui.AnsiPadRight(".."+relPathNoPrefix, 36, "..")
	msg := fmt.Sprintf("🎞️  Placed: %s → %s", outFileName, outRelPath)

	return msg, nil
}

// returns directory to place file, without suffix
func findMetadataMatch(fileName string, index *shared.MetadataIndex, ogcr string) string {

	cfg := shared.LoadConfig()
	baseDir := cfg.TargetDir

	// strayfolder for unmatched videos
	strayfolder := filepath.Join(baseDir, "strayvideos", ogcr)
	// finds season containing chapterRange, returns the seasonFolderName and seasonIndex
	// uses ogcr to find correct season even if its a bundle
	seasonFolderName, seasonIndex := findSeasonForChapter(ogcr, index)
	if seasonFolderName == "" {
		logger.DebugLog(false, "findMetaDataMatch: failed to find Season-folder")
		return strayfolder
	}
	logger.DebugLog(false, "season found for: %s for range %s", seasonFolderName, ogcr)

	// searches the seasonIndex for matching title for chapterRange, tries ogcr first for single-episode seasons
	newFileName := findTitleForChapter(ogcr, seasonIndex)
	if newFileName == "" {
		// if first fails, extract specific chapterRange from fileName
		chapterRange := shared.ExtractChapterRangeFromTitle(fileName)
		if chapterRange == "" {
			// if this extraction fails, try rougher methods
			// TODO: implement these rougher methods
		} else {
			// if extraction succeeded, find title from chapterRange
			newFileName = findTitleForChapter(chapterRange, seasonIndex)
		}
	} else {
		logger.DebugLog(false, "Title match found: ChapterKey: %s - EpisodeTitle: %s", ogcr, fileName)
	}

	seasonDir := filepath.Join(baseDir, seasonFolderName)

	if newFileName == "" {
		logger.DebugLog(false, "Could not determine episode title, sending to stray")
		return strayfolder
	}
	fullPathNoSuffix := filepath.Join(seasonDir, newFileName)

	logger.DebugLog(false, "findMetaDataMatch: returning %s", fullPathNoSuffix)
	return fullPathNoSuffix
}

// exact match, returns title from metadataindex using chapterKey.
func findTitleForChapter(chapterKey string, sindex shared.SeasonIndex) string {
	normKey := shared.NormalizeDash(chapterKey)

	logger.DebugLog(false, "findEpisodeKeyForChapter: chapterKey: %s - normKey: %s ", chapterKey, normKey)

	for epRange, ep := range sindex.EpisodeRange {
		if shared.NormalizeDash(epRange) == normKey {
			return ep.Title
		}
	}

	// no title found based on ChapterKey,
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
