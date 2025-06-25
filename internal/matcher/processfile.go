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

	// collect all paths

	td.PlacementProgress = fmt.Sprintf("🔧 Finding files to place %s", tmpDir)

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
		return
	}

	for _, path := range vidPaths {
		logger.DebugLog(false, "→ Found: %s", path)
		filesChecked++

		// readable src for msg
		readablePath := filepath.Base(path)[10:]
		// upd msg
		td.PlacementProgress = fmt.Sprintf("🔧 Placing ➝ %d/%d - %s", (filesPlaced + 1), len(vidPaths), readablePath)

		// match and place
		msg, err := MatchAndPlaceVideo(path, outDir, index, td.ChapterRange)
		if err != nil {
			logger.DebugLog(true, "Error placing file: %v", err)
		} else if msg != "" {
			filesPlaced++
			//save msg for final summary
			td.PlacementFull = append(td.PlacementFull, msg)
			shared.SaveTorrentDownload(td)
		}
	}

	// placed
	placedMsg := fmt.Sprintf("✅ %d file placed!", filesPlaced)
	if filesPlaced > 1 {
		placedMsg = fmt.Sprintf("✅ %d/%d files placed!", filesPlaced, len(vidPaths))
	}

	td.MarkPlaced(placedMsg)

	logger.DebugLog(false, "File placement done: %d checked, %d placed", filesChecked, filesPlaced)
}
