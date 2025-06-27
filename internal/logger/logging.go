// log/logging.go
package logger

import (
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"sync"
)

var (
	debugEnabled bool
	debugFile    *os.File
	debugLogger  *log.Logger // logger used by all
	logMu        sync.Mutex  // for the log-function
)

// always enabled
func EnableDebugLogging() {
	debugEnabled = true

	f, err := os.OpenFile("debug.log", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0666)
	if err != nil {
		fmt.Fprintf(os.Stderr, "logger: could not open debug.log: %v\n", err)
		return
	}

	debugFile = f
	debugLogger = log.New(f, "", log.LstdFlags|log.Lshortfile)
}

// threadsafe logger
func Log(showUser bool, format string, args ...any) {
	if showUser {
		fmt.Printf(format+"\n", args...)
	}

	if debugEnabled && debugFile != nil {
		logMu.Lock()
		defer logMu.Unlock()

		debugLogger.Printf(format, args...)
	}
}

// shows entries in debug.log if there are any (n = number of lines)
func ShowLogEntries(n int) {

	data, err := os.ReadFile("debug.log")
	if err != nil {
		fmt.Fprintf(os.Stderr, "logger: could not open debug.log: %v\n", err)
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

// redirector
type debugLogWriter struct{}

func (w debugLogWriter) Write(p []byte) (n int, err error) {
	Log(false, "%s", strings.TrimSpace(string(p)))
	return len(p), nil
}

func NewDebugLogWriter() io.Writer {
	return debugLogWriter{}
}
