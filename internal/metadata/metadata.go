// Package metadata
package metadata

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"opforjellyfin/internal/logger"
	"opforjellyfin/internal/shared"
	"opforjellyfin/internal/ui"
)

// FetchAllMetadata downloads the metadata and its generated index.
func FetchAllMetadata(cfg *shared.Config) error {
	return cloneAndCopyRepo(cfg, false)
}

// SyncMetadata clones and syncs metadata updates from GitHub.
func SyncMetadata(cfg *shared.Config) error {
	return cloneAndCopyRepo(cfg, true)
}

// Copy the metadata repository into the configured library.
func cloneAndCopyRepo(cfg *shared.Config, syncOnly bool) error {

	baseDir := cfg.TargetDir

	tmpBase, err := shared.GetTempDir()
	if err != nil {
		return err
	}
	tmpDir, err := os.MkdirTemp(tmpBase, "repo-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmpDir)

	repo := fmt.Sprintf("https://github.com/%s.git", shared.MetadataRepo)

	fmt.Printf("%s", "🌐 Fetching metadata from "+repo+"\n")

	spinner := ui.NewSpinner("🗃️ Downloading.. ", ui.Animations["MetaFetcher"])

	if err = cloneMetadata(repo, tmpDir, cfg.Thumbnails); err != nil {
		spinner.Stop()
		return err
	}

	srcDir := filepath.Join(tmpDir, "One Pace")
	index, err := readIndex(srcDir)
	if err != nil {
		spinner.Stop()
		return err
	}

	if syncOnly {
		err = shared.SyncDir(srcDir, baseDir)
	} else {
		err = shared.CopyDir(srcDir, baseDir)
	}

	if err != nil {
		spinner.Stop()
		return err
	}

	metadataCache = index
	metadataCacheErr = nil

	if err := loadSourceConfig(tmpDir, cfg); err != nil {
		logger.Log(false, "metadata: could not load scraper config from repo: %v", err)
	}

	if err := shared.SaveConfig(*cfg); err != nil {
		spinner.Stop()
		return fmt.Errorf("could not save config: %w", err)
	}

	spinner.Stop()

	path := filepath.Join(baseDir, "metadata-index.json")
	fmt.Println("\n✅ Saved metadata index to", path)

	fmt.Println("✅ Metadata sync complete.")
	return nil
}

// Exclude thumbnail blobs before checkout so they are never requested.
func cloneMetadata(repo, dir string, thumbnails bool) error {
	args := []string{"clone", "--depth=1"}
	if !thumbnails {
		args = append(args, "--filter=blob:none", "--no-checkout")
	}
	args = append(args, repo, dir)
	if output, err := exec.Command("git", args...).CombinedOutput(); err != nil {
		return fmt.Errorf("git clone: %w: %s", err, strings.TrimSpace(string(output)))
	}
	if thumbnails {
		return nil
	}

	sparse := exec.Command("git", "-C", dir, "sparse-checkout", "set", "--no-cone", "--stdin")
	sparse.Stdin = strings.NewReader("/*\n!*-thumb.png\n!*-thumb.jpg\n")
	if output, err := sparse.CombinedOutput(); err != nil {
		return fmt.Errorf("select metadata files: %w: %s", err, strings.TrimSpace(string(output)))
	}
	if output, err := exec.Command("git", "-C", dir, "checkout", "HEAD").CombinedOutput(); err != nil {
		return fmt.Errorf("checkout metadata: %w: %s", err, strings.TrimSpace(string(output)))
	}
	return nil
}

// loadSourceConfig reads config.json from the freshly cloned metadata repo and
// populates cfg.Source (the scraper's site config) with it, so switching trackers
// is a config change in the metadata repo rather than a code change here.
func loadSourceConfig(tmpDir string, cfg *shared.Config) error {
	cfgFile := filepath.Join(tmpDir, "config.json")

	data, err := os.ReadFile(cfgFile)
	if err != nil {
		return fmt.Errorf("could not read %s: %w", cfgFile, err)
	}

	var srcConfig shared.ScraperConfig
	if err := json.Unmarshal(data, &srcConfig); err != nil {
		return fmt.Errorf("could not parse %s: %w", cfgFile, err)
	}

	cfg.Source = srcConfig
	return nil
}
