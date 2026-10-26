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
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := shared.LoadConfig()
		if err != nil {
			return err
		}
		if cfg.TargetDir == "" {
			return fmt.Errorf("no target directory set; use 'setDir' first")
		}
		return metadata.SyncMetadata(cfg)
	},
}

func init() {
	rootCmd.AddCommand(syncCmd)
}
