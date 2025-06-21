// cmd/progress.go
package cmd

import (
	"opforjellyfin/internal/ui"

	"github.com/spf13/cobra"
)

var progressCmd = &cobra.Command{
	Use:   "progress",
	Short: "Show progress for all active downloads",
	Run: func(cmd *cobra.Command, args []string) {
		ui.FollowProgress()
	},
}

// no need for this until background-download is completed

//func init() {
//	rootCmd.AddCommand(progressCmd)
//}
