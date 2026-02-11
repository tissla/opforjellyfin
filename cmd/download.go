package cmd

import (
	"fmt"
	"os"

	"opforjellyfin/internal/logger"
	"opforjellyfin/internal/scraper"
	"opforjellyfin/internal/shared"
	"opforjellyfin/internal/torrent"
	"opforjellyfin/internal/ui"

	"github.com/spf13/cobra"
)

var (
	forceKey string
)

var downloadCmd = &cobra.Command{
	Use:   "download <downloadKey> [downloadKey...]",
	Short: "Download one or more One Pace torrents",
	Run: func(cmd *cobra.Command, args []string) {

		// add spinner
		spinner := ui.NewSpinner("🗃️ Preparing download.. ", ui.Animations["MetaFetcher"])

		if len(args) < 1 {
			logger.Log(true, "⚠️ You must specify atleast one download-key")
			return
		}

		cfg := shared.LoadConfig()
		if cfg.TargetDir == "" {
			logger.Log(true, "⚠️ No target directory set. Use 'setDir <path>' first.")
			return
		}

		if cfg.Source.BaseURL == "" {
			logger.Log(true, "No valid scraper configuration found. Please run 'sync'")
		}

		torrentList, err := scraper.FetchTorrents(cfg)
		if err != nil {
			logger.Log(true, "❌ Error scraping torrents. Site inaccessible? %v", err)
			return
		}

		// stop spinner
		spinner.Stop()

		// Parse arguments with range support (e.g., "10 11-15 20")
		downloadKeys, err := shared.ParseIntListWithRanges(args)
		if err != nil {
			logger.Log(true, "❌ Invalid syntax: %v", args)
			return
		}

		var matches []shared.TorrentEntry
		for _, num := range downloadKeys {

			// sort
			var match *shared.TorrentEntry
			for _, t := range torrentList {
				if t.DownloadKey == num {
					if match == nil || t.Seeders > match.Seeders {
						tmp := t
						match = &tmp
					}
				}
			}

			// no match for download-key
			if match == nil {
				logger.Log(true, "⚠️  No torrent found for key %d", num)
				continue
			}

			// maybe rewrite this part
			if forceKey != "" {
				if len(args) > 1 {
					logger.Log(true, "❌ --forcekey may only be used with a single DownloadKey")
				}
				match.ChapterRange = forceKey
			}

			dKey := ui.StyleFactory(fmt.Sprintf("%4d", match.DownloadKey), ui.Style.Pink)
			title := ui.StyleFactory(match.TorrentName, ui.Style.LBlue)

			logger.Log(true, "🔍 Matched DownloadKey %s → %s (%s) [%s]", dKey, title, match.Quality, match.ChapterRange)
			logger.Log(true, "🎬 Starting download: %s (%s)\n", match.TorrentName, match.Quality)
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
	downloadCmd.Flags().StringVar(&forceKey, "forcekey", "", "Override chapter range (only for single downloadKey)")

	rootCmd.AddCommand(downloadCmd)
}
