package main

import (
	"encoding/binary"
	"fmt"
	"image"
	"image/color"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"os"
	"path/filepath"
	"strings"

	_ "golang.org/x/image/bmp"
	_ "golang.org/x/image/tiff"
	_ "golang.org/x/image/webp"
)

// The pixel readout (Tab, the pixelated view) shows the exact color of the pixel under the pointer. The web view cannot tell it: it
// converts colors for the screen (a Display P3 or ICC-profile picture is not shown with the numbers that are in the file) and its
// JPEG decoder may differ in the last digit. So the picture is decoded a second time here, and the values are those of the decoded
// file, without any color conversion.

// maxPixels is the biggest picture decoded for the readout (200 megapixels would take 800 MB as RGBA).
const maxPixels = 150_000_000

// pixelCache is the decoded picture of one window.
type pixelCache struct {
	seq    int         // the picture it belongs to (App.imgSeq)
	img    image.Image // the decoded picture, as stored in the file
	orient int         // EXIF orientation, 1 to 8: the web view shows the picture turned that way, so the readout has to as well
}

// Pixel is the color of one pixel, not premultiplied, 8 bits per channel (more bits are reduced). Ready is false when the picture is
// not decoded yet or the position is outside it.
type Pixel struct {
	Ready bool `json:"ready"`
	X     int  `json:"x"`
	Y     int  `json:"y"`
	R     int  `json:"r"`
	G     int  `json:"g"`
	B     int  `json:"b"`
	A     int  `json:"a"`
}

// PreparePixels decodes the picture of this window for PixelAt, and says where the readout gets its values:
//
//	"go"      from the decoded file (JPEG, PNG, GIF, BMP, WebP, TIFF)
//	"browser" Go cannot decode this type (AVIF, SVG, ICO): the UI reads the pixel from what the web view drew
//
// A picture too big to decode is an error.
func (a *App) PreparePixels() (string, error) {
	_, source, err := a.decoded()
	return source, err
}

// decoded is the decoded picture of this window (decoding it first if that was not done yet), and "go" or "browser" as above.
func (a *App) decoded() (pixelCache, string, error) {
	a.mu.Lock()
	seq := a.imgSeq
	pc := a.pix
	a.mu.Unlock()
	if pc.seq == seq && pc.img != nil { // also the picture pasted from the clipboard, which is decoded from the start
		return pc, "go", nil
	}
	cur := a.current()
	if cur.Path == "" {
		return pixelCache{}, "", nil
	}
	f, err := os.Open(cur.Path)
	if err != nil {
		return pixelCache{}, "", err
	}
	defer f.Close()
	cfg, _, err := image.DecodeConfig(f)
	if err != nil { // a type that is not registered here
		return pixelCache{}, "browser", nil
	}
	if cfg.Width*cfg.Height > maxPixels {
		return pixelCache{}, "", fmt.Errorf("%s is too big (%d × %d) to read its pixels", cur.Name, cfg.Width, cfg.Height)
	}
	if _, err := f.Seek(0, 0); err != nil {
		return pixelCache{}, "", err
	}
	trace("decoding pixels %dx%d", cfg.Width, cfg.Height)
	img, _, err := image.Decode(f)
	if err != nil {
		return pixelCache{}, "browser", nil
	}
	orient := 1
	if ext := strings.ToLower(filepath.Ext(cur.Path)); ext == ".jpg" || ext == ".jpeg" || ext == ".jpe" || ext == ".jfif" {
		orient = exifOrientation(cur.Path)
	}
	pc = pixelCache{seq: seq, img: img, orient: orient}
	a.mu.Lock()
	if a.imgSeq == seq { // the window may have moved on to another picture meanwhile
		a.pix = pc
	}
	a.mu.Unlock()
	trace("pixels decoded")
	return pc, "go", nil
}

// PixelAt is the color at x, y of the picture as it is shown (the web view's natural size w × h, which has the EXIF turn applied).
func (a *App) PixelAt(x, y, w, h int) Pixel {
	a.mu.Lock()
	pc := a.pix
	ok := pc.img != nil && pc.seq == a.imgSeq
	a.mu.Unlock()
	if !ok {
		return Pixel{}
	}
	b := pc.img.Bounds()
	sx, sy, dw, dh := displayToSource(pc.orient, b.Dx(), b.Dy(), x, y)
	if dw != w || dh != h || x < 0 || y < 0 || x >= dw || y >= dh { // the sizes do not agree: not the picture the web view shows
		return Pixel{}
	}
	c := color.NRGBAModel.Convert(pc.img.At(b.Min.X+sx, b.Min.Y+sy)).(color.NRGBA)
	return Pixel{Ready: true, X: x, Y: y, R: int(c.R), G: int(c.G), B: int(c.B), A: int(c.A)}
}

