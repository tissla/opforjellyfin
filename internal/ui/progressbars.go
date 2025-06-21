// ui/progressbars.go
package ui

import (
	"fmt"
	"opforjellyfin/internal/shared"
	"os"
	"os/signal"
	"time"

	"github.com/charmbracelet/x/ansi"
	"github.com/vbauerster/mpb/v8"
	"github.com/vbauerster/mpb/v8/decor"
)

func FollowProgress() {
	// wait for active.json to fill up
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if len(shared.GetActiveDownloads()) > 0 {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}

	downloads := shared.GetActiveDownloads()
	if len(downloads) == 0 {
		fmt.Println("📭 No active downloads.")
		return
	}

	bars := make(map[int]*mpb.Bar)
	p := mpb.New(mpb.WithWidth(40))

	for _, td := range downloads {
		bar := p.New(
			td.TotalSize,
			mpb.BarStyle().Lbound("[").Filler("▓").Tip("█").Padding("░").Rbound("]"),
			mpb.PrependDecorators(
				decor.Name(ansi.Truncate(td.Title, 20, "..")),
			),
			mpb.AppendDecorators(
				decor.OnComplete(
					decor.Percentage(decor.WCSyncSpace),
					"✔️ Done",
				),
			),
		)
		bars[td.TorrentID] = bar
	}

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)

	done := make(chan struct{})
	go func() {
		p.Wait()
		close(done)
	}()

	for {
		select {
		case <-ticker.C:
			downloads = shared.GetActiveDownloads()

			for _, td := range downloads {
				bar, ok := bars[td.TorrentID]
				if !ok {
					continue
				}
				if td.Done {
					bar.SetTotal(td.TotalSize, true)
					bar.SetCurrent(td.TotalSize)
				} else {
					bar.SetTotal(td.TotalSize, false)
					bar.SetCurrent(td.Progress)
				}
			}

		case <-signalChan:
			fmt.Println("\n🛑 Cancelled by user.")
			return

		case <-done:
			return
		}
	}
}
