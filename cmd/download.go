package cmd

import (
	"fmt"
	"os"
	"strconv"

	"opforjellyfin/internal/logger"
	"opforjellyfin/internal/scraper"
	"opforjellyfin/internal/shared"
	"opforjellyfin/internal/torrent"
	"opforjellyfin/internal/ui"

	"github.com/spf13/cobra"
)

var (
	//follow   bool
	quality  string
	forceKey string
)

var downloadCmd = &cobra.Command{
	Use:   "download <downloadKey> [downloadKey...]",
	Short: "Download one or more One Pace torrents",
	Run: func(cmd *cobra.Command, args []string) {

		// add spinner
		spinner := ui.NewMultirowSpinner(ui.Animations["DownloadPrep"], 3)

		if len(args) < 1 {
			logger.DebugLog(true, "⚠️ You must specify atleast one download-key")
			return
		}

		cfg := shared.LoadConfig()
		if cfg.TargetDir == "" {
			logger.DebugLog(true, "⚠️ No target directory set. Use 'setDir <path>' first.")
			return
		}

		torrentList, err := scraper.FetchOnePaceTorrents()
		if err != nil {
			logger.DebugLog(true, "❌ Error scraping torrents. Site inaccessible? %v", err)
			return
		}

		// stop spinner
		spinner.Stop()
		var matches []shared.TorrentEntry
		for _, arg := range args {
			num, err := strconv.Atoi(arg)
			if err != nil {
				logger.DebugLog(true, "❌ Invalid syntax: %s", arg)
				return
			}

			// sort
			var match *shared.TorrentEntry
			for _, t := range torrentList {
				if t.DownloadKey == num && (quality == "" || t.Quality == quality) {
					if match == nil || t.Seeders > match.Seeders {
						tmp := t
						match = &tmp
					}
				}
			}

			// no match for download-key
			if match == nil {
				logger.DebugLog(true, "⚠️  No torrent found for key %d and quality '%s'", num, quality)
				continue
			}

			// maybe rewrite this part
			if forceKey != "" {
				if len(args) > 1 {
					logger.DebugLog(true, "❌ --forcekey may only be used with a single DownloadKey")
				}
				match.ChapterRange = forceKey
			}

			dKey := ui.StyleFactory(fmt.Sprintf("%4d", match.DownloadKey), ui.Style.Pink)
			title := ui.StyleFactory(match.TorrentName, ui.Style.LBlue)

			logger.DebugLog(true, "🔍 Matched DownloadKey %s → %s (%s) [%s]", dKey, title, match.Quality, match.ChapterRange)
			logger.DebugLog(true, "🎬 Starting download: %s (%s)\n", match.TorrentName, match.Quality)
			matches = append(matches, *match)
		}

		if len(matches) == 0 {
			fmt.Println("⚠️  No downloads to process.")
			os.Exit(0)
		}

		// outsourced to monitoring function
		torrent.HandleDownloadSession(matches, cfg.TargetDir)

	},
}

func init() {
	downloadCmd.Flags().StringVarP(&quality, "quality", "Q", "", "Only download with specific quality (e.g. 1080p)")
	downloadCmd.Flags().StringVar(&forceKey, "forcekey", "", "Override chapter range (only for single downloadKey)")

	rootCmd.AddCommand(downloadCmd)
}
