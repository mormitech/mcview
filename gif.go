package main

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image"
	"image/draw"
	"image/gif"
	"image/png"
	"os"
	"path/filepath"
	"strings"
)

// Animated GIFs: the frames (each one the whole picture as it is at that moment, with the disposal of the frames before it applied) for the frame player of the page.

const maxGifPixels = 25_000_000 // all frames together

type gifSet struct {
	seq    int // the picture of the window they belong to
	frames []*image.NRGBA
	delays []int // milliseconds
}

// GifMeta tells the page whether the picture is an animation it can show frame by frame.
type GifMeta struct {
	Frames int   `json:"frames"` // 0: not an animated GIF (or one that was changed, or too big)
	Delays []int `json:"delays"` // milliseconds of each frame
}

// gifFrames decodes the GIF of the window (cached while the picture stays); nil when the picture is not an unchanged GIF file with more than one frame.
func (a *App) gifFrames() (*gifSet, error) {
	a.mu.Lock()
	seq, set := a.imgSeq, a.gifs
	a.mu.Unlock()
	if set != nil && set.seq == seq {
		return set, nil
	}
	cur := a.current()
	if cur.URL == "" || cur.Path == "" || cur.Modified || strings.ToLower(filepath.Ext(cur.Path)) != ".gif" {
		return nil, nil
	}
	f, err := os.Open(cur.Path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	g, err := gif.DecodeAll(f)
	if err != nil {
		return nil, err
	}
	if len(g.Image) < 2 {
		return nil, nil
	}
	w, h := g.Config.Width, g.Config.Height
	if w < 1 || h < 1 || w*h*len(g.Image) > maxGifPixels {
		return nil, nil
	}
	canvas := image.NewNRGBA(image.Rect(0, 0, w, h))
	set = &gifSet{seq: seq}
	for i, fr := range g.Image {
		var before *image.NRGBA
		if i < len(g.Disposal) && g.Disposal[i] == gif.DisposalPrevious {
			before = cloneNRGBA(canvas)
		}
		draw.Draw(canvas, fr.Bounds(), fr, fr.Bounds().Min, draw.Over)
		set.frames = append(set.frames, cloneNRGBA(canvas))
		d := 100
		if i < len(g.Delay) && g.Delay[i] > 1 { // (a delay of 0 or 1 hundredth is shown as 100 ms, as the browsers do)
			d = g.Delay[i] * 10
		}
		set.delays = append(set.delays, d)
		if i < len(g.Disposal) {
			switch g.Disposal[i] {
			case gif.DisposalBackground:
				draw.Draw(canvas, fr.Bounds(), image.Transparent, image.Point{}, draw.Src)
			case gif.DisposalPrevious:
				canvas = before
			}
		}
	}
	a.mu.Lock()
	if a.imgSeq == seq {
		a.gifs = set
	}
	a.mu.Unlock()
	return set, nil
}

// GifInfo says how many frames the picture has and how long each one is shown (0 frames when it is not an animation that can be stepped through).
func (a *App) GifInfo() (GifMeta, error) {
	set, err := a.gifFrames()
	if err != nil || set == nil {
		return GifMeta{}, err
	}
	return GifMeta{Frames: len(set.frames), Delays: set.delays}, nil
}

// GifFrame is frame i as a PNG data URL.
func (a *App) GifFrame(i int) (string, error) {
	set, err := a.gifFrames()
	if err != nil {
		return "", err
	}
	if set == nil || i < 0 || i >= len(set.frames) {
		return "", fmt.Errorf("there is no frame %d", i)
	}
	var buf bytes.Buffer
	if err := (&png.Encoder{CompressionLevel: png.BestSpeed}).Encode(&buf, set.frames[i]); err != nil {
		return "", err
	}
	return "data:image/png;base64," + base64.StdEncoding.EncodeToString(buf.Bytes()), nil
}

// SetGifFrame makes frame i the one that the pixel tools read and an edit starts from (the page shows it paused); nothing else changes.
func (a *App) SetGifFrame(i int) error {
	set, err := a.gifFrames()
	if err != nil {
		return err
	}
	if set == nil || i < 0 || i >= len(set.frames) {
		return fmt.Errorf("there is no frame %d", i)
	}
	a.mu.Lock()
	a.pix = pixelCache{seq: a.imgSeq, img: cloneNRGBA(set.frames[i]), orient: 1}
	a.mu.Unlock()
	return nil
}
