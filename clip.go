package main

import (
	"bytes"
	"errors"
	"fmt"
	"image"
	"image/draw"
	"image/png"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

// clipboardName is what the title of a window says instead of the name of a file, for a picture pasted from the clipboard.
const clipboardName = "[clipboard image]"

// A picture pasted from the clipboard has no file: it lives in this window (App.clip) until it is saved with Save As. Its "folder" is
// the Pictures folder of the user, and the folder Save As starts in; when it is saved somewhere else, that folder becomes the one the
// window works in (see saveTo).

// workDirLocked is the folder of a picture without a file, App.mu held.
func (a *App) workDirLocked() string {
	if a.workDir == "" {
		a.workDir = picturesDir()
	}
	return a.workDir
}

// picturesDir is the Pictures folder of the user (the "Images" of the system), or their home folder when there is none.
func picturesDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	if runtime.GOOS == "linux" { // it may have another name or place: ~/.config/user-dirs.dirs says
		cfg := os.Getenv("XDG_CONFIG_HOME")
		if cfg == "" {
			cfg = filepath.Join(home, ".config")
		}
		if b, err := os.ReadFile(filepath.Join(cfg, "user-dirs.dirs")); err == nil {
			for _, line := range strings.Split(string(b), "\n") {
				if v, ok := strings.CutPrefix(strings.TrimSpace(line), "XDG_PICTURES_DIR="); ok {
					d := strings.ReplaceAll(strings.Trim(v, `"`), "$HOME", home)
					if fi, err := os.Stat(d); err == nil && fi.IsDir() {
						return d
					}
				}
			}
		}
	}
	if d := filepath.Join(home, "Pictures"); isDir(d) {
		return d
	}
	return home
}

func isDir(path string) bool {
	fi, err := os.Stat(path)
	return err == nil && fi.IsDir()
}

// setClip makes a picture without a file the picture of this window.
func (a *App) setClip(img image.Image) error {
	b := img.Bounds()
	if b.Dx() < 1 || b.Dy() < 1 || b.Dx()*b.Dy() > maxPixels {
		return fmt.Errorf("the picture on the clipboard is %d × %d: too small or too big", b.Dx(), b.Dy())
	}
	n := image.NewNRGBA(image.Rect(0, 0, b.Dx(), b.Dy())) // the colours as they are, not premultiplied
	draw.Draw(n, n.Bounds(), img, b.Min, draw.Src)
	var buf bytes.Buffer
	if err := (&png.Encoder{CompressionLevel: png.BestSpeed}).Encode(&buf, n); err != nil {
		return err
	}
	a.mu.Lock()
	a.image = ""
	a.clip, a.clipPNG = n, buf.Bytes()
	a.imgSeq++
	a.pix = pixelCache{seq: a.imgSeq, img: n, orient: 1} // it is decoded already: the readout and Ctrl+C read from it
	a.workDirLocked()
	a.newStateLocked()
	a.mu.Unlock()
	if a.ctx != nil {
		a.titleWindow()
		a.emit("image", a.Image())
	}
	return nil
}

// PasteImage shows the picture that is on the system's clipboard in this window (Ctrl+V, Cmd+V, Edit → Paste).
func (a *App) PasteImage() error {
	a.mu.Lock()
	if time.Since(a.lastPaste) < 400*time.Millisecond { // the menu and the key can both come for one press
		a.mu.Unlock()
		return nil
	}
	a.lastPaste = time.Now()
	a.mu.Unlock()
	img, err := readClipboardImage()
	if err != nil {
		return err
	}
	return a.setClip(img)
}

// pasteImage is Edit → Paste: the same as PasteImage, with the error shown by the UI.
func (a *App) pasteImage() {
	if err := a.PasteImage(); err != nil {
		a.emit("problem", err.Error())
	}
}

var errNoClipboardPicture = errors.New("there is no picture on the clipboard")

// readClipboardImage asks the system for the picture on its clipboard.
func readClipboardImage() (image.Image, error) {
	f, err := os.CreateTemp("", "mcview-*.img")
	if err != nil {
		return nil, err
	}
	path := f.Name()
	f.Close()
	defer os.Remove(path)

	var cmd *exec.Cmd
	toFile := true // the helper writes the file itself; otherwise its output is the picture
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("osascript",
			"-e", "try", "-e", "set d to (the clipboard as «class PNGf»)", "-e", "on error", "-e", "set d to (the clipboard as «class TIFF»)", "-e", "end try",
			"-e", `set f to open for access POSIX file "`+filepath.ToSlash(path)+`" with write permission`,
			"-e", "set eof f to 0", "-e", "write d to f", "-e", "close access f")
	case "windows":
		cmd = exec.Command("powershell", "-NoProfile", "-NonInteractive", "-STA", "-Command",
			`Add-Type -AssemblyName System.Windows.Forms,System.Drawing; $i = [System.Windows.Forms.Clipboard]::GetImage(); if ($i -eq $null) { exit 3 }; $i.Save('`+filepath.ToSlash(path)+`', [System.Drawing.Imaging.ImageFormat]::Png)`)
	default:
		toFile = false
		if p, err := exec.LookPath("wl-paste"); err == nil && os.Getenv("WAYLAND_DISPLAY") != "" {
			cmd = exec.Command(p, "--type", "image/png")
		} else if p, err := exec.LookPath("xclip"); err == nil {
			cmd = exec.Command(p, "-selection", "clipboard", "-t", "image/png", "-o")
		} else {
			return nil, errors.New("to paste a picture install xclip (X11) or wl-clipboard (Wayland)")
		}
	}
	hideWindow(cmd)
	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		return nil, errNoClipboardPicture
	}
	data := out.Bytes()
	if toFile {
		if data, err = os.ReadFile(path); err != nil {
			return nil, err
		}
	}
	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, errNoClipboardPicture
	}
	return img, nil
}
