package main

import (
	"bytes"
	"errors"
	"image"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// scanCommand is the program that asks a scanner for a picture and leaves it as a PNG file at path (nil when the system has no such thing).
func scanCommand(goos, path string) *exec.Cmd {
	switch goos {
	case "windows":
		script := `$d = New-Object -ComObject WIA.CommonDialog; ` +
			`$i = $d.ShowAcquireImage(1, 1, 65536, '{B96B3CAF-0728-11D3-9D7B-0000F81EF32E}', $false, $true, $false); ` +
			`if ($i -ne $null) { $i.SaveFile('` + strings.ReplaceAll(path, "'", "''") + `') }`
		return exec.Command("powershell", "-NoProfile", "-NonInteractive", "-Command", script)
	case "linux":
		return exec.Command("sh", "-c", `scanimage --format=png > "$1"`, "sh", path)
	}
	return nil
}

// ScanImage asks a scanner for a picture (Windows: the system scan dialog, Linux: SANE's scanimage) and opens it as a new, unsaved picture.
// macOS has no scan command, so Image Capture is opened and the answer says what to do there.
func (a *App) ScanImage() error {
	if runtime.GOOS == "darwin" {
		_ = exec.Command("open", "-a", "Image Capture").Start()
		return errors.New("macOS scans with Image Capture (it is open now): scan, save the picture, then open the file here")
	}
	dir, err := os.MkdirTemp("", "mcview-scan")
	if err != nil {
		return err
	}
	defer os.RemoveAll(dir)
	path := filepath.Join(dir, "scan.png")
	cmd := scanCommand(runtime.GOOS, path)
	if cmd == nil {
		return errors.New("scanning is not available on this system")
	}
	var out bytes.Buffer
	cmd.Stderr = &out
	if err := cmd.Run(); err != nil {
		if msg := strings.TrimSpace(out.String()); msg != "" {
			return errors.New("the scan failed: " + msg)
		}
		return errors.New("the scan failed (is a scanner connected and, on Linux, is SANE installed?)")
	}
	f, err := os.Open(path)
	if err != nil {
		return errors.New("nothing was scanned")
	}
	defer f.Close()
	img, _, err := image.Decode(f)
	if err != nil {
		return errors.New("nothing was scanned")
	}
	return a.setClip(img)
}
