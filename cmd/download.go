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

		cfg, _ := shared.LoadConfig()
		if cfg.TargetDir == "" {
			logger.Log(true, "⚠️ No target directory set. Use 'setDir <path>' first.")
			return
		}

		if cfg.Source.BaseURL == "" {
			logger.Log(true, "No valid scraper configuration found. Please run 'sync'")
		}

		searchCache, err := scraper.LoadSearchCache()
		if err != nil {
			logger.Log(true, "❌ Error loading search cache. Did you run 'list'? - %v", err)
			return
		}

		// stop spinner
		spinner.Stop()
		var matches []shared.TorrentEntry
		for _, arg := range args {
			num, err := strconv.Atoi(arg)
			if err != nil {
				logger.Log(true, "❌ Invalid syntax: %s", arg)
				return
			}

			// sort
			var match *shared.TorrentEntry
			for _, t := range searchCache.Results {
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
