// main.go
package main

import (
	"context"
	"os"

	"opforjellyfin/cmd"
	"opforjellyfin/internal/logger"
	"opforjellyfin/internal/shared"

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

	root := cmd.RootCommand()

	if err := fang.Execute(
		context.Background(),
		root,
		fang.WithoutManpage(),
		fang.WithoutCompletions(),
	); err != nil {
		os.Exit(1)
	}
}
