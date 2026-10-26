package main

import (
	"image"
	"image/color"
	"testing"
)

func TestDrawShape(t *testing.T) {
	a := NewApp()
	if err := a.NewBlankColor(60, 40, "#FFFFFFFF"); err != nil {
		t.Fatal(err)
	}
	at := func(x, y int) color.NRGBA {
		a.mu.Lock()
		defer a.mu.Unlock()
		return a.clip.(*image.NRGBA).NRGBAAt(x, y)
	}
	red, yellow, white := color.NRGBA{255, 0, 0, 255}, color.NRGBA{255, 255, 0, 255}, color.NRGBA{255, 255, 255, 255}
	if _, err := a.DrawShape(0, `{"kind":"rect","x0":10,"y0":5,"x1":29,"y1":24,"color":"#ff0000","bg":"#ffff00","fill":"both","width":3,"style":"solid"}`); err != nil {
		t.Fatal(err)
	}
	if at(10, 5) != red || at(29, 24) != red || at(12, 14) != red || at(20, 14) != yellow || at(9, 5) != white || at(30, 24) != white {
		t.Errorf("the rectangle: corner %v, far corner %v, edge %v, inside %v, outside %v %v", at(10, 5), at(29, 24), at(12, 14), at(20, 14), at(9, 5), at(30, 24))
	}
	if _, err := a.DrawShape(0, `{"kind":"ellipse","x0":35,"y0":5,"x1":54,"y1":24,"color":"#ff0000","fill":"fill","width":3}`); err != nil {
		t.Fatal(err)
	}
	if at(45, 14) != red || at(35, 5) != white {
		t.Errorf("the ellipse: middle %v, corner of its box %v", at(45, 14), at(35, 5))
	}
	if _, err := a.DrawShape(0, `{"kind":"line","x0":0,"y0":35,"x1":59,"y1":35,"color":"#ff0000","width":1,"style":"dashed"}`); err != nil {
		t.Fatal(err)
	}
	if at(0, 35) != red || at(5, 35) != white {
		t.Errorf("the dashed line: start %v, gap %v", at(0, 35), at(5, 35))
	}
	if _, err := a.DrawShape(0, `{"kind":"blob","color":"#ff0000"}`); err == nil {
		t.Errorf("an unknown kind must fail")
	}
	if _, err := a.DrawShape(0, `{"kind":"path","pts":[[0,0]],"color":"#ff0000"}`); err == nil {
		t.Errorf("a path of one point must fail")
	}
}
