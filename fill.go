package main

import (
	"fmt"
	"image"
	"image/color"
)

// FloodFill is the paint bucket: the pixels around x, y (as the picture is shown, rot quarter turns clockwise) that have the color of that pixel,
// give or take tolerance (0 to 255, for the biggest difference of a channel, alpha too), and are connected to it, get the color col (#rrggbb).
// It is a step of the history. When nothing would change (the pixel has the color already), no step is made and the state number stays.
func (a *App) FloodFill(rot, x, y int, col string, tolerance int) (ImageInfo, error) {
	var r, g, b uint8
	if _, err := fmt.Sscanf(col, "#%02x%02x%02x", &r, &g, &b); err != nil {
		return ImageInfo{}, fmt.Errorf("color %q is not #rrggbb", col)
	}
	base, err := a.shownPixels(rot)
	if err != nil {
		return ImageInfo{}, err
	}
	bb := base.Bounds()
	w, h := bb.Dx(), bb.Dy()
	if x < 0 || y < 0 || x >= w || y >= h {
		return a.Image(), nil
	}
	fill := color.NRGBA{r, g, b, 255}
	seed := base.NRGBAAt(x, y)
	if seed == fill {
		return a.Image(), nil
	}
	tol := max(0, min(tolerance, 255))
	near := func(c color.NRGBA) bool {
		return abs(int(c.R)-int(seed.R)) <= tol && abs(int(c.G)-int(seed.G)) <= tol && abs(int(c.B)-int(seed.B)) <= tol && abs(int(c.A)-int(seed.A)) <= tol
	}
	done := make([]bool, w*h) // the pixels that are filled, so that none is looked at twice
	open := func(px, py int) bool { return !done[py*w+px] && near(base.NRGBAAt(px, py)) }
	stack := []image.Point{{X: x, Y: y}}
	for len(stack) > 0 {
		p := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		if !open(p.X, p.Y) {
			continue
		}
		lx, rx := p.X, p.X // the row of this pixel, as far as it goes both ways
		for lx > 0 && open(lx-1, p.Y) {
			lx--
		}
		for rx < w-1 && open(rx+1, p.Y) {
			rx++
		}
		above, below := false, false // a run of the rows beside it is one start, not one for each pixel
		for i := lx; i <= rx; i++ {
			done[p.Y*w+i] = true
			base.SetNRGBA(i, p.Y, fill)
			for _, row := range []struct {
				dy   int
				seen *bool
			}{{-1, &above}, {1, &below}} {
				ny := p.Y + row.dy
				if ny < 0 || ny >= h {
					continue
				}
				if open(i, ny) {
					if !*row.seen {
						stack = append(stack, image.Point{X: i, Y: ny})
						*row.seen = true
					}
				} else {
					*row.seen = false
				}
			}
		}
	}
	if err := a.setPixelsEmit(base, a.current().Path, false); err != nil {
		return ImageInfo{}, err
	}
	return a.Image(), nil
}
