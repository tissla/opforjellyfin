// metadata/metadata.go
package metadata

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

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

	shared.SaveConfig(*cfg)

	spinner.Stop()

	path := filepath.Join(baseDir, "metadata-index.json")
	fmt.Println("\n✅ Saved metadata index to", path)

	fmt.Println("✅ Metadata fetch and indexing complete.")
	return nil
}
