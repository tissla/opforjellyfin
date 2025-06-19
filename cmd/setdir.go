// cmd/setDir.go
package cmd

import (
	"opforjellyfin/internal"

	"github.com/spf13/cobra"
)

var force bool

var setDirCmd = &cobra.Command{
	Use:   "setDir <path>",
	Short: "Set the default target directory",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		internal.SetDir(args[0], force)
	},
}

func init() {
	setDirCmd.Flags().BoolVarP(&force, "force", "f", false, "Force download new metadata")
	rootCmd.AddCommand(setDirCmd)

}
