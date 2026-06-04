package cmd

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"opforjellyfin/internal/metadata"
	"opforjellyfin/internal/shared"

	"github.com/spf13/cobra"
)

var (
	deleteYes    bool
	deleteDryRun bool
	deleteArc    string
)

var deleteCmd = &cobra.Command{
	Use:     "delete <chapterRange>|<arcName> [chapterRange|arcName...]",
	Aliases: []string{"remove", "rm"},
	Short:   "Delete downloaded episode video files",
	Example: "opfor delete 15\nopfor delete 15-15 25-27 --yes\nopfor delete Skypia --dry-run",
	Args: func(cmd *cobra.Command, args []string) error {
		if deleteArc == "" && len(args) == 0 {
			return fmt.Errorf("requires at least one chapter range or --arc")
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		cfg := shared.LoadConfig()
		if cfg.TargetDir == "" {
			fmt.Println("No target directory set. Use 'setDir' first.")
			return
		}

		index := metadata.LoadMetadataCache()
		if index == nil || len(index.Seasons) == 0 {
			fmt.Println("No metadata index found. Run 'sync' first.")
			return
		}

		effectiveArc := deleteArc
		if effectiveArc == "" {
			if positionalName, ok := positionalArcName(args); ok {
				effectiveArc = positionalName
				args = nil
			}
		}

		if effectiveArc != "" {
			arc, err := metadata.FindArc(index, effectiveArc)
			if err != nil {
				fmt.Printf("Error: %v\n", err)
				return
			}

			args = arc.EpisodeRanges
			fmt.Printf("Matched arc %s (%s), chapters %s.\n", arc.Name, arc.Season, arc.Range)
		}

		candidates, missing, err := metadata.FindEpisodeVideosForDelete(cfg.TargetDir, index, args)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}

		for _, chapterRange := range missing {
			fmt.Printf("No metadata episode found for chapter range %s\n", chapterRange)
		}

		if len(candidates) == 0 {
			fmt.Println("No downloaded episode videos found to delete.")
			return
		}

		fmt.Println("Episode videos selected for deletion:")
		for _, candidate := range candidates {
			relPath, err := filepath.Rel(cfg.TargetDir, candidate.Path)
			if err != nil {
				relPath = candidate.Path
			}
			fmt.Printf("   - %s (%s): %s\n", candidate.ChapterRange, candidate.Season, relPath)
		}

		if deleteDryRun {
			fmt.Println("Dry run only. No files were deleted.")
			return
		}

		if !deleteYes && !confirmDelete(len(candidates)) {
			fmt.Println("Cancelled.")
			return
		}

		if err := metadata.DeleteEpisodeVideos(candidates); err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}

		fmt.Printf("Deleted %d episode video file(s).\n", len(candidates))
	},
}

func confirmDelete(count int) bool {
	fmt.Printf("Delete %d file(s)? Type yes to continue: ", count)

	reader := bufio.NewReader(os.Stdin)
	answer, err := reader.ReadString('\n')
	if err != nil {
		return false
	}

	return strings.EqualFold(strings.TrimSpace(answer), "yes")
}

func init() {
	deleteCmd.Flags().BoolVarP(&deleteYes, "yes", "y", false, "Delete without asking for confirmation")
	deleteCmd.Flags().BoolVar(&deleteDryRun, "dry-run", false, "Show files that would be deleted without deleting them")
	deleteCmd.Flags().StringVarP(&deleteArc, "arc", "a", "", "Delete downloaded episode videos for an arc name")
	rootCmd.AddCommand(deleteCmd)
}
