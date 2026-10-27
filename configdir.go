package main

import (
	"os"
	"path/filepath"
)

// configDir is the folder of the app's settings (recent pictures, window, AI settings) in the user's config folder. The app used to be called mcView and kept
// them in "mcview": the first time the new folder is asked for, the old one is moved there, so nothing is lost by the rename.
func configDir() string {
	dir, err := os.UserConfigDir()
	if err != nil {
		return ""
	}
	cur, old := filepath.Join(dir, "mormiview"), filepath.Join(dir, "mcview")
	if _, err := os.Stat(cur); os.IsNotExist(err) {
		if _, err := os.Stat(old); err == nil {
			_ = os.Rename(old, cur)
		}
	}
	return cur
}
