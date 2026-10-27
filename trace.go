package main

import (
	"fmt"
	"os"
	"sync"
	"time"
)

// With MORMIVIEW_TRACE=<file> in the environment mormiView appends a line with the time (epoch milliseconds) to that file at each
// step of starting and showing a picture, from both Go and the UI. It is how a slow start is tracked down; without the variable
// it costs nothing.
var (
	traceMu   sync.Mutex
	tracePath = os.Getenv("MORMIVIEW_TRACE")
)

func trace(format string, args ...any) {
	if tracePath == "" {
		return
	}
	traceMu.Lock()
	defer traceMu.Unlock()
	f, err := os.OpenFile(tracePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return
	}
	defer f.Close()
	fmt.Fprintf(f, "%d pid%d %s\n", time.Now().UnixMilli(), os.Getpid(), fmt.Sprintf(format, args...))
}

// Trace lets the UI write to the trace.
func (a *App) Trace(msg string) { trace("ui: %s", msg) }
