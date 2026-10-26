// Package matcher
package matcher

import (
	"fmt"
	"opforjellyfin/internal/logger"
	"opforjellyfin/internal/shared"
	"opforjellyfin/internal/ui"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"
)

// MatchIndex maps video filename stems to Barry's metadata files.
type MatchIndex struct {
	ByFilename map[string]string       `json:"by_filename"`
	Entries    map[string]MetadataFile `json:"entries"`
}

type MetadataFile struct {
	NFOPath string `json:"nfo_path"`
}

// MatchFile returns the relative NFO path for a video filename.
// Use its parent folder for placement and keep the video's original filename.
func MatchFile(filename string, index *MatchIndex) (string, error) {
	if index == nil {
		return "", fmt.Errorf("metadata index is missing")
	}

	filename = path.Base(strings.ReplaceAll(filename, `\`, "/"))
	ext := path.Ext(filename)
	if !strings.EqualFold(ext, ".mkv") && !strings.EqualFold(ext, ".mp4") {
		return "", fmt.Errorf("unsupported video filename: %s", filename)
	}

	id, found := index.ByFilename[strings.TrimSuffix(filename, ext)]
	if !found {
		return "", fmt.Errorf("no metadata for %s", filename)
	}
	entry, found := index.Entries[id]
	if !found || id == "" {
		return "", fmt.Errorf("metadata entry is missing for %s", filename)
	}

	nfoPath := entry.NFOPath
	if path.Ext(nfoPath) != ".nfo" || path.IsAbs(nfoPath) ||
		strings.HasPrefix(nfoPath, "../") || strings.ContainsAny(nfoPath, "\\:\x00") ||
		path.Clean(nfoPath) != nfoPath {
		return "", fmt.Errorf("invalid metadata path: %q", nfoPath)
	}
	return nfoPath, nil
}

// MatchAndPlaceVideo Matches video-file to metadata, then places it
// No mutex needed here - shared.SafeMoveFile handles all locking
func MatchAndPlaceVideo(videoPath, defaultDir string, index *shared.MetadataIndex, ogcr string) (string, error) {

	if _, err := os.Stat(videoPath); os.IsNotExist(err) {
		return "", nil
	}

	logger.Log(false, "Checking if video file exists: %s", videoPath)
	fileName := filepath.Base(videoPath)
	logger.Log(false, "Placing filename : %s", fileName)

	// strict
	dstPathNoSuffix := findMetadataMatch(fileName, index, ogcr)

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
		timestamp := time.Now().Format("20260102-120304")
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
func findMetadataMatch(fileName string, index *shared.MetadataIndex, ogcr string) string {

	cfg, _ := shared.LoadConfig()
	baseDir := cfg.TargetDir

	// strayfolder for unmatched videos
	strayfolder := filepath.Join(baseDir, "strayvideos", ogcr, fileName)
	// finds season containing chapterRange, returns the seasonFolderName and seasonIndex
	// uses ogcr to find correct season even if its a bundle
	seasonFolderName, seasonIndex := findSeasonForChapter(ogcr, index)
	if seasonFolderName == "" {
		logger.Log(false, "findMetaDataMatch: failed to find Season-folder")
		return strayfolder
	}
	logger.Log(false, "season found for: %s for range %s", seasonFolderName, ogcr)

	// searches the seasonIndex for matching title for chapterRange, tries ogcr first for single-episode seasons
	newFileName := findTitleForChapter(ogcr, seasonIndex)
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
				newFileName = findTitleForChapter(chapterNum, seasonIndex)
			} else {
				// build a matching string from season and rough chapter, eg: seasonNum = 3 and chapternum = 05 => S03E05
				epKey := fmt.Sprintf("S%sE%s", seasonNum, chapterNum)
				newFileName = findTitleRough(epKey, seasonIndex)
			}
		} else {
			// if extraction succeeded, find title from chapterRange
			newFileName = findTitleForChapter(chapterRange, seasonIndex)
		}
	} else {
		logger.Log(false, "Title match found: ChapterKey: %s - EpisodeTitle: %s", ogcr, fileName)
	}

	seasonDir := filepath.Join(baseDir, seasonFolderName)

	if newFileName == "" {
		logger.Log(false, "Could not determine episode title, sending to stray")
		return strayfolder
	}
	fullPathNoSuffix := filepath.Join(seasonDir, newFileName)

	logger.Log(false, "findMetaDataMatch: returning %s", fullPathNoSuffix)
	return fullPathNoSuffix
}

// exact match, returns title from metadataindex using chapterKey.
func findTitleForChapter(chapterKey string, sindex shared.SeasonIndex) string {
	normKey := shared.NormalizeDash(chapterKey)

	logger.Log(false, "findEpisodeKeyForChapter: chapterKey: %s - normKey: %s ", chapterKey, normKey)

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

// rough finder
func findTitleRough(epKey string, sindex shared.SeasonIndex) string {

	for _, ep := range sindex.EpisodeRange {
		if strings.Contains(ep.Title, epKey) {
			logger.Log(false, "roughFindTitle match found: %s > %s", epKey, ep.Title)
			return ep.Title
		}
	}

	logger.Log(false, "roughFindTitle did not find a match. for %s", epKey)
	return ""
}
