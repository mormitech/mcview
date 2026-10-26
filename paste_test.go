package main

import (
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"testing"
)

func solid(w, h int, c color.NRGBA) image.Image {
	img := image.NewNRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			img.SetNRGBA(x, y, c)
		}
	}
	return img
}

var blue = color.NRGBA{B: 255, A: 255}

func TestLayerStretchedOverTheWholePicture(t *testing.T) {
	pic := filepath.Join(t.TempDir(), "pic.png")
	writePNG(t, pic) // 4 × 3, a red pixel at 1,1, the rest transparent
	before, _ := os.ReadFile(pic)
	a := NewApp()
	a.setImage(pic)
	s0 := a.current().State
	// a 2 × 2 layer stretched over the whole 4 × 3 picture: every pixel of it is blue
	modify(a)
	info := a.current()
	if !info.Modified || info.Name != "pic.png" || info.Path != pic || info.State == s0 {
		t.Errorf("after the paste: %+v; want pic.png, modified, a state of its own", info)
	}
	for _, p := range [][2]int{{0, 0}, {1, 1}, {3, 2}} {
		if px := a.PixelAt(p[0], p[1], 4, 3); !px.Ready || px.B != 255 || px.A != 255 || px.R != 0 {
			t.Errorf("pixel %v = %+v, want blue", p, px)
		}
	}
	if after, _ := os.ReadFile(pic); string(after) != string(before) {
		t.Error("the file must not change before it is saved")
	}
	// the modified picture is what the web view gets
	rec := get(a.imageHandler(), info.URL)
	if img, err := png.Decode(rec.Body); err != nil || img.Bounds().Dx() != 4 {
		t.Errorf("the handler serves %v, %v", img, err)
	}
	// and undo brings the file back
	back, err := a.RestoreState(s0)
	if err != nil || back.Modified || a.clip != nil {
		t.Errorf("RestoreState(%d) = %+v, %v; want the picture of the file", s0, back, err)
	}
	if again, err := a.RestoreState(info.State); err != nil || !again.Modified || a.PixelAt(0, 0, 4, 3).B != 255 {
		t.Errorf("redo: %+v, %v", again, err)
	}
}

func TestPasteTurnsTheViewForGood(t *testing.T) {
	a := NewApp()
	a.setImage(gridPNG(t)) // 10 × 6: the pixel x, y has R = x*20, G = y*40
	// shown a quarter turn clockwise it is 6 × 10, and its top left is the bottom left of the file's picture (x 0, y 5)
	if _, err := a.pasteClipboard(solid(1, 1, blue), 1, 5, 9); err != nil {
		t.Fatal(err)
	}
	if err := a.bakeFloat(false); err != nil {
		t.Fatal(err)
	}
	if px := a.PixelAt(0, 0, 6, 10); !px.Ready || px.R != 0 || px.G != 200 {
		t.Errorf("shown 0,0 = %+v, want the file's pixel 0,5 (R 0, G 200): the turn is part of the picture now", px)
	}
	if px := a.PixelAt(5, 9, 6, 10); px.B != 255 || px.R != 0 {
		t.Errorf("the pasted pixel = %+v", px)
	}
}

func TestPasteIsALayerAtTheTopLeft(t *testing.T) {
	a := NewApp()
	a.setImage(gridPNG(t)) // 10 × 6
	asked := 0
	a.confirmLarge = func(w, h, pw, ph int) largeChoice { asked++; return pasteCancel }
	L, err := a.pasteClipboard(solid(2, 2, blue), 0, 3, 2) // small: no question, and it is laid where the UI says
	if err != nil || asked != 0 || L.W != 2 || L.X != 3 || L.Y != 2 || !a.floating() || !a.current().Modified {
		t.Fatalf("a small paste: %+v, %v, asked %d, floating %v", L, err, asked, a.floating())
	}
	if L2, _ := a.pasteClipboard(solid(2, 2, blue), 0, 9, 5); L2.X != 8 || L2.Y != 4 { // moved in so that it fits; the first layer is applied
		t.Errorf("a layer must not stick out: %+v", L2)
	}
	if px := a.PixelAt(3, 2, 10, 6); px.B != 255 || px.R != 0 {
		t.Errorf("a new paste applies the layer that was floating: %+v", px)
	}
}

