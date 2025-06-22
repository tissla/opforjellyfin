package cmd

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"opforjellyfin/internal/logger"
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
		if len(args) < 1 {
			log.Fatalf("You must specify at least one downloadKey")
		}

		cfg := shared.LoadConfig()
		if cfg.TargetDir == "" {
			log.Fatalf("⚠️  No target directory set. Use 'setDir <path>' first.")
		}

		torrentList, err := torrent.FetchOnePaceTorrents()
		if err != nil {
			logger.DebugLog(false, "❌ Failed to fetch torrents: %v", err)
			log.Fatalf("❌ Failed to fetch torrents: %v", err)
		}

		var matches []shared.TorrentEntry
		for _, arg := range args {
			num, err := strconv.Atoi(arg)
			if err != nil {
				log.Fatalf("Invalid downloadKey: %s", arg)
			}

			var match *shared.TorrentEntry
			for _, t := range torrentList {
				if t.DownloadKey == num && (quality == "" || t.Quality == quality) {
					if match == nil || t.Seeders > match.Seeders {
						tmp := t
						match = &tmp
					}
				}
			}

			if match == nil {
				logger.DebugLog(true, "⚠️  No torrent found for key %d and quality '%s'", num, quality)
				continue
			}

			if forceKey != "" {
				if len(args) > 1 {
					log.Fatalf("❌ --forcekey may only be used with a single DownloadKey")
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
	//downloadCmd.Flags().BoolVarP(&follow, "follow", "f", false, "Follow download progress live")
	downloadCmd.Flags().StringVarP(&quality, "quality", "Q", "", "Only download with specific quality (e.g. 1080p)")
	downloadCmd.Flags().StringVar(&forceKey, "forcekey", "", "Override chapter range (only for single downloadKey)")

	rootCmd.AddCommand(downloadCmd)
}
