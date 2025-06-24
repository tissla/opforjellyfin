// cmd/info.go
package cmd

import (
	"fmt"
	"opforjellyfin/internal/shared"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/spf13/cobra"
)

var verboseInfo bool

var infoCmd = &cobra.Command{
	Use:   "info",
	Short: "Show current configuration and library status",
	Run: func(cmd *cobra.Command, args []string) {
		cfg := shared.LoadConfig()
		fmt.Println("🔧 Current Configuration:")
		fmt.Printf("📂 Target Directory: %s\n", cfg.TargetDir)

		if cfg.TargetDir == "" {
			fmt.Println("⚠️  No target directory set. Use 'opforjellyfin setDir <path>'")
			return
		}

		files, err := os.ReadDir(cfg.TargetDir)
		if err != nil {
			fmt.Printf("❌ Could not read target directory: %v\n", err)
			return
		}

		if verboseInfo {
			fmt.Printf("📡 Torrent Provider: %s\n", cfg.TorrentAPIURL)
			fmt.Printf("🐙 Metadata Source:  https://github.com/%s\n", cfg.GitHubRepo)
		}

		var seasonFolders []string

		for _, f := range files {
			if !f.IsDir() {
				continue
			}

			subdir := filepath.Join(cfg.TargetDir, f.Name())
			entries, err := os.ReadDir(subdir)
			if err != nil {
				continue
			}

			for _, entry := range entries {
				if !entry.IsDir() && (strings.HasSuffix(entry.Name(), ".mkv") || strings.HasSuffix(entry.Name(), ".mp4")) {
					seasonFolders = append(seasonFolders, f.Name())
					break
				}
			}
		}

		sort.Strings(seasonFolders)
		fmt.Printf("📦 Seasons Downloaded: %d\n", len(seasonFolders))

		if verboseInfo && len(seasonFolders) > 0 {
			fmt.Println("\n📁 Season folders:")
			for _, s := range seasonFolders {
				fmt.Printf("   - %s\n", s)
			}
		}
	},
}

func init() {
	infoCmd.Flags().BoolVarP(&verboseInfo, "verbose", "v", false, "Show season folder names")
	rootCmd.AddCommand(infoCmd)
}
