package main

import (
	"testing"
	"time"
)

func TestChangedName(t *testing.T) {
	at := time.Date(2026, 10, 26, 15, 30, 12, 0, time.UTC)
	for in, want := range map[string]string{
		"photo.jpg": "photo-20261026-153012.jpg",
		"a.b.PNG":   "a.b-20261026-153012.PNG",
		"noext":     "noext-20261026-153012",
	} {
		if got := changedName(in, at); got != want {
			t.Errorf("changedName(%q) = %q, want %q", in, got, want)
		}
	}
}
