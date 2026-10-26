package cmd

import (
	"fmt"
	"strconv"

	"opforjellyfin/internal/logger"
	"opforjellyfin/internal/metadata"
	"opforjellyfin/internal/scraper"
	"opforjellyfin/internal/shared"
	"opforjellyfin/internal/torrent"
	"opforjellyfin/internal/ui"

	"github.com/spf13/cobra"
)

var seed, includeExtended, includeSpecials bool

var downloadCmd = &cobra.Command{
	Use:   "download <all | downloadKey [downloadKey...]>",
	Short: "Download One Pace torrents, or all missing episodes",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return fmt.Errorf("specify 'all' or at least one download key")
		}
		if args[0] == "all" {
			if len(args) != 1 {
				return fmt.Errorf("'all' cannot be combined with download keys")
			}
			if seed {
				return fmt.Errorf("--seed cannot be used with 'all': seeding would prevent the queue from advancing")
			}
			return nil
		}
		if cmd.Flags().Changed("include-extended") || cmd.Flags().Changed("include-specials") {
			return fmt.Errorf("--include-extended and --include-specials require 'download all'")
		}
		for _, arg := range args {
			if _, err := strconv.Atoi(arg); err != nil {
				return fmt.Errorf("invalid download key: %s", arg)
			}
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := shared.LoadConfig()
		if err != nil {
			return err
		}
		if cfg.TargetDir == "" {
			return fmt.Errorf("no target directory set; use 'setDir <path>' first")
		}
		if cfg.Source.BaseURL == "" {
			return fmt.Errorf("no valid scraper configuration found; run 'sync'")
		}

		spinner := ui.NewSpinner("🗃️ Preparing downloads.. ", ui.Animations["MetaFetcher"])
		defer spinner.Stop()
		var matches []shared.TorrentEntry
		if args[0] == "all" {
			index, err := metadata.LoadMetadataCache()
			if err != nil {
				return err
			}
			entries, err := scraper.FetchTorrents(cfg)
			if err != nil {
				return err
			}
			matches, err = torrent.SelectAllDownloads(cmd.Context(), entries, cfg, index, includeExtended, includeSpecials)
			if err != nil {
				return err
			}
			spinner.Stop()
			files := 0
			for _, match := range matches {
				files += len(match.SelectedFiles)
			}
			logger.Log(true, "Selected %d missing files at 1080p or higher from %d torrents (up to %d downloads at once).", files, len(matches), torrent.MaxConcurrent)
		} else {
			searchCache, err := scraper.LoadSearchCache()
			if err != nil {
				return fmt.Errorf("load search cache; run 'list' first: %w", err)
			}
			spinner.Stop()
			for _, arg := range args {
				num, _ := strconv.Atoi(arg)
				var match *shared.TorrentEntry
				for _, t := range searchCache.Results {
					if t.DownloadKey == num && (match == nil || t.Seeders > match.Seeders) {
						tmp := t
						match = &tmp
					}
				}
				if match == nil {
					logger.Log(true, "⚠️ No torrent found for key %d", num)
					continue
				}
				dKey := ui.StyleFactory(fmt.Sprintf("%4d", match.DownloadKey), ui.Style.Pink)
				title := ui.StyleFactory(match.TorrentName, ui.Style.LBlue)
				logger.Log(true, "🔍 Matched DownloadKey %s → %s (%s) [%s]", dKey, title, match.Quality, match.ChapterRange)
				matches = append(matches, *match)
			}
		}

		if len(matches) == 0 {
			logger.Log(true, "No downloads to process.")
			return nil
		}
		if seed && len(matches) > torrent.MaxConcurrent {
			logger.Log(true, "⚠️ --seed with more than %d keys: only the first %d will start this run. Seeding workers stay occupied until Ctrl+C.", torrent.MaxConcurrent, torrent.MaxConcurrent)
		}
		torrent.HandleDownloadSession(matches, cfg.TargetDir, seed)
		return nil
	},
}

func init() {
	downloadCmd.Flags().BoolVar(&seed, "seed", false, "Keep seeding downloaded torrents until you stop the program (Ctrl+C)")
	downloadCmd.Flags().BoolVar(&includeExtended, "include-extended", false, "Also download extended versions with 'download all'")
	downloadCmd.Flags().BoolVar(&includeSpecials, "include-specials", false, "Also download specials with 'download all'")
	rootCmd.AddCommand(downloadCmd)
}
