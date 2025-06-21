package matcher

import (
	"fmt"
	"io"
	"opforjellyfin/internal/logger"
	"opforjellyfin/internal/shared"
	"os"
	"path/filepath"
	"strings"
)

func ProcessTorrentFiles(tmpDir, outDir string, td *shared.TorrentDownload) {
	filesChecked := 0
	filesPlaced := 0

	// collect all paths
	var vidPaths []string
	err := filepath.Walk(tmpDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			logger.DebugLog(true, "Failed walking file: %v", err)
			return nil
		}
		if info.IsDir() || !strings.HasSuffix(strings.ToLower(info.Name()), ".mkv") && !strings.HasSuffix(strings.ToLower(info.Name()), ".mp4") {
			return nil
		}
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

		msg, err := MatchAndPlaceVideo(path, outDir)
		if err != nil {
			logger.DebugLog(true, "Error placing file: %v", err)
			td.Error += fmt.Sprintf("%s\n", err)
		} else if msg != "" {
			filesPlaced++
			td.Messages = append(td.Messages, msg)
			shared.SaveTorrentDownload(td)
		}
	}

	logger.DebugLog(false, "📦 File placement done: %d checked, %d placed", filesChecked, filesPlaced)
}

// safeMoveFile moves a file safely, creates the directory if it does not exist
func SafeMoveFile(src, dst string) error {
	logger.DebugLog(false, "sfm: 🔧 starting move from %s to %s", src, dst)

	// Steg 1: Skapa målmapp om den inte finns
	logger.DebugLog(false, "sfm: 📁 ensuring destination directory exists: %s", filepath.Dir(dst))
	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		logger.DebugLog(true, "sfm: MkdirAll failed: %v", err)
		return err
	}

	// Steg 2: Försök rena
	logger.DebugLog(false, "sfm: 🔄 trying os.Rename")
	if err := os.Rename(src, dst); err == nil {
		logger.DebugLog(false, "sfm: os.Rename succeeded")
		return nil
	} else {
		logger.DebugLog(false, "sfm: os.Rename failed, falling back to manual copy: %v", err)
	}

	// Steg 3: Öppna src
	logger.DebugLog(false, "sfm: opening source file: %s", src)
	in, err := os.Open(src)
	if err != nil {
		logger.DebugLog(true, "sfm: failed to open src: %v", err)
		return err
	}
	defer func() {
		logger.DebugLog(false, "sfm: closed source file")
		in.Close()
	}()

	// Steg 4: Skapa destination
	logger.DebugLog(false, "sfm: creating destination file: %s", dst)
	out, err := os.Create(dst)
	if err != nil {
		logger.DebugLog(true, "sfm: failed to create dst: %v", err)
		return err
	}
	defer func() {
		logger.DebugLog(false, "sfm: closed destination file")
		out.Close()
	}()

	// Steg 5: Kopiera innehåll
	logger.DebugLog(false, "sfm: copying data...")
	nBytes, err := io.Copy(out, in)
	if err != nil {
		logger.DebugLog(true, "sfm: copy failed: %v", err)
		return err
	}
	logger.DebugLog(false, "sfm: copied %d bytes", nBytes)

	// Steg 6: Sätt rättigheter
	logger.DebugLog(false, "sfm: chmod destination")
	if err := out.Chmod(0644); err != nil {
		logger.DebugLog(true, "sfm: chmod failed: %v", err)
		return err
	}

	// Steg 7: Radera src
	logger.DebugLog(false, "sfm: removing src-file: %s", src)
	if err := os.Remove(src); err != nil {
		logger.DebugLog(true, "sfm: failed to remove src: %v", err)
		return err
	}
	logger.DebugLog(false, "sfm: successfully removed src")

	return nil
}
