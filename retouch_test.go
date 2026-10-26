package main

import (
	"image"
	"image/color"
	"strings"
	"testing"
	"time"
)

// testPicture makes the picture of a window from f.
func testPicture(t *testing.T, w, h int, f func(x, y int) color.NRGBA) *App {
	t.Helper()
	a := NewApp()
	img := image.NewNRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			img.SetNRGBA(x, y, f(x, y))
		}
	}
	if err := a.setPixelsEmit(img, "", false); err != nil {
		t.Fatal(err)
	}
	return a
}

func pixelOf(a *App, x, y int) color.NRGBA {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.clip.(*image.NRGBA).NRGBAAt(x, y)
}

func TestRetouchClone(t *testing.T) {
	red, blue, white := color.NRGBA{255, 0, 0, 255}, color.NRGBA{0, 0, 255, 255}, color.NRGBA{255, 255, 255, 255}
	a := testPicture(t, 30, 10, func(x, y int) color.NRGBA {
		switch {
		case x < 5 && y < 5:
			return red
		case x < 5:
			return blue
		}
		return white
	})
	// a plain copy of the box
	if _, err := a.Retouch(0, `{"op":"clone","x":0,"y":0,"w":5,"h":10,"tx":10,"ty":0}`); err != nil {
		t.Fatal(err)
	}
	if pixelOf(a, 10, 0) != red || pixelOf(a, 14, 9) != blue || pixelOf(a, 15, 0) != white {
		t.Errorf("the copy: %v %v %v", pixelOf(a, 10, 0), pixelOf(a, 14, 9), pixelOf(a, 15, 0))
	}
	// flipped top to bottom: the red part is at the bottom now
	if _, err := a.Retouch(0, `{"op":"clone","x":0,"y":0,"w":5,"h":10,"tx":20,"ty":0,"flip_v":true}`); err != nil {
		t.Fatal(err)
	}
	if pixelOf(a, 20, 9) != red || pixelOf(a, 20, 0) != blue {
		t.Errorf("the flip: %v %v", pixelOf(a, 20, 9), pixelOf(a, 20, 0))
	}
	// a polygon (a triangle of the red part): only the pixels inside are copied
	if _, err := a.Retouch(0, `{"op":"clone","pts":[[0,0],[4,0],[0,4]],"tx":25,"ty":5}`); err != nil {
		t.Fatal(err)
	}
	if pixelOf(a, 25, 5) != red || pixelOf(a, 29, 9) != white {
		t.Errorf("the triangle: %v %v", pixelOf(a, 25, 5), pixelOf(a, 29, 9))
	}
	// a soft edge: the colors between the two meet at the edge of the copy
	b := testPicture(t, 40, 20, func(x, y int) color.NRGBA {
		if x < 10 {
			return red
		}
		return white
	})
	if _, err := b.Retouch(0, `{"op":"clone","x":0,"y":0,"w":10,"h":20,"tx":20,"ty":0,"feather":6}`); err != nil {
		t.Fatal(err)
	}
	if e := pixelOf(b, 20, 10); e == red || e == white {
		t.Errorf("the edge of a soft copy is a mix: %v", e)
	}
	if p := pixelOf(b, 25, 10); p.R < 240 || p.G > 60 {
		t.Errorf("the inside of the copy is nearly whole: %v", p)
	}
	// what is not allowed
	if _, err := a.Retouch(0, `{"op":"clone","x":0,"y":0,"w":50,"h":10,"tx":0,"ty":0}`); err == nil {
		t.Errorf("a box that sticks out must fail")
	}
	if _, err := a.Retouch(0, `{"op":"clone","x":0,"y":0,"w":5,"h":5,"tx":0,"ty":0,"scale":50}`); err == nil {
		t.Errorf("a scale of 50 must fail")
	}
	if _, err := a.Retouch(0, `{"op":"clone","x":0,"y":0,"w":5,"h":5,"tx":0,"ty":0,"blend":"magic"}`); err == nil {
		t.Errorf("an unknown blend must fail")
	}
	if _, err := a.Retouch(0, `{"op":"smudge"}`); err == nil {
		t.Errorf("an unknown retouch must fail")
	}
}

