package main

import "testing"

func TestApp(t *testing.T) {
	a := NewApp()
	if a.Version() != version() {
		t.Errorf("Version() = %q, want %q", a.Version(), version())
	}
	if a.Mode() != "" {
		t.Errorf("Mode() = %q, want the main window", a.Mode())
	}
	a.mode = "about"
	if a.Mode() != "about" {
		t.Errorf("Mode() = %q, want about", a.Mode())
	}
}
