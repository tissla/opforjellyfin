package cmd

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"opforjellyfin/internal"

	"github.com/spf13/cobra"
)

var (
	//follow   bool
	quality  string
	debug    bool
	forceKey string
)

var downloadCmd = &cobra.Command{
	Use:   "download <downloadKey> [downloadKey...]",
	Short: "Download one or more One Pace torrents",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 1 {
			log.Fatalf("You must specify at least one downloadKey")
		}

		cfg := internal.LoadConfig()
		if cfg.TargetDir == "" {
			log.Fatalf("⚠️  No target directory set. Use 'setDir <path>' first.")
		}

		torrentList, err := internal.FetchOnePaceTorrents()
		if err != nil {
			internal.DebugLog(false, "❌ Failed to fetch torrents: %v", err)
			log.Fatalf("❌ Failed to fetch torrents: %v", err)
		}

		var matches []internal.TorrentEntry
		for _, arg := range args {
			num, err := strconv.Atoi(arg)
			if err != nil {
				log.Fatalf("Invalid downloadKey: %s", arg)
			}

			var match *internal.TorrentEntry
			for _, t := range torrentList {
				if t.DownloadKey == num && (quality == "" || t.Quality == quality) {
					tmp := t
					match = &tmp
					break
				}
			}

			if match == nil {
				internal.DebugLog(true, "⚠️  No torrent found for key %d and quality '%s'", num, quality)
				continue
			}

			if forceKey != "" {
				if len(args) > 1 {
					log.Fatalf("❌ --forcekey may only be used with a single DownloadKey")
				}
				match.ChapterRange = forceKey
			}

			dKey := internal.StyleFactory(fmt.Sprintf("%4d", match.DownloadKey), internal.Style.Pink)
			title := internal.StyleFactory(match.SeasonName, internal.Style.LBlue)

			internal.DebugLog(true, "🔍 Matched DownloadKey %s → %s (%s) [%s]", dKey, title, match.Quality, match.ChapterRange)
			internal.DebugLog(true, "🎬 Starting download: %s (%s)\n", match.SeasonName, match.Quality)
			matches = append(matches, *match)
		}

		if len(matches) == 0 {
			fmt.Println("⚠️  No downloads to process.")
			os.Exit(0)
		}

		// clears current downloads from active.json so as not to confuse followprogress.
		internal.ClearActiveDownloads()

		// needs to start as a go-routine, but with context.Background()?
		go internal.StartMultipleDownloads(context.Background(), matches, cfg.TargetDir)

		fmt.Println("🚀 Downloads started!")
		time.Sleep(1 * time.Second)
		internal.FollowProgress()

	},
}

func init() {
	//downloadCmd.Flags().BoolVarP(&follow, "follow", "f", false, "Follow download progress live")
	downloadCmd.Flags().StringVarP(&quality, "quality", "Q", "", "Only download with specific quality (e.g. 1080p)")
	downloadCmd.Flags().BoolVar(&debug, "debug", false, "Tail debug.log live during download (not implemented)")
	downloadCmd.Flags().StringVar(&forceKey, "forcekey", "", "Override chapter range (only for single downloadKey)")

	rootCmd.AddCommand(downloadCmd)
}
