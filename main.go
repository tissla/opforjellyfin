// main.go
package main

import (
	"context"
	"log"
	"os"

	"opforjellyfin/cmd"
	"opforjellyfin/internal/logger"
	"opforjellyfin/internal/shared"

	"github.com/charmbracelet/fang"
)

func main() {

	shared.EnsureConfigExists()

	root := cmd.RootCommand()

	if err := fang.Execute(
		context.Background(),
		root,
		fang.WithoutManpage(),
		fang.WithVersion("1.0.1"),
		fang.WithoutCompletions(),
	); err != nil {
		os.Exit(1)
	}
}

// routes external loggers through debug log
func init() {
	log.SetOutput(logger.NewDebugLogWriter())
}
