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
	// based on tests
	const maxConcurrent = 5

	// Create a context that can be cancelled with Ctrl+C
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

	// Load metadata index once
	metadataIndex := metadata.LoadMetadataCache()

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

	// Channel to collect placement results
	placementResults := make(chan *shared.TorrentDownload, len(entries))

	// Start worker goroutines
	var wg sync.WaitGroup
	for w := 0; w < maxConcurrent; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := range workQueue {
				select {
				case <-ctx.Done():
					return // Exit if cancelled
				default:
					td := allTDs[i]

					// Download
					downloadCtx, downloadCancel := context.WithTimeout(ctx, 30*time.Minute)
					err := StartTorrent(downloadCtx, td)
					downloadCancel()

					if err != nil {
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
						placementResults <- td
						continue
					}

					// Place immediately after download completes
					tmpDir := filepath.Join(os.TempDir(), fmt.Sprintf("opfor-tmp-%d", td.TorrentID))
					matcher.ProcessTorrentFiles(tmpDir, outDir, td, metadataIndex)

					// Clean up temp directory immediately
					if err := os.RemoveAll(tmpDir); err != nil {
						logger.Log(false, "Failed to remove temp dir %s: %v", tmpDir, err)
					}

					placementResults <- td
				}
			}
		}()
	}

	// Collect results
	go func() {
		wg.Wait()
		close(placementResults)
	}()

	// Collect all placement results
	var placedTorrents []*shared.TorrentDownload
	for td := range placementResults {
		placedTorrents = append(placedTorrents, td)
	}

	// Signal UI that downloads are done
	doneChan <- struct{}{}

	// Wait for UI to finish
	select {
	case <-doneChan:
	case <-time.After(1 * time.Second):
		// Don't wait forever for UI
	}

	// Print placement results
	for _, td := range placedTorrents {
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
