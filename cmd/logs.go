// cmd/logs.go
package cmd

import (
	"opforjellyfin/internal"

	"github.com/spf13/cobra"
)

var lines int

var logsCmd = &cobra.Command{
	Use:   "logs",
	Short: "Show last [n] last entries in debug.log",
	Run: func(cmd *cobra.Command, args []string) {

		if lines <= 0 {
			lines = 20 
		}
		internal.ShowLogEntries(lines)

	},
}

func init() {
	rootCmd.AddCommand(logsCmd)
	logsCmd.Flags().IntVarP(&lines, "lines", "l", 20, "Number of lines to show. Default 20")
}
