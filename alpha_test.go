package main

import (
	"bytes"
	"image"
	"image/color"
	"image/gif"
	"image/jpeg"
	"image/png"
	"os"
	"path/filepath"
	"testing"
)

// a 4 × 2 picture whose left column is transparent
func writeHoleyPNG(t *testing.T, path string) {
	t.Helper()
	img := image.NewNRGBA(image.Rect(0, 0, 4, 2))
	for y := 0; y < 2; y++ {
		for x := 0; x < 4; x++ {
			if x > 0 {
				img.SetNRGBA(x, y, color.NRGBA{R: 200, G: 50, B: 50, A: 255})
			}
		}
	}
	var b bytes.Buffer
	png.Encode(&b, img)
	os.WriteFile(path, b.Bytes(), 0o644)
}

func TestHasTransparency(t *testing.T) {
	dir := t.TempDir()
	holey, opaque := filepath.Join(dir, "holey.png"), filepath.Join(dir, "opaque.jpg")
	writeHoleyPNG(t, holey)
	var b bytes.Buffer
	jpeg.Encode(&b, solid(4, 4, blue), nil)
	os.WriteFile(opaque, b.Bytes(), 0o644)
	a := NewApp()
	if a.HasTransparency() {
		t.Error("no picture, no transparency")
	}
	a.setImage(holey)
	if !a.HasTransparency() {
		t.Error("a PNG with a transparent column has transparency")
	}
	a.setImage(opaque)
	if a.HasTransparency() {
		t.Error("a JPEG has none")
	}
	a.setImage(pngOf(t, solid(3, 3, blue)))
	if a.HasTransparency() {
		t.Error("an opaque PNG has none: no squares behind it")
	}
	a.setClip(image.NewNRGBA(image.Rect(0, 0, 2, 2))) // all transparent
	if !a.HasTransparency() {
		t.Error("a transparent picture from the clipboard has transparency")
	}
}

func pngOf(t *testing.T, img image.Image) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "p.png")
	var b bytes.Buffer
	png.Encode(&b, img)
	os.WriteFile(path, b.Bytes(), 0o644)
	return path
}

func TestLiftedPartLeavesTransparencyWhereTheTypeCanHaveIt(t *testing.T) {
	// a PNG that is opaque stays opaque: white
	o := NewApp()
	o.setImage(pngOf(t, solid(6, 4, blue)))
	o.SetFloat(0, 0, 0, 2, 2, 4, 2, 2, 2, true)
	o.BakeFloat()
	if px := o.PixelAt(0, 0, 6, 4); px.A != 255 || px.R != 255 || o.HasTransparency() {
		t.Errorf("the place a part of an opaque PNG came from must be white: %+v", px)
	}
	// a PNG that has transparency: transparent
	a := NewApp()
	img := solid(6, 4, blue).(*image.NRGBA)
	img.SetNRGBA(5, 3, color.NRGBA{A: 0})
	a.setImage(pngOf(t, img))
	a.SetFloat(0, 0, 0, 2, 2, 4, 2, 2, 2, true)
	a.BakeFloat()
	if px := a.PixelAt(0, 0, 6, 4); px.A != 0 {
		t.Errorf("the place a part of a PNG with transparency came from must be transparent: %+v", px)
	}
	if !a.HasTransparency() || a.current().Name != "p.png" {
		t.Error("the picture has transparency, and keeps its name")
	}
	// a JPEG: white
	dir := t.TempDir()
	jp := filepath.Join(dir, "j.jpg")
	var b bytes.Buffer
	jpeg.Encode(&b, solid(6, 4, blue), nil)
	os.WriteFile(jp, b.Bytes(), 0o644)
	c := NewApp()
	c.setImage(jp)
	c.SetFloat(0, 0, 0, 2, 2, 4, 2, 2, 2, true)
	c.BakeFloat()
	if px := c.PixelAt(0, 0, 6, 4); px.A != 255 || px.R < 250 || px.G < 250 || px.B < 250 {
		t.Errorf("the place a JPEG's part came from must be white: %+v", px)
	}
	if info := c.current(); info.CanAlpha {
		t.Errorf("a JPEG cannot have transparency: %+v", info)
	}
}

func TestSaveKeepsTransparency(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "holey.png")
	writeHoleyPNG(t, src)
	a := NewApp()
	a.setImage(src)

	// PNG keeps it
	out := filepath.Join(dir, "copy.png")
	if _, err := a.saveTo(out, 1); err != nil { // turned, so that it is written again, not copied
		t.Fatal(err)
	}
	f, _ := os.Open(out)
	img, err := png.Decode(f)
	f.Close()
	if err != nil {
		t.Fatal(err)
	}
	if got := color.NRGBAModel.Convert(img.At(0, 0)).(color.NRGBA); got.A != 0 {
		t.Errorf("turned PNG: the top left pixel was in the transparent column, got %+v", got)
	}

	// GIF keeps it too (a transparent color)
	b := NewApp()
	b.setImage(src)
	gout := filepath.Join(dir, "copy.gif")
	if _, err := b.saveTo(gout, 0); err != nil {
		t.Fatal(err)
	}
	gf, _ := os.Open(gout)
	g, err := gif.Decode(gf)
	gf.Close()
	if err != nil {
		t.Fatal(err)
	}
	if _, _, _, al := g.At(0, 0).RGBA(); al != 0 {
		t.Errorf("GIF: the transparent column must stay transparent, alpha %d", al)
	}
	if _, _, _, al := g.At(2, 1).RGBA(); al == 0 {
		t.Error("GIF: the opaque part must stay opaque")
	}

	// JPEG: white, and it says so
	c := NewApp()
	c.setImage(src)
	msg, err := c.saveTo(filepath.Join(dir, "copy.jpg"), 0)
	if err != nil || !bytes.Contains([]byte(msg), []byte("transparent parts are white")) {
		t.Errorf("JPEG save: %q, %v", msg, err)
	}
}
