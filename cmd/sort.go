package cmd

import (
	"github.com/spf13/cobra"
)

var sortCmd = &cobra.Command{
	Use:   "sort",
	Short: "Sorts the One Pace video files in the target directory into the set opfor directory.",
	Run: func(cmd *cobra.Command, args []string) {

	},
}

func init() {
	rootCmd.AddCommand(sortCmd)
}
