//go:build !darwin

package main

// windowID is only known on macOS (see windowid_darwin.go).
func windowID() int { return 0 }
