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

type season struct {
	sNum      int
	sName     string
	videos    int
	totalnfos int
}

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

		var seasonFolders []season
		metaIndex := metadata.LoadMetadataCache()

		for _, f := range files {
			if !f.IsDir() {
				continue
			}

			if f.Name() == "strayvideos" {
				continue
			}

			subdir := filepath.Join(cfg.TargetDir, f.Name())

			v, nfo := metadata.CountVideosAndTotal(subdir)

			extNum := shared.ExtractSeasonNumber(f.Name())
			sNum, _ := strconv.Atoi(extNum)

			sName := ""
			if metaIndex != nil {
				if seasonData, exists := metaIndex.Seasons[f.Name()]; exists {
					sName = seasonData.Name
				}
			}

			seasonFolders = append(seasonFolders, season{
				sNum:      sNum,
				sName:     sName,
				videos:    v,
				totalnfos: nfo,
			})

		}

		sort.Slice(seasonFolders, func(i, j int) bool {
			return seasonFolders[i].sNum < seasonFolders[j].sNum
		})

		fmt.Printf("📦 Seasons Downloaded: %d\n", len(seasonFolders))

		fmt.Println("\n📁 Season folders:")
		for _, s := range seasonFolders {
			formattedPrint := styleSeasonPrint(s)
			fmt.Printf("   - %s\n", formattedPrint)
		}

	},
}

func styleSeasonPrint(s season) string {

	vidStr := fmt.Sprintf("%4d", s.videos)
	nfoStr := fmt.Sprintf("%-3d", s.totalnfos)

	vids := ui.StyleByRange(vidStr, 0, s.totalnfos)
	nfos := ui.StyleByRange(nfoStr, 0, s.totalnfos)

	stringnum := fmt.Sprintf("%d", s.sNum)
	snum := ui.AnsiPadLeft(ui.StyleFactory(stringnum, ui.Style.Pink), 3)

	sname := ui.StyleFactory(s.sName, ui.Style.LBlue)

	if s.sNum == 0 {
		return fmt.Sprintf("Specials  : %s / %s", vids, nfos)
	}

	leftPart := fmt.Sprintf("Season %s: %s / %s", snum, vids, nfos)

	paddedLeft := ui.AnsiPadRight(leftPart, 25)

	return fmt.Sprintf("%s | %s", paddedLeft, sname)
}

func init() {
	infoCmd.Flags().BoolVarP(&verboseInfo, "verbose", "v", false, "Show season folder names")
	rootCmd.AddCommand(infoCmd)
}
