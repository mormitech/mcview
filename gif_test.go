package main

import (
	"image"
	"image/color"
	"image/gif"
	"os"
	"path/filepath"
	"testing"
)

func TestGifFramesComposeAndDispose(t *testing.T) {
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
	a := &App{image: path}
	a.imgSeq = 1
	meta, err := a.GifInfo()
	if err != nil || meta.Frames != 2 || meta.Delays[0] != 50 || meta.Delays[1] != 100 {
		t.Fatalf("%+v %v", meta, err)
	}
	set, _ := a.gifFrames()
	if c := set.frames[1].NRGBAAt(0, 0); c != (color.NRGBA{255, 0, 0, 255}) {
		t.Fatalf("frame 2 keeps the red of frame 1: %v", c)
	}
	if c := set.frames[1].NRGBAAt(1, 1); c != (color.NRGBA{0, 0, 255, 255}) {
		t.Fatalf("frame 2 has the blue square: %v", c)
	}
	if err := a.SetGifFrame(1); err != nil || a.pix.img == nil {
		t.Fatalf("set frame: %v", err)
	}
	if _, err := a.GifFrame(5); err == nil {
		t.Fatal("a frame that is not there")
	}
}
