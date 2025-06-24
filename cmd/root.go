// cmd/root.go
package cmd

import (
	"fmt"
	"opforjellyfin/internal/logger"

	"github.com/spf13/cobra"
)

var debugMode bool

var rootCmd = &cobra.Command{
	Use:   "opfor",
	Short: "Automates download and metadata for One Pace to Jellyfin",
	Long:  "A CLI tool to download One Pace releases and organize them for use with Jellyfin.",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {

		if debugMode {

			logger.EnableDebugLogging()
		}
	},
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("📦 Use a subcommand, e.g. 'download', 'progress' or 'list'")
	},
}

func init() {
	rootCmd.PersistentFlags().BoolVar(&debugMode, "debug", false, "Enable debug logging")
}

func RootCommand() *cobra.Command {
	return rootCmd
}