func TestRetouchFill(t *testing.T) {
	cA, cB := color.NRGBA{200, 30, 30, 255}, color.NRGBA{30, 30, 200, 255}
	stripe := func(x int) color.NRGBA {
		if (x/4)%2 == 0 {
			return cA
		}
		return cB
	}
	a := testPicture(t, 80, 50, func(x, y int) color.NRGBA { return stripe(x) })
	// a hole in the middle (painted green): the fill must continue the stripes
	if _, err := a.DrawShape(0, `{"kind":"rect","x0":30,"y0":15,"x1":47,"y1":32,"color":"#00ff00","fill":"fill","width":1}`); err != nil {
		t.Fatal(err)
	}
	if _, err := a.Retouch(0, `{"op":"fill","x":30,"y":15,"w":18,"h":18}`); err != nil {
		t.Fatal(err)
	}
	right, total := 0, 0
	for y := 15; y < 33; y++ {
		for x := 30; x < 48; x++ {
			p := pixelOf(a, x, y)
			if p != cA && p != cB {
				t.Fatalf("pixel %d,%d is %v: not a color of the surroundings", x, y, p)
			}
			total++
			if p == stripe(x) {
				right++
			}
		}
	}
	if right*100 < total*85 {
		t.Errorf("only %d of %d pixels continue the stripes", right, total)
	}
	if pixelOf(a, 29, 20) != stripe(29) || pixelOf(a, 48, 20) != stripe(48) {
		t.Errorf("the pixels around the hole were changed")
	}
}

func TestRetouchBlend(t *testing.T) {
	grey := func(v uint8) color.NRGBA { return color.NRGBA{v, v, v, 255} }
	inBox := func(x, y int) bool { return x >= 20 && x < 40 && y >= 10 && y < 30 }
	patch := testPicture(t, 60, 40, func(x, y int) color.NRGBA {
		if inBox(x, y) {
			return grey(180)
		}
		return grey(100)
	})
	// poisson: the brighter patch is pulled to the surroundings, also in the middle
	if _, err := patch.Retouch(0, `{"op":"blend","x":20,"y":10,"w":20,"h":20,"method":"poisson"}`); err != nil {
		t.Fatal(err)
	}
	if e := pixelOf(patch, 20, 20).R; e < 95 || e > 110 {
		t.Errorf("poisson: edge pixel %d, want about 100", e)
	}
	if m := pixelOf(patch, 30, 20).R; m < 95 || m > 125 {
		t.Errorf("poisson: middle pixel %d, want about 100", m)
	}
	if o := pixelOf(patch, 10, 20).R; o != 100 {
		t.Errorf("poisson changed the outside: %d", o)
	}
	// color_match: the average inside moves to the one around
	cm := testPicture(t, 60, 40, func(x, y int) color.NRGBA {
		if inBox(x, y) {
			return grey(200)
		}
		return grey(100)
	})
	if _, err := cm.Retouch(0, `{"op":"blend","x":20,"y":10,"w":20,"h":20,"method":"color_match","width":6}`); err != nil {
		t.Fatal(err)
	}
	if m := pixelOf(cm, 30, 20).R; m > 130 {
		t.Errorf("color_match: middle pixel %d, want near 100", m)
	}
	// smooth: a hard line becomes a ramp
	sm := testPicture(t, 40, 20, func(x, y int) color.NRGBA {
		if x < 20 {
			return grey(0)
		}
		return grey(255)
	})
	if _, err := sm.Retouch(0, `{"op":"blend","x":20,"y":0,"w":20,"h":20,"method":"smooth","width":6}`); err != nil {
		t.Fatal(err)
	}
	if e := pixelOf(sm, 19, 10).R; e == 0 {
		t.Errorf("smooth: the pixel at the line is still %d", e)
	}
	if far := pixelOf(sm, 2, 10).R; far != 0 {
		t.Errorf("smooth changed a pixel far from the line: %d", far)
	}
	if _, err := sm.Retouch(0, `{"op":"blend","x":0,"y":0,"w":5,"h":5,"method":"blur"}`); err == nil {
		t.Errorf("an unknown method must fail")
	}
}

