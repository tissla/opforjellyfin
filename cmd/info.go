// cmd/info.go
package cmd

import (
	"fmt"
	"opforjellyfin/internal/metadata"
	"opforjellyfin/internal/shared"
	"opforjellyfin/internal/ui"
	"path/filepath"
	"sort"

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
		cfg, err := shared.LoadConfig()
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Println("🔧 Current Configuration:")
		fmt.Printf("📂 Target Directory: %s\n", cfg.TargetDir)

		if cfg.TargetDir == "" {
			fmt.Println("⚠️ No target directory set. Use 'opforjellyfin setDir <path>'")
			return
		}

		if verboseInfo {
			fmt.Printf("📡 Torrent Provider: %s\n", cfg.Source.BaseURL)
			fmt.Printf("🐙 Metadata Source:  https://github.com/%s\n", cfg.GitHubRepo)
		}

		var seasonFolders []season
		metaIndex, err := metadata.LoadMetadataCache()
		if err != nil {
			fmt.Println(err)
			return
		}
		for _, arc := range metaIndex.Arcs {
			v, total := metadata.CountVideosAndTotal(filepath.Join(cfg.TargetDir, arc.Folder))
			seasonFolders = append(seasonFolders, season{sNum: arc.SeasonNumber, sName: arc.Name, videos: v, totalnfos: total})
		}

		sort.Slice(seasonFolders, func(i, j int) bool {
			return seasonFolders[i].sNum < seasonFolders[j].sNum
		})

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
