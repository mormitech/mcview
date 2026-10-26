package main

import (
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"math"
)

// Lines, rectangles, ellipses and freehand paths put into the picture by Go, at once: the chat panel's draw tools (see mcp.go). They use no pointer and
// no overlay; they make the same shapes as the tools of the window (see drawShape in main.js), with the same round pen, dashes and fills.

type shapeSpec struct {
	Kind  string   `json:"kind"` // line, rect, ellipse, path
	X0    int      `json:"x0"`   // line, rect, ellipse: the two corners (or ends), both pixels of the picture
	Y0    int      `json:"y0"`
	X1    int      `json:"x1"`
	Y1    int      `json:"y1"`
	Pts   [][2]int `json:"pts"`   // path: the points
	Color string   `json:"color"` // #rrggbb, the outline
	Bg    string   `json:"bg"`    // #rrggbb, what fills a rect or an ellipse when Fill is "both"
	Width int      `json:"width"`
	Style string   `json:"style"` // solid, dashed, dotted
	Fill  string   `json:"fill"`  // outline, both, fill
}

// pen puts down the tip: round (a circle as wide as the width, like the pencil) or square (the corners of a rectangle).
type pen struct {
	img    *image.NRGBA
	c      color.NRGBA
	width  int
	square bool
}

func (p *pen) put(x, y int) {
	w, bb := p.width, p.img.Bounds()
	lim := (float64(w)/2 - 0.25) * (float64(w)/2 - 0.25) // the same round tip as the one the UI shows (see stampOffsets in main.js)
	mid := float64(w-1) / 2
	off := (w - 1) / 2
	if p.square {
		off = w / 2
	}
	for j := 0; j < w; j++ {
		for i := 0; i < w; i++ {
			if !p.square && (float64(i)-mid)*(float64(i)-mid)+(float64(j)-mid)*(float64(j)-mid) > lim {
				continue
			}
			if q := image.Pt(x+i-off, y+j-off); q.In(bb) {
				p.img.SetNRGBA(q.X, q.Y, p.c)
			}
		}
	}
}

// walk gives the pixels from (x0, y0) to (x1, y1), both included (Bresenham).
func walk(x0, y0, x1, y1 int, f func(x, y int)) {
	dx, dy := abs(x1-x0), -abs(y1-y0)
	sx, sy := 1, 1
	if x0 > x1 {
		sx = -1
	}
	if y0 > y1 {
		sy = -1
	}
	e := dx + dy
	for {
		f(x0, y0)
		if x0 == x1 && y0 == y1 {
			return
		}
		e2 := 2 * e
		if e2 >= dy {
			e += dy
			x0 += sx
		}
		if e2 <= dx {
			e += dx
			y0 += sy
		}
	}
}

// path draws through the points with the tip; dash is {on, off} in pixels along the path (nil: a solid line), an "on" of 0 is a dot at the start of each period.
func (p *pen) path(pts [][2]int, dash []float64) {
	if len(pts) == 0 {
		return
	}
	if len(pts) == 1 {
		p.put(pts[0][0], pts[0][1])
		return
	}
	var line [][2]int // every pixel of the way, a joint once
	for i := 1; i < len(pts); i++ {
		walk(pts[i-1][0], pts[i-1][1], pts[i][0], pts[i][1], func(x, y int) {
			if i > 1 && len(line) > 0 && line[len(line)-1] == [2]int{x, y} {
				return
			}
			line = append(line, [2]int{x, y})
		})
	}
	for d, q := range line {
		if dash == nil {
			p.put(q[0], q[1])
			continue
		}
		ph := math.Mod(float64(d), dash[0]+dash[1])
		if (dash[0] == 0 && ph < 1) || (dash[0] > 0 && ph < dash[0]) {
			p.put(q[0], q[1])
		}
	}
}

func dashPattern(w int, style string) []float64 { // the same as dashes() in main.js
	switch style {
	case "dashed":
		return []float64{math.Max(4, float64(w*3)), math.Max(3, float64(w*2))}
	case "dotted":
		return []float64{0, math.Max(3, float64(w*2))}
	}
	return nil
}

func parseRGB(s string) (color.NRGBA, error) {
	var r, g, b uint8
	if _, err := fmt.Sscanf(s, "#%02x%02x%02x", &r, &g, &b); err != nil {
		return color.NRGBA{}, fmt.Errorf("color %q is not #rrggbb", s)
	}
	return color.NRGBA{r, g, b, 255}, nil
}

