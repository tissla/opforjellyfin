package ui

import (
	"fmt"
	"opforjellyfin/internal/logger"
	"strings"
	"sync"
	"time"

	"github.com/charmbracelet/x/ansi"
)

type Spinner struct {
	stop chan struct{}
	done chan struct{}
	once sync.Once
}

func NewSpinner(message string, frames []string) *Spinner {
	return newSpinner(message, frames, 1, 150*time.Millisecond)
}

func NewMultirowSpinner(frames []string, rows int) *Spinner {
	return newSpinner("", frames, rows, 200*time.Millisecond)
}

func newSpinner(message string, frames []string, rows int, interval time.Duration) *Spinner {
	if len(frames) == 0 {
		frames = []string{"⏳"}
	}
	pictures := make([]string, len(frames))
	height := max(1, rows)
	for i, frame := range frames {
		if message != "" {
			frame = message + " " + frame
		}
		pictures[i] = strings.TrimRight(frame, "\n")
		height = max(height, strings.Count(pictures[i], "\n")+1)
	}

	s := &Spinner{stop: make(chan struct{}), done: make(chan struct{})}
	logger.BufferUserOutput(true)
	go func() {
		defer close(s.done)
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		previous, drawn, i := "", false, 0

		for {
			select {
			case <-s.stop:
				if drawn {
					fmt.Printf("\033[%dA\r\033[J", height)
				}
				logger.BufferUserOutput(false)
				return
			case <-ticker.C:
				lines := strings.Split(pictures[i%len(pictures)], "\n")
				i++
				width := max(0, GetTerminalWidth()-1)
				var frame strings.Builder
				for row := 0; row < height; row++ {
					line := ""
					if row < len(lines) {
						line = lines[row]
					}
					fmt.Fprintf(&frame, "\r%s\033[K\n", ansi.Truncate(line, width, ""))
				}
				next := frame.String()
				messages := logger.TakeUserOutput()
				if next == previous && messages == "" {
					continue
				}

				var output strings.Builder
				if drawn {
					fmt.Fprintf(&output, "\033[%dA", height)
				}
				output.WriteString("\r")
				output.WriteString(strings.ReplaceAll(messages, "\n", "\033[K\r\n"))
				output.WriteString(next)
				fmt.Print(output.String())
				previous, drawn = next, true
			}
		}
	}()
	return s
}

func (s *Spinner) Stop() {
	s.once.Do(func() { close(s.stop) })
	<-s.done
}
