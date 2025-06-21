package torrent

import (
	"context"
	"fmt"
	"opforjellyfin/internal/shared"
	"opforjellyfin/internal/ui"
	"time"

	"github.com/charmbracelet/x/ansi"
)

func HandleDownloadSession(entries []shared.TorrentEntry, outDir string) {

	// cleanup eventual residue
	shared.ClearActiveDownloads()

	// start downloads
	go StartMultipleDownloads(context.Background(), entries, outDir)

	// start progress-UI
	fmt.Println("🚀 Downloads started!")
	time.Sleep(1 * time.Second)
	ui.FollowProgress()

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
