// main.go
package main

import (
	"context"
	"os"
	"time"

	"opforjellyfin/cmd"
	"opforjellyfin/internal/logger"
	"opforjellyfin/internal/shared"

	"github.com/briandowns/spinner"
	"github.com/charmbracelet/fang"
)

func main() {
	shared.EnsureConfigExists()

	for _, arg := range os.Args {
		if arg == "--debug" {
			logger.EnableDebugLogging()
			break
		}
	}

	s := spinner.New(spinner.CharSets[14], 100*time.Millisecond)
	s.Prefix = "⏳ Loading commands "
	s.Start()
	root := cmd.RootCommand()
	s.Stop()

	if err := fang.Execute(
		context.Background(),
		root,
		fang.WithoutManpage(),
		fang.WithoutCompletions(),
	); err != nil {
		os.Exit(1)
	}
}