// inEllipse says whether the middle of the pixel is inside the ellipse with the middle cx, cy and the radii rx, ry.
func inEllipse(x, y int, cx, cy, rx, ry float64) bool {
	if rx <= 0 || ry <= 0 {
		return false
	}
	dx, dy := (float64(x)+0.5-cx)/rx, (float64(y)+0.5-cy)/ry
	return dx*dx+dy*dy <= 1
}

// DrawShape puts a shape into the picture of this window (as shown, turned rot quarters) and makes one step of the history. spec is JSON (shapeSpec).
func (a *App) DrawShape(rot int, spec string) (ImageInfo, error) {
	var s shapeSpec
	if err := json.Unmarshal([]byte(spec), &s); err != nil {
		return ImageInfo{}, fmt.Errorf("the shape is not understood: %v", err)
	}
	c1, err := parseRGB(s.Color)
	if err != nil {
		return ImageInfo{}, err
	}
	c2 := c1
	if s.Bg != "" {
		if c2, err = parseRGB(s.Bg); err != nil {
			return ImageInfo{}, err
		}
	}
	base, err := a.shownPixels(rot)
	if err != nil {
		return ImageInfo{}, err
	}
	if base == nil {
		return ImageInfo{}, fmt.Errorf("there is no picture in this window")
	}
	bb := base.Bounds()
	w := max(1, min(s.Width, 200))
	dash := dashPattern(w, s.Style)
	switch s.Kind {
	case "path":
		if len(s.Pts) < 2 {
			return ImageInfo{}, fmt.Errorf("a path needs two points or more")
		}
		(&pen{img: base, c: c1, width: w}).path(s.Pts, nil)
	case "line":
		(&pen{img: base, c: c1, width: w}).path([][2]int{{s.X0, s.Y0}, {s.X1, s.Y1}}, dash)
	case "rect", "ellipse":
		xa, ya := min(s.X0, s.X1), min(s.Y0, s.Y1)
		W, H := abs(s.X1-s.X0)+1, abs(s.Y1-s.Y0)+1
		round := s.Kind == "ellipse"
		cx, cy, rx, ry := float64(xa)+float64(W)/2, float64(ya)+float64(H)/2, float64(W)/2, float64(H)/2
		area := func(c color.NRGBA, in func(x, y int) bool) {
			for y := max(ya, bb.Min.Y); y < min(ya+H, bb.Max.Y); y++ {
				for x := max(xa, bb.Min.X); x < min(xa+W, bb.Max.X); x++ {
					if in(x, y) {
						base.SetNRGBA(x, y, c)
					}
				}
			}
		}
		whole := func(x, y int) bool { return !round || inEllipse(x, y, cx, cy, rx, ry) }
		switch {
		case s.Fill == "fill" || W <= w || H <= w: // (a shape that is all outline is a filled one)
			area(c1, whole)
		default:
			if s.Fill == "both" {
				area(c2, whole)
			}
			switch {
			case !round && dash == nil: // the outline is inside the box
				ring := func(x, y int) bool { return x < xa+w || x >= xa+W-w || y < ya+w || y >= ya+H-w }
				area(c1, ring)
			case !round:
				o := w / 2
				p := &pen{img: base, c: c1, width: w, square: s.Style == "dashed"}
				x0, y0, x1, y1 := xa+o, ya+o, xa+W-w+o, ya+H-w+o
				p.path([][2]int{{x0, y0}, {x1, y0}, {x1, y1}, {x0, y1}, {x0, y0}}, dash)
			case dash == nil:
				area(c1, func(x, y int) bool {
					return inEllipse(x, y, cx, cy, rx, ry) && !inEllipse(x, y, cx, cy, rx-float64(w), ry-float64(w))
				})
			default: // the middle of the outline, as many points as the ellipse is long
				mrx, mry := rx-float64(w)/2, ry-float64(w)/2
				n := max(16, int(2*math.Pi*math.Max(mrx, mry)))
				pts := make([][2]int, 0, n+1)
				for i := 0; i <= n; i++ {
					t := 2 * math.Pi * float64(i) / float64(n)
					pts = append(pts, [2]int{int(math.Floor(cx + mrx*math.Cos(t))), int(math.Floor(cy + mry*math.Sin(t)))})
				}
				(&pen{img: base, c: c1, width: w, square: s.Style == "dashed"}).path(pts, dash)
			}
		}
	default:
		return ImageInfo{}, fmt.Errorf("a shape is a line, a rect, an ellipse or a path, not %q", s.Kind)
	}
	if err := a.setPixelsEmit(base, a.current().Path, false); err != nil {
		return ImageInfo{}, err
	}
	return a.Image(), nil
}
