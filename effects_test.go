package main

import (
	"image"
	"image/color"
	"testing"
)

func testPic(w, h int, f func(x, y int) color.NRGBA) *image.NRGBA {
	img := image.NewNRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			img.SetNRGBA(x, y, f(x, y))
		}
	}
	return img
}

func boxShape(x, y, w, h int) *shape {
	s := &shape{r: image.Rect(x, y, x+w, y+h), m: make([]uint8, w*h)}
	for i := range s.m {
		s.m[i] = 255
	}
	return s
}

func fl(v float64) *float64 { return &v }

func TestEffectsStayInTheShape(t *testing.T) {
	for _, op := range effectOps {
		img := testPic(20, 20, func(x, y int) color.NRGBA { return color.NRGBA{uint8(x * 12), uint8(y * 12), 90, 255} })
		orig := testPic(20, 20, func(x, y int) color.NRGBA { return img.NRGBAAt(x, y) })
		sp := effectSpec{Op: op, Color: "#ff0000", Color2: "#0000ff", Angle: 90}
		if _, err := applyEffect(img, boxShape(5, 5, 8, 8), sp); err != nil {
			t.Fatalf("%s: %v", op, err)
		}
		changed := false
		for y := 0; y < 20; y++ {
			for x := 0; x < 20; x++ {
				in := x >= 5 && x < 13 && y >= 5 && y < 13
				if !in && img.NRGBAAt(x, y) != orig.NRGBAAt(x, y) {
					t.Fatalf("%s changed %d,%d outside the box", op, x, y)
				}
				changed = changed || (in && img.NRGBAAt(x, y) != orig.NRGBAAt(x, y))
			}
		}
		if !changed && op != "auto_levels" && op != "gamma" && op != "levels" { // (levels and gamma with their defaults may leave a picture)
			t.Errorf("%s changed nothing in the box", op)
		}
	}
	if _, err := applyEffect(testPic(4, 4, func(x, y int) color.NRGBA { return color.NRGBA{} }), boxShape(0, 0, 4, 4), effectSpec{Op: "nope"}); err == nil {
		t.Errorf("an unknown effect is an error")
	}
}

func TestEffectResults(t *testing.T) {
	px := func(c color.NRGBA) *image.NRGBA { return testPic(8, 8, func(x, y int) color.NRGBA { return c }) }
	all := boxShape(0, 0, 8, 8)
	img := px(color.NRGBA{200, 100, 50, 255})
	applyEffect(img, all, effectSpec{Op: "negative"})
	if c := img.NRGBAAt(3, 3); c != (color.NRGBA{55, 155, 205, 255}) {
		t.Errorf("negative = %v", c)
	}
	img = px(color.NRGBA{200, 100, 50, 255})
	applyEffect(img, all, effectSpec{Op: "grayscale"})
	if c := img.NRGBAAt(3, 3); c.R != c.G || c.G != c.B {
		t.Errorf("grayscale = %v", c)
	}
	img = px(color.NRGBA{10, 20, 30, 255})
	applyEffect(img, all, effectSpec{Op: "brightness", Amount: fl(50)})
	if c := img.NRGBAAt(0, 0); c.R <= 100 {
		t.Errorf("brightness = %v", c)
	}
	// pixelize: every block is one color
	img = testPic(8, 8, func(x, y int) color.NRGBA { return color.NRGBA{uint8(x * 30), 0, 0, 255} })
	applyEffect(img, all, effectSpec{Op: "pixelize", Size: 4})
	if img.NRGBAAt(0, 0) != img.NRGBAAt(3, 3) || img.NRGBAAt(0, 0) == img.NRGBAAt(4, 0) {
		t.Errorf("pixelize blocks: %v %v %v", img.NRGBAAt(0, 0), img.NRGBAAt(3, 3), img.NRGBAAt(4, 0))
	}
	// blur: a white dot spreads
	img = testPic(9, 9, func(x, y int) color.NRGBA {
		if x == 4 && y == 4 {
			return color.NRGBA{255, 255, 255, 255}
		}
		return color.NRGBA{0, 0, 0, 255}
	})
	applyEffect(img, boxShape(0, 0, 9, 9), effectSpec{Op: "blur", Radius: 4})
	if img.NRGBAAt(4, 4).R >= 255 || img.NRGBAAt(3, 4).R == 0 {
		t.Errorf("blur: %v %v", img.NRGBAAt(4, 4), img.NRGBAAt(3, 4))
	}
	// censor with a bar
	img = px(color.NRGBA{200, 100, 50, 255})
	applyEffect(img, boxShape(2, 2, 3, 3), effectSpec{Op: "censor", Bar: true})
	if img.NRGBAAt(3, 3) != (color.NRGBA{0, 0, 0, 255}) || img.NRGBAAt(0, 0) != (color.NRGBA{200, 100, 50, 255}) {
		t.Errorf("censor bar: %v", img.NRGBAAt(3, 3))
	}
	// a polygon: only its inside (a triangle) changes; a feather softens the edge
	tri := &shape{r: image.Rect(0, 0, 8, 8), m: make([]uint8, 64)}
	for y := 0; y < 8; y++ {
		for x := 0; x <= y; x++ {
			tri.m[y*8+x] = 255
		}
	}
	img = px(color.NRGBA{200, 100, 50, 255})
	applyEffect(img, tri, effectSpec{Op: "negative"})
	if img.NRGBAAt(0, 7) == (color.NRGBA{200, 100, 50, 255}) || img.NRGBAAt(7, 0) != (color.NRGBA{200, 100, 50, 255}) {
		t.Errorf("polygon: inside %v outside %v", img.NRGBAAt(0, 7), img.NRGBAAt(7, 0))
	}
	img = px(color.NRGBA{0, 0, 0, 255})
	applyEffect(img, boxShape(8, 8, 8, 8), effectSpec{Op: "fill", Color: "#ffffff"}) // outside the picture: nothing to do, no crash
	// a gradient goes from one color to the other
	img = px(color.NRGBA{0, 0, 0, 255})
	applyEffect(img, all, effectSpec{Op: "gradient", Color: "#000000", Color2: "#ffffff"})
	if !(img.NRGBAAt(0, 4).R < 60 && img.NRGBAAt(7, 4).R > 190) {
		t.Errorf("gradient: %v … %v", img.NRGBAAt(0, 4), img.NRGBAAt(7, 4))
	}
	// an explosion brightens the middle and keeps the size
	img = testPic(21, 21, func(x, y int) color.NRGBA { return color.NRGBA{100, 100, 100, 255} })
	applyEffect(img, boxShape(0, 0, 21, 21), effectSpec{Op: "explosion"})
	if img.NRGBAAt(10, 10).R <= img.NRGBAAt(0, 0).R {
		t.Errorf("explosion: middle %v corner %v", img.NRGBAAt(10, 10), img.NRGBAAt(0, 0))
	}
}