func TestSnapshotZoomAndOutline(t *testing.T) {
	a := testPicture(t, 20, 20, func(x, y int) color.NRGBA { return color.NRGBA{uint8(x * 10), uint8(y * 10), 0, 255} })
	res, err := a.snapshot(map[string]any{"x": 2.0, "y": 3.0, "w": 10.0, "h": 8.0, "zoom": 4.0, "outline": []any{[]any{2.0, 3.0}, []any{11.0, 3.0}, []any{11.0, 10.0}}})
	if err != nil {
		t.Fatal(err)
	}
	if res["width"] != 40 || res["height"] != 32 || res["zoom"] != 4 {
		t.Errorf("zoomed: %v × %v, zoom %v", res["width"], res["height"], res["zoom"])
	}
	res, err = a.snapshot(map[string]any{"zoom": 4.0, "max": 30.0}) // the zoom stops where the answer would be over the limit
	if err != nil || res["zoom"] != 1 {
		t.Errorf("a zoom over the limit: %v %v", res["zoom"], err)
	}
}

func TestRetouchFillBigHole(t *testing.T) {
	if testing.Short() {
		t.Skip("slow")
	}
	a := testPicture(t, 1830, 1220, func(x, y int) color.NRGBA {
		v := uint8((x*7+y*13)%97 + (x/9+y/11)%53)
		return color.NRGBA{v, v / 2, 255 - v, 255}
	})
	start := time.Now()
	if _, err := a.Retouch(0, `{"op":"fill","x":300,"y":400,"w":400,"h":300}`); err != nil {
		t.Fatal(err)
	}
	t.Logf("a hole of 400 × 300 in a picture of 1830 × 1220: %v", time.Since(start))
	if time.Since(start) > 40*time.Second {
		t.Errorf("too slow: %v", time.Since(start))
	}
}

func TestCloneScaleAngleSeamless(t *testing.T) {
	red, white := color.NRGBA{255, 0, 0, 255}, color.NRGBA{255, 255, 255, 255}
	a := testPicture(t, 80, 60, func(x, y int) color.NRGBA {
		if x < 10 && y < 10 {
			return red
		}
		return white
	})
	// scaled up 3 times, the middle of the box stays where it would have been: a box of 10 × 10 at 40,20 is centred at 45,25
	if _, err := a.Retouch(0, `{"op":"clone","x":0,"y":0,"w":10,"h":10,"tx":40,"ty":20,"scale":3}`); err != nil {
		t.Fatal(err)
	}
	if pixelOf(a, 45, 25) != red || pixelOf(a, 31, 11) != red || pixelOf(a, 28, 25) != white {
		t.Errorf("scale: middle %v, near the corner %v, outside %v", pixelOf(a, 45, 25), pixelOf(a, 31, 11), pixelOf(a, 28, 25))
	}
	// turned 45 degrees: the corners of the square are cut off at the middle of the sides of the diamond
	b := testPicture(t, 80, 60, func(x, y int) color.NRGBA {
		if x < 20 && y < 20 {
			return red
		}
		return white
	})
	if _, err := b.Retouch(0, `{"op":"clone","x":0,"y":0,"w":20,"h":20,"tx":40,"ty":20,"angle":45}`); err != nil {
		t.Fatal(err)
	}
	if pixelOf(b, 50, 30) != red || pixelOf(b, 41, 21) != white { // (the corner of the box is outside the turned square)
		t.Errorf("angle: middle %v, corner of the box %v", pixelOf(b, 50, 30), pixelOf(b, 41, 21))
	}
	// seamless: a dark patch (value 40) put on a bright area (200) with a texture: the patch takes the brightness around, and keeps its own texture
	grey := func(v int) color.NRGBA { return color.NRGBA{uint8(v), uint8(v), uint8(v), 255} }
	c := testPicture(t, 100, 40, func(x, y int) color.NRGBA {
		if x < 30 {
			return grey(40 + (x/3+y/3)%2*30) // dark, with a checker
		}
		return grey(200)
	})
	if _, err := c.Retouch(0, `{"op":"clone","x":0,"y":5,"w":30,"h":30,"tx":50,"ty":5,"blend":"seamless"}`); err != nil {
		t.Fatal(err)
	}
	inside := []int{int(pixelOf(c, 55, 20).R), int(pixelOf(c, 56, 20).R), int(pixelOf(c, 70, 20).R), int(pixelOf(c, 75, 25).R)}
	for _, v := range inside {
		if v < 150 {
			t.Errorf("seamless: the patch stays dark: %v", inside)
			break
		}
	}
	lo, hi := 255, 0
	for x := 52; x < 78; x++ {
		for y := 8; y < 32; y++ {
			v := int(pixelOf(c, x, y).R)
			lo, hi = min(lo, v), max(hi, v)
		}
	}
	if hi-lo < 15 {
		t.Errorf("seamless: the texture of the patch is lost (range %d)", hi-lo)
	}
	// a plain copy shows the seam; the seamless one does not (the numbers of seam_score say so)
	d := testPicture(t, 100, 40, func(x, y int) color.NRGBA {
		if x < 30 {
			return grey(40 + (x/3+y/3)%2*30)
		}
		return grey(200)
	})
	d.Retouch(0, `{"op":"clone","x":0,"y":5,"w":30,"h":30,"tx":50,"ty":5}`)
	plain, _ := d.seamCommand(map[string]any{"x": 50.0, "y": 5.0, "w": 30.0, "h": 30.0})
	seam, _ := c.seamCommand(map[string]any{"x": 50.0, "y": 5.0, "w": 30.0, "h": 30.0})
	t.Logf("plain: %v\nseamless: %v", plain["report"], seam["report"])
	if !strings.Contains(plain["report"].(string), "obvious") && !strings.Contains(plain["report"].(string), "visible") {
		t.Errorf("a plain copy of a dark patch must show: %v", plain["report"])
	}
	if strings.Contains(seam["report"].(string), "obvious") {
		t.Errorf("the seamless copy must not be obvious: %v", seam["report"])
	}
}

