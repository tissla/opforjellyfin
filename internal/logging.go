// internal/logging.go
package internal

import (
	"fmt"
	"log"
	"os"
	"strings"
)

var debugEnabled bool


// InitDebugLogging initializes logging to debug.log
func InitDebugLogging(enabled bool) {
	debugEnabled = enabled

	if !debugEnabled {
		return
	}

	f, err := os.OpenFile("debug.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Printf("❌ Could not open debug.log: %v", err)
		return
	}

	log.SetOutput(f)
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}


func ShowLogEntries(n int) {
	if !debugEnabled {
		return
	}

	data, err := os.ReadFile("debug.log")
	if err != nil {
		log.Printf("❌ Could not open debug.log: %v", err)
		return
	}

	lines := strings.Split(string(data), "\n")
	if len(lines) > n {
		lines = lines[len(lines)-n:]
	}

	for _, line := range lines {
		if line != "" {
			fmt.Println(line)
		}
	}
}



