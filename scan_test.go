package main

import "testing"

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
