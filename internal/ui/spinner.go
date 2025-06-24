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
				ClearLines(rows + 1)
				s.done <- struct{}{}
				return
			default:
				frame := frames[i%len(frames)]
				frameLineCount := strings.Count(frame, "\n") + 1
				ClearLines(frameLineCount + 1)

				PrintMultiline(frame)
				time.Sleep(200 * time.Millisecond)
				i++

			}
		}
	}()

	return s
}
