package cmd

// import (
// 	"fmt"
// 	"github.com/spf13/cobra"
// 	"opforjellyfin/internal/logger"
// 	"opforjellyfin/internal/matcher"
// 	"opforjellyfin/internal/metadata"
// 	"opforjellyfin/internal/scraper"
// 	"opforjellyfin/internal/shared"
// 	"opforjellyfin/internal/torrent"
// 	"opforjellyfin/internal/ui"
// 	"os"
// 	"strconv"
// )

// var (
// 	forceChapter string
// )

// var importCmd = &cobra.Command{
// 	Use:   "import [folder...]",
// 	Short: "Import One Pace episodes from one or more files/folders. By default, it searches the strayvideos files.",
// 	Run: func(cmd *cobra.Command, args []string) {

// add spinner
// spinner := ui.NewSpinner("🗃️ Preparing import.. ", ui.Animations["MetaFetcher"])

// sourcePaths := make([]string, 0, max(len(args), 1))

// if len(args) < 1 {
// 	sourcePaths[0] = "strayvideos"
// }

// cfg := shared.LoadConfig()
// if cfg.TargetDir == "" {
// 	logger.Log(true, "⚠️ No target directory set. Use 'setDir <path>' first.")
// 	return
// }

// if cfg.Source.BaseURL == "" {
// 	logger.Log(true, "No valid scraper configuration found. Please run 'sync'")
// }

// // stop spinner
// spinner.Stop()
// var matches map[string]shared.EpisodeData

// for _, arg := range args {

// 	chapterRange := "0-0"

// 	// maybe rewrite this part
// 	if forceChapter != "" {
// 		if len(args) > 1 {
// 			logger.Log(true, "❌ --forceChapter may only be used with a single folder/file")
// 		}
// 		chapterRange = forceChapter
// 	}

// 	//dKey := ui.StyleFactory(fmt.Sprintf("%4d", match.DownloadKey), ui.Style.Pink)
// 	//title := ui.StyleFactory(match.TorrentName, ui.Style.LBlue)

// 	//logger.Log(true, "🔍 Matched DownloadKey %s → %s (%s) [%s]", dKey, title, match.Quality, match.ChapterRange)
// 	//logger.Log(true, "🎬 Starting download: %s (%s)\n", match.TorrentName, match.Quality)
// 	//matches = append(matches, *match)
// }

// if len(matches) == 0 {
// 	fmt.Println("⚠️  No files to process.")
// 	os.Exit(0)
// }

// // Load metadata index once
// metadataIndex := metadata.LoadMetadataCache()

// matcher.ProcessTorrentFiles(tmpDir, cfg.TargetDir, td, metadataIndex)
// 	},
// }

// func init() {
// 	importCmd.Flags().StringVar(&forceChapter, "forceChapter", "", "Override chapter range (only for single downloadKey)")

// 	rootCmd.AddCommand(importCmd)
// }
