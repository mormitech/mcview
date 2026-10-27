package main

import (
	"image"
	"image/color"
	"testing"
)

func TestRoundRectAndPolygon(t *testing.T) {
	if inRoundRect(0, 0, 0, 0, 10, 10, 4) || !inRoundRect(5, 5, 0, 0, 10, 10, 4) || !inRoundRect(4, 0, 0, 0, 10, 10, 4) {
		t.Errorf("the corner is cut off, the middle and the straight edge are in")
	}
	img := image.NewNRGBA(image.Rect(0, 0, 10, 10))
	fillPolygon(img, [][2]int{{0, 0}, {9, 0}, {0, 9}}, color.NRGBA{255, 0, 0, 255})
	if img.NRGBAAt(1, 1).A == 0 || img.NRGBAAt(8, 8).A != 0 {
		t.Errorf("a triangle: inside %v, outside %v", img.NRGBAAt(1, 1), img.NRGBAAt(8, 8))
	}
	fillPolygon(img, [][2]int{{-20, -20}, {40, -20}, {40, 40}}, color.NRGBA{0, 255, 0, 255}) // hanging over the edge: no crash
}

func TestCalloutAndLayer(t *testing.T) {
	img := image.NewNRGBA(image.Rect(0, 0, 100, 100))
	s := shapeSpec{Kind: "callout", X0: 20, Y0: 10, X1: 80, Y1: 40, Radius: 8, TX: 50, TY: 70, Width: 2, Fill: "both"}
	drawCallout(img, s, color.NRGBA{0, 0, 0, 255}, color.NRGBA{255, 255, 255, 255}, 2)
	if img.NRGBAAt(50, 25) != (color.NRGBA{255, 255, 255, 255}) {
		t.Errorf("the inside is filled: %v", img.NRGBAAt(50, 25))
	}
	if img.NRGBAAt(50, 10).A == 0 || img.NRGBAAt(50, 10) != (color.NRGBA{0, 0, 0, 255}) {
		t.Errorf("the top edge is outlined: %v", img.NRGBAAt(50, 10))
	}
	if img.NRGBAAt(50, 60).A == 0 {
		t.Errorf("the tail reaches down to the tip: %v", img.NRGBAAt(50, 60))
	}
	if img.NRGBAAt(5, 5).A != 0 || img.NRGBAAt(20, 10).A != 0 {
		t.Errorf("outside, and the rounded corner, are empty")
	}
	base := image.NewNRGBA(image.Rect(0, 0, 100, 100))
	half := 0.5
	if err := layerOver(base, img, shapeSpec{Opacity: &half, Shadow: &struct {
		DX    int    `json:"dx"`
		DY    int    `json:"dy"`
		Blur  int    `json:"blur"`
		Color string `json:"color"`
	}{DX: 6, DY: 6, Blur: 4}}, 0.5); err != nil {
		t.Fatal(err)
	}
	if c := base.NRGBAAt(30, 12); c.A < 100 || c.A > 140 {
		t.Errorf("half see-through: %v", c)
	}
	if base.NRGBAAt(60, 44).A == 0 { // below the box, where only the shadow reaches
		t.Errorf("a shadow under the box")
	}
}
