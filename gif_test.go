package main

import (
	"bytes"
	"image"
	"image/color"
	"image/gif"
	"os"
	"path/filepath"
	"testing"
)

func writeTestGif(t *testing.T) string {
	pal := color.Palette{color.Transparent, color.RGBA{255, 0, 0, 255}, color.RGBA{0, 0, 255, 255}}
	f1 := image.NewPaletted(image.Rect(0, 0, 4, 4), pal)
	for i := range f1.Pix {
		f1.Pix[i] = 1 // all red
	}
	f2 := image.NewPaletted(image.Rect(1, 1, 3, 3), pal)
	for i := range f2.Pix {
		f2.Pix[i] = 2 // a blue square
	}
	g := &gif.GIF{Image: []*image.Paletted{f1, f2}, Delay: []int{5, 0}, Disposal: []byte{gif.DisposalNone, gif.DisposalNone}, Config: image.Config{ColorModel: pal, Width: 4, Height: 4}}
	path := filepath.Join(t.TempDir(), "a.gif")
	out, _ := os.Create(path)
	if err := gif.EncodeAll(out, g); err != nil {
		t.Fatal(err)
	}
	out.Close()
	return path
}

func TestGifFramesComposeAndEdit(t *testing.T) {
	path := writeTestGif(t)
	a := &App{image: path}
	a.imgSeq = 1
	meta, err := a.GifInfo()
	if err != nil || meta.Frames != 2 || meta.Delays[0] != 50 || meta.Delays[1] != 100 || !meta.Fresh || meta.Edited {
		t.Fatalf("%+v %v", meta, err)
	}
	if c := a.gifs.frames[1].NRGBAAt(0, 0); c != (color.NRGBA{255, 0, 0, 255}) {
		t.Fatalf("frame 2 keeps the red of frame 1: %v", c)
	}
	if c := a.gifs.frames[1].NRGBAAt(1, 1); c != (color.NRGBA{0, 0, 255, 255}) {
		t.Fatalf("frame 2 has the blue square: %v", c)
	}
	if err := a.SetGifFrame(1); err != nil || a.pix.img == nil {
		t.Fatalf("set frame: %v", err)
	}
	// an edit goes into the frame that is shown, and undo brings the frame back
	a.mu.Lock()
	a.newStateLocked()
	before := a.stateID
	a.mu.Unlock()
	edited := image.NewNRGBA(image.Rect(0, 0, 4, 4))
	edited.SetNRGBA(0, 0, color.NRGBA{0, 255, 0, 255})
	a.mu.Lock()
	a.gifNoteEditLocked(edited, path)
	a.newStateLocked()
	a.mu.Unlock()
	meta, _ = a.GifInfo()
	if !meta.Edited || a.gifs.frames[1] != edited || a.gifs.frames[0] != a.gifs.orig[0] {
		t.Fatalf("the edit is in frame 2 only: %+v", meta)
	}
	if _, err := a.RestoreState(before); err != nil {
		t.Fatal(err)
	}
	if a.gifs.frames[1] != a.gifs.orig[1] {
		t.Fatal("undo must bring the frame back")
	}
	// a picture of another size ends the session
	a.mu.Lock()
	a.gifNoteEditLocked(image.NewNRGBA(image.Rect(0, 0, 2, 2)), path)
	a.mu.Unlock()
	if a.gifs != nil {
		t.Fatal("a resize ends the session")
	}
}

func TestGifSaveAnimationAndFrame(t *testing.T) {
	path := writeTestGif(t)
	a := &App{image: path}
	a.imgSeq = 1
	if _, err := a.GifInfo(); err != nil {
		t.Fatal(err)
	}
	a.confirmGif = func(int, int) gifSaveChoice { return gifSaveAll }
	var buf bytes.Buffer
	if err := encodeGifAnimation(&buf, a.gifs.frames, a.gifs.delays); err != nil {
		t.Fatal(err)
	}
	g, err := gif.DecodeAll(&buf)
	if err != nil || len(g.Image) != 2 {
		t.Fatalf("the animation has 2 frames: %v %v", len(g.Image), err)
	}
	out := filepath.Join(t.TempDir(), "frame.png")
	f, _ := os.Create(out)
	a.gifMode = gifSaveFrame
	if err := a.writePicture(f, a.current(), "png", 0); err != nil {
		t.Fatal(err)
	}
	f.Close()
	fi, err := os.Stat(out)
	if err != nil || fi.Size() == 0 {
		t.Fatal("the frame was not written")
	}
}

func TestAnimationFromAPicture(t *testing.T) {
	a := &App{}
	img := image.NewNRGBA(image.Rect(0, 0, 8, 8))
	img.SetNRGBA(1, 1, color.NRGBA{255, 0, 0, 255})
	if err := a.setPixelsEmit(img, "", false); err != nil {
		t.Fatal(err)
	}
	if _, err := a.GifStart(0); err != nil {
		t.Fatal(err)
	}
	if _, err := a.GifDuplicateFrame(0, 0, 1); err != nil { // a copy of frame 1 as frame 2
		t.Fatal(err)
	}
	if _, err := a.GifAddFrame(0, true, -2); err != nil { // an empty one at the end
		t.Fatal(err)
	}
	meta, _ := a.GifInfo()
	if meta.Frames != 3 || meta.Cur != 2 || !meta.Edited {
		t.Fatalf("%+v", meta)
	}
	if a.gifs.frames[1].NRGBAAt(1, 1).R != 255 || a.gifs.frames[2].NRGBAAt(1, 1).A != 0 {
		t.Fatal("frame 2 is a copy, frame 3 is empty")
	}
	if _, err := a.GifMoveFrame(2, 0); err != nil {
		t.Fatal(err)
	}
	if a.gifs.frames[0].NRGBAAt(1, 1).A != 0 || a.gifs.cur != 0 {
		t.Fatal("the empty frame is first now")
	}
	if _, err := a.GifSetDelay(-1, 250); err != nil || a.gifs.delays[2] != 250 {
		t.Fatalf("delay: %v %v", err, a.gifs.delays)
	}
	if _, err := a.GifDeleteFrame(0); err != nil || len(a.gifs.frames) != 2 {
		t.Fatalf("delete: %v", err)
	}
	// every step has its frames: undo goes back through them
	for n := len(a.gifs.frames); n < 5; n++ {
		a.mu.Lock()
		st := a.stateID
		a.mu.Unlock()
		if _, err := a.RestoreState(st - 1); err != nil {
			t.Fatal(err)
		}
		break
	}
	if len(a.gifs.frames) != 3 {
		t.Fatalf("undo of the delete brings the frame back: %d", len(a.gifs.frames))
	}
	dir := t.TempDir()
	if _, err := a.ExportFrames(dir, "f"); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(dir, "f-003.png")); err != nil {
		t.Fatal("three frames are exported")
	}
	sheet, err := a.GifSheet(512)
	if err != nil || sheet["frames"] != 3 {
		t.Fatalf("sheet: %v %v", sheet, err)
	}
	var buf bytes.Buffer
	if err := encodeGifAnimation(&buf, a.gifs.frames, a.gifs.delays); err != nil {
		t.Fatal(err)
	}
	if g, err := gif.DecodeAll(&buf); err != nil || len(g.Image) != 3 {
		t.Fatalf("the animation: %v", err)
	}
}
