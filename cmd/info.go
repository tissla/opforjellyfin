// cmd/info.go
package cmd

import (
	"fmt"
	"opforjellyfin/internal/shared"
	"opforjellyfin/internal/ui"
	"os"
	"path/filepath"
	"sort"
	"strconv"
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

			v, nfo := CountVideosAndTotal(subdir)
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

// move this
// counts videos that match with .nfo files. returns: matching videos, total .nfo files (excluding season.nfo)
func CountVideosAndTotal(dir string) (matched int, totalNFO int) {
	videoFiles := map[string]bool{}

	err := filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			return nil
		}

		lower := strings.ToLower(d.Name())
		base := strings.TrimSuffix(d.Name(), filepath.Ext(d.Name()))

		if strings.HasSuffix(lower, ".mkv") || strings.HasSuffix(lower, ".mp4") {
			videoFiles[base] = false
		}

		if shared.IsEpisodeNFO(lower) {
			totalNFO++
			if _, exists := videoFiles[base]; exists {
				videoFiles[base] = true //
			}
		}

		return nil
	})

	if err != nil {
		return 0, 0
	}

	// count matched
	for _, matchedFlag := range videoFiles {
		if matchedFlag {
			matched++
		}
	}

	return matched, totalNFO
}

func init() {
	infoCmd.Flags().BoolVarP(&verboseInfo, "verbose", "v", false, "Show season folder names")
	rootCmd.AddCommand(infoCmd)
}
