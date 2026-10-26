package main

import (
	"bytes"
	"encoding/base64"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"testing"
)

func TestDiscardChangesPastedPicture(t *testing.T) {
	a := NewApp()
	if err := a.setClip(solid(6, 4, blue)); err != nil {
		t.Fatal(err)
	}
	st0 := a.current().State
	// nothing changed: nothing to discard, and no new step
	if info, err := a.DiscardChanges(); err != nil || info.State != st0 {
		t.Fatalf("nothing to discard: state %d (was %d), %v", info.State, st0, err)
	}
	if _, err := a.PaintPixels(0, `[[0,0,"#FF0000FF"]]`); err != nil {
		t.Fatal(err)
	}
	st1 := a.current().State
	info, err := a.DiscardChanges()
	if err != nil || info.State == st1 || info.State == st0 || info.Modified {
		t.Fatalf("discard: state %d (was %d, %d), modified %v, %v", info.State, st1, st0, info.Modified, err)
	}
	if p := pixelOf(a, 0, 0); p != blue {
		t.Errorf("the pixel is %v after discard, want the pasted color", p)
	}
	// undo brings the change back, redo discards it again (the page does that with RestoreState)
	if _, err := a.RestoreState(st1); err != nil {
		t.Fatal(err)
	}
	if p := pixelOf(a, 0, 0); p != (color.NRGBA{255, 0, 0, 255}) || !a.current().Modified {
		t.Errorf("undo of the discard: %v, modified %v", p, a.current().Modified)
	}
	if _, err := a.RestoreState(info.State); err != nil {
		t.Fatal(err)
	}
	if p := pixelOf(a, 0, 0); p != blue || a.current().Modified {
		t.Errorf("redo of the discard: %v, modified %v", p, a.current().Modified)
	}
	// a new picture whose size was changed goes back to its size too
	b := NewApp()
	b.NewBlankColor(10, 8, "#FFFFFFFF")
	b.ResizeCanvas(0, 30, 30, "", "")
	if info, err := b.DiscardChanges(); err != nil || info.Modified {
		t.Fatalf("discard of a resize: %+v %v", info, err)
	}
	if bw := func() int { b.mu.Lock(); defer b.mu.Unlock(); return b.clip.Bounds().Dx() }(); bw != 10 {
		t.Errorf("the width is %d after discard, want 10", bw)
	}
	if _, err := NewApp().DiscardChanges(); err == nil {
		t.Errorf("an empty window has nothing to discard")
	}
}

func TestDiscardChangesFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "a.png")
	img := image.NewNRGBA(image.Rect(0, 0, 5, 5))
	for i := range img.Pix {
		img.Pix[i] = 255
	}
	var buf bytes.Buffer
	png.Encode(&buf, img)
	os.WriteFile(path, buf.Bytes(), 0o644)
	a := NewApp()
	if err := a.setImage(path); err != nil {
		t.Fatal(err)
	}
	if a.current().Modified {
		t.Fatal("a picture just opened is not modified")
	}
	if _, err := a.PaintPixels(0, `[[1,1,"#00FF00FF"]]`); err != nil {
		t.Fatal(err)
	}
	if !a.current().Modified {
		t.Fatal("a painted picture is modified")
	}
	info, err := a.DiscardChanges()
	if err != nil || info.Modified || info.Path != path {
		t.Fatalf("discard: %+v %v", info, err)
	}
	a.mu.Lock()
	clipNil := a.clip == nil
	a.mu.Unlock()
	if !clipNil {
		t.Errorf("the picture is the file again, not a copy in memory")
	}
}

func TestPolygonPNG(t *testing.T) {
	red := color.NRGBA{255, 0, 0, 255}
	a := testPicture(t, 20, 20, func(x, y int) color.NRGBA { return red })
	b64, err := a.PolygonPNG(0, `[[2,2],[10,2],[2,10]]`)
	if err != nil {
		t.Fatal(err)
	}
	raw, _ := base64.StdEncoding.DecodeString(b64)
	im, err := png.Decode(bytes.NewReader(raw))
	if err != nil {
		t.Fatal(err)
	}
	if im.Bounds().Dx() != 9 || im.Bounds().Dy() != 9 {
		t.Errorf("the box of the triangle is %v", im.Bounds())
	}
	if _, _, _, al := im.At(1, 1).RGBA(); al == 0 {
		t.Errorf("a pixel inside the triangle is transparent")
	}
	if _, _, _, al := im.At(8, 8).RGBA(); al != 0 {
		t.Errorf("a pixel outside the triangle is not transparent")
	}
	if _, err := a.PolygonPNG(0, `[[2,2],[40,2],[2,10]]`); err == nil {
		t.Errorf("a polygon that sticks out must fail")
	}
}
