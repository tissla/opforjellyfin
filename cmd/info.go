// cmd/info.go
package cmd

import (
	"fmt"
	"opforjellyfin/internal/metadata"
	"opforjellyfin/internal/shared"
	"opforjellyfin/internal/ui"
	"os"
	"path/filepath"
	"sort"
	"strconv"

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
			fmt.Println("⚠️ No target directory set. Use 'opforjellyfin setDir <path>'")
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

		type season struct {
			sNum  int
			label string
		}
		var seasonFolders []season

		for _, f := range files {
			if !f.IsDir() {
				continue
			}

			subdir := filepath.Join(cfg.TargetDir, f.Name())

			v, nfo := metadata.CountVideosAndTotal(subdir)
			if v != 0 {
				extNum := shared.ExtractSeasonNumber(f.Name())
				sNum, _ := strconv.Atoi(extNum)

				dNum := ui.AnsiPadLeft(ui.StyleFactory(extNum, ui.Style.Pink), 3)
				sLabel := fmt.Sprintf("Season %s: %d/%d", dNum, v, nfo)
				seasonFolders = append(seasonFolders, season{
					sNum:  sNum,
					label: sLabel,
				})

			}
		}

		sort.Slice(seasonFolders, func(i, j int) bool {
			return seasonFolders[i].sNum < seasonFolders[j].sNum
		})

		fmt.Printf("📦 Seasons Downloaded: %d\n", len(seasonFolders))

		if len(seasonFolders) > 0 {
			fmt.Println("\n📁 Season folders:")
			for _, s := range seasonFolders {
				fmt.Printf("   - %s\n", s.label)
			}
		}
	},
}

func init() {
	infoCmd.Flags().BoolVarP(&verboseInfo, "verbose", "v", false, "Show season folder names")
	rootCmd.AddCommand(infoCmd)
}
