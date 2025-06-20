// main.go
package main

import (
	"context"
	"os"
	"time"

	"opforjellyfin/cmd"
	"opforjellyfin/internal"

	"github.com/briandowns/spinner"
	"github.com/charmbracelet/fang"
)

func main() {
	internal.EnsureConfigExists()

	for _, arg := range os.Args {
		if arg == "--debug" {
			internal.EnableDebugLogging()
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
