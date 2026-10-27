package main

import (
	"os"
	"path/filepath"
	"strings"
)

// configDir is the folder of the app's settings (recent pictures, window, AI settings) in the user's config folder. The app used to be called mcView and kept
// them in "mcview": the first time the new folder is asked for, the old one is moved there, so nothing is lost by the rename.
func configDir() string {
	if demoDir != "" {
		return demoDir
	}
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

// demoDir is the settings folder of a demo run (--demo or --mock), a new empty one in the temporary folder: the recent pictures, the window and the AI
// settings of the person who makes the screenshots neither show in them nor change. "" in every other run.
var demoDir string

// demoFlag says whether the arguments ask for a demo run: control mode (dialogs answered by the program that drives the window) on a clean slate.
func demoFlag(args []string) bool {
	for _, a := range args {
		switch a {
		case "--demo", "-demo", "--mock", "-mock":
			return true
		}
	}
	return false
}

// startDemo makes the settings folder of a demo run.
func startDemo(args []string) {
	if !demoFlag(args) {
		return
	}
	if d, err := os.MkdirTemp("", "mormiview-demo-"); err == nil {
		demoDir = d
	}
}

// underHome says whether the path is in the user's home folder, where the names of files and folders are personal.
func underHome(path string) bool {
	home, err := os.UserHomeDir()
	if err != nil || home == "" {
		return false
	}
	for _, h := range []string{home, evalLinks(home)} {
		for _, p := range []string{path, evalLinks(path)} {
			if rel, err := filepath.Rel(h, p); err == nil && rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
				return true
			}
		}
	}
	return false
}

func evalLinks(p string) string {
	if r, err := filepath.EvalSymlinks(p); err == nil {
		return r
	}
	return p
}
