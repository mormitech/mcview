package main

import (
	"bytes"
	"fmt"
	"image/png"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"golang.org/x/image/tiff"
)

// imageTypes are the file types mcView opens, with the type each is served as. The web view shows all of them itself,
// except TIFF, which Go decodes and hands over as PNG (WebKit shows TIFF, WebView2 and WebKitGTK do not). The same list is in
// wails.json (macOS), packaging/windows/installer.iss and packaging/linux/mcview.desktop; a test keeps them equal.
var imageTypes = map[string]string{
	".jpg":  "image/jpeg",
	".jpeg": "image/jpeg",
	".jpe":  "image/jpeg",
	".jfif": "image/jpeg",
	".png":  "image/png",
	".gif":  "image/gif",
	".bmp":  "image/bmp",
	".webp": "image/webp",
	".ico":  "image/x-icon",
	".svg":  "image/svg+xml",
	".avif": "image/avif",
	".tif":  "image/tiff",
	".tiff": "image/tiff",
}

func isImage(path string) bool {
	_, ok := imageTypes[strings.ToLower(filepath.Ext(path))]
	return ok
}

// fileArg is the file the system asked us to open: the first argument that is not a flag. Explorer, the Linux file managers
// and `mcview picture.jpg` pass it this way; macOS sends an open-file event instead (see onFileOpen).
func fileArg(args []string) string {
	if len(args) == 0 {
		return ""
	}
	return args[0]
}

// ImageInfo is the picture the window shows. The URL changes with every picture, so the web view never shows a cached one.
// Dir is the folder of the picture; Index (counting from 1) and Count place it among the pictures of that folder.
type ImageInfo struct {
	Name  string `json:"name"`
	Path  string `json:"path"`
	Dir   string `json:"dir"`
	URL   string `json:"url"`
	Index int    `json:"index"`
	Count int    `json:"count"`
	State int    `json:"state"` // which picture state this is, for undo (see undo.go)
}

// current is the picture of this window without its place in the folder, which needs the folder to be read.
func (a *App) current() ImageInfo {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.image == "" {
		if a.clip == nil {
			return ImageInfo{}
		}
		return ImageInfo{Name: clipboardName, Dir: a.workDirLocked(), URL: "/image/" + strconv.Itoa(a.imgSeq), State: a.stateID}
	}
	return ImageInfo{Name: filepath.Base(a.image), Path: a.image, Dir: filepath.Dir(a.image), URL: "/image/" + strconv.Itoa(a.imgSeq), State: a.stateID}
}

// Image is the picture of this window, or an empty ImageInfo.
func (a *App) Image() ImageInfo {
	trace("Image() asked")
	info := a.current()
	if info.Path == "" {
		return info
	}
	if names, err := listPictures(info.Dir); err == nil {
		info.Count = len(names)
		if i := indexOf(names, info.Name); i >= 0 {
			info.Index = i + 1
		}
	}
	return info
}

// setImage makes path the picture of this window, and tells the UI if it is already running.
func (a *App) setImage(path string) error {
	abs, err := filepath.Abs(path)
	if err != nil {
		return err
	}
	if !isImage(abs) {
		return fmt.Errorf("%s is not a picture type mcView opens", filepath.Base(abs))
	}
	if fi, err := os.Stat(abs); err != nil || fi.IsDir() {
		return fmt.Errorf("cannot open %s", filepath.Base(abs))
	}
	a.mu.Lock()
	a.image = abs
	a.clip, a.clipPNG = nil, nil
	a.imgSeq++
	a.pix = pixelCache{} // the decoded picture of the one before is not needed any more
	a.newStateLocked()
	a.mu.Unlock()
	if a.ctx != nil { // not yet during startup: the window asks for the picture itself
		a.titleWindow()
		a.emit("image", a.Image())
	}
	return nil
}

// OpenImage opens a picture in this window.
func (a *App) OpenImage(path string) error { return a.setImage(path) }

// onFileOpen is what macOS calls when a file is opened with the app: double-click, Open With, dropped on the Dock icon. It
// comes to the running copy of the app, so the first picture goes into its window and every other one into a window of its own.
//
// Care is needed: macOS also reports a file given on the command line as an open-file event, so a window started with
// `mcview picture.jpg` hears about picture.jpg again. That must neither open the same picture a second time nor start yet
// another window (it did, without end). A window that was started for a picture never starts windows itself.
func (a *App) onFileOpen(path string) {
	trace("open-file event %s", path)
	a.reveal()
	if !isImage(path) {
		return
	}
	abs, err := filepath.Abs(path)
	if err != nil {
		return
	}
	a.mu.Lock()
	current, started, pasted := a.image, a.child, a.clip != nil
	a.mu.Unlock()
	switch {
	case abs == current: // the picture this window shows already
	case (current == "" && !pasted) || started: // an empty window takes it; a pasted picture, which has no file, is not thrown away
		a.setImage(abs)
	default:
		a.openWindow(abs)
	}
}

// imageHandler serves the picture of this window, and nothing else: no other path, no other file.
func (a *App) imageHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		a.mu.Lock()
		path, seq, pasted := a.image, a.imgSeq, a.clipPNG
		a.mu.Unlock()
		if (path == "" && pasted == nil) || r.URL.Path != "/image/"+strconv.Itoa(seq) {
			http.NotFound(w, r)
			return
		}
		if path == "" { // a picture pasted from the clipboard
			w.Header().Set("Cache-Control", "no-store")
			w.Header().Set("Content-Type", "image/png")
			http.ServeContent(w, r, "", time.Time{}, bytes.NewReader(pasted))
			return
		}
		trace("serving %s", r.URL.Path)
		defer trace("served %s", r.URL.Path)
		f, err := os.Open(path)
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		defer f.Close()
		w.Header().Set("Cache-Control", "no-store")
		typ := imageTypes[strings.ToLower(filepath.Ext(path))]
		if typ == "image/tiff" {
			img, err := tiff.Decode(f)
			if err != nil {
				http.Error(w, err.Error(), http.StatusUnprocessableEntity)
				return
			}
			w.Header().Set("Content-Type", "image/png")
			png.Encode(w, img)
			return
		}
		fi, err := f.Stat()
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", typ)
		http.ServeContent(w, r, "", fi.ModTime(), f)
	})
}
