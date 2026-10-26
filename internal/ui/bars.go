package ui

import (
	"fmt"
	"opforjellyfin/internal/logger"
	"opforjellyfin/internal/shared"
	"strings"
	"time"

	"github.com/charmbracelet/x/ansi"
)

func FollowProgress(doneChan chan struct{}) {
	ticker := time.NewTicker(300 * time.Millisecond)
	defer ticker.Stop()
	logger.BufferUserOutput(true)

	previous, rows := "", 0
	draw := func() {
		downloads := shared.GetActiveDownloads()
		frame := renderAllBars(downloads)
		if placed, total := shared.GetDownloadProgress(); total > 0 {
			frame += fmt.Sprintf("\r%s\033[K\n", renderTotalBar(placed, total))
		}
		messages := logger.TakeUserOutput()
		if frame == previous && messages == "" {
			return
		}

		var output strings.Builder
		if rows > 0 {
			fmt.Fprintf(&output, "\033[%dA", rows)
		}
		output.WriteString("\r")
		output.WriteString(strings.ReplaceAll(messages, "\n", "\033[K\r\n"))
		output.WriteString(frame)
		output.WriteString("\r\033[J")
		fmt.Print(output.String())
		previous, rows = frame, strings.Count(frame, "\n")
	}

	for {
		select {
		case <-ticker.C:
			draw()
		case <-doneChan:
			draw()
			logger.BufferUserOutput(false)
			doneChan <- struct{}{}
			return
		}
	}
}

func renderTotalBar(placed, total int) string {
	counts := fmt.Sprintf("%d/%d episodes | %d remaining", placed, total, max(0, total-placed))
	width := max(0, GetTerminalWidth()-1)
	barWidth := min(24, max(0, width-ansi.StringWidth(counts)-len("Total [] ")))
	filled := 0
	if total > 0 {
		filled = barWidth * min(total, max(0, placed)) / total
	}
	bar := strings.Repeat("█", filled) + strings.Repeat("░", barWidth-filled)
	return ansi.Truncate(fmt.Sprintf("Total [%s] %s", bar, counts), width, "")
}

func renderSingleBar(title, msg string, progress, total int64, titlewidth, barwidth int) string {
	percent := 0.0
	if total > 0 {
		percent = min(1.0, max(0.0, float64(progress)/float64(total)))
	}
	filled := int(percent * float64(barwidth))
	bar := strings.Repeat("█", filled) + strings.Repeat("░", barwidth-filled)
	percentStr := fmt.Sprintf("%3.0f%%", percent*100)
	if percent >= 1.0 {
		percentStr = " ✅ "
	}
	if msg != "" {
		msg = "| " + msg
	}
	output := fmt.Sprintf("%s [%s] %s %s", AnsiPadRight(title, titlewidth), bar, AnsiPadLeft(percentStr, 4), msg)
	// Leave the last column free to avoid terminal line wrapping.
	return ansi.Truncate(output, max(0, GetTerminalWidth()-1), "")
}

func renderAllBars(downloads []*shared.TorrentDownload) string {
	var frame strings.Builder
	for _, td := range downloads {
		bar := renderSingleBar(td.Title, td.PlacementProgress, td.Progress, td.TotalSize, 15, 40)
		fmt.Fprintf(&frame, "\r%s\033[K\n", bar)
	}
	return frame.String()
}
