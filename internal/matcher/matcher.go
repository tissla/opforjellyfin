// matcher/matcher.go
package matcher

import (
	"encoding/json"
	"fmt"
	"opforjellyfin/internal/logger"
	"opforjellyfin/internal/shared"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/x/ansi"
)

func MatchAndPlaceVideo(videoPath, metadataDir string) (string, error) {

	if _, err := os.Stat(videoPath); os.IsNotExist(err) {
		return "", nil
	}

	logger.DebugLog(false, "Checking if video file exists: %s", videoPath)
	fileName := filepath.Base(videoPath)
	logger.DebugLog(false, "Placing filename : %s", fileName)

	// strict
	chapterKey, err := shared.ExtractChapterKeyFromTitle(fileName)

	if err != nil {
		return "", fmt.Errorf("⚠️ Could not extract manga chapter: %w", err)
	}
	logger.DebugLog(false, "chapterkey extracted: %s from %s", chapterKey, fileName)

	index, err := loadMetadataIndex(metadataDir)
	if err != nil {
		return "", fmt.Errorf("❌ Could not load metadata index: %w", err)
	}

	dstPathNoSuffix := findMetadataMatch(chapterKey, index)

	logger.DebugLog(false, "dstPath for chapterKey %s will be %s", chapterKey, dstPathNoSuffix)

	ext := filepath.Ext(fileName)
	finalPath := dstPathNoSuffix + ext

	if err := SafeMoveFile(videoPath, finalPath); err != nil {
		logger.DebugLog(false, "sfm Error: %s", err)
		return "", err
	}

	//relative path for logs
	relPath, _ := filepath.Rel(metadataDir, finalPath)
	//debug
	logger.DebugLog(false, fmt.Sprintf("🎞️  Placed: %s → %s", fileName, relPath))

	// truncate for outmessage
	outFileName := ansi.Truncate(fileName, 36, "..")
	outRelPath := ansi.Truncate(relPath, 36, "..")
	msg := fmt.Sprintf("🎞️  Placed: %s → %s", outFileName, outRelPath)

	return msg, nil
}

// sets pointer to MetadataIndex, from file
func loadMetadataIndex(metadataDir string) (*shared.MetadataIndex, error) {
	data, err := os.ReadFile(filepath.Join(metadataDir, "metadata-index.json"))
	if err != nil {
		return nil, err
	}
	var index shared.MetadataIndex
	if err := json.Unmarshal(data, &index); err != nil {
		return nil, err
	}
	return &index, nil
}

// returns directory to place file, without suffix
func findMetadataMatch(chapterKey string, index *shared.MetadataIndex) string {

	cfg := shared.LoadConfig()
	baseDir := cfg.TargetDir
	strayfolder := filepath.Join(baseDir, "strayvideos", chapterKey)

	seasonFolder, seasonIndex := findSeasonForChapter(chapterKey, index)
	if seasonFolder == "" {
		return strayfolder
	}

	episodeKey := findEpisodeKeyForChapter(chapterKey, seasonIndex)
	if episodeKey == "" {
		return strayfolder
	}

	logger.DebugLog(false, "EpisodeKey match found: ChapterKey: %s - EpisodeKey%s", chapterKey, episodeKey)

	seasonDir := filepath.Join(baseDir, seasonFolder)

	files, err := os.ReadDir(seasonDir)
	if err != nil {
		return strayfolder
	}

	for _, f := range files {
		if !f.IsDir() && strings.HasSuffix(f.Name(), ".nfo") && strings.Contains(f.Name(), episodeKey) {
			filename := strings.TrimSuffix(f.Name(), ".nfo")
			return filepath.Join(seasonDir, filename)
		}
	}

	return strayfolder
}

// exact match
func findEpisodeKeyForChapter(chapterKey string, sindex shared.SeasonIndex) string {
	for episodeKey, epRange := range sindex.Episodes {
		if epRange == chapterKey {
			return episodeKey
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
