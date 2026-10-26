// cmd/setDir.go
package cmd

import (
	"fmt"
	"log"
	"opforjellyfin/internal/metadata"
	"opforjellyfin/internal/shared"
	"path/filepath"

	"github.com/spf13/cobra"
)

var (
	force      bool
	thumbnails bool
)

var setDirCmd = &cobra.Command{
	Use:   "setDir <path>",
	Short: "Set the default target directory",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		abs, err := filepath.Abs(args[0])
		if err != nil {
			return err
		}
		cfg, err := shared.LoadConfig()
		if err != nil {
			return err
		}
		cfg.TargetDir = abs
		if cmd.Flags().Changed("thumbnails") {
			cfg.Thumbnails = thumbnails
		}
		if err := shared.SaveConfig(*cfg); err != nil {
			return err
		}
		cmd.Printf("Default target directory: %s (thumbnails: %t)\n", abs, cfg.Thumbnails)
		if force {
			return metadata.FetchAllMetadata(cfg)
		}
		return metadata.SyncMetadata(cfg)
	},
}

func init() {
	setDirCmd.Flags().BoolVarP(&force, "force", "f", false, "Force download new metadata")
	setDirCmd.Flags().BoolVar(&thumbnails, "thumbnails", false, "Download episode thumbnails; saves the choice for future syncs")
	rootCmd.AddCommand(setDirCmd)
}
