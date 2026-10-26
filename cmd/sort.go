package cmd

import (
	"fmt"
	"opforjellyfin/internal/matcher"
	"opforjellyfin/internal/metadata"
	"opforjellyfin/internal/shared"
	"path/filepath"

	"github.com/spf13/cobra"
)

var sortCmd = &cobra.Command{
	Use:   "sort <path>",
	Short: "Move One Pace videos into their arc folders, keeping their filenames",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := shared.LoadConfig()
		if err != nil {
			return err
		}
		if cfg.TargetDir == "" {
			return fmt.Errorf("set a library directory with 'opfor setDir <path>' first")
		}
		index, err := metadata.LoadMetadataCache()
		if err != nil {
			return err
		}
		source, err := filepath.Abs(args[0])
		if err != nil {
			return err
		}
		files, err := matcher.FindVideoFiles(source)
		if err != nil {
			return err
		}
		placed, alreadyPlaced, failed := 0, 0, 0
		for _, file := range files {
			message, err := matcher.MatchAndPlaceVideo(file, cfg.TargetDir, index)
			if err != nil {
				failed++
				cmd.Printf("Left unchanged: %s (%v)\n", file, err)
			} else if message == "" {
				alreadyPlaced++
			} else {
				placed++
				cmd.Println(message)
			}
		}
		cmd.Printf("%d placed, %d already in place, %d left unchanged.\n", placed, alreadyPlaced, failed)
		if failed > 0 {
			return fmt.Errorf("%d file(s) could not be placed", failed)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(sortCmd)
}
