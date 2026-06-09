// matcher/matcher.go
package matcher

import (
	"fmt"
	"opforjellyfin/internal/logger"
	"opforjellyfin/internal/shared"
	"opforjellyfin/internal/ui"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Matches video-file to metadata, then places it
// No mutex needed here - shared.SafeMoveFile handles all locking
func MatchAndPlaceVideo(videoPath, defaultDir string, index *shared.MetadataIndex, td *shared.TorrentDownload) (string, error) {

	if _, err := os.Stat(videoPath); os.IsNotExist(err) {
		return "", nil
	}

	logger.Log(false, "Checking if video file exists: %s", videoPath)
	fileName := filepath.Base(videoPath)
	logger.Log(false, "Placing filename : %s", fileName)

	// strict
	dstPathNoSuffix := findMetadataMatch(fileName, index, td)

	logger.Log(false, "dstPath for fileName %s will be %s", fileName, dstPathNoSuffix)

	// extract suffix from original file
	ext := filepath.Ext(fileName)
	finalPath := dstPathNoSuffix + ext

	var msg string

	// SafeMoveFile now handles all locking internally
	if err := shared.SafeMoveFile(videoPath, finalPath); err != nil {
		logger.Log(false, "sfm Error: %s, moving to strayvideos", err)

		// Create strayvideos directory using the thread-safe function
		strayDir := filepath.Join(defaultDir, "strayvideos")
		if err := shared.CreateDirectory(strayDir); err != nil {
			logger.Log(true, "Failed to create strayvideos directory: %v", err)
			return "", fmt.Errorf("failed to create strayvideos: %w", err)
		}

		// Add timestamp to filename to avoid collisions
		nameWithoutExt := strings.TrimSuffix(fileName, ext)
		timestamp := time.Now().Format("20060102-150405")
		strayFileName := fmt.Sprintf("%s_%s%s", nameWithoutExt, timestamp, ext)
		strayPath := filepath.Join(strayDir, strayFileName)

		// SafeMoveFile handles locking
		if err := shared.SafeMoveFile(videoPath, strayPath); err != nil {
			logger.Log(true, "Failed to move to strayvideos: %v", err)
			return "", fmt.Errorf("failed to place file anywhere: %w", err)
		}

		// Format message for strayvideos
		outFileName := ui.AnsiPadRight(fileName, 26, "..")
		outRelPath := ui.AnsiPadRight("strayvideos/"+strayFileName, 36, "..")
		msg = fmt.Sprintf("⚠️  Placed in stray: %s → %s", outFileName, outRelPath)

	} else {
		//relative path for logs
		relPath, _ := filepath.Rel(defaultDir, finalPath)
		//debug
		logger.Log(false, "%s", fmt.Sprintf("placed: %s → %s", fileName, relPath))

		// some formatting
		fileNameNoPrefix := fileName
		if len(fileName) > 10 {
			fileNameNoPrefix = fileName[10:]
		}
		relPathNoPrefix := filepath.Base(relPath)
		if len(relPathNoPrefix) > 10 {
			relPathNoPrefix = relPathNoPrefix[10:]
		}
		outFileName := ui.AnsiPadRight(fileNameNoPrefix, 26, "..")
		outRelPath := ui.AnsiPadRight(".."+relPathNoPrefix, 36, "..")
		msg = fmt.Sprintf("🎞️  Placed: %s → %s", outFileName, outRelPath)
	}

	return msg, nil
}

// returns directory to place file, without suffix
func findMetadataMatch(fileName string, index *shared.MetadataIndex, td *shared.TorrentDownload) string {

	cfg := shared.LoadConfig()
	baseDir := cfg.TargetDir

	// strayfolder for unmatched videos
	strayfolder := filepath.Join(baseDir, "strayvideos", td.ChapterRange, fileName)
	// finds season containing chapterRange, returns the seasonFolderName and seasonIndex
	// uses ogcr to find correct season even if its a bundle
	seasonStartIdx := -1
	newFileName := ""
	seasonDir := ""

	sortedIndex := shared.SortMetadataSeasons(index)

	// loop to repeat season lookups on episode match failures
	// some seasons have overlapping chapters (i.e. The Trials of Koby-Meppo cover pages span Ch. 83-119)
	for seasonStartIdx < len(sortedIndex.Seasons) {

		seasonFolderName, seasonIndex, startIdx := findSeasonForChapter(td.ChapterRange, sortedIndex, seasonStartIdx)
		seasonStartIdx = startIdx

		if seasonFolderName == "" {
			logger.Log(false, "findMetaDataMatch: failed to find Season-folder")
			return strayfolder
		}
		logger.Log(false, "season found for: %s for range %s", seasonFolderName, td.ChapterRange)

		var filters []string
		if td.Entry.IsExtended {
			filters = append(filters, "extended")
		}

		// searches the seasonIndex for matching title for chapterRange, tries ogcr first for single-episode seasons
		newFileName = findTitleForChapter(td.ChapterRange, seasonIndex, filters...)
		if newFileName == "" {
			// if first fails, extract specific chapterRange from fileName
			chapterRange := shared.ExtractChapterRangeFromTitle(fileName)
			if chapterRange == "" {
				logger.Log(false, "findMetaDataMatch - trying rough extraction for: %s", fileName)
				// use ogcr + file regex
				// if this extraction fails, try rougher methods
				seasonZ := shared.ExtractSeasonNumber(seasonFolderName)
				seasonNum := fmt.Sprintf("%02s", seasonZ)

				// rough extract can find chapterRange or rough chapter(in relation to season) if lucky.
				chapterNum, isRange := shared.RoughExtractChapterFromTitle(fileName)
				logger.Log(false, "findMetaDataMatch - rough extracted chapterNum: %s", chapterNum)

				if isRange {
					newFileName = findTitleForChapter(chapterNum, seasonIndex, filters...)
				} else {
					// build a matching string from season and rough chapter, eg: seasonNum = 3 and chapternum = 05 => S03E05
					epKey := fmt.Sprintf("S%sE%s", seasonNum, chapterNum)
					newFileName = findTitleRough(epKey, seasonIndex, filters...)
				}
			} else {
				// if extraction succeeded, find title from chapterRange
				newFileName = findTitleForChapter(chapterRange, seasonIndex, filters...)
			}
		} else {
			logger.Log(false, "Title match found: ChapterKey: %s - EpisodeTitle: %s", td.ChapterRange, fileName)
		}

		if newFileName == "" {
			matches := shared.FindMatchingEpisodes(td.ChapterRange, index)
			// Pick the first match that belongs to the same season
			for _, m := range matches {
				if m.SeasonKey == seasonFolderName && m.EpisodeTitle != "" {
					logger.Log(false, "findMetaDataMatch: overlap match found: %s -> %s (%d)", td.ChapterRange, m.EpisodeTitle, m.MatchType)
					newFileName = m.EpisodeTitle
					break
				}
			}
		}

		seasonDir = filepath.Join(baseDir, seasonFolderName)

		if newFileName == "" {
			logger.Log(false, "Could not determine episode title, restarting season lookup")
			continue
		}
		// found a valid episode title for the matched season
		break
	}
	if newFileName == "" || seasonDir == "" {
		logger.Log(false, "Could not determine episode title, sending to stray")
		return strayfolder
	}
	fullPathNoSuffix := filepath.Join(seasonDir, newFileName)

	logger.Log(false, "findMetaDataMatch: returning %s", fullPathNoSuffix)
	return fullPathNoSuffix
}

// exact match, returns title from metadataindex using chapterKey.
func findTitleForChapter(chapterKey string, sindex shared.SeasonIndex, nameFilters ...string) string {
	normKey := shared.NormalizeDash(chapterKey)

	logger.Log(false, "findEpisodeKeyForChapter: chapterKey: %s - normKey: %s ", chapterKey, normKey)

	for epRange, epData := range sindex.EpisodeRange {
		epKey := shared.NormalizeDash(epRange)
		logger.Log(false, "checkingTitle for chapter[%s]: [%s]", normKey, epKey)
		if epKey == normKey {
			var bestMatch string
			bestCount := -1

			for _, ep := range epData {
				filterCount := 0
				for _, nameFilter := range nameFilters {
					if strings.Contains(strings.ToLower(ep.Title), strings.ToLower(nameFilter)) {
						filterCount++
					}
				}
				if filterCount > bestCount {
					bestCount = filterCount
					bestMatch = ep.Title
				}
			}
			logger.Log(false, "foundTitle for chapter[%s]: [%s] - %s", normKey, epKey, bestMatch)

			return bestMatch
		}
	}

	logger.Log(false, "findTitleForChapter no title found from chapterKey: %s", chapterKey)

	// no title found based on ChapterKey,
	return ""
}

// finds the season a ChapterKey belongs to. returns the season name as a string, also returns the whole SeasonIndex struct.
// When multiple seasons contain the range (e.g. overlapping recap seasons), picks the narrowest fit.
func findSeasonForChapter(chapterKey string, index *shared.SortedMetadataIndex, startIdx int) (string, shared.SeasonIndex, int) {
	chStart, chEnd := shared.ParseRange(chapterKey)

	// loop through seasons offset by startIdx
	for idx, season := range index.Seasons[startIdx+1:] {
		seasonStart, seasonEnd := shared.ParseRange(season.SeasonIndex.Range)

		if chStart >= seasonStart && chEnd <= seasonEnd {
			logger.Log(false, "foundSeason from channel[%d-%d]: %s [%d-%d]", chStart, chEnd, season.Title, seasonStart, seasonEnd)
			return season.Title, season.SeasonIndex, idx + startIdx + 1
		}
	}

	return "", shared.SeasonIndex{}, -1
}

// rough finder
func findTitleRough(epKey string, sindex shared.SeasonIndex, nameFilters ...string) string {

	for _, epData := range sindex.EpisodeRange {
		bestMatch := ""
		bestCount := -1

		for _, ep := range epData {
			if strings.Contains(ep.Title, epKey) {
				filterCount := 0
				for _, nameFilter := range nameFilters {
					if strings.Contains(strings.ToLower(ep.Title), strings.ToLower(nameFilter)) {
						filterCount++
					}
				}
				if filterCount > bestCount {
					bestCount = filterCount
					bestMatch = ep.Title
				}
			}
		}
		if bestMatch != "" {
			logger.Log(false, "findTitleRough found rough match for epKey %s: %s", epKey, bestMatch)
			return bestMatch
		}
	}

	logger.Log(false, "roughFindTitle did not find a match. for %s", epKey)
	return ""
}
