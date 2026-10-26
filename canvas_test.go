package main

import (
	"image"
	"image/color"
	"testing"
)

func TestResizeCanvasAndCrop(t *testing.T) {
	a := NewApp()
	if err := a.NewBlankColor(4, 3, "#FF0000FF"); err != nil { // red, no file: transparency is possible
		t.Fatal(err)
	}
	at := func(x, y int) color.NRGBA {
		a.mu.Lock()
		defer a.mu.Unlock()
		return a.clip.(*image.NRGBA).NRGBAAt(x, y)
	}
	size := func() (int, int) {
		a.mu.Lock()
		defer a.mu.Unlock()
		b := a.clip.(*image.NRGBA).Bounds()
		return b.Dx(), b.Dy()
	}
	red, clear := color.NRGBA{255, 0, 0, 255}, color.NRGBA{}
	// wider and taller, the default anchor (bottom-left): the new room is to the right and above
	if _, err := a.ResizeCanvas(0, 8, 5, "", ""); err != nil {
		t.Fatal(err)
	}
	if w, h := size(); w != 8 || h != 5 {
		t.Fatalf("the canvas is %d × %d", w, h)
	}
	if at(0, 2) != red || at(3, 4) != red || at(4, 4) != clear || at(0, 1) != clear {
		t.Errorf("bottom-left: kept %v %v, new room %v %v", at(0, 2), at(3, 4), at(4, 4), at(0, 1))
	}
	// top-right with a color: the picture (8 × 5, red at the bottom left) moves to the top right of a 10 × 6 canvas
	if _, err := a.ResizeCanvas(0, 10, 6, "top-right", "#0000FFFF"); err != nil {
		t.Fatal(err)
	}
	blue := color.NRGBA{0, 0, 255, 255}
	if at(0, 0) != blue || at(0, 5) != blue || at(2, 0) != clear || at(2, 2) != red {
		t.Errorf("top-right: new room %v %v, picture %v %v", at(0, 0), at(0, 5), at(2, 0), at(2, 2))
	}
	if _, err := a.ResizeCanvas(0, 10, 6, "sideways", ""); err == nil {
		t.Errorf("an anchor that is not one must fail")
	}
	if _, err := a.ResizeCanvas(0, 0, 6, "", ""); err == nil {
		t.Errorf("a canvas of no width must fail")
	}
	// undo goes back through the history (the states are there): the crop makes a new, smaller picture
	if _, err := a.CropPicture(0, 2, 1, 4, 3); err != nil {
		t.Fatal(err)
	}
	if w, h := size(); w != 4 || h != 3 {
		t.Errorf("the crop is %d × %d", w, h)
	}
	if _, err := a.CropPicture(0, 3, 1, 4, 3); err == nil {
		t.Errorf("a part that sticks out of the picture must fail")
	}
}

func TestCanvasOffset(t *testing.T) {
	for _, c := range []struct {
		anchor       string
		w, h, W, H   int
		wantX, wantY int
	}{
		{"", 30, 30, 60, 30, 0, 0},
		{"", 30, 30, 60, 60, 0, 30},
		{"top-left", 30, 30, 60, 60, 0, 0},
		{"center", 30, 30, 60, 60, 15, 15},
		{"bottom-right", 30, 30, 60, 60, 30, 30},
		{"top", 30, 30, 60, 60, 15, 0},
		{"top-left", 30, 30, 10, 10, 0, 0},
		{"bottom-right", 30, 30, 10, 10, -20, -20},
	} {
		x, y, err := canvasOffset(c.anchor, c.w, c.h, c.W, c.H)
		if err != nil || x != c.wantX || y != c.wantY {
			t.Errorf("%q %dx%d in %dx%d: %d,%d %v, want %d,%d", c.anchor, c.w, c.h, c.W, c.H, x, y, err, c.wantX, c.wantY)
		}
	}
}
