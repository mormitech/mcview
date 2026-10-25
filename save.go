package main

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/gif"
	"image/jpeg"
	"image/png"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"golang.org/x/image/bmp"
	"golang.org/x/image/tiff"

	wruntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

// saveFormats are the types Save As writes, by extension, with the group each belongs to (jpg, jpeg and so on are one type).
var saveFormats = map[string]string{
	".png": "png", ".jpg": "jpeg", ".jpeg": "jpeg", ".jpe": "jpeg", ".jfif": "jpeg", ".gif": "gif", ".bmp": "bmp", ".tif": "tiff", ".tiff": "tiff",
}

// SaveAs is the S key (and File → Save As…): it asks where to save the picture, and saves it there, in the type the name ends with
// (PNG, JPEG, GIF, BMP or TIFF; a name without one gets .png). A picture that was pasted becomes the file it was saved to. Saving to
// another folder makes that folder the one this window works in: the arrow keys go through its pictures from then on. The text
// returned is for the UI to show; it is empty when the dialog was cancelled.
func (a *App) SaveAs() (string, error) {
	a.mu.Lock()
	if a.saving {
		a.mu.Unlock()
		return "", nil
	}
	a.saving = true
	a.mu.Unlock()
	defer func() { a.mu.Lock(); a.saving = false; a.mu.Unlock() }()

	cur := a.current()
	if cur.URL == "" {
		return "", nil
	}
	name := cur.Name
	if cur.Path == "" {
		name = "clipboard-" + time.Now().Format("20060102-150405") + ".png"
	}
	target, err := wruntime.SaveFileDialog(a.ctx, wruntime.SaveDialogOptions{
		Title:                "Save picture as",
		DefaultDirectory:     cur.Dir,
		DefaultFilename:      name,
		CanCreateDirectories: true,
		Filters:              []wruntime.FileFilter{{DisplayName: "PNG, JPEG, GIF, BMP, TIFF", Pattern: "*.png;*.jpg;*.jpeg;*.gif;*.bmp;*.tif;*.tiff"}},
	})
	if err != nil || target == "" {
		return "", err
	}
	return a.saveTo(target)
}

// saveAs is File → Save As…: the same as SaveAs, with the result shown by the UI.
func (a *App) saveAs() {
	msg, err := a.SaveAs()
	switch {
	case err != nil:
		a.emit("problem", err.Error())
	case msg != "":
		a.emit("note", msg)
	}
}

func samePath(a, b string) bool {
	return filepath.Clean(a) == filepath.Clean(b) // two names for one file are not told apart; saving over it is then refused by the rename check
}

// saveTo writes the picture of this window to target and, when that is a new folder (or the picture had no file), makes the saved file
// the picture of the window.
func (a *App) saveTo(target string) (string, error) {
	ext := strings.ToLower(filepath.Ext(target))
	if ext == "" {
		target, ext = target+".png", ".png"
	}
	format, ok := saveFormats[ext]
	if !ok {
		return "", fmt.Errorf("mcView saves PNG, JPEG, GIF, BMP and TIFF; %s is none of them", ext)
	}
	cur := a.current()
	if cur.URL == "" {
		return "", nil
	}
	if cur.Path != "" && samePath(cur.Path, target) {
		return cur.Name + " is this picture already", nil
	}
	tmp, err := os.CreateTemp(filepath.Dir(target), ".mcview-save-*")
	if err != nil {
		return "", err
	}
	defer os.Remove(tmp.Name()) // nothing is left behind when it did not work, and nothing when it is renamed
	err = a.writePicture(tmp, cur, format)
	if cerr := tmp.Close(); err == nil {
		err = cerr
	}
	if err != nil {
		return "", err
	}
	if err := os.Rename(tmp.Name(), target); err != nil {
		return "", err
	}
	os.Chmod(target, 0o644)

	dir := filepath.Dir(target)
	if cur.Path == "" || !samePath(filepath.Dir(cur.Path), dir) {
		a.mu.Lock()
		a.workDir = dir
		a.mu.Unlock()
		if err := a.setImage(target); err != nil {
			return "", err
		}
		return fmt.Sprintf("Saved %s; %s is the folder now", filepath.Base(target), dir), nil
	}
	return "Saved " + filepath.Base(target), nil
}

// writePicture writes the picture to f in the given format. A file that is of that type already is copied as it is, which keeps
// everything in it (EXIF, profiles, animation); anything else is decoded and written again, turned the way it is shown.
func (a *App) writePicture(f *os.File, cur ImageInfo, format string) error {
	if cur.Path != "" && saveFormats[strings.ToLower(filepath.Ext(cur.Path))] == format {
		src, err := os.Open(cur.Path)
		if err != nil {
			return err
		}
		defer src.Close()
		_, err = io.Copy(f, src)
		return err
	}
	pc, source, err := a.decoded()
	if err != nil {
		return err
	}
	if source != "go" {
		return fmt.Errorf("mcView cannot convert %s to another type", cur.Name)
	}
	return encodePicture(f, shownImage(pc), format)
}

// shownImage is the picture the way it is shown, with the EXIF orientation applied.
func shownImage(pc pixelCache) image.Image {
	if pc.orient == 1 {
		return pc.img
	}
	b := pc.img.Bounds()
	_, _, dw, dh := displayToSource(pc.orient, b.Dx(), b.Dy(), 0, 0)
	n := image.NewNRGBA(image.Rect(0, 0, dw, dh))
	for y := 0; y < dh; y++ {
		for x := 0; x < dw; x++ {
			sx, sy, _, _ := displayToSource(pc.orient, b.Dx(), b.Dy(), x, y)
			n.SetNRGBA(x, y, color.NRGBAModel.Convert(pc.img.At(b.Min.X+sx, b.Min.Y+sy)).(color.NRGBA))
		}
	}
	return n
}

func encodePicture(w io.Writer, img image.Image, format string) error {
	switch format {
	case "png":
		return png.Encode(w, img)
	case "jpeg": // no transparency in JPEG: a transparent pixel becomes white
		b := img.Bounds()
		flat := image.NewRGBA(b)
		draw.Draw(flat, b, image.White, image.Point{}, draw.Src)
		draw.Draw(flat, b, img, b.Min, draw.Over)
		return jpeg.Encode(w, flat, &jpeg.Options{Quality: 95})
	case "gif":
		return gif.Encode(w, img, &gif.Options{NumColors: 256, Drawer: draw.FloydSteinberg})
	case "bmp":
		return bmp.Encode(w, img)
	case "tiff":
		return tiff.Encode(w, img, &tiff.Options{Compression: tiff.Deflate, Predictor: true})
	}
	return fmt.Errorf("cannot write %s", format)
}
