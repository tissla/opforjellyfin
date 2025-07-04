package matcher

import (
	"fmt"
	"opforjellyfin/internal/logger"
	"opforjellyfin/internal/shared"
	"os"
	"path/filepath"
	"strings"
)

// walks through downloaded files and tries to place them in correct dir
func ProcessTorrentFiles(tmpDir, outDir string, td *shared.TorrentDownload, index *shared.MetadataIndex) {
	filesChecked := 0
	filesPlaced := 0
	var lastError error

	// collect all paths
	td.PlacementProgress = fmt.Sprintf("🔧 Finding files to place %s", tmpDir)

	var vidPaths []string
	err := filepath.Walk(tmpDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			logger.Log(true, "Failed walking file: %v", err)
			return nil
		}
		if info.IsDir() || (!strings.HasSuffix(strings.ToLower(info.Name()), ".mkv") && !strings.HasSuffix(strings.ToLower(info.Name()), ".mp4")) {
			return nil
		}
		logger.Log(false, "added path: %s", path)
		vidPaths = append(vidPaths, path)
		return nil
	})

	if err != nil {
		logger.Log(true, "Error walking tmpDir: %v", err)
		return
	}

	// Handle case where no video files found
	if len(vidPaths) == 0 {
		td.MarkPlaced("⚠️ No video files found to place!")
		return
	}

	for _, path := range vidPaths {
		logger.Log(false, "→ Found: %s", path)
		filesChecked++

		// readable src for msg
		fileName := filepath.Base(path)
		readablePath := fileName
		if len(fileName) > 10 {
			readablePath = fileName[10:]
		}

		// upd msg
		td.PlacementProgress = fmt.Sprintf("🔧 Placing ➝ %d/%d - %s", (filesPlaced + 1), len(vidPaths), readablePath)

		// match and place
		msg, err := MatchAndPlaceVideo(path, outDir, index, td.ChapterRange)
		if err != nil {
			logger.Log(true, "Error placing file: %v", err)
			lastError = err
		} else if msg != "" {
			filesPlaced++
			//save msg for final summary
			td.PlacementFull = append(td.PlacementFull, msg)
			shared.SaveTorrentDownload(td)
		}
	}

	// Create appropriate message based on results
	var placedMsg string

	if filesPlaced == 0 && lastError != nil {
		placedMsg = fmt.Sprintf("❌ Failed to place any files! Last error: %v", lastError)
	} else if filesPlaced == 0 {
		placedMsg = "❌ No files could be placed!"
	} else if filesPlaced == len(vidPaths) {
		if filesPlaced == 1 {
			placedMsg = "✅ 1 file placed!"
		} else {
			placedMsg = fmt.Sprintf("✅ All %d files placed!", filesPlaced)
		}
	} else {
		// Partial success
		placedMsg = fmt.Sprintf("⚠️ %d/%d files placed!", filesPlaced, len(vidPaths))
	}

	td.MarkPlaced(placedMsg)
	logger.Log(false, "File placement done: %d checked, %d placed", filesChecked, filesPlaced)
}
