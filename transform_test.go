package main

import (
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"testing"
)

func TestResizeFlipRotate(t *testing.T) {
	src := testPic(8, 4, func(x, y int) color.NRGBA { return color.NRGBA{uint8(x * 30), uint8(y * 60), 0, 255} })
	out, err := resizePicture(src, transformSpec{W: 4})
	if err != nil || out.Bounds().Dx() != 4 || out.Bounds().Dy() != 2 {
		t.Fatalf("resize to w=4 keeps the proportions: %v %v", out.Bounds(), err)
	}
	out, _ = resizePicture(src, transformSpec{Scale: 2, Filter: "nearest"})
	if out.Bounds().Dx() != 16 || out.NRGBAAt(2, 0) != src.NRGBAAt(1, 0) {
		t.Errorf("nearest scale 2: %v %v", out.Bounds(), out.NRGBAAt(2, 0))
	}
	f, _ := flipPicture(src, transformSpec{Horizontal: true})
	if f.NRGBAAt(0, 0) != src.NRGBAAt(7, 0) {
		t.Errorf("flip: %v", f.NRGBAAt(0, 0))
	}
	if _, err := flipPicture(src, transformSpec{}); err == nil {
		t.Errorf("flip needs a direction")
	}
	r, _ := rotatePicture(src, transformSpec{Angle: 90}, true)
	if r.Bounds().Dx() != 4 || r.Bounds().Dy() != 8 {
		t.Errorf("a quarter turn swaps the sides: %v", r.Bounds())
	}
	r, _ = rotatePicture(src, transformSpec{Angle: 45}, true)
	if r.Bounds().Dx() <= 8 || r.NRGBAAt(0, 0).A != 0 {
		t.Errorf("a turn of 45° makes a bigger canvas with transparent corners: %v %v", r.Bounds(), r.NRGBAAt(0, 0))
	}
	sq := testPic(8, 8, func(x, y int) color.NRGBA { return color.NRGBA{0, 200, 0, 255} })
	r, _ = rotatePicture(sq, transformSpec{Angle: 45, Crop: true, Bg: "#ff0000ff"}, true)
	if r.Bounds().Dx() != 8 || r.NRGBAAt(0, 0) != (color.NRGBA{255, 0, 0, 255}) {
		t.Errorf("crop: %v %v", r.Bounds(), r.NRGBAAt(0, 0))
	}
	mid := r.NRGBAAt(4, 4)
	if mid.A != 255 {
		t.Errorf("the middle stays covered: %v", mid)
	}
}

func TestPlacePicture(t *testing.T) {
	dir := t.TempDir()
	logo := testPic(4, 4, func(x, y int) color.NRGBA { return color.NRGBA{255, 0, 0, 255} })
	path := filepath.Join(dir, "logo.png")
	f, _ := os.Create(path)
	png.Encode(f, logo)
	f.Close()
	base := testPic(10, 10, func(x, y int) color.NRGBA { return color.NRGBA{0, 0, 255, 255} })
	half := 0.5
	out, err := placePicture(base, transformSpec{Path: path, X: 8, Y: 8, Opacity: &half})
	if err != nil {
		t.Fatal(err)
	}
	if c := out.NRGBAAt(9, 9); c.R == 0 || c.B == 0 {
		t.Errorf("half opaque over blue: %v", c)
	}
	if out.NRGBAAt(0, 0) != (color.NRGBA{0, 0, 255, 255}) || out.Bounds() != image.Rect(0, 0, 10, 10) {
		t.Errorf("the rest stays: %v", out.NRGBAAt(0, 0))
	}
	if _, err := placePicture(base, transformSpec{Path: path, X: 50, Y: 50}); err == nil {
		t.Errorf("a picture outside is an error")
	}
	if _, err := placePicture(base, transformSpec{Path: filepath.Join(dir, "nope.txt")}); err == nil {
		t.Errorf("not a picture file")
	}
}
