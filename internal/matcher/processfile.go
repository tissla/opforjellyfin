package matcher

import (
	"fmt"
	"opforjellyfin/internal/shared"
	"os"
	"path/filepath"
)

// FindVideoFiles accepts a single video or collects videos from a directory.
func FindVideoFiles(source string) ([]string, error) {
	info, err := os.Lstat(source)
	if err != nil {
		return nil, err
	}
	if !info.IsDir() {
		if info.Mode().IsRegular() && shared.IsVideoFile(source) {
			return []string{source}, nil
		}
		return nil, fmt.Errorf("not a video file or directory: %s", source)
	}
	var paths []string
	err = filepath.WalkDir(source, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			if path != source && (entry.Name() == ".opfor-tmp" || entry.Name() == ".git") {
				return filepath.SkipDir
			}
			return nil
		}
		if entry.Type().IsRegular() && shared.IsVideoFile(entry.Name()) {
			paths = append(paths, path)
		}
		return nil
	})
	return paths, err
}

// ProcessTorrentFiles places the actual video paths supplied by the torrent.
func ProcessTorrentFiles(outDir string, td *shared.TorrentDownload, index *shared.MetadataIndex) error {
	if len(td.VideoPaths) == 0 {
		td.PlacementProgress = "No video files found"
		shared.SaveTorrentDownload(td)
		return fmt.Errorf("no video files found in %s", td.FullTitle)
	}
	failed := 0
	for i, videoPath := range td.VideoPaths {
		td.PlacementProgress = fmt.Sprintf("Placing %d/%d", i+1, len(td.VideoPaths))
		message, err := MatchAndPlaceVideo(videoPath, outDir, index)
		if err != nil {
			failed++
			message = fmt.Sprintf("Left in %s: %v", videoPath, err)
		} else {
			shared.RecordPlacedVideo()
		}
		if message != "" {
			td.PlacementFull = append(td.PlacementFull, message)
		}
		shared.SaveTorrentDownload(td)
	}
	if failed > 0 {
		td.PlacementProgress = fmt.Sprintf("%d file(s) left in download directory", failed)
		shared.SaveTorrentDownload(td)
		return fmt.Errorf("%d file(s) could not be placed", failed)
	}
	td.MarkPlaced(fmt.Sprintf("%d video(s) placed", len(td.VideoPaths)))
	return nil
}
