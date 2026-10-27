package main

import (
	"image"
	"image/color"
	"testing"
)

func TestRotatePartQuarterFillsFrame(t *testing.T) {
	img := image.NewNRGBA(image.Rect(0, 0, 20, 10))
	for y := 0; y < 10; y++ {
		for x := 0; x < 20; x++ {
			img.SetNRGBA(x, y, color.NRGBA{255, 0, 0, 255})
		}
	}
	box := image.Rect(0, 0, 10, 10)
	rotatePart(img, box, 90)
	if got := img.NRGBAAt(15, 5); got != (color.NRGBA{255, 0, 0, 255}) {
		t.Fatalf("outside the box changed: %v", got)
	}
	if got := img.NRGBAAt(5, 5); got.A != 255 {
		t.Fatalf("the frame is not filled: %v", got)
	}
}

func TestRotatePartFineHasNoEmptyCorners(t *testing.T) {
	img := image.NewNRGBA(image.Rect(0, 0, 40, 30))
	for i := range img.Pix {
		img.Pix[i] = 255
	}
	rotatePart(img, image.Rect(5, 5, 35, 25), 17)
	for _, p := range [][2]int{{5, 5}, {34, 5}, {5, 24}, {34, 24}} {
		if a := img.NRGBAAt(p[0], p[1]).A; a != 255 {
			t.Fatalf("corner %v is empty (alpha %d)", p, a)
		}
	}
}

func TestRotationBaseContinues(t *testing.T) {
	a := &App{}
	img := image.NewNRGBA(image.Rect(0, 0, 4, 4))
	orig, total := a.rotationBase(img, "k", 10, false)
	a.keepRotation(orig, "k", total)
	if _, tot := a.rotationBase(img, "k", 20, true); tot != 30 {
		t.Fatalf("total %v, want 30", tot)
	}
	if _, tot := a.rotationBase(img, "k", 20, false); tot != 20 {
		t.Fatalf("a turn that does not continue: %v", tot)
	}
	if _, tot := a.rotationBase(img, "other", 20, true); tot != 20 {
		t.Fatalf("another part: %v", tot)
	}
}

func TestWholeRotationsAreMadeFromTheFirstPicture(t *testing.T) {
	a := &App{}
	img := image.NewNRGBA(image.Rect(0, 0, 40, 20))
	for i := range img.Pix {
		img.Pix[i] = 200
	}
	a.mu.Lock()
	a.newStateLocked()
	a.mu.Unlock()
	orig, total := a.rotationBase(img, "whole", 10, true)
	a.keepRotation(orig, "whole", total)
	if o, tot := a.rotationBase(img, "whole", -10, true); tot != 0 || o != orig {
		t.Fatalf("a turn back must give the first picture again: %v", tot)
	}
	a.mu.Lock()
	a.newStateLocked() // something else changed the picture
	a.mu.Unlock()
	if _, tot := a.rotationBase(img, "whole", 5, true); tot != 5 {
		t.Fatalf("the cache must not outlive a change: %v", tot)
	}
}
