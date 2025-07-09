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
func FetchAllMetadata(baseDir string, cfg shared.Config) error {
	return cloneAndCopyRepo(baseDir, cfg, false)
}

// SyncMetadata clones and syncs metadata updates from GitHub.
func SyncMetadata(baseDir string, cfg shared.Config) error {
	return cloneAndCopyRepo(baseDir, cfg, true)
}

// Main dataobtainer, builds or rebuilds index when complete.
func cloneAndCopyRepo(baseDir string, cfg shared.Config, syncOnly bool) error {
	tmpDir := filepath.Join(os.TempDir(), "repo-tmp")
	defer os.RemoveAll(tmpDir)

	repo := fmt.Sprintf("https://github.com/%s.git", cfg.GitHubRepo)

	fmt.Printf("%s", "🌐 Fetching metadata from "+repo+"\n")

	spinner := ui.NewSpinner("🗃️ Downloading.. ", ui.Animations["MetaFetcher"])

	if err := exec.Command("git", "clone", "--depth=1", repo, tmpDir).Run(); err != nil {
		spinner.Stop()
		return fmt.Errorf("git clone failed: %w", err)
	}

	srcDir := filepath.Join(tmpDir, "One Pace")
	var err error

	if syncOnly {
		err = shared.SyncDir(srcDir, baseDir)
	} else {
		err = shared.CopyDir(srcDir, baseDir)
	}

	if err != nil {
		spinner.Stop()
		return fmt.Errorf("failed to copy metadata: %w", err)
	}

	if err := BuildMetadataIndex(baseDir); err != nil {
		spinner.Stop()
		return fmt.Errorf("failed to build metadata index: %w", err)
	}

	updateConfig(tmpDir, cfg)

	spinner.Stop()

	path := filepath.Join(baseDir, "metadata-index.json")
	fmt.Println("\n✅ Saved metadata index to", path)

	fmt.Println("✅ Metadata fetch and indexing complete.")
	return nil
}

func updateConfig(tmpDir string, cfg shared.Config) {
	cfgFile := filepath.Join(tmpDir, "config.json")

	data, _ := os.ReadFile(cfgFile)

	var srcConfig shared.ScraperConfig
	if err := json.Unmarshal(data, &srcConfig); err != nil {
		logger.Log(false, "Error updating source config: %v", err)
	}

	cfg.Source = srcConfig
	shared.SaveConfig(cfg)
}