func TestHighlightKeepsThePartAndPushesBackTheRest(t *testing.T) {
	img := testPic(60, 60, func(x, y int) color.NRGBA { return color.NRGBA{uint8(x * 4), uint8(y * 4), 200, 255} })
	orig := testPic(60, 60, func(x, y int) color.NRGBA { return img.NRGBAAt(x, y) })
	if _, err := applyEffect(img, boxShape(20, 20, 15, 15), effectSpec{Op: "highlight"}); err != nil {
		t.Fatal(err)
	}
	for y := 20; y < 35; y++ {
		for x := 20; x < 35; x++ {
			if img.NRGBAAt(x, y) != orig.NRGBAAt(x, y) {
				t.Fatalf("the part changed at %d,%d", x, y)
			}
		}
	}
	far, was := img.NRGBAAt(2, 50), orig.NRGBAAt(2, 50)
	if lum(float64(far.R), float64(far.G), float64(far.B)) >= lum(float64(was.R), float64(was.G), float64(was.B)) {
		t.Fatalf("the rest is not darker: %v, was %v", far, was)
	}
	if _, err := applyEffect(img, boxShape(0, 0, 60, 60), effectSpec{Op: "highlight"}); err == nil {
		t.Fatal("highlighting all of the picture should say to select a part")
	}
}

func TestHighlightLevelsPushBackMoreAndMore(t *testing.T) {
	var prev float64 = 256
	for _, level := range []string{"light", "medium", "full"} {
		img := testPic(60, 60, func(x, y int) color.NRGBA { return color.NRGBA{200, 120, 60, 255} })
		if _, err := applyEffect(img, boxShape(20, 20, 15, 15), effectSpec{Op: "highlight", Level: level}); err != nil {
			t.Fatal(err)
		}
		c := img.NRGBAAt(2, 50)
		l := lum(float64(c.R), float64(c.G), float64(c.B))
		if l >= prev {
			t.Fatalf("%s is not darker than the level before: %v", level, c)
		}
		prev = l
	}
	if _, err := applyEffect(testPic(10, 10, func(x, y int) color.NRGBA { return color.NRGBA{A: 255} }), boxShape(2, 2, 3, 3), effectSpec{Op: "highlight", Level: "loud"}); err == nil {
		t.Fatal("an unknown level should be refused")
	}
}