func TestMasks(t *testing.T) {
	red, white := color.NRGBA{200, 20, 20, 255}, color.NRGBA{250, 250, 250, 255}
	a := testPicture(t, 40, 30, func(x, y int) color.NRGBA {
		if x >= 10 && x < 20 && y >= 10 && y < 20 { // a red square, with a white dot in it
			if x == 14 && y == 14 {
				return white
			}
			return red
		}
		return white
	})
	res, err := a.maskCommand(map[string]any{"op": "color", "x": 12.0, "y": 12.0, "tolerance": 10.0})
	if err != nil {
		t.Fatal(err)
	}
	if res["pixels"] != 100 { // the dot is a hole of the area, filled
		t.Errorf("select_by_color: %v pixels, want 100 (the hole filled)", res["pixels"])
	}
	id := res["mask"].(string)
	res, err = a.maskCommand(map[string]any{"op": "color", "x": 12.0, "y": 12.0, "tolerance": 10.0, "fill_holes": false})
	if err != nil || res["pixels"] != 99 {
		t.Errorf("without filling the hole: %v %v", res["pixels"], err)
	}
	if res, _ = a.maskCommand(map[string]any{"op": "color", "x": 0.0, "y": 0.0, "tolerance": 10.0, "grow": 2.0}); res["pixels"] == nil {
		t.Errorf("a mask of the background: %v", res)
	}
	// grow and shrink
	if res, _ = a.maskCommand(map[string]any{"op": "edit", "mask": id, "grow": 1.0}); res["pixels"] != 140 {
		t.Errorf("grown by 1: %v pixels, want 140", res["pixels"])
	}
	if res, _ = a.maskCommand(map[string]any{"op": "edit", "mask": id, "shrink": 2.0}); res["pixels"] != 64 {
		t.Errorf("shrunk by 2 from 12 × 12 (corners rounded): %v pixels", res["pixels"])
	}
	// a shape, and operations between masks
	sq, err := a.maskCommand(map[string]any{"op": "create", "x": 0.0, "y": 0.0, "w": 15.0, "h": 15.0})
	if err != nil || sq["pixels"] != 225 {
		t.Fatalf("create: %v %v", sq, err)
	}
	other, _ := a.maskCommand(map[string]any{"op": "create", "pts": [][2]int{{10, 10}, {30, 10}, {30, 25}, {10, 25}}})
	inter, _ := a.maskCommand(map[string]any{"op": "edit", "mask": sq["mask"], "intersect_mask": other["mask"]})
	if inter["pixels"] != 25 {
		t.Errorf("intersect: %v pixels, want 25", inter["pixels"])
	}
	if _, err := a.maskCommand(map[string]any{"op": "edit", "mask": "m99"}); err == nil {
		t.Errorf("a mask that does not exist must fail")
	}
	// a mask is used as the shape of a tool, and snapshot tints it
	if _, err := a.Retouch(0, `{"op":"fill","mask":"`+id+`"}`); err != nil {
		t.Fatal(err)
	}
	snap, err := a.snapshot(map[string]any{"mask": id})
	if err != nil || snap["width"] != 40 {
		t.Errorf("snapshot with a mask: %v %v", snap["width"], err)
	}
	if _, err := a.snapshot(map[string]any{"mask": "m99"}); err == nil {
		t.Errorf("a snapshot with a mask that does not exist must fail")
	}
	// a mask is for the size of the picture it was made for
	if _, err := a.ResizeCanvas(0, 50, 30, "", ""); err != nil {
		t.Fatal(err)
	}
	if _, err := a.Retouch(0, `{"op":"fill","mask":"`+id+`"}`); err == nil || !strings.Contains(err.Error(), "make it again") {
		t.Errorf("an old mask on a picture of another size: %v", err)
	}
}

