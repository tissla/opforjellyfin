// cmd/list.go
package cmd

import (
	"fmt"
	"opforjellyfin/internal/shared"
	"opforjellyfin/internal/torrent"
	"opforjellyfin/internal/ui"
	"sort"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
)

var (
	rangeFilter  string
	titleFilter  string
	onlySpecials bool
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all available One Pace seasons and specials",
	Run: func(cmd *cobra.Command, args []string) {

		spinner := ui.NewMultirowSpinner(ui.Animations["Searcher"], 4)

		allTorrents, err := torrent.FetchOnePaceTorrents()
		if err != nil {
			fmt.Printf("❌ Error: %v\n", err)
			return
		}

		// Apply filters after keys are assigned
		var filtered []shared.TorrentEntry
		for _, t := range allTorrents {
			if applyFilters(t) {
				filtered = append(filtered, t)
			}
		}

		// Sort by SeasonKey, then seeders descending
		sort.SliceStable(filtered, func(i, j int) bool {
			if filtered[i].DownloadKey == filtered[j].DownloadKey {
				return filtered[i].Seeders > filtered[j].Seeders
			}
			return filtered[i].DownloadKey < filtered[j].DownloadKey
		})

		alternate := false

		spinner.Stop()

		fmt.Println("📚 Filtered Download List:\n")
		for _, t := range filtered {

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

			truncatedTitle := shared.Truncate(t.TorrentName, 20)

			row := ui.RenderRow(
				"%s - %s: %-30s Have? %s | Meta: %s | %-9s | %s | %s seeders",
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

			fmt.Println(row)

			alternate = !alternate
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
		parts := strings.Split(rangeFilter, "-")
		if len(parts) == 2 {
			min, _ := strconv.Atoi(parts[0])
			max, _ := strconv.Atoi(parts[1])
			if t.DownloadKey < min || t.DownloadKey > max {
				return false
			}
		}
	}
	// title filter
	if titleFilter != "" && !strings.Contains(strings.ToLower(t.TorrentName), strings.ToLower(titleFilter)) {
		return false
	}
	return true
}

func init() {
	listCmd.Flags().StringVarP(&rangeFilter, "range", "r", "", "Show seasons in range, e.g. 10-20")
	listCmd.Flags().StringVarP(&titleFilter, "title", "t", "", "Filter by title keyword")
	listCmd.Flags().BoolVarP(&onlySpecials, "specials", "s", false, "Show only specials")
	rootCmd.AddCommand(listCmd)
}
