package main

import (
	"bytes"
	"encoding/base64"
	"encoding/binary"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func decodePNG(t *testing.T, b64 string) image.Image {
	t.Helper()
	raw, err := base64.StdEncoding.DecodeString(b64)
	if err != nil {
		t.Fatal(err)
	}
	img, err := png.Decode(bytes.NewReader(raw))
	if err != nil {
		t.Fatal(err)
	}
	return img
}

// a 10 × 6 picture where the pixel at x, y has the colour (x*20, y*40, 77) and the transparency 200 + x
func gridPNG(t *testing.T) string {
	t.Helper()
	img := image.NewNRGBA(image.Rect(0, 0, 10, 6))
	for y := 0; y < 6; y++ {
		for x := 0; x < 10; x++ {
			img.SetNRGBA(x, y, color.NRGBA{R: uint8(x * 20), G: uint8(y * 40), B: 77, A: uint8(200 + x)})
		}
	}
	path := filepath.Join(t.TempDir(), "grid.png")
	var b bytes.Buffer
	png.Encode(&b, img)
	os.WriteFile(path, b.Bytes(), 0o644)
	return path
}

func TestSelectionPNG(t *testing.T) {
	a := NewApp()
	a.setImage(gridPNG(t))
	b64, err := a.SelectionPNG(3, 2, 4, 3, 10, 6)
	if err != nil || b64 == "" {
		t.Fatalf("SelectionPNG = %q, %v", b64, err)
	}
	got := decodePNG(t, b64)
	if got.Bounds().Dx() != 4 || got.Bounds().Dy() != 3 {
		t.Fatalf("the part is %v, want 4 × 3", got.Bounds())
	}
	for y := 0; y < 3; y++ {
		for x := 0; x < 4; x++ { // the colours are those of the file, transparency included, not premultiplied or converted
			want := color.NRGBA{R: uint8((x + 3) * 20), G: uint8((y + 2) * 40), B: 77, A: uint8(200 + x + 3)}
			if c := color.NRGBAModel.Convert(got.At(x, y)).(color.NRGBA); c != want {
				t.Errorf("pixel %d,%d = %v, want %v", x, y, c, want)
			}
		}
	}
	whole, err := a.SelectionPNG(0, 0, 10, 6, 10, 6)
	if err != nil || decodePNG(t, whole).Bounds().Dx() != 10 {
		t.Errorf("the whole picture: %v", err)
	}
	// outside the picture, empty, or a picture that is not the one the web view shows
	for _, c := range [][6]int{{8, 0, 3, 1, 10, 6}, {0, 5, 1, 2, 10, 6}, {-1, 0, 2, 2, 10, 6}, {0, 0, 0, 2, 10, 6}} {
		if _, err := a.SelectionPNG(c[0], c[1], c[2], c[3], c[4], c[5]); err == nil {
			t.Errorf("SelectionPNG%v must be an error", c)
		}
	}
	if s, err := a.SelectionPNG(0, 0, 2, 2, 6, 10); s != "" || err != nil {
		t.Errorf("another size than the picture's: %q, %v; want nothing", s, err)
	}
}

func TestSelectionOfTurnedJPEG(t *testing.T) {
	// stored 8 × 4, left half red and right half blue; shown turned clockwise (orientation 6) as 4 × 8: red on top, blue at the bottom
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
	b64, err := a.SelectionPNG(0, 4, 4, 4, 4, 8) // the lower half of what is shown
	if err != nil || b64 == "" {
		t.Fatalf("SelectionPNG = %q, %v", b64, err)
	}
	got := decodePNG(t, b64)
	if got.Bounds().Dx() != 4 || got.Bounds().Dy() != 4 {
		t.Fatalf("the part is %v, want 4 × 4", got.Bounds())
	}
	if r, _, bl, _ := got.At(1, 1).RGBA(); r>>8 > 60 || bl>>8 < 200 {
		t.Errorf("the lower half should be blue, got r=%d b=%d", r>>8, bl>>8)
	}
}

func TestSelectionOfTypeGoCannotDecode(t *testing.T) {
	path := filepath.Join(t.TempDir(), "x.svg")
	os.WriteFile(path, []byte("<svg/>"), 0o644)
	a := NewApp()
	a.setImage(path)
	if s, err := a.SelectionPNG(0, 0, 1, 1, 1, 1); s != "" || err != nil {
		t.Errorf("SelectionPNG = %q, %v; want nothing, so that the UI cuts it from the screen", s, err)
	}
}

// Puts a picture on the real clipboard of the machine, so it only runs when asked: MCVIEW_TEST_CLIPBOARD=1 go test -run Clipboard
func TestCopyImageNative(t *testing.T) {
	if os.Getenv("MCVIEW_TEST_CLIPBOARD") == "" {
		t.Skip("set MCVIEW_TEST_CLIPBOARD=1 to run: it overwrites the clipboard")
	}
	if runtime.GOOS != "darwin" {
		t.Skip("the check below reads the clipboard the macOS way")
	}
	a := NewApp()
	a.setImage(gridPNG(t))
	b64, _ := a.SelectionPNG(1, 1, 5, 4, 10, 6)
	if err := a.CopyImageNative(b64); err != nil {
		t.Fatal(err)
	}
	out, err := exec.Command("osascript", "-e", "clipboard info").Output()
	if err != nil || !strings.Contains(string(out), "PNGf") {
		t.Errorf("clipboard info = %q, %v; want a PNG on it", out, err)
	}
}

func TestRotatePNG(t *testing.T) {
	// 3 × 2: the top left is red, the rest differs by its place
	src := image.NewNRGBA(image.Rect(0, 0, 3, 2))
	for y := 0; y < 2; y++ {
		for x := 0; x < 3; x++ {
			src.SetNRGBA(x, y, color.NRGBA{R: uint8(x*50 + 10), G: uint8(y*100 + 10), B: 5, A: uint8(200 + x)})
		}
	}
	var b bytes.Buffer
	png.Encode(&b, src)
	in := base64.StdEncoding.EncodeToString(b.Bytes())
	a := NewApp()
	at := func(img image.Image, x, y int) color.NRGBA {
		return color.NRGBAModel.Convert(img.At(x, y)).(color.NRGBA)
	}
	for q, c := range map[int]struct{ w, h int }{1: {2, 3}, 2: {3, 2}, 3: {2, 3}} {
		out, err := a.RotatePNG(in, q)
		if err != nil {
			t.Fatal(err)
		}
		img := decodePNG(t, out)
		if img.Bounds().Dx() != c.w || img.Bounds().Dy() != c.h {
			t.Errorf("%d quarters: size %v, want %d × %d", q, img.Bounds(), c.w, c.h)
		}
		// the top left pixel of the source goes to the top right (1), the bottom right (2), the bottom left (3)
		corner := map[int][2]int{1: {1, 0}, 2: {2, 1}, 3: {0, 2}}[q]
		if got, want := at(img, corner[0], corner[1]), at(src, 0, 0); got != want {
			t.Errorf("%d quarters: the top left pixel is at %v as %v, want %v", q, corner, got, want)
		}
	}
	if out, _ := a.RotatePNG(in, 4); out != in {
		t.Error("four quarters are no turn")
	}
	if out, _ := a.RotatePNG(in, -1); decodePNG(t, out).Bounds().Dx() != 2 {
		t.Error("-1 is three quarters")
	}
	if _, err := a.RotatePNG("not base64!", 1); err == nil {
		t.Error("a bad picture must be an error")
	}
}
