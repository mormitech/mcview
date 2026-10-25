//go:build !windows

package main

// windowsDiscardChoice is only for Windows (see discard_windows.go).
func windowsDiscardChoice(string, string) discardChoice { return choiceCancel }
