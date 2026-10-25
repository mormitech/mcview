package main

import (
	"bytes"
	"encoding/binary"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"os"
	"path/filepath"
	"testing"
)

func TestPixelAtPNG(t *testing.T) {
	img := image.NewNRGBA(image.Rect(0, 0, 3, 2))
	img.SetNRGBA(0, 0, color.NRGBA{R: 201, G: 77, B: 13, A: 255})
	img.SetNRGBA(1, 0, color.NRGBA{R: 10, G: 20, B: 30, A: 128}) // half transparent: must come back as it is, not premultiplied
	img.SetNRGBA(2, 1, color.NRGBA{R: 255, G: 255, B: 255, A: 0})
	path := filepath.Join(t.TempDir(), "p.png")
	var b bytes.Buffer
	png.Encode(&b, img)
	os.WriteFile(path, b.Bytes(), 0o644)

	a := NewApp()
	a.setImage(path)
	if p := a.PixelAt(0, 0, 3, 2); p.Ready {
		t.Error("PixelAt before PreparePixels must not be ready")
	}
	if src, err := a.PreparePixels(); src != "go" || err != nil {
		t.Fatalf("PreparePixels = %q, %v", src, err)
	}
	for _, c := range []struct {
		x, y       int
		r, g, b, a int
	}{{0, 0, 201, 77, 13, 255}, {1, 0, 10, 20, 30, 128}, {2, 1, 255, 255, 255, 0}, {1, 1, 0, 0, 0, 0}} {
		p := a.PixelAt(c.x, c.y, 3, 2)
		if !p.Ready || p.R != c.r || p.G != c.g || p.B != c.b || p.A != c.a {
			t.Errorf("PixelAt(%d,%d) = %+v, want %d %d %d %d", c.x, c.y, p, c.r, c.g, c.b, c.a)
		}
	}
	// outside the picture, and a size that is not the picture the web view shows
	for _, c := range [][4]int{{3, 0, 3, 2}, {0, 2, 3, 2}, {-1, 0, 3, 2}, {0, 0, 2, 3}} {
		if a.PixelAt(c[0], c[1], c[2], c[3]).Ready {
			t.Errorf("PixelAt%v must not be ready", c)
		}
	}
	// the next picture starts without the decoded one of the picture before
	other := filepath.Join(filepath.Dir(path), "q.png")
	os.WriteFile(other, b.Bytes(), 0o644)
	a.setImage(other)
	if a.PixelAt(0, 0, 3, 2).Ready {
		t.Error("a new picture must not read from the decoded one before")
	}
}

func TestPixelAtJPEG(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 16, 16))
	for y := 0; y < 16; y++ {
		for x := 0; x < 16; x++ {
			img.Set(x, y, color.RGBA{R: 200, G: 40, B: 90, A: 255})
		}
	}
	path := filepath.Join(t.TempDir(), "p.jpg")
	var b bytes.Buffer
	jpeg.Encode(&b, img, &jpeg.Options{Quality: 100})
	os.WriteFile(path, b.Bytes(), 0o644)
	a := NewApp()
	a.setImage(path)
	a.PreparePixels()
	p := a.PixelAt(5, 5, 16, 16)
	near := func(got, want int) bool { return got-want <= 3 && want-got <= 3 }
	if !p.Ready || !near(p.R, 200) || !near(p.G, 40) || !near(p.B, 90) || p.A != 255 {
		t.Errorf("PixelAt = %+v, want about 200 40 90 255", p)
	}
}

func TestUndecodableTypeIsLeftToTheBrowser(t *testing.T) {
	path := filepath.Join(t.TempDir(), "x.avif")
	os.WriteFile(path, []byte("not something Go decodes"), 0o644)
	a := NewApp()
	a.setImage(path)
	if src, err := a.PreparePixels(); src != "browser" || err != nil {
		t.Errorf("PreparePixels = %q, %v; want browser", src, err)
	}
}

