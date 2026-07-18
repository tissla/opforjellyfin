// metadata/metadata.go
package metadata

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"opforjellyfin/internal/logger"
	"opforjellyfin/internal/shared"
	"opforjellyfin/internal/ui"
)

// BuildMetadataIndex constructs and caches metadata index.
func BuildMetadataIndex(baseDir string) error {
	index, err := buildIndexFromDir(baseDir)
	if err != nil {
		return err
	}

	return saveMetadataIndex(index, baseDir)
}

// FetchAllMetadata clones and indexes metadata from GitHub.
func FetchAllMetadata(cfg *shared.Config) error {
	return cloneAndCopyRepo(cfg, false)
}

// SyncMetadata clones and syncs metadata updates from GitHub.
func SyncMetadata(cfg *shared.Config) error {
	return cloneAndCopyRepo(cfg, true)
}

// Main dataobtainer, builds or rebuilds index when complete.
func cloneAndCopyRepo(cfg *shared.Config, syncOnly bool) error {

	baseDir := cfg.TargetDir

	tmpBase, err := shared.GetTempDir()
	if err != nil {
		return err
	}
	tmpDir := filepath.Join(tmpBase, "repo-tmp")
	defer os.RemoveAll(tmpDir)

	repo := fmt.Sprintf("https://github.com/%s.git", cfg.GitHubRepo)

	fmt.Printf("%s", "🌐 Fetching metadata from "+repo+"\n")

	spinner := ui.NewSpinner("🗃️ Downloading.. ", ui.Animations["MetaFetcher"])

	if err = exec.Command("git", "clone", "--depth=1", repo, tmpDir).Run(); err != nil {
		spinner.Stop()
		fmt.Println("⚠️  Git clone failed: %w", err)
		return err
	}

	srcDir := filepath.Join(tmpDir, "One Pace")

	if syncOnly {
		err = shared.SyncDir(srcDir, baseDir)
	} else {
		err = shared.CopyDir(srcDir, baseDir)
	}

	if err != nil {
		spinner.Stop()
		return err
	}

	if err := BuildMetadataIndex(baseDir); err != nil {
		spinner.Stop()
		return err
	}

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

	fmt.Println("✅ Metadata fetch and indexing complete.")
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
