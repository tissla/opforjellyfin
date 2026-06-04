package metadata

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"opforjellyfin/internal/shared"
)

type DeleteCandidate struct {
	ChapterRange string
	Season       string
	Title        string
	Path         string
}

func FindEpisodeVideosForDelete(baseDir string, index *shared.MetadataIndex, chapterRanges []string) ([]DeleteCandidate, []string, error) {
	var candidates []DeleteCandidate
	var missing []string
	seen := make(map[string]bool)

	baseAbs, err := filepath.Abs(baseDir)
	if err != nil {
		return nil, nil, fmt.Errorf("could not resolve target directory: %w", err)
	}

	for _, rawRange := range chapterRanges {
		chapterRange, err := normalizeDeleteRange(rawRange)
		if err != nil {
			return nil, nil, err
		}

		found := false
		for seasonKey, season := range index.Seasons {
			for epRange, ep := range season.EpisodeRange {
				if shared.NormalizeDash(epRange) != chapterRange {
					continue
				}

				found = true
				for _, ext := range []string{".mp4", ".mkv"} {
					path := filepath.Join(baseDir, seasonKey, ep.Title+ext)
					if !shared.FileExists(path) {
						continue
					}

					if err := ensurePathInDir(baseAbs, path); err != nil {
						return nil, nil, err
					}

					if seen[path] {
						continue
					}
					seen[path] = true

					candidates = append(candidates, DeleteCandidate{
						ChapterRange: chapterRange,
						Season:       seasonKey,
						Title:        ep.Title,
						Path:         path,
					})
				}
			}
		}

		if !found {
			missing = append(missing, chapterRange)
		}
	}

	return candidates, missing, nil
}

func DeleteEpisodeVideos(candidates []DeleteCandidate) error {
	for _, candidate := range candidates {
		if err := os.Remove(candidate.Path); err != nil {
			return fmt.Errorf("failed to delete %s: %w", candidate.Path, err)
		}
	}

	return nil
}

func normalizeDeleteRange(value string) (string, error) {
	value = strings.TrimSpace(shared.NormalizeDash(value))
	if value == "" {
		return "", fmt.Errorf("empty chapter range")
	}

	if !strings.Contains(value, "-") {
		chapter, err := strconv.Atoi(value)
		if err != nil {
			return "", fmt.Errorf("invalid chapter range: %s", value)
		}
		return fmt.Sprintf("%d-%d", chapter, chapter), nil
	}

	parts := strings.Split(value, "-")
	if len(parts) != 2 {
		return "", fmt.Errorf("invalid chapter range: %s", value)
	}

	start, err := strconv.Atoi(strings.TrimSpace(parts[0]))
	if err != nil {
		return "", fmt.Errorf("invalid chapter range: %s", value)
	}
	end, err := strconv.Atoi(strings.TrimSpace(parts[1]))
	if err != nil {
		return "", fmt.Errorf("invalid chapter range: %s", value)
	}
	if start > end {
		return "", fmt.Errorf("range start cannot be greater than range end: %s", value)
	}

	return fmt.Sprintf("%d-%d", start, end), nil
}

func ensurePathInDir(baseAbs, path string) error {
	pathAbs, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("could not resolve path %s: %w", path, err)
	}

	rel, err := filepath.Rel(baseAbs, pathAbs)
	if err != nil {
		return fmt.Errorf("could not compare %s to target directory: %w", path, err)
	}

	if rel == "." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) || rel == ".." || filepath.IsAbs(rel) {
		return fmt.Errorf("refusing to delete path outside target directory: %s", path)
	}

	return nil
}
