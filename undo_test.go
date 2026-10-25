package main

import (
	"image"
	"path/filepath"
	"testing"
)

func TestRestoreState(t *testing.T) {
	dir := t.TempDir()
	one, two := filepath.Join(dir, "one.png"), filepath.Join(dir, "two.png")
	writePNG(t, one)
	writePNG(t, two)
	a := NewApp()
	empty := a.current().State
	if empty != 0 {
		t.Fatalf("an empty window is state 0, got %d", empty)
	}
	a.setImage(one)
	s1 := a.current().State
	a.setImage(two)
	s2 := a.current().State
	if s1 == 0 || s2 == s1 {
		t.Fatalf("every picture gets a state of its own: %d, %d", s1, s2)
	}
	a.setClip(image.NewNRGBA(image.Rect(0, 0, 6, 4)))
	s3 := a.current().State
	if a.current().Name != clipboardName {
		t.Fatal("the pasted picture is the current one")
	}

	// back to the first file, and the URL is new, or the web view would show the one it has
	before := a.current().URL
	info, err := a.RestoreState(s1)
	if err != nil || info.Path != one || info.State != s1 || info.URL == before || a.current().Path != one {
		t.Errorf("RestoreState(%d) = %+v, %v; want %s", s1, info, err, one)
	}
	// the pasted picture comes back with its pixels, and is decoded for the readout already
	info, err = a.RestoreState(s3)
	if err != nil || info.Name != clipboardName || a.clip == nil || a.clip.Bounds().Dx() != 6 || a.pix.img == nil {
		t.Errorf("RestoreState(%d) = %+v, %v", s3, info, err)
	}
	// and no picture at all
	info, err = a.RestoreState(0)
	if err != nil || info.URL != "" || info.Name != "" || a.clip != nil || a.image != "" {
		t.Errorf("RestoreState(0) = %+v, %v; want the empty window", info, err)
	}
	if _, err := a.RestoreState(999); err == nil {
		t.Error("an unknown state must be an error")
	}

	// forgetting: only what the history lists stays (and the current picture, and the empty window)
	a.RestoreState(s2)
	a.KeepStates([]int{s1})
	for n, want := range map[int]bool{0: true, s1: true, s2: true, s3: false} {
		if _, err := a.RestoreState(n); (err == nil) != want {
			t.Errorf("after KeepStates, state %d restorable = %v, want %v", n, err == nil, want)
		}
		a.RestoreState(s2)
	}
}
