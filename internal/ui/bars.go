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
	fmt.Print("\033[?25l")

	previous, latest := "", ""
	rows, slots, previousWidth := 0, 0, 0
	var messages strings.Builder
	draw := func(finishing bool) {
		width := max(0, GetTerminalWidth()-1)
		downloads := shared.GetActiveDownloads()
		slots = max(slots, len(downloads))
		if pending := logger.TakeUserOutput(); pending != "" {
			messages.WriteString(pending)
			for _, line := range strings.Split(pending, "\n") {
				if strings.TrimSpace(line) != "" {
					latest = singleLine(line)
				}
			}
		}

		frame := renderAllBars(downloads, width)
		frame += strings.Repeat("\r\033[K\n", slots-len(downloads))
		if placed, total := shared.GetDownloadProgress(); total > 0 {
			frame += fmt.Sprintf("\r%s\033[K\n", renderTotalBar(placed, total, width))
		}
		status := "Status: Downloading"
		if finishing {
			status = "Status: Finished"
		}
		if latest != "" {
			status = "Latest: " + latest
		}
		frame += fmt.Sprintf("\r%s\033[K\n", ansi.Truncate(status, width, "…"))
		if frame == previous {
			return
		}

		// A narrower terminal can reflow the previous frame onto more rows.
		if previousWidth != width && previous != "" {
			rows = strings.Count(ansi.Hardwrap(previous, width+1, true), "\n")
		}
		var output strings.Builder
		if rows > 0 {
			fmt.Fprintf(&output, "\033[%dA", rows)
		}
		output.WriteString("\r")
		output.WriteString(frame)
		output.WriteString("\r\033[J")
		fmt.Print(output.String())
		previous, rows, previousWidth = frame, strings.Count(frame, "\n"), width
	}

	for {
		select {
		case <-ticker.C:
			draw(false)
		case <-doneChan:
			draw(true)
			fmt.Print("\033[?25h", messages.String())
			logger.BufferUserOutput(false)
			doneChan <- struct{}{}
			return
		}
	}
}

func renderTotalBar(placed, total, width int) string {
	titleWidth, barWidth, detailWidth := progressColumns(width)
	title := fmt.Sprintf("Total: %d/%d episodes", placed, total)
	percent := 0.0
	if total > 0 {
		percent = min(1.0, max(0.0, float64(placed)/float64(total)))
	}
	line := fmt.Sprintf("%s %s %4.0f%%", AnsiPadRight(title, titleWidth, "…"), progressBar(int64(placed), int64(total), barWidth, Style.Green), percent*100)
	if detailWidth > 0 {
		line += " " + AnsiPadRight(fmt.Sprintf("%d remaining", max(0, total-placed)), detailWidth, "…")
	}
	return ansi.Truncate(line, width, "")
}

func renderDownload(td *shared.TorrentDownload, width int) string {
	percent := 0.0
	if td.TotalSize > 0 {
		percent = min(1.0, max(0.0, float64(td.Progress)/float64(td.TotalSize)))
	}
	detail := td.PlacementProgress
	if detail == "" {
		switch {
		case td.TotalSize == 0:
			detail = "Preparing"
		case !td.Done && td.Progress >= td.TotalSize:
			detail = "Verifying"
		default:
			detail = downloadSize(td.Progress) + " / " + downloadSize(td.TotalSize)
		}
	}
	titleWidth, barWidth, detailWidth := progressColumns(width)
	line := fmt.Sprintf("%s %s %4.0f%%", AnsiPadRight(singleLine(td.Title), titleWidth, "…"), progressBar(td.Progress, td.TotalSize, barWidth, Style.LBlue), percent*100)
	if detailWidth > 0 {
		line += " " + AnsiPadRight(singleLine(detail), detailWidth, "…")
	}
	return ansi.Truncate(line, width, "")
}

func progressColumns(width int) (titleWidth, barWidth, detailWidth int) {
	barWidth = min(20, max(4, width/5))
	if width >= 60 {
		detailWidth = min(24, max(20, width/4))
	}
	titleWidth = max(0, width-barWidth-7)
	if detailWidth > 0 {
		titleWidth = max(0, titleWidth-detailWidth-1)
	}
	return
}

func progressBar(progress, total int64, width int, style ColorStyle) string {
	filled := 0
	if total > 0 {
		filled = int(float64(width) * min(1.0, max(0.0, float64(progress)/float64(total))))
	}
	return StyleFactory(strings.Repeat("━", filled)+strings.Repeat("·", width-filled), style)
}

func downloadSize(bytes int64) string {
	if bytes >= 1<<30 {
		return fmt.Sprintf("%.1f GiB", float64(bytes)/(1<<30))
	}
	return fmt.Sprintf("%.0f MiB", float64(max(0, bytes))/(1<<20))
}

func singleLine(text string) string {
	return strings.Join(strings.Fields(text), " ")
}

func renderAllBars(downloads []*shared.TorrentDownload, width int) string {
	var frame strings.Builder
	for _, td := range downloads {
		fmt.Fprintf(&frame, "\r%s\033[K\n", renderDownload(td, width))
	}
	return frame.String()
}
