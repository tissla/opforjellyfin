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
	"path/filepath"
	"sync"
	"time"
)

func HandleDownloadSession(entries []shared.TorrentEntry, outDir string) {

	// start downloads (with UIprogress)
	allTDs := []*shared.TorrentDownload{}
	var wg sync.WaitGroup

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

	//start ui
	go ui.FollowProgress()

	for i, entry := range entries {
		wg.Add(1)
		go func(i int, entry shared.TorrentEntry) {
			defer wg.Done()
			td := allTDs[i]
			_ = StartTorrent(context.Background(), td)

		}(i, entry)
	}

	wg.Wait()
	// cool spinner
	//spinner := ui.NewSpinner(" 🗃️ Placing files", ui.Animations["MoviePlacement"])

	// placing files
	StartPlacement(allTDs, outDir)

	// stop spinner
	//spinner.Stop()

	// print placement data
	for _, td := range allTDs {
		if len(td.PlacementFull) > 0 {
			fmt.Printf("🎞️  %s\n", ui.AnsiPadRight(td.Title, 36, ".."))
			for _, line := range td.PlacementFull {
				fmt.Printf("   → %s\n", line)
			}
		}
	}

	shared.ClearActiveDownloads()

	logger.DebugLog(true, "\n✅ All downloads finished and placed.")
}

// sequential
func StartPlacement(allTDs []*shared.TorrentDownload, outDir string) {
	index := metadata.LoadMetadataCache()

	for _, td := range allTDs {
		tmpDir := filepath.Join(os.TempDir(), fmt.Sprintf("opfor-tmp-%d", td.TorrentID))
		matcher.ProcessTorrentFiles(tmpDir, outDir, td, index)
	}
}
