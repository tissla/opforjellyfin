package ui

import (
	"fmt"
	"opforjellyfin/internal/shared"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/charmbracelet/x/ansi"
)

func FollowProgress() {

	first := true
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	// ctrl+C cancel, wont cancel downloads tho
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)

	for {
		select {
		case <-ticker.C:

			downloads := shared.GetActiveDownloads()
			num := len(downloads)

			if len(downloads) == 0 {
				fmt.Println("📭 No active downloads.")
				return
			}

			// print initial lines
			if first {
				for i := 0; i < num+1; i++ {
					fmt.Print("\n")
				}
				first = false
			}

			// +1 for the marker, +1 for the trailing \n in renderAllBars
			ClearLines(num + 2)

			allDone := true
			for _, td := range downloads {
				if !td.Placed {
					allDone = false
					break
				}
			}
			if allDone {
				renderAllBars(downloads)
				return
			}

			// render current status
			renderAllBars(downloads)

		case <-signalChan:
			fmt.Println("\n🛑 Cancelled by user.")
			return
		}
	}
}

// render bars

func renderSingleBar(title, msg string, progress, total int64, width int) string {
	if total == 0 {
		return fmt.Sprintf("%s [%s] 0%%", ansi.Truncate(title, 20, ""), strings.Repeat("░", width))
	}
	percent := float64(progress) / float64(total)
	filled := int(percent * float64(width))
	bar := strings.Repeat("█", filled) + strings.Repeat("░", width-filled)

	maxW := GetTerminalWidth()

	output := fmt.Sprintf("%-20s [%s] %3.0f%% %s", AnsiPadRight(title, 20), bar, percent*100, msg)
	return AnsiPadRight(output, maxW)
}

// renderall
func renderAllBars(downloads []*shared.TorrentDownload) {
	allbars := ""
	for _, td := range downloads {

		bar := renderSingleBar(td.Title, td.ProgressMessage, td.Progress, td.TotalSize, 40)
		allbars = allbars + bar + "\n"
	}
	PrintMultiline(allbars)
}
