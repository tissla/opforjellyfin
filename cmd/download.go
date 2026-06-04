package cmd

import (
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"

	"opforjellyfin/internal/logger"
	"opforjellyfin/internal/metadata"
	"opforjellyfin/internal/scraper"
	"opforjellyfin/internal/shared"
	"opforjellyfin/internal/torrent"
	"opforjellyfin/internal/ui"

	"github.com/spf13/cobra"
)

var (
	forceKey string
	arcName  string
)

var downloadCmd = &cobra.Command{
	Use:   "download <downloadKey>|<arcName> [downloadKey|arcName...]",
	Short: "Download one or more One Pace torrents",
	Args: func(cmd *cobra.Command, args []string) error {
		if arcName == "" && len(args) == 0 {
			return fmt.Errorf("requires at least one download-key or --arc")
		}
		if arcName != "" && forceKey != "" {
			return fmt.Errorf("--forcekey cannot be used with --arc")
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {

		// add spinner
		spinner := ui.NewSpinner("🗃️ Preparing download.. ", ui.Animations["MetaFetcher"])

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

		effectiveArc := arcName
		if effectiveArc == "" {
			if positionalName, ok := positionalArcName(args); ok {
				effectiveArc = positionalName
				args = nil
			}
		}

		if effectiveArc != "" {
			index := metadata.LoadMetadataCache()
			if index == nil || len(index.Seasons) == 0 {
				logger.Log(true, "⚠️ No metadata index found. Please run 'sync'")
				return
			}

			arc, err := metadata.FindArc(index, effectiveArc)
			if err != nil {
				logger.Log(true, "❌ %v", err)
				return
			}

			matches := selectArcDownloads(torrentList, arc)
			if len(matches) == 0 {
				logger.Log(true, "⚠️ No torrents found for arc %s (%s)", arc.Name, arc.Range)
				return
			}

			logger.Log(true, "🎬 Matched arc %s (%s), chapters %s", arc.Name, arc.Season, arc.Range)
			for _, match := range matches {
				dKey := ui.StyleFactory(fmt.Sprintf("%4d", match.DownloadKey), ui.Style.Pink)
				title := ui.StyleFactory(match.TorrentName, ui.Style.LBlue)
				logger.Log(true, "🔍 Matched DownloadKey %s → %s (%s) [%s]", dKey, title, match.Quality, match.ChapterRange)
			}

			torrent.HandleDownloadSession(matches, cfg.TargetDir)
			return
		}

		expandedArgs, err := expandEpisodeArgs(args)
		if err != nil {
			logger.Log(true, "❌ %v", err)
			return
		}

		var matches []shared.TorrentEntry
		for _, arg := range expandedArgs {
			num, err := strconv.Atoi(arg)
			if err != nil {
				logger.Log(true, "❌ Invalid syntax: %s", arg)
				return
			}

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
				if len(expandedArgs) > 1 {
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
	downloadCmd.Flags().StringVarP(&arcName, "arc", "a", "", "Download all torrents for an arc name")

	rootCmd.AddCommand(downloadCmd)
}

type rangedTorrentEntry struct {
	entry shared.TorrentEntry
	start int
	end   int
}

func selectArcDownloads(torrentList []shared.TorrentEntry, arc *metadata.ArcMatch) []shared.TorrentEntry {
	arcStart, arcEnd := shared.ParseRange(arc.Range)
	if arcStart < 0 || arcEnd < 0 {
		return nil
	}

	bestByRange := make(map[string]rangedTorrentEntry)
	for _, entry := range torrentList {
		if entry.IsSpecial || entry.ChapterRange == "" {
			continue
		}

		start, end := shared.ParseRange(entry.ChapterRange)
		if start < arcStart || end > arcEnd || start < 0 || end < 0 {
			continue
		}

		rangeKey := shared.NormalizeDash(entry.ChapterRange)
		current, exists := bestByRange[rangeKey]
		if !exists || entry.Seeders > current.entry.Seeders {
			bestByRange[rangeKey] = rangedTorrentEntry{
				entry: entry,
				start: start,
				end:   end,
			}
		}
	}

	var candidates []rangedTorrentEntry
	for _, candidate := range bestByRange {
		candidates = append(candidates, candidate)
	}

	sort.Slice(candidates, func(i, j int) bool {
		if candidates[i].start == candidates[j].start {
			iLen := candidates[i].end - candidates[i].start
			jLen := candidates[j].end - candidates[j].start
			if iLen == jLen {
				return candidates[i].entry.Seeders > candidates[j].entry.Seeders
			}
			return iLen > jLen
		}
		return candidates[i].start < candidates[j].start
	})

	covered := make(map[int]bool)
	var selected []shared.TorrentEntry
	for _, candidate := range candidates {
		overlapsSelected := false
		for chapter := candidate.start; chapter <= candidate.end; chapter++ {
			if covered[chapter] {
				overlapsSelected = true
				break
			}
		}
		if overlapsSelected {
			continue
		}

		selected = append(selected, candidate.entry)
		for chapter := candidate.start; chapter <= candidate.end; chapter++ {
			covered[chapter] = true
		}
	}

	return selected
}

// expandEpisodeArgs takes raw CLI arguments and expands range formats (like "15-17").
func expandEpisodeArgs(args []string) ([]string, error) {
	var expanded []string

	for _, arg := range args {
		// If the argument contains a hyphen, treat it as a range
		if strings.Contains(arg, "-") {
			parts := strings.Split(arg, "-")
			if len(parts) != 2 {
				return nil, fmt.Errorf("invalid range format: %s", arg)
			}

			// Parse the start and end of the range
			start, err1 := strconv.Atoi(parts[0])
			end, err2 := strconv.Atoi(parts[1])

			if err1 != nil || err2 != nil {
				return nil, fmt.Errorf("range must contain valid numbers: %s", arg)
			}
			if start > end {
				return nil, fmt.Errorf("range start cannot be greater than range end: %s", arg)
			}

			// Expand the range and append each episode to our result list
			for i := start; i <= end; i++ {
				expanded = append(expanded, strconv.Itoa(i))
			}
		} else {
			// If it's a normal single number, just append it
			expanded = append(expanded, arg)
		}
	}

	return expanded, nil
}