func TestFillSourceAndAvoid(t *testing.T) {
	green, red := color.NRGBA{20, 200, 20, 255}, color.NRGBA{200, 20, 20, 255}
	// the left half is green, the right half red; the gap is on the border, so a plain fill would copy both
	a := testPicture(t, 80, 40, func(x, y int) color.NRGBA {
		if x < 40 {
			return green
		}
		return red
	})
	if _, err := a.DrawShape(0, `{"kind":"rect","x0":34,"y0":10,"x1":45,"y1":25,"color":"#000000","fill":"fill","width":1}`); err != nil {
		t.Fatal(err)
	}
	src, _ := a.maskCommand(map[string]any{"op": "create", "x": 0.0, "y": 0.0, "w": 30.0, "h": 40.0})
	if _, err := a.Retouch(0, `{"op":"fill","x":34,"y":10,"w":12,"h":16,"patch":5,"source_mask":"`+src["mask"].(string)+`"}`); err != nil {
		t.Fatal(err)
	}
	for y := 10; y < 26; y++ {
		for x := 34; x < 46; x++ {
			if p := pixelOf(a, x, y); p != green {
				t.Fatalf("pixel %d,%d is %v: it was copied from outside the source", x, y, p)
			}
		}
	}
	// avoid: the same, from the other side
	b := testPicture(t, 80, 40, func(x, y int) color.NRGBA {
		if x < 40 {
			return green
		}
		return red
	})
	b.DrawShape(0, `{"kind":"rect","x0":34,"y0":10,"x1":45,"y1":25,"color":"#000000","fill":"fill","width":1}`)
	av, _ := b.maskCommand(map[string]any{"op": "create", "x": 40.0, "y": 0.0, "w": 40.0, "h": 40.0})
	if _, err := b.Retouch(0, `{"op":"fill","x":34,"y":10,"w":12,"h":16,"patch":5,"avoid_mask":"`+av["mask"].(string)+`"}`); err != nil {
		t.Fatal(err)
	}
	if p := pixelOf(b, 43, 20); p != green {
		t.Errorf("avoid: pixel 43,20 is %v", p)
	}
	// a source that cannot give a whole patch
	tiny, _ := b.maskCommand(map[string]any{"op": "create", "x": 0.0, "y": 0.0, "w": 3.0, "h": 3.0})
	if _, err := b.Retouch(0, `{"op":"fill","x":34,"y":10,"w":4,"h":4,"patch":9,"source_mask":"`+tiny["mask"].(string)+`"}`); err == nil {
		t.Errorf("a source that is too small must fail")
	}
}

func TestCloneRepeatNoteAndCompare(t *testing.T) {
	a := testPicture(t, 100, 40, func(x, y int) color.NRGBA { return color.NRGBA{uint8(x), uint8(y * 3), 0, 255} })
	info, err := a.Retouch(0, `{"op":"clone","x":0,"y":0,"w":20,"h":20,"tx":40,"ty":0}`)
	if err != nil || info.Note != "" {
		t.Fatalf("the first copy: %q %v", info.Note, err)
	}
	info, err = a.Retouch(0, `{"op":"clone","x":2,"y":2,"w":18,"h":18,"tx":70,"ty":10}`)
	if err != nil || !strings.Contains(info.Note, "already copied") {
		t.Errorf("the second copy of the same place must be told: %q %v", info.Note, err)
	}
	// compare: a picture with no file has no original
	if _, err := a.snapshot(map[string]any{"compare": true}); err == nil {
		t.Errorf("compare without a file must fail")
	}
}

