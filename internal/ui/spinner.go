package ui

import (
	"fmt"
	"time"
)

// cool animation
var frames = []string{
	"📁 🎞️ ➝         📂",
	"📁  🎞️ ➝        📂",
	"📁   🎞️ ➝       📂",
	"📁    🎞️ ➝      📂",
	"📁     🎞️ ➝     📂",
	"📁      🎞️ ➝    📂",
	"📁       🎞️ ➝   📂",
	"📁        🎞️ ➝  📂",
	"📁         🎞️ ➝ 📂",
}

// spinner struct
type Spinner struct {
	stop chan struct{}
	done chan struct{}
}

// spinner creator
func NewFileMoveSpinner(message string) *Spinner {
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
				time.Sleep(200 * time.Millisecond)
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
