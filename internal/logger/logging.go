// log/logging.go
package logger

import (
	"fmt"
	"log"
	"os"
	"strings"
)

var debugEnabled bool = false
var debugFile *os.File

func EnableDebugLogging() {
	debugEnabled = true

	f, err := os.OpenFile("debug.log", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0666)
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Could not open debug.log: %v\n", err)
		return
	}

	debugFile = f
	log.SetOutput(f)
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}

func DebugLog(showUser bool, format string, args ...any) {
	if showUser {
		fmt.Printf(format+"\n", args...)
	}

	if debugEnabled && debugFile != nil {
		log.SetOutput(debugFile)
		log.Printf(format, args...)
	}
}

func ShowLogEntries(n int) {
	if !debugEnabled {
		return
	}

	data, err := os.ReadFile("debug.log")
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Could not open debug.log: %v\n", err)
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
