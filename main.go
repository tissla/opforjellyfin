// main.go
package main

import (
	"context"
	"os"

	"opforjellyfin/cmd"
	"opforjellyfin/internal"

	"github.com/charmbracelet/fang"
)

func main() {
	internal.EnsureConfigExists()
	internal.CleanupStaleDownloads()

	for _, arg := range os.Args {
		if arg == "--debug" {
			internal.InitDebugLogging(true)
			break
		}
	}

	if err := fang.Execute(
		context.Background(),
		cmd.RootCommand(),
		fang.WithoutManpage(),
		fang.WithoutCompletions(),
	); err != nil {
		os.Exit(1)
	}
}
