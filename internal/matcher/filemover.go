package matcher

import (
	"fmt"
	"opforjellyfin/internal/logger"
	"opforjellyfin/internal/shared"
	"os"
	"path/filepath"
	"strings"
)

func ProcessTorrentFiles(tmpDir, outDir string, td *shared.TorrentDownload, index *shared.MetadataIndex) {
	filesChecked := 0
	filesPlaced := 0

	// collect all paths

	td.ProgressMessage = fmt.Sprintf("Walking through directory %s", tmpDir)

	var vidPaths []string
	err := filepath.Walk(tmpDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			logger.DebugLog(true, "Failed walking file: %v", err)
			return nil
		}
		if info.IsDir() || !strings.HasSuffix(strings.ToLower(info.Name()), ".mkv") && !strings.HasSuffix(strings.ToLower(info.Name()), ".mp4") {
			return nil
		}
		logger.DebugLog(false, "added path: %s", path)
		vidPaths = append(vidPaths, path)
		return nil
	})

	if err != nil {
		logger.DebugLog(true, "Error walking tmpDir: %v", err)
		td.Error += fmt.Sprintf(" Walk error: %v\n", err)
		return
	}

	for _, path := range vidPaths {
		logger.DebugLog(false, "→ Found: %s", path)
		filesChecked++

		td.ProgressMessage = fmt.Sprintf("Placing.. %s", filepath.Base(path))

		msg, err := MatchAndPlaceVideo(path, outDir, index)
		if err != nil {
			logger.DebugLog(true, "Error placing file: %v", err)
			td.Error += fmt.Sprintf("%s\n", err)
		} else if msg != "" {
			filesPlaced++
			td.ProgressMessage = msg
			td.Messages = append(td.Messages, msg)
			shared.SaveTorrentDownload(td)
		}
	}

	td.Placed = true
	shared.SaveTorrentDownload(td)

	logger.DebugLog(false, "File placement done: %d checked, %d placed", filesChecked, filesPlaced)
}

// safeMoveFile moves a file safely, creates the directory if it does not exist
func SafeMoveFile(src, dst string) error {
	logger.DebugLog(false, "sfm: starting move from %s to %s", src, dst)

	dstDir := filepath.Dir(dst)
	logger.DebugLog(false, "sfm: ensuring destination directory exists: %s", dstDir)
	if err := os.MkdirAll(dstDir, 0755); err != nil {
		logger.DebugLog(true, "sfm: MkdirAll failed: %v", err)
		return err
	}

	logger.DebugLog(false, "sfm: copying file from %s to %s", src, dst)
	if err := shared.CopyFile(src, dst, 0644); err != nil {
		logger.DebugLog(true, "sfm: copyFile failed: %v", err)
		return err
	}
	logger.DebugLog(false, "sfm: copyFile succeeded")

	if err := os.Chmod(dst, 0644); err != nil {
		logger.DebugLog(true, "sfm: chmod failed: %v", err)
		return err
	}

	logger.DebugLog(false, "sfm: removing source file: %s", src)
	if err := os.Remove(src); err != nil {
		logger.DebugLog(true, "sfm: failed to remove src: %v", err)
		return err
	}
	logger.DebugLog(false, "sfm: source file removed")

	return nil
}