// displayToSource turns a position of the picture as shown into the one in the picture as stored, for an EXIF orientation o. The
// shown picture is sw × sh stored, or sh × sw when o turns it by a quarter (5 to 8): that size is returned as dw × dh.
//
//	1 as stored          2 mirrored left-right     3 turned 180°         4 mirrored top-bottom
//	5 mirrored along the main diagonal   6 turned 90° clockwise   7 mirrored along the other diagonal   8 turned 90° counter-clockwise
func displayToSource(o, sw, sh, x, y int) (sx, sy, dw, dh int) {
	switch o {
	case 2:
		return sw - 1 - x, y, sw, sh
	case 3:
		return sw - 1 - x, sh - 1 - y, sw, sh
	case 4:
		return x, sh - 1 - y, sw, sh
	case 5:
		return y, x, sh, sw
	case 6:
		return y, sh - 1 - x, sh, sw
	case 7:
		return sw - 1 - y, sh - 1 - x, sh, sw
	case 8:
		return sw - 1 - y, x, sh, sw
	}
	return x, y, sw, sh
}

// exifOrientation reads the Orientation tag of the EXIF block of a JPEG; 1 (as stored) when there is none or it cannot be read.
func exifOrientation(path string) int {
	f, err := os.Open(path)
	if err != nil {
		return 1
	}
	defer f.Close()
	head := make([]byte, 128*1024) // the EXIF block is one of the first segments
	n, _ := f.Read(head)
	return orientationOf(head[:n])
}

func orientationOf(b []byte) int {
	if len(b) < 4 || b[0] != 0xFF || b[1] != 0xD8 {
		return 1
	}
	for i := 2; i+4 <= len(b); {
		if b[i] != 0xFF {
			return 1
		}
		marker, size := b[i+1], int(binary.BigEndian.Uint16(b[i+2:]))
		if marker == 0xDA || size < 2 { // the picture data starts: no EXIF before it
			return 1
		}
		if marker == 0xE1 && i+4+size-2 <= len(b) {
			if o := tiffOrientation(b[i+4 : i+2+size]); o != 0 {
				return o
			}
		}
		i += 2 + size
	}
	return 1
}

func tiffOrientation(seg []byte) int {
	if len(seg) < 14 || string(seg[:6]) != "Exif\x00\x00" {
		return 0
	}
	t := seg[6:]
	var bo binary.ByteOrder
	switch string(t[:2]) {
	case "II":
		bo = binary.LittleEndian
	case "MM":
		bo = binary.BigEndian
	default:
		return 0
	}
	ifd := int(bo.Uint32(t[4:]))
	if ifd < 8 || ifd+2 > len(t) {
		return 0
	}
	count := int(bo.Uint16(t[ifd:]))
	for k := 0; k < count; k++ {
		e := ifd + 2 + 12*k
		if e+12 > len(t) {
			return 0
		}
		if bo.Uint16(t[e:]) == 0x0112 { // Orientation, a SHORT
			if o := int(bo.Uint16(t[e+8:])); o >= 1 && o <= 8 {
				return o
			}
			return 0
		}
	}
	return 0
}

// HasTransparency tells whether the picture has a pixel that is not opaque: the UI then shows a chequered background, as Gimp does. It is
// worked out after the picture is shown (it decodes the file, which the pixel readout needs anyway). JPEG and BMP have none; SVG and ICO
// are taken to have (their background is empty); AVIF is not looked into.
func (a *App) HasTransparency() bool {
	cur := a.current()
	switch ext := strings.ToLower(filepath.Ext(cur.Path)); {
	case cur.URL == "":
		return false
	case cur.Path != "" && !canAlpha(cur.Path):
		return false
	case ext == ".svg" || ext == ".ico":
		return true
	case ext == ".avif":
		return false
	}
	pc, source, err := a.decoded()
	if err != nil || source != "go" {
		return false
	}
	if o, ok := pc.img.(interface{ Opaque() bool }); ok {
		return !o.Opaque()
	}
	return false
}
