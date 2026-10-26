// cmd/clear.go

package cmd

import (
	"fmt"
	"opforjellyfin/internal/shared"

	"github.com/spf13/cobra"
)

var clearCmd = &cobra.Command{
	Use:   "clear",
	Short: "Delete temporary downloads, including unplaced files",
	Run: func(cmd *cobra.Command, args []string) {

		shared.ClearActiveDownloads()
		if err := shared.CleanupTempDirs(); err != nil {
			fmt.Println(err)
			return
		}

		fmt.Println("✅ Cleared temporary files.")
	},
}

func init() {
	rootCmd.AddCommand(clearCmd)
}
