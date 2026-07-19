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

// MaxConcurrent is the number of torrents downloaded (or, with seed=true,
// downloaded-and-seeded) at once. With seed=true, a worker never returns to
// pick up more work until the whole session is stopped (Ctrl+C) - so at most
// MaxConcurrent of the requested entries will ever start seeding in one run.
const MaxConcurrent = 5

func HandleDownloadSession(entries []shared.TorrentEntry, outDir string, seed bool) {

	// Create a context that can be cancelled with Ctrl+C
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle Ctrl+C
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigChan
		logger.Log(true, "\n❌ Received interrupt signal, cancelling downloads...")
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
	for w := 0; w < MaxConcurrent; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := range workQueue {
				select {
				case <-ctx.Done():
					return // Exit if cancelled
				default:
					td := allTDs[i]

					// Download (and, if seed is set, keep uploading afterward
					// until ctx is cancelled - StartTorrent blocks for that).
					err := StartTorrent(ctx, td, seed)

					// A cancel that arrives *after* the download already
					// finished just means the user stopped a --seed session -
					// that's a successful download and should still be
					// placed, not treated as a failure.
					seedingStoppedAfterSuccess := err == context.Canceled && td.Done

					if err != nil && !seedingStoppedAfterSuccess {
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

					tmpBase, err := shared.GetTempDir()
					if err != nil {
						logger.Log(false, "failed to find temp dir: %v", err)
					}
					// Place immediately after download completes
					tmpDir := filepath.Join(tmpBase, fmt.Sprintf("opfor-tmp-%d", td.TorrentID))
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
		logger.Log(true, "\n❌ Downloads cancelled.")
	} else {
		logger.Log(true, "\n✅ All downloads finished and placed.")
	}
}
