package shared

import (
	"io"
	"opforjellyfin/internal/logger"
	"os"
	"path/filepath"
)

// safeMoveFile moves a file safely, creates the directory if it does not exist
func SafeMoveFile(src, dst string) error {
	logger.Log(false, "sfm: starting move from %s to %s", src, dst)

	logger.Log(false, "sfm: copying file from %s to %s", src, dst)
	if err := CopyFile(src, dst, 0644); err != nil {
		logger.Log(true, "sfm: copyFile failed: %v", err)
		return err
	}
	logger.Log(false, "sfm: copyFile succeeded")

	if err := os.Chmod(dst, 0644); err != nil {
		logger.Log(true, "sfm: chmod failed: %v", err)
		return err
	}

	logger.Log(false, "sfm: removing source file: %s", src)
	if err := os.Remove(src); err != nil {
		logger.Log(true, "sfm: failed to remove src: %v", err)
		return err
	}
	logger.Log(false, "sfm: source file removed")

	return nil
}

// copyFile copies from src to dst with permissions using io.Copy. use os.Stat for permissions or 0644
func CopyFile(src, dst string, perm os.FileMode) error {
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

	return os.Chmod(dst, perm)
}

// bool
func FileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

// copies all files (overwrites)
func CopyDir(src, dst string) error {
	return walkAndCopy(src, dst, false)
}

// SyncDir copies new/changed files from src to dst
func SyncDir(src, dst string) error {
	return walkAndCopy(src, dst, true)
}
func walkAndCopy(src, dst string, onlyIfChanged bool) error {
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

		return CopyFile(path, destPath, info.Mode())
	})
}
