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

var force bool

var setDirCmd = &cobra.Command{
	Use:   "setDir <path>",
	Short: "Set the default target directory",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		SetDir(args[0], force)
	},
}

// sets a directory as the configs default directory, then fills it with metadata
func SetDir(dir string, force bool) {
	abs, err := filepath.Abs(dir)
	if err != nil {
		log.Fatalf("❌ Invalid directory: %v", err)
	}

	cfg := shared.LoadConfig()
	cfg.TargetDir = abs
	shared.SaveConfig(cfg)

	fmt.Println("✅ Default target directory set to:", abs)

	if force {
		metadata.FetchAllMetadata(abs, cfg)
	} else {
		metadata.SyncMetadata(abs, cfg)
	}
}

func init() {
	setDirCmd.Flags().BoolVarP(&force, "force", "f", false, "Force download new metadata")
	rootCmd.AddCommand(setDirCmd)

}
