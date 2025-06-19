// cmd/sync.go
package cmd

import (
	"fmt"
	"opforjellyfin/internal"

	"github.com/spf13/cobra"
)

var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Update metadata library with new content from GitHub",
	Run: func(cmd *cobra.Command, args []string) {
		cfg := internal.LoadConfig()
		if cfg.TargetDir == "" {
			fmt.Println("⚠️  No target directory set. Use 'setDir' first.")
			return
		}
		internal.SyncMetadata(cfg.TargetDir, cfg)
	},
}

func init() {
	rootCmd.AddCommand(syncCmd)
}
