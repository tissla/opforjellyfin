package shared

import (
	"fmt"
	"io"
	"opforjellyfin/internal/logger"
	"os"
	"path/filepath"
	"sync"
)

var (
	// Mutex for directory creation and file movement operations
	dirMutex sync.Mutex
)

// CreateTempTorrentDir safely creates a temporary directory for torrent downloads
func CreateTempTorrentDir(torrentID int) (string, error) {
	dirMutex.Lock()
	defer dirMutex.Unlock()

	tmpDir := filepath.Join(os.TempDir(), fmt.Sprintf("opfor-tmp-%d", torrentID))

	// Check if it already exists
	if info, err := os.Stat(tmpDir); err == nil && info.IsDir() {
		logger.Log(false, "Temp dir already exists: %s", tmpDir)
		return tmpDir, nil
	}

	// Create with MkdirAll (safe for concurrent calls)
	if err := os.MkdirAll(tmpDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create temp dir: %w", err)
	}

	logger.Log(false, "Created temp dir: %s", tmpDir)
	return tmpDir, nil
}

// SafeMoveFile moves a file safely, creates the directory if it does not exist
// This function is thread-safe and handles concurrent file operations
func SafeMoveFile(src, dst string) error {
	// Lock for the entire move operation to ensure atomicity
	dirMutex.Lock()
	defer dirMutex.Unlock()

	logger.Log(false, "sfm: starting move from %s to %s", src, dst)

	// Ensure destination directory exists
	dstDir := filepath.Dir(dst)
	if err := os.MkdirAll(dstDir, 0755); err != nil {
		logger.Log(true, "sfm: failed to create dst dir: %v", err)
		return err
	}

	logger.Log(false, "sfm: copying file from %s to %s", src, dst)
	if err := copyFileInternal(src, dst, 0644); err != nil {
		logger.Log(true, "sfm: copyFile failed: %v", err)
		return err
	}
	logger.Log(false, "sfm: copyFile succeeded")

	logger.Log(false, "sfm: removing source file: %s", src)
	if err := os.Remove(src); err != nil {
		logger.Log(true, "sfm: failed to remove src: %v", err)
		return err
	}
	logger.Log(false, "sfm: source file removed")

	return nil
}

// copyFileInternal is the internal non-locked version for use within already locked functions
func copyFileInternal(src, dst string, perm os.FileMode) error {
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

	return out.Chmod(perm)
}

// CopyFile copies from src to dst with permissions using io.Copy. use os.Stat for permissions or 0644
// This is the public version that locks
func CopyFile(src, dst string, perm os.FileMode) error {
	dirMutex.Lock()
	defer dirMutex.Unlock()

	return copyFileInternal(src, dst, perm)
}

// CreateDirectory safely creates a directory with proper locking
func CreateDirectory(path string) error {
	dirMutex.Lock()
	defer dirMutex.Unlock()

	return os.MkdirAll(path, 0755)
}

// FileExists checks if a file exists (no locking needed for read operation)
func FileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

// CopyDir copies all files (overwrites)
func CopyDir(src, dst string) error {
	dirMutex.Lock()
	defer dirMutex.Unlock()

	return walkAndCopyInternal(src, dst, false)
}

// SyncDir copies new/changed files from src to dst
func SyncDir(src, dst string) error {
	dirMutex.Lock()
	defer dirMutex.Unlock()

	return walkAndCopyInternal(src, dst, true)
}

// walkAndCopyInternal is the internal non-locked version
func walkAndCopyInternal(src, dst string, onlyIfChanged bool) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		destPath := filepath.Join(dst, relPath)

		if info.IsDir() {
			return os.MkdirAll(destPath, 0755)
		}

		if onlyIfChanged && FileExists(destPath) {
			old, err1 := os.ReadFile(destPath)
			new, err2 := os.ReadFile(path)
			if err1 == nil && err2 == nil && string(old) == string(new) {
				return nil
			}
		}

		return copyFileInternal(path, destPath, info.Mode())
	})
}
