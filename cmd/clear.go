// cmd/clear.go

package cmd

import (
	"fmt"
	"opforjellyfin/internal/shared"

	"github.com/spf13/cobra"
)

var clearCmd = &cobra.Command{
	Use:   "clear",
	Short: "Clear all active downloads and temporary files",
	Run: func(cmd *cobra.Command, args []string) {

		shared.ClearActiveDownloads()

		fmt.Println("✅ Cleared active downloads and temporary files.")
	},
}

func init() {
	rootCmd.AddCommand(clearCmd)
}
