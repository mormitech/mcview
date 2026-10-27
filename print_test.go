package main

import "testing"

func TestPrintCommand(t *testing.T) {
	if c := printCommand("darwin", "/x.png"); c == nil || c.Args[0] != "lp" {
		t.Fatalf("darwin: %v", c)
	}
	if printCommand("plan9", "/x.png") != nil {
		t.Fatal("unknown system")
	}
}
