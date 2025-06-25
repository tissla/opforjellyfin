package ui

import (
	"fmt"
	"opforjellyfin/internal/logger"
	"opforjellyfin/internal/shared"
	"os"
	"os/signal"
	"strings"
	"time"
)

func FollowProgress() {

	first := true
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	// ctrl+C cancels bars, wont cancel downloads tho
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)

	for {
		select {
		case <-ticker.C:

			downloads := shared.GetActiveDownloads()
			num := len(downloads)

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
				logger.DebugLog(false, "ALLDONE! RENDERING ALL BARS")
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

func renderSingleBar(title, msg string, progress, total int64, titlewidth, barwidth int) string {
	if total == 0 {
		return fmt.Sprintf("%s [%s] %s", AnsiPadRight(title, titlewidth), strings.Repeat("░", barwidth), AnsiPadLeft("0%", 4))
	}
	percent := float64(progress) / float64(total)
	filled := int(percent * float64(barwidth))
	bar := strings.Repeat("█", filled) + strings.Repeat("░", barwidth-filled)

	maxW := GetTerminalWidth()

	percentStr := fmt.Sprintf("%3.0f%%", percent*100)
	if percent >= 1.0 {
		percentStr = " ✅ "
	}

	outMsg := msg
	if msg != "" {
		outMsg = "| " + msg
	}

	output := fmt.Sprintf("%s [%s] %s %s", AnsiPadRight(title, titlewidth), bar, AnsiPadLeft(percentStr, 4), outMsg)
	return AnsiPadRight(output, maxW)
}

// renderall
func renderAllBars(downloads []*shared.TorrentDownload) {
	allbars := ""
	for _, td := range downloads {

		bar := renderSingleBar(td.Title, td.PlacementProgress, td.Progress, td.TotalSize, 15, 40)
		allbars = allbars + bar + "\n"
	}
	PrintMultiline(allbars)
}
