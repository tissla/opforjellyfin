// cmd/root.go
package cmd

import (
	"fmt"
	"log"
	"os"

	"github.com/spf13/cobra"
)

var debugMode bool

var rootCmd = &cobra.Command{
	Use:   "opfor",
	Short: "Automates download and metadata for One Pace to Jellyfin",
	Long:  "A CLI tool to download One Pace releases and organize them for use with Jellyfin.",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		if debugMode {
			f, err := os.OpenFile("debug.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
			if err != nil {
				fmt.Printf("❌ Could not open debug.log: %v\n", err)
			} else {
				log.SetOutput(f)
			}
		}
	},
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("📦 Use a subcommand, e.g. 'download', 'progress' or 'list'")
	},
}

func init() {
	rootCmd.PersistentFlags().BoolVar(&debugMode, "debug", false, "Enable debug logging")

	rootCmd.AddCommand(downloadCmd)
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(progressCmd)

}

func RootCommand() *cobra.Command {
	return rootCmd
}
