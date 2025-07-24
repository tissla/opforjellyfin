// cmd/sync.go
package cmd

import (
	"fmt"
	"opforjellyfin/internal/metadata"
	"opforjellyfin/internal/shared"

	"github.com/spf13/cobra"
)

var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Update metadata library with new content from GitHub",
	Run: func(cmd *cobra.Command, args []string) {
		cfg := shared.LoadConfig()
		if cfg.TargetDir == "" {
			fmt.Println("⚠️  No target directory set. Use 'setDir' first.")
			return
		}
		err := metadata.SyncMetadata(cfg.TargetDir, cfg)
		if err != nil {
			fmt.Println("⚠️  Unable to sync metadata. (Is git installed?)")
		}
	},
}

func init() {
	rootCmd.AddCommand(syncCmd)
}
