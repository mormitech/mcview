package main

import (
	"image/color"
	"testing"
)

func TestFloatMoveLeavesWhiteAndKeepsPixels(t *testing.T) {
	a := NewApp()
	a.setImage(gridPNG(t)) // 10 × 6: the pixel x, y is (x*20, y*40, 77) with the transparency 200+x
	s0 := a.current().State
	// lifted but not moved: nothing is modified
	if err := a.SetFloat(0, 2, 1, 3, 2, 2, 1, 3, 2, true); err != nil || a.current().Modified {
		t.Fatalf("a lifted part that was not moved: %v, modified %v", err, a.current().Modified)
	}
	// moved: it is modified at once (closing asks), and the picture itself is not changed yet
	if err := a.SetFloat(0, 2, 1, 3, 2, 6, 3, 3, 2, true); err != nil || !a.current().Modified {
		t.Fatalf("a moved part: %v, modified %v", err, a.current().Modified)
	}
	if px := a.PixelAt(2, 1, 10, 6); px.R != 40 {
		t.Errorf("the picture changes only when applied: %+v", px)
	}
	info, err := a.BakeFloat()
	if err != nil || !info.Modified || info.Name != "grid.png" || info.State == s0 {
		t.Fatalf("BakeFloat = %+v, %v", info, err)
	}
	if a.fl != nil {
		t.Error("the session ends when it is applied")
	}
	if px := a.PixelAt(2, 1, 10, 6); px.A != 0 { // a PNG can be transparent
		t.Errorf("the place it came from must be transparent: %+v", px)
	}
	if px := a.PixelAt(4, 2, 10, 6); px.A != 0 { // the part is x 2 to 4, y 1 to 2
		t.Errorf("4,2 is part of it, so transparent too: %+v", px)
	}
	if px := a.PixelAt(5, 1, 10, 6); px.R != 100 {
		t.Errorf("a pixel beside the lifted part must stay: %+v", px)
	}
	// 2,1 now at 6,3: its color (40, 40, 77)
	if px := a.PixelAt(6, 3, 10, 6); px.R != 40 || px.G != 40 || px.B != 77 {
		t.Errorf("the dropped pixel 6,3 = %+v, want the one that was 2,1", px)
	}
	if px := a.PixelAt(8, 4, 10, 6); px.R != 80 || px.G != 80 {
		t.Errorf("the dropped pixel 8,4 = %+v, want the one that was 4,2 (80, 80)", px)
	}
}

func TestFloatStretchIsWorkedOutFromTheStart(t *testing.T) {
	a := NewApp()
	a.setImage(gridPNG(t))
	// stretched wide and then back to its size: nothing is lost, the part is as it was
	a.SetFloat(0, 0, 0, 2, 2, 0, 0, 8, 6, false)
	a.SetFloat(0, 0, 0, 2, 2, 5, 2, 2, 2, false)
	if _, err := a.BakeFloat(); err != nil {
		t.Fatal(err)
	}
	for _, c := range []struct{ x, y int }{{5, 2}, {6, 3}} {
		px := a.PixelAt(c.x, c.y, 10, 6)
		want := color.NRGBA{R: uint8((c.x - 5) * 20), G: uint8((c.y - 2) * 40), B: 77, A: uint8(200 + c.x - 5)}
		if px.R != int(want.R) || px.G != int(want.G) || px.A != int(want.A) {
			t.Errorf("pixel %d,%d = %+v, want %v", c.x, c.y, px, want)
		}
	}
}

func TestFloatIsAppliedBeforeSaveOrPaste(t *testing.T) {
	a := NewApp()
	a.setImage(gridPNG(t))
	a.SetFloat(0, 0, 0, 2, 2, 7, 3, 2, 2, true)
	if _, err := a.saveTo(t.TempDir()+"/out.png", 0); err != nil {
		t.Fatal(err)
	}
	if a.floating() || a.current().Modified {
		t.Error("saving applies the lifted part and then the file is the picture")
	}
	a.PreparePixels()
	if px := a.PixelAt(0, 0, 10, 6); px.A != 0 {
		t.Errorf("the saved picture has the place it came from transparent: %+v", px)
	}

	b := NewApp()
	b.setImage(gridPNG(t))
	b.SetFloat(0, 0, 0, 2, 2, 7, 3, 2, 2, true)
	if _, err := b.pasteClipboard(solid(1, 1, blue), 0, 4, 4); err != nil {
		t.Fatal(err)
	}
	b.bakeFloat(false)
	if px := b.PixelAt(0, 0, 10, 6); px.A != 0 { // lifted and applied before the paste
		t.Errorf("a paste applies what was lifted: %+v", px)
	}
	if px := b.PixelAt(4, 4, 10, 6); px.B != 255 || px.R != 0 {
		t.Errorf("the pasted pixel: %+v", px)
	}
}

func TestFloatEndsWithAnotherPicture(t *testing.T) {
	a := NewApp()
	a.setImage(gridPNG(t))
	a.SetFloat(0, 0, 0, 2, 2, 7, 3, 2, 2, true)
	a.setImage(gridPNG(t))
	if a.fl != nil || a.current().Modified {
		t.Error("a new picture ends the session without applying it")
	}
}
