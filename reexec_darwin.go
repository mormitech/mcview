package main

import (
	"os"
	"syscall"
)

// startThroughRealPath starts the program again, at once and in the same process, by the path of the file itself when it was started through a link
// (see resolvedPath). The environment says it was done, so that it never repeats.
func startThroughRealPath() {
	if os.Getenv("MORMIVIEW_REEXEC") != "" {
		return
	}
	exe, err := os.Executable()
	if err != nil {
		return
	}
	real, changed := resolvedPath(exe)
	if !changed {
		return
	}
	args := append([]string{real}, os.Args[1:]...)
	syscall.Exec(real, args, append(os.Environ(), "MORMIVIEW_REEXEC=1")) // (when it fails, the program goes on as it is)
}
