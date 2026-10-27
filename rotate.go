package main

import (
	"image"
	"image/draw"
	"math"

	xdraw "golang.org/x/image/draw"
	"golang.org/x/image/math/f64"
)

// A rotation cache for a selected part: turning it again right after a turn (while the selection is kept) starts from the picture as it was before the first
// turn and turns it by the angles added up, so the pixels are resampled once, however many times it is turned (and a turn back is exact). The whole picture
// has no such cache: it changes its size, and what is turned is always the picture as it is.
type rotCache struct {
	after int          // the state that the last turn made: the cache is good only while that is the picture
	orig  *image.NRGBA // the picture before the first turn
	key   string       // what is turned (the whole picture or a part) and how
	total float64      // degrees clockwise, all the turns added up
}

func cloneNRGBA(src *image.NRGBA) *image.NRGBA {
	dst := image.NewNRGBA(src.Bounds())
	copy(dst.Pix, src.Pix)
	return dst
}

// rotationBase says what to turn and by how much: the cached original and the angles added up when this turn continues the last one, else base itself and angle.
func (a *App) rotationBase(base *image.NRGBA, key string, angle float64, again bool) (*image.NRGBA, float64) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if c := a.rotc; again && c != nil && c.after == a.stateID && c.key == key {
		return c.orig, c.total + angle
	}
	return cloneNRGBA(base), angle
}

// keepRotation remembers where the turn that just made the current state started from.
func (a *App) keepRotation(orig *image.NRGBA, key string, total float64) {
	a.mu.Lock()
	a.rotc = &rotCache{after: a.stateID, orig: orig, key: key, total: total}
	a.mu.Unlock()
}

// rotatePart turns what is inside box of img by angle degrees clockwise, staying inside box: a quarter turn is stretched to fill the box again (a box that is not
// square), and any other angle is blown up just enough that the turned picture reaches every edge (no empty corners).
func rotatePart(img *image.NRGBA, box image.Rectangle, angle float64) {
	w, h := box.Dx(), box.Dy()
	deg := math.Mod(angle, 360)
	if deg < 0 {
		deg += 360
	}
	if w < 1 || h < 1 || deg == 0 {
		return
	}
	src := image.NewNRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		copy(src.Pix[y*src.Stride:y*src.Stride+4*w], img.Pix[(box.Min.Y+y-img.Rect.Min.Y)*img.Stride+4*(box.Min.X-img.Rect.Min.X):])
	}
	dst := image.NewNRGBA(src.Bounds())
	switch deg {
	case 90, 270:
		t := turnImage(src, int(deg/90))
		xdraw.BiLinear.Scale(dst, dst.Bounds(), t, t.Bounds(), xdraw.Src, nil)
	case 180:
		dst = turnImage(src, 2)
	default:
		rad := deg * math.Pi / 180
		cs, sn := math.Cos(rad), math.Sin(rad)
		fw, fh := float64(w), float64(h)
		s := math.Max((fw*math.Abs(cs)+fh*math.Abs(sn))/fw, (fw*math.Abs(sn)+fh*math.Abs(cs))/fh) * (1 + 2/math.Min(fw, fh))
		aff := f64.Aff3{s * cs, -s * sn, fw/2 - s*(cs*fw/2-sn*fh/2), s * sn, s * cs, fh/2 - s*(sn*fw/2+cs*fh/2)}
		xdraw.BiLinear.Transform(dst, aff, src, src.Bounds(), xdraw.Src, nil)
	}
	draw.Draw(img, box, dst, image.Point{}, draw.Src)
}
