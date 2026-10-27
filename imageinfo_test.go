package main

import (
	"encoding/binary"
	"image"
	"image/color"
	"testing"
)

// exifJPEG is the start of a JPEG with an EXIF block: Make "Acme", ISO 400, an exposure of 1/250 s and a GPS block.
func exifJPEG() []byte {
	le := binary.LittleEndian
	t := []byte("II*\x00")
	t = le.AppendUint32(t, 8)
	entry := func(tag, typ uint16, count, val uint32) []byte {
		e := le.AppendUint16(nil, tag)
		e = le.AppendUint16(e, typ)
		e = le.AppendUint32(e, count)
		return le.AppendUint32(e, val)
	}
	// IFD0 at 8: Make (5 bytes, at 62), ExifIFD (at 68), GPS IFD (at 100); 3 entries = 2 + 36 + 4 bytes, so 8 + 42 = 50 .. data from 50
	t = le.AppendUint16(t, 3)
	t = append(t, entry(0x010F, 2, 5, 50)...)
	t = append(t, entry(0x8769, 4, 1, 56)...)
	t = append(t, entry(0x8825, 4, 1, 100)...)
	t = le.AppendUint32(t, 0)
	t = append(t, "Acme\x00"...) // at 50
	t = append(t, 0)             // pad to 56
	// ExifIFD at 56: 2 entries = 2 + 24 + 4 = 30 -> data at 86
	t = le.AppendUint16(t, 2)
	t = append(t, entry(0x8827, 3, 1, 400)...)
	t = append(t, entry(0x829A, 5, 1, 86)...)
	t = le.AppendUint32(t, 0)
	t = le.AppendUint32(t, 1) // 1 / 250
	t = le.AppendUint32(t, 250)
	t = append(t, 0, 0, 0, 0) // pad to 100
	t = le.AppendUint16(t, 0) // an empty GPS IFD at 100
	seg := append([]byte("Exif\x00\x00"), t...)
	b := []byte{0xFF, 0xD8, 0xFF, 0xE1}
	b = binary.BigEndian.AppendUint16(b, uint16(len(seg)+2))
	return append(b, seg...)
}

func TestReadExif(t *testing.T) {
	ex := readExif(exifJPEG())
	if ex["make"] != "Acme" || ex["iso"] != 400 || ex["exposureTime"] != "1/250 s" || ex["hasGPS"] != true {
		t.Errorf("exif = %v", ex)
	}
	if readExif([]byte("not a jpeg")) != nil || readExif(nil) != nil {
		t.Errorf("no EXIF in what is not a JPEG")
	}
}

func TestColorStats(t *testing.T) {
	img := image.NewNRGBA(image.Rect(0, 0, 4, 2))
	for x := 0; x < 4; x++ {
		img.SetNRGBA(x, 0, color.NRGBA{255, 0, 0, 255})
	}
	img.SetNRGBA(0, 1, color.NRGBA{0, 0, 255, 255})
	img.SetNRGBA(1, 1, color.NRGBA{0, 0, 255, 255})
	// the other two pixels stay transparent
	st := colorStats(img, nil)
	if st["pixels"] != 8 || st["differentColors"] != 3 || st["transparentPct"] != 25.0 || st["grayscale"] != false {
		t.Errorf("stats = %v", st)
	}
	main := st["mainColors"].([]map[string]any)
	if len(main) != 2 || main[0]["color"] != "#FF0000" || main[0]["pct"] != 66.7 || main[1]["color"] != "#0000FF" {
		t.Errorf("main colors = %v", main)
	}
}

func TestHistoryReport(t *testing.T) {
	step := func(n int, what, by string) any { return map[string]any{"n": float64(n), "label": what, "by": by} }
	r := historyReport(map[string]any{"at": float64(2), "steps": []any{step(0, "start", "user"), step(1, "pen stroke", "user"), step(2, "rect drawn", "AI"), step(3, "cropped", "AI")}})
	un, re := r["undoable"].([]any), r["redoable"].([]any)
	if len(un) != 2 || un[0].(map[string]any)["what"] != "rect drawn" || un[1].(map[string]any)["by"] != "user" {
		t.Errorf("undoable = %v", un)
	}
	if len(re) != 1 || re[0].(map[string]any)["what"] != "cropped" || r["start"] != "start" {
		t.Errorf("redoable = %v, start = %v", re, r["start"])
	}
}
