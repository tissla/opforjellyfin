// cmd/clear.go

package cmd

import (
	"fmt"
	"os"

	"opforjellyfin/internal"

	"github.com/spf13/cobra"
)

var clearCmd = &cobra.Command{
	Use:   "clear",
	Short: "Clear all active downloads and temporary files",
	Run: func(cmd *cobra.Command, args []string) {
		if err := internal.ClearActiveDownloads(); err != nil {
			fmt.Fprintf(os.Stderr, "❌ Failed to clear active downloads: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("✅ Cleared active downloads and temporary files.")
	},
}

func init() {
	rootCmd.AddCommand(clearCmd)
}
