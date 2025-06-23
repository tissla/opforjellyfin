// ui/progressbars.go
package ui

import (
	"fmt"
	"opforjellyfin/internal/logger"
	"opforjellyfin/internal/shared"
	"os"
	"os/signal"
	"sync"
	"time"

	"github.com/vbauerster/mpb/v8"
	"github.com/vbauerster/mpb/v8/decor"
)

// follow progress of downloads and store in cool progress bars
func FollowProgress() {
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if len(shared.GetActiveDownloads()) > 0 {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}

	if len(shared.GetActiveDownloads()) == 0 {
		logger.DebugLog(true, "📭 No active downloads.")
		return
	}

	var (
		barMu = sync.Mutex{}
		bars  = make(map[int]*mpb.Bar)
		seen  = make(map[int]bool)
	)

	// TODO: make dynamic for terminal window width
	// more styling in general!

	p := mpb.New(mpb.WithWidth(40))
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)

	done := make(chan struct{})

	go func() {
		for {
			time.Sleep(1 * time.Second)
			allDone := true
			for _, td := range shared.GetActiveDownloads() {
				if !td.Done {
					allDone = false
					break
				}
			}
			if allDone {
				p.Wait()
				close(done)
				return
			}
		}
	}()

	for {
		select {
		case <-ticker.C:
			downloads := shared.GetActiveDownloads()

			for _, td := range downloads {
				if _, exists := seen[td.TorrentID]; !exists && td.TotalSize > 0 {
					barMu.Lock()
					bar := p.New(
						td.TotalSize,
						mpb.BarStyle().Lbound("[").Filler("▓").Tip("█").Padding("░").Rbound("]"),
						mpb.PrependDecorators(
							decor.Name(AnsiPadRight(td.Title, 20)),
						),
						mpb.AppendDecorators(
							decor.OnComplete(
								decor.Percentage(decor.WCSyncSpace),
								"✔️ Done",
							),
						),
					)
					bars[td.TorrentID] = bar
					seen[td.TorrentID] = true
					barMu.Unlock()
				}
			}

			for _, td := range downloads {
				barMu.Lock()
				if bar, ok := bars[td.TorrentID]; ok {
					if td.Done {
						bar.SetTotal(td.TotalSize, true)
						bar.SetCurrent(td.TotalSize)
					} else {
						bar.SetTotal(td.TotalSize, false)
						bar.SetCurrent(td.Progress)
					}
				}
				barMu.Unlock()
			}

		case <-signalChan:
			fmt.Println("\n🛑 Cancelled by user.")
			return

		case <-done:
			return
		}
	}
}