// Where the first stored pixel (0,0) shows up in the picture as shown, for each EXIF orientation (the table of the EXIF standard),
// and that the mapping loses no pixel.
func TestDisplayToSource(t *testing.T) {
	const sw, sh = 4, 3
	corner := map[int][2]int{ // the display position of the stored pixel (0, 0), as a corner: 0 left/top, 1 right/bottom
		1: {0, 0}, 2: {1, 0}, 3: {1, 1}, 4: {0, 1}, 5: {0, 0}, 6: {1, 0}, 7: {1, 1}, 8: {0, 1},
	}
	for o := 1; o <= 8; o++ {
		seen := map[[2]int]bool{}
		_, _, dw, dh := displayToSource(o, sw, sh, 0, 0)
		if (o >= 5) != (dw == sh && dh == sw) {
			t.Errorf("orientation %d: shown size %d × %d", o, dw, dh)
		}
		for y := 0; y < dh; y++ {
			for x := 0; x < dw; x++ {
				sx, sy, _, _ := displayToSource(o, sw, sh, x, y)
				if sx < 0 || sy < 0 || sx >= sw || sy >= sh {
					t.Fatalf("orientation %d: (%d,%d) -> (%d,%d) is outside the stored picture", o, x, y, sx, sy)
				}
				seen[[2]int{sx, sy}] = true
				if sx == 0 && sy == 0 {
					want := corner[o]
					if (x == dw-1) != (want[0] == 1) || (y == dh-1) != (want[1] == 1) || (x != 0 && x != dw-1) || (y != 0 && y != dh-1) {
						t.Errorf("orientation %d: stored (0,0) is shown at (%d,%d) of %d × %d, want corner %v", o, x, y, dw, dh, want)
					}
				}
			}
		}
		if len(seen) != sw*sh {
			t.Errorf("orientation %d: %d of %d stored pixels are shown", o, len(seen), sw*sh)
		}
	}
}

// withOrientation puts an EXIF block with an Orientation tag after the start of a JPEG.
func withOrientation(t *testing.T, jpg []byte, o uint16, order binary.AppendByteOrder, tag string) []byte {
	t.Helper()
	tiff := []byte(tag)
	tiff = order.AppendUint16(tiff, 42)
	tiff = order.AppendUint32(tiff, 8)
	tiff = order.AppendUint16(tiff, 1) // one entry
	tiff = order.AppendUint16(tiff, 0x0112)
	tiff = order.AppendUint16(tiff, 3) // SHORT
	tiff = order.AppendUint32(tiff, 1)
	tiff = order.AppendUint16(tiff, o)
	tiff = append(tiff, 0, 0)
	tiff = order.AppendUint32(tiff, 0)
	seg := append([]byte("Exif\x00\x00"), tiff...)
	out := []byte{0xFF, 0xD8, 0xFF, 0xE1}
	out = binary.BigEndian.AppendUint16(out, uint16(len(seg)+2))
	out = append(out, seg...)
	return append(out, jpg[2:]...)
}

func TestExifOrientation(t *testing.T) {
	var b bytes.Buffer
	jpeg.Encode(&b, image.NewRGBA(image.Rect(0, 0, 8, 4)), nil)
	plain := b.Bytes()
	if o := orientationOf(plain); o != 1 {
		t.Errorf("a JPEG without EXIF: orientation %d, want 1", o)
	}
	for o := uint16(1); o <= 8; o++ {
		for _, c := range []struct {
			order binary.AppendByteOrder
			tag   string
		}{{binary.LittleEndian, "II"}, {binary.BigEndian, "MM"}} {
			data := withOrientation(t, plain, o, c.order, c.tag)
			if got := orientationOf(data); got != int(o) {
				t.Errorf("%s orientation %d read as %d", c.tag, o, got)
			}
			if _, err := jpeg.Decode(bytes.NewReader(data)); err != nil {
				t.Errorf("the test JPEG with orientation %d does not decode: %v", o, err)
			}
		}
	}
	if o := orientationOf([]byte("not a jpeg")); o != 1 {
		t.Errorf("garbage: orientation %d, want 1", o)
	}
}

func TestPixelAtFollowsExifOrientation(t *testing.T) {
	// stored 8 × 4: left half red, right half blue; orientation 6 shows it turned clockwise, 4 × 8: red at the top, blue at the bottom
	img := image.NewRGBA(image.Rect(0, 0, 8, 4))
	for y := 0; y < 4; y++ {
		for x := 0; x < 8; x++ {
			if x < 4 {
				img.Set(x, y, color.RGBA{R: 255, A: 255})
			} else {
				img.Set(x, y, color.RGBA{B: 255, A: 255})
			}
		}
	}
	var b bytes.Buffer
	jpeg.Encode(&b, img, &jpeg.Options{Quality: 100})
	path := filepath.Join(t.TempDir(), "turned.jpg")
	os.WriteFile(path, withOrientation(t, b.Bytes(), 6, binary.LittleEndian, "II"), 0o644)
	a := NewApp()
	a.setImage(path)
	a.PreparePixels()
	if p := a.PixelAt(1, 1, 8, 4); p.Ready { // the size of the stored picture is not the one shown
		t.Error("the stored size must not be accepted for a turned picture")
	}
	top, bottom := a.PixelAt(1, 1, 4, 8), a.PixelAt(1, 6, 4, 8)
	if !top.Ready || top.R < 200 || top.B > 60 || !bottom.Ready || bottom.B < 200 || bottom.R > 60 {
		t.Errorf("turned picture: top %+v should be red, bottom %+v blue", top, bottom)
	}
}
