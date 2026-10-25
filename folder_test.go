package main

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func touch(t *testing.T, dir string, names ...string) {
	t.Helper()
	for _, n := range names {
		if err := os.WriteFile(filepath.Join(dir, n), []byte("x"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
}

func TestListPictures(t *testing.T) {
	dir := t.TempDir()
	touch(t, dir, "b.jpg", "A.png", "c.GIF", "a.txt", ".hidden.jpg", "._b.jpg", "notes", "D.tiff")
	os.Mkdir(filepath.Join(dir, "sub.png"), 0o755)
	got, err := listPictures(dir)
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"A.png", "b.jpg", "c.GIF", "D.tiff"} // alphabetical, upper and lower case alike; no folders, dotfiles, other types
	if !reflect.DeepEqual(got, want) {
		t.Errorf("listPictures = %v, want %v", got, want)
	}
}

func TestStep(t *testing.T) {
	dir := t.TempDir()
	touch(t, dir, "b.png", "a.png", "c.png", "readme.txt")
	// another folder's pictures must never show up
	other := t.TempDir()
	touch(t, other, "z.png")

	a := NewApp()
	if note, err := a.Step(1); note != "" || err != nil {
		t.Errorf("Step with no picture = %q, %v; want nothing", note, err)
	}
	a.setImage(filepath.Join(dir, "a.png"))
	if info := a.Image(); info.Dir != dir || info.Index != 1 || info.Count != 3 {
		t.Errorf("Image() = %+v, want folder %s, 1 of 3", info, dir)
	}
	// it goes round: before the first picture is the last, after the last the first
	if note, err := a.Step(-1); note != "" || err != nil || a.current().Name != "c.png" || a.Image().Index != 3 {
		t.Errorf("before the first: note %q, err %v, picture %q; want c.png, 3 of 3", note, err, a.current().Name)
	}
	if note, err := a.Step(1); note != "" || err != nil || a.current().Name != "a.png" || a.Image().Index != 1 {
		t.Errorf("after the last: note %q, err %v, picture %q; want a.png, 1 of 3", note, err, a.current().Name)
	}
	for _, want := range []string{"b.png", "c.png"} {
		if note, err := a.Step(1); note != "" || err != nil || a.current().Name != want {
			t.Fatalf("Step(1): note %q, err %v, picture %q; want %s", note, err, a.current().Name, want)
		}
	}
	a.Step(-1)
	if a.current().Name != "b.png" || a.Image().Index != 2 {
		t.Errorf("Step(-1) from c: picture %q, index %d; want b.png, 2", a.current().Name, a.Image().Index)
	}
	// a folder with one picture stays where it is
	solo := t.TempDir()
	touch(t, solo, "only.png")
	b := NewApp()
	b.setImage(filepath.Join(solo, "only.png"))
	seq := b.imgSeq
	if _, err := b.Step(1); err != nil || b.imgSeq != seq {
		t.Errorf("a folder with one picture must not reload it: err %v, seq %d -> %d", err, seq, b.imgSeq)
	}
	// the picture was deleted behind our back: Step starts again at the first one
	os.Remove(filepath.Join(dir, "b.png"))
	a.Step(1)
	if a.current().Name != "a.png" {
		t.Errorf("after the picture disappeared Step shows %q, want a.png", a.current().Name)
	}
}
