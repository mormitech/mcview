package main

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image"
	"image/color"
	"image/color/palette"
	"image/draw"
	"image/gif"
	"image/png"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	wruntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

// The GIF frame editor. An animated GIF that is open (unchanged) becomes a session: its frames are decoded and composed (each one the whole picture as it is at that
// moment), the page shows them with a player, and what is drawn goes into the frame that is shown. The frames are kept as a list of pictures that are never
// changed (an edit puts a new picture in the place of the frame), and every state of the undo history keeps its own list, so undo and redo bring the frames back.
// Saving offers the whole animation or only the frame that is shown.

const maxGifPixels = 25_000_000 // all frames together

type gifSession struct {
	path   string         // the file, and the picture of the window as long as the window shows it
	w, h   int            // every frame has this size: a change of the size (resize, crop, canvas) ends the session
	orig   []*image.NRGBA // the frames of the file
	frames []*image.NRGBA // the frames now
	delays []int          // milliseconds
	cur    int            // the frame that is shown (and that the tools work on)
}

// GifMeta tells the page whether the picture is an animation it can show frame by frame, and where it is.
type GifMeta struct {
	Frames int      `json:"frames"` // 0: not an animated GIF (or too big)
	Delays []int    `json:"delays"` // milliseconds of each frame
	Keys   []string `json:"keys"`   // a frame that is changed gets a new key (the page keeps the pictures it has made by key)
	Cur    int      `json:"cur"`    // the frame the tools work on
	Fresh  bool     `json:"fresh"`  // the session has just begun
	Edited bool     `json:"edited"` // a frame was changed
}

// gifSnapshotLocked is the frames as they are now, for the state the window gets (App.mu held); nil without a session.
func (a *App) gifSnapshotLocked() []*image.NRGBA {
	g := a.gifs
	if g == nil || g.path != a.image {
		return nil
	}
	return append([]*image.NRGBA(nil), g.frames...)
}

// gifNoteEditLocked is called when the window gets new pixels (App.mu held): they are the new picture of the frame that is shown; a picture of another size ends the session.
func (a *App) gifNoteEditLocked(img *image.NRGBA, file string) {
	g := a.gifs
	if g == nil || file == "" || file != g.path {
		return
	}
	if b := img.Bounds(); b.Dx() != g.w || b.Dy() != g.h {
		a.gifs = nil
		return
	}
	frames := append([]*image.NRGBA(nil), g.frames...)
	frames[g.cur] = img
	g.frames = frames
}

// gifOpen decodes and composes the frames of a GIF file; nil when it has one frame only, or is too big.
func gifOpen(path string) (*gifSession, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	g, err := gif.DecodeAll(f)
	if err != nil {
		return nil, err
	}
	w, h := g.Config.Width, g.Config.Height
	if len(g.Image) < 2 || w < 1 || h < 1 || w*h*len(g.Image) > maxGifPixels {
		return nil, nil
	}
	canvas := image.NewNRGBA(image.Rect(0, 0, w, h))
	s := &gifSession{path: path, w: w, h: h}
	for i, fr := range g.Image {
		var before *image.NRGBA
		if i < len(g.Disposal) && g.Disposal[i] == gif.DisposalPrevious {
			before = cloneNRGBA(canvas)
		}
		draw.Draw(canvas, fr.Bounds(), fr, fr.Bounds().Min, draw.Over)
		s.orig = append(s.orig, cloneNRGBA(canvas))
		d := 100
		if i < len(g.Delay) && g.Delay[i] > 1 { // (a delay of 0 or 1 hundredth is shown as 100 ms, as the browsers do)
			d = g.Delay[i] * 10
		}
		s.delays = append(s.delays, d)
		if i < len(g.Disposal) {
			switch g.Disposal[i] {
			case gif.DisposalBackground:
				draw.Draw(canvas, fr.Bounds(), image.Transparent, image.Point{}, draw.Src)
			case gif.DisposalPrevious:
				canvas = before
			}
		}
	}
	s.frames = append([]*image.NRGBA(nil), s.orig...)
	return s, nil
}

// gifSession is the session of the picture of the window: the one that is going (it goes on while the window shows its file), or a new one when the window shows
// an unchanged GIF file. nil when the picture is not an animation.
func (a *App) gifSession() (*gifSession, bool, error) {
	a.mu.Lock()
	g, image := a.gifs, a.image
	a.mu.Unlock()
	if g != nil && g.path == image {
		return g, false, nil
	}
	cur := a.current()
	if cur.URL == "" || cur.Path == "" || cur.Modified || strings.ToLower(filepath.Ext(cur.Path)) != ".gif" {
		return nil, false, nil
	}
	s, err := gifOpen(cur.Path)
	if err != nil || s == nil {
		return nil, false, err
	}
	a.mu.Lock()
	if a.image == cur.Path {
		a.gifs = s
	}
	a.mu.Unlock()
	return s, true, nil
}

