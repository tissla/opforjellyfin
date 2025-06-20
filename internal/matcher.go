// internal/matcher.go
package internal

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

func MatchAndPlaceVideo(videoPath, metadataDir string, td *TorrentDownload) (string, error) {

	if _, err := os.Stat(videoPath); os.IsNotExist(err) {
		return "", nil
	}

	fileName := filepath.Base(videoPath)

	chapterKey, err := extractChapterKey(td.ChapterRange)
	if err != nil {
		return "", fmt.Errorf("⚠️ Could not extract manga chapter: %w", err)
	}

	index, err := loadMetadataIndex(metadataDir)
	if err != nil {
		return "", fmt.Errorf("❌ Could not load metadata index: %w", err)
	}

	match, err := findMetadataMatch(chapterKey, index)
	if err != nil {
		return "", fmt.Errorf("❌ %v", err)
	}

	finalPath, err := prepareDestination(metadataDir, videoPath, match, chapterKey)
	if err != nil {
		return "", fmt.Errorf("❌ Could not place video file: %w", err)
	}

	relPath, _ := filepath.Rel(metadataDir, finalPath)
	msg := fmt.Sprintf("🎞️  Placed: %s → %s", fileName, relPath)

	return msg, nil
}

func extractChapterKey(chapterRange string) (string, error) {
	re := regexp.MustCompile(`(\d+)`)
	matches := re.FindStringSubmatch(chapterRange)
	if len(matches) < 2 {
		return "", fmt.Errorf("could not extract first chapter from range: %s", chapterRange)
	}
	return matches[1], nil
}

func loadMetadataIndex(metadataDir string) (*MetadataIndex, error) {
	data, err := os.ReadFile(filepath.Join(metadataDir, "metadata-index.json"))
	if err != nil {
		return nil, err
	}
	var index MetadataIndex
	if err := json.Unmarshal(data, &index); err != nil {
		return nil, err
	}
	return &index, nil
}

func findMetadataMatch(chapterKey string, index *MetadataIndex) (struct {
	seasonName string
	episodeKey string
	seasonNum  string
	episodeNum string
}, error) {
	for seasonName, season := range index.Seasons {
		for episodeKey, rangeStr := range season.Episodes {
			if strings.HasPrefix(rangeStr, chapterKey+"-") || rangeStr == chapterKey+"-"+chapterKey {
				return struct {
					seasonName string
					episodeKey string
					seasonNum  string
					episodeNum string
				}{
					seasonName: seasonName,
					episodeKey: episodeKey,
					seasonNum:  extractSeasonNumber(seasonName),
					episodeNum: extractEpisodeNumber(episodeKey),
				}, nil
			}
		}
	}
	return struct {
		seasonName string
		episodeKey string
		seasonNum  string
		episodeNum string
	}{}, fmt.Errorf("no metadata match found for chapter key: %s", chapterKey)
}

func prepareDestination(metadataDir, videoPath string, match struct {
	seasonName string
	episodeKey string
	seasonNum  string
	episodeNum string
}, chapterKey string) (string, error) {
	title := sanitizeTitle(fmt.Sprintf("Chapter %s", chapterKey))

	nfoPattern := filepath.Join(metadataDir, match.seasonName, fmt.Sprintf("One Pace - S%02sE%02s - *.nfo", match.seasonNum, match.episodeNum))
	nfoFiles, err := filepath.Glob(nfoPattern)
	if err != nil {
		return "", fmt.Errorf("failed to glob nfo files: %w", err)
	}
	if len(nfoFiles) > 0 {
		nfoData, err := os.ReadFile(nfoFiles[0])
		if err != nil {
			return "", fmt.Errorf("failed to read nfo file: %w", err)
		}
		if parsedTitle := extractXMLTag(nfoData, "title"); parsedTitle != "" {
			title = sanitizeTitle(parsedTitle)
		}
	}

	newName := fmt.Sprintf("One Pace - S%02sE%02s - %s%s", match.seasonNum, match.episodeNum, title, filepath.Ext(videoPath))
	finalDir := filepath.Join(metadataDir, match.seasonName)
	if err := os.MkdirAll(finalDir, 0755); err != nil {
		return "", err
	}

	dstPath := filepath.Join(finalDir, newName)
	if err := safeMoveFile(videoPath, dstPath); err != nil {
		return "", err
	}
	return dstPath, nil
}

func extractSeasonNumber(seasonKey string) string {
	// "Season 02" -> "02"
	parts := strings.Fields(seasonKey)
	if len(parts) == 2 {
		return parts[1]
	}
	return "00"
}

func extractEpisodeNumber(episodeKey string) string {
	// "S02E03" -> "03"
	re := regexp.MustCompile(`E(\d+)$`)
	matches := re.FindStringSubmatch(episodeKey)
	if len(matches) == 2 {
		return matches[1]
	}
	return "00"
}

// safeMoveFile moves a file safely
func safeMoveFile(src, dst string) error {
	if err := os.Rename(src, dst); err == nil {
		return nil
	}
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	if _, err := io.Copy(out, in); err != nil {
		return err
	}

	return os.Remove(src)
}

func extractXMLTag(data []byte, tag string) string {
	re := regexp.MustCompile(fmt.Sprintf(`(?i)<%s>(.*?)</%s>`, tag, tag))
	matches := re.FindSubmatch(data)
	if len(matches) >= 2 {
		return strings.TrimSpace(string(matches[1]))
	}
	return ""
}

func sanitizeTitle(title string) string {
	return strings.Map(func(r rune) rune {
		if strings.ContainsRune(`\/:*?"<>|`, r) {
			return -1
		}
		return r
	}, title)
}
