package torrent

import (
	"context"
	"fmt"
	"opforjellyfin/internal/logger"
	"opforjellyfin/internal/matcher"
	"opforjellyfin/internal/metadata"
	"opforjellyfin/internal/shared"
	"opforjellyfin/internal/ui"
	"os"
	"os/signal"
	"path/filepath"
	"sync"
	"syscall"
	"time"
)

func HandleDownloadSession(entries []shared.TorrentEntry, outDir string) {
	// based on testing, might need to change
	const maxConcurrent = 5

	// context (for ctrl C cancellation)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle Ctrl+C
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigChan
		logger.Log(true, "\n🛑 Received interrupt signal, cancelling downloads...")
		cancel()
	}()

	// Prepare all download metadata first
	allTDs := []*shared.TorrentDownload{}
	for _, entry := range entries {
		dKey := ui.StyleFactory(fmt.Sprintf("%4d", entry.DownloadKey), ui.Style.Pink)
		title := ui.StyleFactory(entry.TorrentName, ui.Style.LBlue)

		td := &shared.TorrentDownload{
			Title:        fmt.Sprintf("%s: %s (%s)", dKey, title, entry.Quality),
			TorrentID:    entry.TorrentID,
			FullTitle:    entry.Title,
			Started:      time.Now(),
			ChapterRange: entry.ChapterRange,
		}

		shared.SaveTorrentDownload(td)
		allTDs = append(allTDs, td)
	}

	// Start UI progress monitoring
	doneChan := make(chan struct{})
	go ui.FollowProgress(doneChan)

	// Create work queue
	workQueue := make(chan int, len(entries))
	for i := range entries {
		workQueue <- i
	}
	close(workQueue)

	// Start worker goroutines
	var wg sync.WaitGroup
	for w := 0; w < maxConcurrent; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := range workQueue {
				select {
				case <-ctx.Done():
					return
				default:
					td := allTDs[i]
					// 30 min timeout / download
					downloadCtx, downloadCancel := context.WithTimeout(ctx, 30*time.Minute)
					if err := StartTorrent(downloadCtx, td); err != nil {
						if err == context.DeadlineExceeded {
							logger.Log(true, "Download timeout for %s (no progress in 30 min)", td.Title)
							td.PlacementProgress = "❌ Timeout - no seeders?"
						} else if err == context.Canceled {
							td.PlacementProgress = "❌ Cancelled"
						} else {
							logger.Log(true, "Download failed for %s: %v", td.Title, err)
							td.PlacementProgress = "❌ Failed"
						}
						shared.SaveTorrentDownload(td)
					}
					downloadCancel()
				}
			}
		}()
	}

	// Wait for all downloads to complete or context cancellation
	downloadsDone := make(chan struct{})
	go func() {
		wg.Wait()
		close(downloadsDone)
	}()

	select {
	case <-downloadsDone:
		// do nothing
	case <-ctx.Done():
		// when cancelled manually
		logger.Log(true, "Waiting for downloads to stop...")
		select {
		case <-downloadsDone:
		case <-time.After(5 * time.Second):
			logger.Log(true, "Force stopping...")
		}
	}

	// only place if not cancelled
	if ctx.Err() == nil {
		StartPlacement(allTDs, outDir)
	}

	// UI signal
	doneChan <- struct{}{}

	// Wait for UI to finish (for a sec)
	select {
	case <-doneChan:
	case <-time.After(1 * time.Second):
	}

	// Print placement results
	for _, td := range allTDs {
		if len(td.PlacementFull) > 0 {
			fmt.Printf("🎞️  %s\n", ui.AnsiPadRight(td.Title, 36, ".."))
			for _, line := range td.PlacementFull {
				fmt.Printf("   → %s\n", line)
			}
		}
	}

	shared.ClearActiveDownloads()

	if ctx.Err() != nil {
		logger.Log(true, "\n❌ Downloads cancelled by user.")
	} else {
		logger.Log(true, "\n✅ All downloads finished and placed.")
	}
}

// sequential
func StartPlacement(allTDs []*shared.TorrentDownload, outDir string) {
	index := metadata.LoadMetadataCache()

	for _, td := range allTDs {
		tmpDir := filepath.Join(os.TempDir(), fmt.Sprintf("opfor-tmp-%d", td.TorrentID))
		matcher.ProcessTorrentFiles(tmpDir, outDir, td, index)
	}
}
