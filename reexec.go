package main

import "path/filepath"

// resolvedPath is the path of the program with its symbolic links followed, and whether that is a different path. macOS finds the app bundle of a
// program (its identity, which the Open and Save dialogs need) from the path the program was started by: started through a link, such as the
// `mormiview` command that Homebrew puts in /opt/homebrew/bin, it has no bundle, and the Open dialog collapses at once.
func resolvedPath(exe string) (string, bool) {
	real, err := filepath.EvalSymlinks(exe)
	if err != nil || real == exe {
		return exe, false
	}
	return real, true
}