// GifInfo says how many frames the picture has and how long each one is shown (0 frames when it is not an animation that can be stepped through).
func (a *App) GifInfo() (GifMeta, error) {
	g, fresh, err := a.gifSession()
	if err != nil || g == nil {
		return GifMeta{}, err
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	m := GifMeta{Frames: len(g.frames), Delays: g.delays, Cur: g.cur, Fresh: fresh}
	for i, f := range g.frames {
		m.Keys = append(m.Keys, fmt.Sprintf("%p", f))
		m.Edited = m.Edited || f != g.orig[i]
	}
	return m, nil
}

// GifFrame is frame i as a PNG data URL.
func (a *App) GifFrame(i int) (string, error) {
	g, _, err := a.gifSession()
	if err != nil {
		return "", err
	}
	a.mu.Lock()
	var fr *image.NRGBA
	if g != nil && i >= 0 && i < len(g.frames) {
		fr = g.frames[i]
	}
	a.mu.Unlock()
	if fr == nil {
		return "", fmt.Errorf("there is no frame %d", i)
	}
	var buf bytes.Buffer
	if err := (&png.Encoder{CompressionLevel: png.BestSpeed}).Encode(&buf, fr); err != nil {
		return "", err
	}
	return "data:image/png;base64," + base64.StdEncoding.EncodeToString(buf.Bytes()), nil
}

// SetGifFrame makes frame i the one that the pixel tools read and an edit goes into (the page shows it paused); the picture and the history do not change.
func (a *App) SetGifFrame(i int) error {
	g, _, err := a.gifSession()
	if err != nil {
		return err
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	if g == nil || i < 0 || i >= len(g.frames) {
		return fmt.Errorf("there is no frame %d", i)
	}
	g.cur = i
	a.pix = pixelCache{seq: a.imgSeq, img: g.frames[i], orient: 1}
	return nil
}

// restoreGifLocked brings the frames of a state back (App.mu held): the picture of the window is what the player shows, the frame that is shown.
func (a *App) restoreGifLocked(st picState) {
	g := a.gifs
	if g == nil || g.path != st.image {
		return
	}
	if st.gif != nil && len(st.gif) == len(g.orig) {
		g.frames = append([]*image.NRGBA(nil), st.gif...)
	} else {
		g.frames = append([]*image.NRGBA(nil), g.orig...) // the state from before the first change
	}
	a.pix = pixelCache{seq: a.imgSeq, img: g.frames[g.cur], orient: 1}
}

// ---- saving ----

type gifSaveChoice int

const (
	gifSaveCancel gifSaveChoice = iota
	gifSaveAll                  // the whole animation
	gifSaveFrame                // the frame that is shown, as a picture
)

// gifSaving is whether the window has a GIF session with a frame list to save (the session of the picture that is shown).
func (a *App) gifSaving() *gifSession {
	a.mu.Lock()
	defer a.mu.Unlock()
	if g := a.gifs; g != nil && g.path == a.image {
		return g
	}
	return nil
}

// askGifSave asks what to save when the picture is an animation: the whole GIF or the frame that is shown.
func (a *App) askGifSave(frame, count int) gifSaveChoice {
	if v, ok := a.mockAnswer("gifsave", fmt.Sprint(frame+1)); ok { // control mode: all, frame or cancel
		switch v {
		case "all":
			return gifSaveAll
		case "frame":
			return gifSaveFrame
		}
		return gifSaveCancel
	}
	if a.confirmGif != nil {
		return a.confirmGif(frame, count)
	}
	if a.ctx == nil {
		return gifSaveAll
	}
	text := fmt.Sprintf("This is an animation of %d frames; frame %d is shown.\n\nSave the whole animation, or only this frame?", count, frame+1)
	switch runtime.GOOS {
	case "windows":
		switch windowsDiscardChoice("Save animation", text+"\n\nYes: the whole animation   No: only this frame") {
		case choiceSave:
			return gifSaveAll
		case choiceDiscard:
			return gifSaveFrame
		}
		return gifSaveCancel
	case "darwin":
		res, err := wruntime.MessageDialog(a.ctx, wruntime.MessageDialogOptions{
			Type: wruntime.QuestionDialog, Title: "Save animation", Message: text,
			Buttons: []string{"Whole animation", "This frame", "Cancel"}, DefaultButton: "Whole animation", CancelButton: "Cancel",
		})
		switch {
		case err != nil:
			return gifSaveCancel
		case res == "Whole animation":
			return gifSaveAll
		case res == "This frame":
			return gifSaveFrame
		}
		return gifSaveCancel
	}
	res, err := wruntime.MessageDialog(a.ctx, wruntime.MessageDialogOptions{Type: wruntime.QuestionDialog, Title: "Save animation", Message: "Save the whole animation? (No: only this frame)", DefaultButton: "Yes"})
	switch {
	case err != nil:
		return gifSaveCancel
	case res == "Yes":
		return gifSaveAll
	}
	return gifSaveFrame
}

// encodeGifAnimation writes frames (all of one size) as an animated GIF that repeats for ever; delays are milliseconds.
func encodeGifAnimation(w io.Writer, frames []*image.NRGBA, delays []int) error {
	g := &gif.GIF{LoopCount: 0}
	for i, f := range frames {
		opaque := f.Opaque()
		var pal color.Palette = palette.Plan9
		disposal := byte(gif.DisposalNone)
		if !opaque { // a transparent color of its own, and each frame is cleared before the next (the frames are whole pictures)
			pal = transparentPlan9{}.Quantize(nil, f)
			disposal = gif.DisposalBackground
		}
		p := image.NewPaletted(f.Bounds(), pal)
		draw.FloydSteinberg.Draw(p, f.Bounds(), f, image.Point{})
		g.Image = append(g.Image, p)
		d := 10
		if i < len(delays) {
			d = max(2, delays[i]/10)
		}
		g.Delay = append(g.Delay, d)
		g.Disposal = append(g.Disposal, disposal)
	}
	return gif.EncodeAll(w, g)
}
