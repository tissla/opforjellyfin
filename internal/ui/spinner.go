package ui

import (
	"fmt"
	"strings"
	"time"
)

// spinner struct
type Spinner struct {
	stop chan struct{}
	done chan struct{}
}

// spinner creator
func NewSpinner(message string, frames []string) *Spinner {
	s := &Spinner{
		stop: make(chan struct{}),
		done: make(chan struct{}),
	}

	go func() {
		i := 0
		for {
			select {
			case <-s.stop:
				fmt.Print("\r")
				s.done <- struct{}{}
				return
			default:
				fmt.Printf("\r%s %s", message, frames[i%len(frames)])
				time.Sleep(150 * time.Millisecond)
				i++
			}
		}
	}()

	return s
}

// send stop signal
func (s *Spinner) Stop() {
	close(s.stop)
	<-s.done
	fmt.Print("\r" + blankLine(40) + "\r\n")
}

func blankLine(width int) string {
	return fmt.Sprintf("%-*s", width, "")
}

// Creates a multi-row spinner, AnimationFreames, Number of rows,
func NewMultirowSpinner(frames []string, rows int) *Spinner {
	if len(frames) == 0 {
		frames = []string{"⏳"}
	}

	s := &Spinner{
		stop: make(chan struct{}),
		done: make(chan struct{}),
	}

	for i := 0; i < rows+1; i++ {
		fmt.Print("\n")
	}

	go func() {
		i := 0
		for {
			select {
			case <-s.stop:
				clearLines(rows + 1)
				s.done <- struct{}{}
				return
			default:
				frame := frames[i%len(frames)]
				frameLineCount := strings.Count(frame, "\n") + 1
				clearLines(frameLineCount + 1)

				printMultiline("", frame)
				time.Sleep(200 * time.Millisecond)
				i++

			}
		}
	}()

	return s
}

// helper
func printMultiline(message string, frame string) {
	lines := strings.Split(frame, "\n")
	fmt.Printf("%s", message)
	for _, line := range lines {
		fmt.Print(line + "\n")
	}
}

// Clears n lines from current cursor pos upwards
func clearLines(n int) {
	for i := 0; i < n; i++ {
		fmt.Print("\r\033[K") // Clear current line
		if i < n-1 {
			fmt.Print("\033[1A") // Move cursor up (except last line)
		}
	}
}
