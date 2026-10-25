package main

import (
	"bytes"
	"encoding/binary"
	"image"
	"image/jpeg"
	"image/png"
	"os"
	"path/filepath"
	"testing"
)

func TestColourProfileSurvivesEditing(t *testing.T) {
	profile := bytes.Repeat([]byte("profile!"), 40)
	var jb bytes.Buffer
	jpeg.Encode(&jb, image.NewRGBA(image.Rect(0, 0, 8, 8)), nil)
	j := jb.Bytes()
	seg := append([]byte("ICC_PROFILE\x00\x01\x01"), profile...)
	app2 := append([]byte{0xff, 0xe2, 0, 0}, seg...)
	binary.BigEndian.PutUint16(app2[2:], uint16(len(seg)+2))
	withICC := append(append(append([]byte(nil), j[:2]...), app2...), j[2:]...)
	path := filepath.Join(t.TempDir(), "p3.jpg")
	os.WriteFile(path, withICC, 0o644)

	a := NewApp()
	if err := a.setImage(path); err != nil {
		t.Fatal(err)
	}
	if err := a.setPixels(image.NewNRGBA(image.Rect(0, 0, 3, 3)), path); err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(a.clipPNG, []byte("iCCP")) {
		t.Error("the PNG made of an edited picture must carry the profile of the file")
	}
	if _, err := png.Decode(bytes.NewReader(a.clipPNG)); err != nil {
		t.Errorf("the PNG with the profile must still decode: %v", err)
	}
	if got := colourChunks(filepath.Join(t.TempDir(), "none.jpg")); got != nil {
		t.Error("no file, no chunks")
	}
}
