package torrent

import (
	"context"
	"fmt"
	"opforjellyfin/internal/matcher"
	"opforjellyfin/internal/metadata"
	"opforjellyfin/internal/shared"
	"opforjellyfin/internal/ui"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/charmbracelet/x/ansi"
)

func HandleDownloadSession(entries []shared.TorrentEntry, outDir string) {

	// start downloads (with UIprogress)
	allTDs, _ := StartDownloads(entries, outDir)

	// cool spinner
	spinner := ui.NewSpinner(" 🗃️ Placing files", ui.Animations["MoviePlacement"])

	// placing files
	StartPlacement(allTDs, outDir)

	// stop spinner
	spinner.Stop()

	// get placement data
	for _, td := range allTDs {
		if len(td.Messages) > 0 {
			fmt.Printf("🎞️  %s\n", ansi.Truncate(td.Title, 36, ".."))
			for _, line := range td.Messages {
				fmt.Printf("   → %s\n", line)
			}
		}
	}

	fmt.Println("\n✅ All downloads finished.")
}

func StartDownloads(entries []shared.TorrentEntry, outDir string) ([]*shared.TorrentDownload, []error) {
	var wg sync.WaitGroup
	var mu sync.Mutex
	var allTDs []*shared.TorrentDownload
	var allErrors []error

	// start downloads one at a time
	for _, e := range entries {
		wg.Add(1)
		go func(entry shared.TorrentEntry) {
			defer wg.Done()
			td, err := StartTorrent(context.Background(), entry, outDir)
			mu.Lock()
			defer mu.Unlock()
			if err != nil {
				allErrors = append(allErrors, fmt.Errorf("download %d (%s): %w", entry.DownloadKey, entry.TorrentName, err))
				return
			}
			allTDs = append(allTDs, td)
		}(e)
	}

	fmt.Println("🚀 Downloads started!")

	//wait for downloads to kick it up
	for {
		if len(shared.GetActiveDownloads()) >= len(entries) {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}
	//start ui
	ui.FollowProgress()

	wg.Wait()

	return allTDs, allErrors
}

// sequential
func StartPlacement(allTDs []*shared.TorrentDownload, outDir string) {
	index := metadata.LoadMetadataCache()

	for _, td := range allTDs {
		tmpDir := filepath.Join(os.TempDir(), fmt.Sprintf("opfor-tmp-%d", td.TorrentID))
		matcher.ProcessTorrentFiles(tmpDir, outDir, td, index)
	}
}
