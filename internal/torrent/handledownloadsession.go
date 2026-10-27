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
	"strings"
	"sync"
	"syscall"
	"time"
)

// MaxConcurrent is the number of torrents downloaded (or, with seed=true,
// downloaded-and-seeded) at once. With seed=true, a worker never returns to
// pick up more work until the whole session is stopped (Ctrl+C) - so at most
// MaxConcurrent of the requested entries will ever start seeding in one run.
const MaxConcurrent = 5

func HandleDownloadSession(entries []shared.TorrentEntry, outDir string, seed, all bool) {
	var started time.Time
	if all {
		started = time.Now()
	}
	// Create a context that can be cancelled with Ctrl+C
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle Ctrl+C
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(sigChan)
	go func() {
		select {
		case <-sigChan:
			logger.Log(true, "\n❌ Received interrupt signal, cancelling downloads...")
			cancel()
		case <-ctx.Done():
		}
	}()

	// Load metadata index once
	metadataIndex, err := metadata.LoadMetadataCache()
	if err != nil {
		logger.Log(true, "%v", err)
		return
	}

	// Prepare all download metadata first
	allTDs := []*shared.TorrentDownload{}
	totalVideos, totalKnown := 0, true
	for _, entry := range entries {
		totalVideos += len(entry.SelectedFiles)
		if entry.SelectedFiles == nil {
			totalKnown = false
		}
		dKey := ui.StyleFactory(fmt.Sprintf("%4d", entry.DownloadKey), ui.Style.Pink)
		title := ui.StyleFactory(entry.TorrentName, ui.Style.LBlue)
		displayTitle := fmt.Sprintf("%s: %s", dKey, title)
		if entry.Quality != "" && entry.Quality != "n/a" {
			displayTitle += fmt.Sprintf(" (%s)", entry.Quality)
		}

		td := &shared.TorrentDownload{
			Title:        displayTitle,
			TorrentID:    entry.TorrentID,
			FullTitle:    entry.Title,
			ChapterRange: entry.ChapterRange,
		}

		allTDs = append(allTDs, td)
	}
	if !totalKnown {
		totalVideos = 0
	}
	shared.SetDownloadTotal(totalVideos)

	// Start UI progress monitoring
	doneChan := make(chan struct{})
	logger.BufferUserOutput(true)
	go ui.FollowProgress(doneChan, started)

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
					shared.SaveTorrentDownload(td)

					// Download (and, if seed is set, keep uploading afterward
					// until ctx is cancelled - StartTorrent blocks for that).
					err := StartTorrent(ctx, td, seed, entries[i].SelectedFiles)

					// A cancel that arrives *after* the download already
					// finished just means the user stopped a --seed session -
					// that's a successful download and should still be
					// placed, not treated as a failure.
					seedingStoppedAfterSuccess := err == context.Canceled && td.Done

					if err != nil && !seedingStoppedAfterSuccess {
						if err == context.DeadlineExceeded {
							logger.Log(true, "Download timed out after 30 minutes: %s", td.Title)
							td.PlacementProgress = "Timed out"
						} else if err == context.Canceled {
							td.PlacementProgress = "Cancelled"
						} else {
							logger.Log(true, "Download failed for %s: %v", td.Title, err)
							td.PlacementProgress = "Failed"
						}
						shared.SaveTorrentDownload(td)
						shared.RemoveActiveDownload(td.TorrentID)
						placementResults <- td
						continue
					}

					tmpBase, err := shared.GetTempDir()
					if err != nil {
						logger.Log(true, "failed to find temp dir: %v", err)
						shared.RemoveActiveDownload(td.TorrentID)
						placementResults <- td
						continue
					}
					// Place immediately after download completes
					tmpDir := filepath.Join(tmpBase, fmt.Sprintf("opfor-tmp-%d", td.TorrentID))
					if err := matcher.ProcessTorrentFiles(outDir, td, metadataIndex); err != nil {
						logger.Log(true, "%v; retained download in %s", err, tmpDir)
					} else if err := os.RemoveAll(tmpDir); err != nil {
						logger.Log(false, "Failed to remove temp dir %s: %v", tmpDir, err)
					}

					shared.RemoveActiveDownload(td.TorrentID)
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

	printDownloadResults(placedTorrents, all, ctx.Err() != nil, time.Since(started))
	shared.ClearActiveDownloads()
}

func printDownloadResults(placedTorrents []*shared.TorrentDownload, all, cancelled bool, elapsed time.Duration) {
	for _, td := range placedTorrents {
		if all {
			for _, line := range td.PlacementFull {
				logger.Log(!strings.HasPrefix(line, "Placed: "), "%s", line)
			}
			continue
		}
		if len(td.PlacementFull) > 0 {
			fmt.Printf("🎞️  %s\n", ui.AnsiPadRight(td.Title, 36, ".."))
			for _, line := range td.PlacementFull {
				fmt.Printf("   → %s\n", line)
			}
		}
	}

	if all {
		placed, total := shared.GetDownloadProgress()
		status := "Finished"
		if cancelled {
			status = "Cancelled"
		}
		logger.Log(true, "%s %d/%d files in %s; %d remaining.", status, placed, total, elapsed.Truncate(time.Second), max(0, total-placed))
		if placed < total {
			logger.Log(true, "Run 'download all' again to retry; partial files are kept in .opfor-tmp.")
		}
		return
	}

	allPlaced := true
	for _, td := range placedTorrents {
		if !td.Placed {
			allPlaced = false
		}
	}
	if cancelled {
		logger.Log(true, "\n❌ Downloads cancelled.")
	} else if !allPlaced {
		logger.Log(true, "\nSome downloads could not be placed. Remaining files were kept in .opfor-tmp.")
	} else {
		logger.Log(true, "\n✅ All downloads finished and placed.")
	}
}
