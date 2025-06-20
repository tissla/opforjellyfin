// cmd/status.go
package cmd

import (
	"fmt"
	"opforjellyfin/internal"

	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show currently active downloads",
	Run: func(cmd *cobra.Command, args []string) {
		downloads := internal.GetActiveDownloads()
		if len(downloads) == 0 {
			fmt.Println("📭 No active downloads.")
			return
		}
		fmt.Println("📦 Active Downloads:")
		for _, d := range downloads {
			fmt.Printf("- %s: %.2f%%\n", d.Title, (float64(d.Progress)/float64(d.TotalSize))*100)
		}
	},
}

func init() {
	rootCmd.AddCommand(statusCmd)
}
