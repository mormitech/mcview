package main

import (
	"encoding/base64"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"time"
)

// printCommand is the program that sends the picture file at path to the default printer (nil when the system has none).
func printCommand(goos, path string) *exec.Cmd {
	switch goos {
	case "darwin", "linux":
		return exec.Command("lp", path)
	case "windows":
		return exec.Command("powershell", "-NoProfile", "-NonInteractive", "-Command", "Start-Process -FilePath $args[0] -Verb Print -Wait", path)
	}
	return nil
}

// PrintPNG prints a picture (base64 PNG, e.g. the selected part) on the default printer and says where it went.
func (a *App) PrintPNG(b64 string) (string, error) {
	data, err := base64.StdEncoding.DecodeString(b64)
	if err != nil || len(data) == 0 {
		return "", errors.New("nothing to print")
	}
	dir, err := os.MkdirTemp("", "mormiview-print")
	if err != nil {
		return "", err
	}
	path := filepath.Join(dir, "mormiview.png")
	if err := os.WriteFile(path, data, 0o600); err != nil {
		os.RemoveAll(dir)
		return "", err
	}
	cmd := printCommand(runtime.GOOS, path)
	if cmd == nil {
		os.RemoveAll(dir)
		return "", errors.New("printing is not available on this system")
	}
	out, err := cmd.CombinedOutput()
	go func() { time.Sleep(2 * time.Minute); os.RemoveAll(dir) }() // (the spooler may still read the file for a while)
	if err != nil {
		if len(out) > 0 {
			return "", errors.New("printing failed: " + string(out))
		}
		return "", errors.New("printing failed (is a printer set up?)")
	}
	return "Sent to the printer", nil
}
