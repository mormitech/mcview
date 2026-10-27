package main

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestScanCommand(t *testing.T) {
	if scanCommand("darwin", "/x.png") != nil {
		t.Fatal("macOS has no scan command")
	}
	if c := scanCommand("linux", "/x.png"); c == nil || c.Args[len(c.Args)-1] != "/x.png" {
		t.Fatalf("linux: %v", c)
	}
	if c := scanCommand("windows", "C:\\it's.png"); c == nil || c.Path == "" {
		t.Fatal("windows")
	}
}

func TestWaitForScan(t *testing.T) {
	dir := t.TempDir()
	old := filepath.Join(dir, "old.png")
	os.WriteFile(old, []byte("x"), 0o644)
	os.Chtimes(old, time.Now().Add(-time.Hour), time.Now().Add(-time.Hour))
	since := time.Now().Add(-time.Minute)
	if got := waitForScan([]string{dir}, since, 30*time.Millisecond, 5*time.Millisecond); got != "" {
		t.Fatalf("an old file was taken: %s", got)
	}
	want := filepath.Join(dir, "scan.JPG")
	os.WriteFile(want, []byte("data"), 0o644)
	os.WriteFile(filepath.Join(dir, "note.txt"), []byte("data"), 0o644)
	if got := waitForScan([]string{dir}, since, time.Second, 5*time.Millisecond); got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}