func TestFillPlainMethods(t *testing.T) {
	grey := func(v int) color.NRGBA { return color.NRGBA{uint8(v), uint8(v), uint8(v), 255} }
	ramp := func() *App { // a smooth ramp from left to right, and ink (black) in the middle of it
		return testPicture(t, 60, 20, func(x, y int) color.NRGBA {
			if x >= 25 && x < 35 && y >= 5 && y < 15 {
				return grey(0)
			}
			return grey(x * 4)
		})
	}
	// smooth: the ramp comes back inside
	a := ramp()
	if _, err := a.Retouch(0, `{"op":"fill","method":"smooth","x":25,"y":5,"w":10,"h":10}`); err != nil {
		t.Fatal(err)
	}
	for _, x := range []int{25, 30, 34} {
		if v := int(pixelOf(a, x, 10).R); v < x*4-12 || v > x*4+12 {
			t.Errorf("smooth: pixel %d,10 is %d, want about %d", x, v, x*4)
		}
	}
	// nearest: every pixel has the color of a neighbour on the border
	b := ramp()
	if _, err := b.Retouch(0, `{"op":"fill","method":"nearest","x":25,"y":5,"w":10,"h":10}`); err != nil {
		t.Fatal(err)
	}
	if pixelOf(b, 25, 10) != grey(24*4) || pixelOf(b, 34, 10) != grey(35*4) {
		t.Errorf("nearest: %v %v", pixelOf(b, 25, 10), pixelOf(b, 34, 10))
	}
	// average and solid: one color
	c := ramp()
	if _, err := c.Retouch(0, `{"op":"fill","method":"average","x":25,"y":5,"w":10,"h":10,"ring":4}`); err != nil {
		t.Fatal(err)
	}
	if v := pixelOf(c, 30, 10).R; v < 100 || v > 130 || pixelOf(c, 26, 7) != pixelOf(c, 33, 12) {
		t.Errorf("average: %d, flat %v", v, pixelOf(c, 26, 7) == pixelOf(c, 33, 12))
	}
	if _, err := c.Retouch(0, `{"op":"fill","method":"solid","color":"#336699","pts":[[0,0],[8,0],[0,8]]}`); err != nil {
		t.Fatal(err)
	}
	if pixelOf(c, 1, 1) != (color.NRGBA{0x33, 0x66, 0x99, 255}) || pixelOf(c, 7, 7) == (color.NRGBA{0x33, 0x66, 0x99, 255}) {
		t.Errorf("solid in a triangle: %v %v", pixelOf(c, 1, 1), pixelOf(c, 7, 7))
	}
	if _, err := c.Retouch(0, `{"op":"fill","method":"solid","x":0,"y":0,"w":3,"h":3}`); err == nil {
		t.Errorf("solid with no color must fail")
	}
	if _, err := c.Retouch(0, `{"op":"fill","method":"magic","x":0,"y":0,"w":3,"h":3}`); err == nil {
		t.Errorf("an unknown method must fail")
	}
	// grain: the surroundings have noise, the flat fill gets it too
	noisy := testPicture(t, 80, 40, func(x, y int) color.NRGBA {
		v := 128 + (x*7919+y*104729)%21 - 10
		if x >= 30 && x < 50 && y >= 10 && y < 30 {
			v = 0
		}
		return grey(v)
	})
	if _, err := noisy.Retouch(0, `{"op":"fill","method":"average","grain":true,"x":30,"y":10,"w":20,"h":20}`); err != nil {
		t.Fatal(err)
	}
	lo, hi := 255, 0
	for y := 10; y < 30; y++ {
		for x := 30; x < 50; x++ {
			v := int(pixelOf(noisy, x, y).R)
			lo, hi = min(lo, v), max(hi, v)
		}
	}
	if hi-lo < 8 {
		t.Errorf("grain: the fill is flat (range %d)", hi-lo)
	}
	// the source limits what is used: only the left half is allowed, which is dark
	d := testPicture(t, 60, 20, func(x, y int) color.NRGBA {
		if x < 30 {
			return grey(20)
		}
		return grey(220)
	})
	src, _ := d.maskCommand(map[string]any{"op": "create", "x": 0.0, "y": 0.0, "w": 30.0, "h": 20.0})
	if _, err := d.Retouch(0, `{"op":"fill","method":"smooth","x":26,"y":5,"w":8,"h":10,"source_mask":"`+src["mask"].(string)+`"}`); err != nil {
		t.Fatal(err)
	}
	if v := pixelOf(d, 33, 10).R; v > 40 {
		t.Errorf("source_mask: the fill uses pixels outside it (%d)", v)
	}
}
