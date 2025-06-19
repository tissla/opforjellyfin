// internal/progressbars.go
package internal

import (
	"fmt"
	"os"
	"os/signal"
	"time"

	"github.com/vbauerster/mpb/v8"
	"github.com/vbauerster/mpb/v8/decor"
)

type DownloadGui struct {
	Bar *mpb.Bar
}

// unused
func CreateProgressBar(p *mpb.Progress, td *TorrentDownload) *mpb.Bar {
	return p.New(
		td.TotalSize,
		mpb.BarStyle().
			Filler("▓").
			Padding("░").
			Rbound("█"),
		mpb.PrependDecorators(
			decor.Name(truncate(td.Title, 20)),
		),
		mpb.AppendDecorators(
			decor.CountersKibiByte("% .0f / % .0f"),
			decor.Percentage(decor.WCSyncSpace),
			decor.Any(func(decor.Statistics) string {
				return td.Message
			}),
		),
	)
}

func FollowProgress() {
	downloads := loadActiveDownloadsFromFile()
	if len(downloads) == 0 {
		fmt.Println("📭 No active downloads.")
		return
	}

	lastMessages := make(map[int]string)
	p := mpb.New(mpb.WithWidth(40))
	bars := make(map[int]*mpb.Bar)
	messages := make(map[int]*string)

	for _, td := range downloads {
		msg := td.Message
		bar := p.New(
			td.TotalSize,
			mpb.BarStyle().Lbound("[").Filler("▓").Tip("█").Padding("░").Rbound("]"),
			mpb.PrependDecorators(
				decor.Name(truncate(td.Title, 15)),
			),
			mpb.AppendDecorators(
				decor.Any(func(_ decor.Statistics) string {
					return msg
				}),
			),
		)
		bars[td.TorrentID] = bar
		messages[td.TorrentID] = &msg
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
			downloads = loadActiveDownloadsFromFile()
			for _, td := range downloads {
				if bar, exists := bars[td.TorrentID]; exists {
					bar.SetCurrent(td.Progress)
					bar.SetTotal(td.TotalSize, td.Done)
					if msgPtr, ok := messages[td.TorrentID]; ok {
						if td.Message != lastMessages[td.TorrentID] {
							*msgPtr = td.Message
							lastMessages[td.TorrentID] = td.Message
						}
					}
				}
			}
		case <-signalChan:
			fmt.Println("\n🛑 Cancelled by user.")
			return
		case <-done:
			fmt.Println("\n✅ All downloads finished.")
			return
		}
	}
}

func WaitForActiveDownloads(timeout time.Duration) bool {
	start := time.Now()
	for time.Since(start) < timeout {
		if len(GetActiveDownloads()) > 0 {
			return true
		}
		time.Sleep(250 * time.Millisecond)
	}
	return false
}

func truncate(s string, n int) string {
	if len(s) > n {
		return s[:n-1] + "…"
	}
	return s
}
