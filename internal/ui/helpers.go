package ui

import (
	"fmt"
	"strings"
)

// helper
func PrintMultiline(frame string) {
	lines := strings.Split(frame, "\n")
	for _, line := range lines {
		fmt.Print(line + "\n")
	}
}

// Clears n lines from current cursor pos upwards
func ClearLines(n int) {
	for i := 0; i < n; i++ {
		fmt.Print("\r\033[K") // Clear current line
		if i < n-1 {
			fmt.Print("\033[1A") // Move cursor up (except last line)
		}
	}
}