func TestLargePasteAsksWhatItIs(t *testing.T) {
	pic := filepath.Join(t.TempDir(), "pic.png")
	writePNG(t, pic) // 4 × 3
	for _, c := range []struct {
		answer largeChoice
		newImg bool
		layer  bool
	}{{pasteCancel, false, false}, {pasteNew, true, false}, {pasteOver, false, true}} {
		a := NewApp()
		a.setImage(pic)
		asked := 0
		a.confirmLarge = func(w, h, pw, ph int) largeChoice { asked++; return c.answer }
		a.confirmDiscard = func(name, action string) discardChoice { return choiceDiscard }
		L, err := a.pasteClipboard(solid(3, 2, blue), 0, 0, 0) // 6 pixels of 12: half of the picture
		if err != nil || asked != 1 {
			t.Fatalf("answer %v: %v, asked %d", c.answer, err, asked)
		}
		if got := a.current().Name == clipboardName; got != c.newImg {
			t.Errorf("answer %v: new image %v, want %v", c.answer, got, c.newImg)
		}
		if got := L.W > 0 && a.floating(); got != c.layer {
			t.Errorf("answer %v: layer %v, want %v", c.answer, got, c.layer)
		}
	}
	// nothing loaded: the paste is the picture, nobody is asked
	a := NewApp()
	a.confirmLarge = func(w, h, pw, ph int) largeChoice { t.Error("asked in an empty window"); return pasteCancel }
	if _, err := a.pasteClipboard(solid(3, 2, blue), 0, 0, 0); err != nil || a.current().Name != clipboardName {
		t.Errorf("an empty window: %v, %+v", err, a.current())
	}
	for _, c := range []struct {
		w, h, pw, ph int
		large        bool
	}{{2, 2, 10, 6, false}, {7, 4, 10, 6, false}, {5, 6, 10, 6, true}, {11, 1, 10, 6, true}, {1, 7, 10, 6, true}} {
		if got := isLargePaste(c.w, c.h, c.pw, c.ph); got != c.large {
			t.Errorf("isLargePaste(%d×%d in %d×%d) = %v, want %v", c.w, c.h, c.pw, c.ph, got, c.large)
		}
	}
}

func TestSaveModifiedAndTurnedPictures(t *testing.T) {
	dir := t.TempDir()
	pic := filepath.Join(dir, "pic.png")
	writePNG(t, pic)
	a := NewApp()
	a.setImage(pic)
	modify(a)
	// saving over the file itself is not "this picture already": the modified picture is what gets written
	if msg, err := a.saveTo(pic, 0); err != nil || msg == "pic.png is this picture already" {
		t.Fatalf("saveTo = %q, %v", msg, err)
	}
	f, _ := os.Open(pic)
	img, err := png.Decode(f)
	f.Close()
	if err != nil || color.NRGBAModel.Convert(img.At(0, 0)).(color.NRGBA) != blue {
		t.Errorf("the saved file: %v, %v; want blue", img, err)
	}
	if info := a.current(); info.Modified || info.Path != pic || a.clip != nil {
		t.Errorf("after saving the picture is the file again: %+v", info)
	}

	// a turned picture is saved turned, next to where it came from, and then is the picture
	grid := gridPNG(t)
	b := NewApp()
	b.setImage(grid)
	out := filepath.Join(filepath.Dir(grid), "turned.png")
	if _, err := b.saveTo(out, 1); err != nil {
		t.Fatal(err)
	}
	f, _ = os.Open(out)
	cfg, _, err := image.DecodeConfig(f)
	f.Close()
	if err != nil || cfg.Width != 6 || cfg.Height != 10 {
		t.Errorf("saved turned picture %v, %v; want 6 × 10", cfg, err)
	}
	if info := b.current(); info.Path != out || info.Modified {
		t.Errorf("after saving a turned picture: %+v", info)
	}
}

func TestPastedLayerFloatsUntilApplied(t *testing.T) {
	a := NewApp()
	a.setImage(gridPNG(t)) // 10 × 6, opaque colors with alpha 200+
	if err := a.setPasted(0, solidNRGBA(2, 2, blue), image.Rect(1, 1, 3, 3), true); err != nil {
		t.Fatal(err)
	}
	if !a.current().Modified {
		t.Error("a pasted layer makes the picture modified")
	}
	if px := a.PixelAt(1, 1, 10, 6); px.B == 255 && px.R == 0 {
		t.Error("the picture is not changed before the layer is applied")
	}
	if err := a.MovePasted(4, 2, 2, 2, true); err != nil {
		t.Fatal(err)
	}
	if err := a.bakeFloat(false); err != nil {
		t.Fatal(err)
	}
	if px := a.PixelAt(4, 2, 10, 6); px.B != 255 || px.R != 0 {
		t.Errorf("the layer is where it was moved to: %+v", px)
	}
	if px := a.PixelAt(1, 1, 10, 6); px.R != 20 || px.A == 0 {
		t.Errorf("nothing is left white or transparent where the layer was first laid: %+v", px)
	}
	if err := a.MovePasted(0, 0, 1, 1, true); err == nil {
		t.Error("there is no layer to move any more")
	}
}

func solidNRGBA(w, h int, c color.NRGBA) *image.NRGBA { return solid(w, h, c).(*image.NRGBA) }

// modify makes the picture of a modified one: a blue layer over all of it, applied.
func modify(a *App) {
	base, _ := a.shownPixels(0)
	r := base.Bounds()
	a.setPasted(0, solidNRGBA(2, 2, blue), r, true)
	a.bakeFloat(false)
}
