// cmd/list.go
package cmd

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"opforjellyfin/internal/flags"
	"opforjellyfin/internal/logger"
	"opforjellyfin/internal/scraper"
	"opforjellyfin/internal/shared"
	"opforjellyfin/internal/ui"

	"github.com/spf13/cobra"
)

var (
	rangeFilter          string
	titleFilter          string
	minimumQualityFilter = flags.StringChoice([]string{"480p", "720p", "1080p"})
	qualityFilter        = flags.StringChoice([]string{"480p", "720p", "1080p"})

	onlySpecials bool
	verboseList  bool

	alternate bool
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all available One Pace seasons and specials",
	Run: func(cmd *cobra.Command, args []string) {
		spinner := ui.NewMultirowSpinner(ui.Animations["Searcher"], 4)

		cfg := shared.LoadConfig()

		if cfg.Source.BaseURL == "" {
			spinner.Stop()
			logger.Log(true, "⚠️ No valid scraper configuration found. Please run 'sync' or 'setDir'")
			return
		}

		allTorrents, err := scraper.FetchTorrents(cfg)
		if err != nil {
			spinner.Stop()
			logger.Log(true, "❌ Error scraping torrents. Site inaccessible? %v", err)
			return
		}

		// Apply filters after keys are assigned
		var filtered []shared.TorrentEntry
		for _, t := range allTorrents {
			if applyFilters(t) {
				filtered = append(filtered, t)
			}
		}

		// Sort by DownloadKey, then seeders descending
		sort.SliceStable(filtered, func(i, j int) bool {
			if filtered[i].DownloadKey == filtered[j].DownloadKey {
				return filtered[i].Seeders > filtered[j].Seeders
			}
			return filtered[i].DownloadKey < filtered[j].DownloadKey
		})

		spinner.Stop()

		fmt.Println("📚 Filtered Download List:\n")
		for _, t := range filtered {
			if verboseList {
				renderVerboseRow(t)
			} else {
				renderRow(t)
			}
		}
	},
}

func applyFilters(t shared.TorrentEntry) bool {
	// --specials only
	if onlySpecials && !t.IsSpecial {
		return false
	}
	// range filter
	if rangeFilter != "" {

		// defaults
		cMin, cMax, min, max := 0, 0, 0, 0

		// chapterRange of torrent
		cMin, cMax = shared.ParseRange(t.ChapterRange)

		// chosen chapterRange
		min, max = shared.ParseRange(rangeFilter)

		if cMin < 0 || cMax < 0 || !shared.RangesOverlap(cMin, cMax, min, max) {
			return false
		}
	}
	// title filter
	if titleFilter != "" && !strings.Contains(strings.ToLower(t.TorrentName), strings.ToLower(titleFilter)) {
		return false
	}

	if qualityFilter.String() != "" && t.Quality != qualityFilter.Value {
		return false
	}

	if minimumQualityFilter.String() != "" {
		// Not sure about error checking here, both values should be guaranteed parsable if everything else works as intended.
		quality, _ := strconv.Atoi(strings.TrimSuffix(t.Quality, "p"))
		minimumQuality, _ := strconv.Atoi(strings.TrimSuffix(minimumQualityFilter.Value, "p"))
		if quality < minimumQuality {
			return false
		}
	}

	return true
}

// rowrender
func renderRow(t shared.TorrentEntry) {
	// bools
	metaMark := "❌"

	haveMark := map[int]string{
		0: "❌",
		1: "🟠",
		2: "✅",
	}[t.HaveIt]

	if t.MetaDataAvail {
		metaMark = "✅"
	}

	// styling and render
	truncatedTitle := ui.AnsiPadRight(t.TorrentName, 30)

	row := ui.RenderRow(
		"%s - %s: %s Have? %s | Meta: %s | %-9s | %s | %s seeders",
		alternate,
		ui.StyleFactory("DKEY", ui.Style.LBlue),
		ui.StyleFactory(fmt.Sprintf("%4d", t.DownloadKey), ui.Style.Pink),
		ui.StyleFactory(truncatedTitle, ui.Style.LBlue),
		haveMark,
		metaMark,
		t.ChapterRange,
		ui.AnsiPadLeft(ui.StyleByRange(t.Quality, 400, 1000), 5),
		ui.AnsiPadLeft(ui.StyleByRange(t.Seeders, 0, 10), 3),
	)

	// set flag
	alternate = !alternate

	fmt.Println(row)
}

// render verbose row
func renderVerboseRow(t shared.TorrentEntry) {
	metaMark := "❌"
	if t.MetaDataAvail {
		metaMark = "✅"
	}

	haveMark := map[int]string{
		0: "❌",
		1: "🟠",
		2: "✅",
	}[t.HaveIt]

	fullTitle := ui.AnsiPadRight(t.Title, 60)

	row := ui.RenderRow(
		"%s - %s: %s H:%s M:%s | %s seeders",
		alternate,
		ui.StyleFactory("DKEY", ui.Style.LBlue),
		ui.StyleFactory(fmt.Sprintf("%4d", t.DownloadKey), ui.Style.Pink),
		ui.StyleFactory(fullTitle, ui.Style.LBlue),
		haveMark,
		metaMark,
		ui.AnsiPadLeft(ui.StyleByRange(t.Seeders, 0, 10), 3),
	)

	alternate = !alternate
	fmt.Println(row)
}

// init
func init() {
	listCmd.Flags().StringVarP(&rangeFilter, "range", "r", "", "Show seasons in range, e.g. 10-20")
	listCmd.Flags().StringVarP(&titleFilter, "title", "t", "", "Filter by title keyword")
	listCmd.Flags().VarP(qualityFilter, "quality", "q", "Filter by quality, e.g. 1080p")
	listCmd.Flags().Var(minimumQualityFilter, "minquality", "Filter by minimum quality, e.g. 720p will only list 720p and 1080p.")

	listCmd.Flags().BoolVarP(&onlySpecials, "specials", "s", false, "Show only specials")
	listCmd.Flags().BoolVarP(&verboseList, "verbose", "v", false, "Show full titles")
	rootCmd.AddCommand(listCmd)
}
