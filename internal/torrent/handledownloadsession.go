package torrent

import (
	"context"
	"fmt"
	"opforjellyfin/internal/shared"
	"opforjellyfin/internal/ui"
	"sync"
	"time"

	"github.com/charmbracelet/x/ansi"
)

func HandleDownloadSession(entries []shared.TorrentEntry, outDir string) {

	// cleanup eventual residue
	shared.ClearActiveDownloads()

	// start downloads
	var wg sync.WaitGroup
	wg.Add(len(entries))

	for _, e := range entries {
		go func(entry shared.TorrentEntry) {
			defer wg.Done()
			_ = StartTorrent(context.Background(), entry, outDir)
		}(e)
	}

	// start progress-UI
	fmt.Println("🚀 Downloads started!")
	time.Sleep(1 * time.Second)
	ui.FollowProgress()

	//wait for all downloads to finish, and files to get placed
	wg.Wait()

	// get placement data
	downloads := shared.GetActiveDownloads()

	for _, td := range downloads {
		if len(td.Messages) > 0 {
			fmt.Printf("🎞️  %s\n", ansi.Truncate(td.Title, 30, ".."))
			for _, line := range td.Messages {
				fmt.Printf("   → %s\n", line)
			}
		}
	}

	//cleanup
	shared.ClearActiveDownloads()

	fmt.Println("\n✅ All downloads finished.")
}
