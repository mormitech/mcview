package main

import (
	"fmt"
	"os"
	"path/filepath"
)

// SaveToPath saves the picture of this window to path without a dialog (the chat tools' save_image and the control command saveto); rot is how the view is turned.
// A file that exists is only replaced with overwrite.
func (a *App) SaveToPath(path string, rot int, overwrite bool) (string, error) {
	if !filepath.IsAbs(path) {
		return "", fmt.Errorf("%q is not a full path", path)
	}
	if fi, err := os.Stat(path); err == nil && (fi.IsDir() || !overwrite) {
		return "", fmt.Errorf("%s exists already", path)
	}
	if cur := a.current(); cur.URL == "" {
		return "", fmt.Errorf("there is no picture to save")
	}
	return a.saveTo(path, rot)
}
