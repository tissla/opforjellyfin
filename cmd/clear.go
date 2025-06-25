// cmd/clear.go

package cmd

import (
	"fmt"
	"opforjellyfin/internal/shared"

	"github.com/spf13/cobra"
)

var clearCmd = &cobra.Command{
	Use:   "clear",
	Short: "Clear all temporary files, in case something stuck",
	Run: func(cmd *cobra.Command, args []string) {

		shared.ClearActiveDownloads()

		fmt.Println("✅ Cleared temporary files.")
	},
}

func init() {
	rootCmd.AddCommand(clearCmd)
}
