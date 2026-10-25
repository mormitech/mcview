package main

import (
	"context"
	"image"
	"os"
	"os/exec"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	wruntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

// App holds what the UI can call: the version, the window mode and the picture of the window.
type App struct {
	ctx      context.Context
	mu       sync.Mutex
	winW     int // the size the window was asked for, before it is fitted to the screen
	winH     int
	theme    string               // --theme: light or dark instead of the system's setting
	mode     string               // "" for the viewer, "about" for the About window
	image    string               // the picture of this window, as a full path
	imgSeq   int                  // counts the pictures, so that each has its own URL
	child    bool                 // this window was started with a picture (by the system, or by another window)
	pix      pixelCache           // the decoded picture for the pixel readout (see pixel.go)
	infoWins map[string]*exec.Cmd // the windows of their own that are open, by key

	clip      image.Image // a picture pasted from the clipboard, which has no file yet (image is "" then); see clip.go
	clipPNG   []byte      // the same as a PNG, for the web view
	workDir   string      // the folder of a pasted picture and the one Save As starts in: the Pictures folder, until a picture is saved elsewhere
	lastPaste time.Time   // the menu and the key can both ask for one paste
	lastNew   time.Time   // the same for a new window
	saving    bool        // the Save As dialog is open
	spawned   int         // the windows this one started, so that each opens a little further down and right
	cascade   int         // --cascade: how many steps of that this window was started at

	states    map[int]picState // every picture state this window has had, by number, for undo (see undo.go)
	stateID   int              // the number of the one it is in; 0 is the empty window
	nextState int

	openWindow func(path string) error // starts a window for a picture; spawn, except in the tests
}

func NewApp() *App {
	a := &App{}
	a.openWindow = a.spawn
	return a
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	trace("startup")
	fitWindow(ctx, a.winW, a.winH)     // every window, the small ones too
	if a.cascade > 0 && a.mode == "" { // a window started by another one opens a little down and to the right of it, not on top of it
		x, y := wruntime.WindowGetPosition(ctx)
		wruntime.WindowSetPosition(ctx, x+28*a.cascade, y+28*a.cascade)
	}
	trace("window fitted")
	// macOS starts an app hidden when it is launched with a file (Open With, double-click, `open -a`), and Wails does not unhide
	// it: the window came up about 5 seconds later, and the picture with it. Both calls are needed to show it at once (either
	// alone did not).
	a.reveal()
	a.titleWindow()
}

// reveal shows the window and the app. It is called at several steps of the start, because it is a race with macOS (see startup).
func (a *App) reveal() {
	if a.ctx == nil || a.mode != "" {
		return
	}
	wruntime.WindowShow(a.ctx)
	wruntime.Show(a.ctx)
	trace("revealed")
}

// Reveal is for the UI: it asks again while the page still finds itself hidden.
func (a *App) Reveal() { a.reveal() }

// ToggleFullscreen is the Enter key: the window goes full screen, and back. On macOS that is the system's full screen, in a Space
// of its own.
func (a *App) ToggleFullscreen() {
	if a.ctx == nil {
		return
	}
	if wruntime.WindowIsFullscreen(a.ctx) {
		wruntime.WindowUnfullscreen(a.ctx)
	} else {
		wruntime.WindowFullscreen(a.ctx)
	}
}

// CloseWindow is Cmd+W (Ctrl+W): it closes the window the way its close button does, which ends this program (one window per process).
func (a *App) CloseWindow() {
	if a.ctx != nil {
		wruntime.Quit(a.ctx)
	}
}

// IsFullscreen tells the UI whether the window is in full screen (it can also be put there with the green button of macOS).
func (a *App) IsFullscreen() bool { return a.ctx != nil && wruntime.WindowIsFullscreen(a.ctx) }

// ExitFullscreen is the Esc key: it leaves full screen, and does nothing otherwise.
func (a *App) ExitFullscreen() {
	if a.ctx != nil && wruntime.WindowIsFullscreen(a.ctx) {
		wruntime.WindowUnfullscreen(a.ctx)
	}
}

// Version is the app version from wails.json.
func (a *App) Version() string { return version() }

// Mode tells the UI which window it is in: "" for the viewer, "about" for the About window.
func (a *App) Mode() string { return a.mode }

// titleWindow names the window after its picture.
func (a *App) titleWindow() {
	if a.ctx == nil || a.mode != "" {
		return
	}
	if name := a.current().Name; name != "" {
		wruntime.WindowSetTitle(a.ctx, name+" – mcView v"+version())
	} else { // undo can take the window back to empty
		wruntime.WindowSetTitle(a.ctx, "mcView v"+version())
	}
}

func (a *App) emit(event string, data ...interface{}) {
	if a.ctx != nil {
		wruntime.EventsEmit(a.ctx, event, data...)
	}
}

// imageFilter is the file type list of the Open dialog, from imageTypes.
func imageFilter() wruntime.FileFilter {
	exts := make([]string, 0, len(imageTypes))
	for e := range imageTypes {
		exts = append(exts, "*"+e)
	}
	sort.Strings(exts)
	return wruntime.FileFilter{DisplayName: "Pictures", Pattern: strings.Join(exts, ";")}
}

// ChooseImage shows the Open dialog and opens the picture in this window. Cancelling is not an error.
func (a *App) ChooseImage() error {
	dir := a.current().Dir
	path, err := wruntime.OpenFileDialog(a.ctx, wruntime.OpenDialogOptions{
		Title:            "Open picture",
		DefaultDirectory: dir,
		Filters:          []wruntime.FileFilter{imageFilter()},
	})
	if err != nil || path == "" {
		return err
	}
	return a.setImage(path)
}

// chooseImage is File → Open…: the same as ChooseImage, with the error shown by the UI.
func (a *App) chooseImage() {
	if err := a.ChooseImage(); err != nil {
		a.emit("problem", err.Error())
	}
}

// spawn opens a window of its own: a second copy of this program, because Wails v2 has one window per process. It shows the picture
// at path, or is empty when path is "". Unlike the About window it lives on when this window is closed. Each window it starts opens
// a little further down and to the right than the one before.
func (a *App) spawn(path string) error {
	exe, err := os.Executable()
	if err != nil {
		return err
	}
	a.mu.Lock()
	a.spawned = a.spawned%8 + 1
	n := a.cascade + a.spawned
	a.mu.Unlock()
	args := []string{"--cascade", strconv.Itoa(n)}
	if a.theme != "" {
		args = append(args, "--theme", a.theme)
	}
	if path != "" {
		args = append(args, path)
	}
	cmd := exec.Command(exe, args...)
	if err := cmd.Start(); err != nil {
		return err
	}
	go cmd.Wait()
	return nil
}

// newWindow is File → New Window: an empty window, for a picture to be pasted into or opened.
func (a *App) newWindow() {
	a.mu.Lock()
	if time.Since(a.lastNew) < 400*time.Millisecond { // the menu's key and the page's key can both come for one press
		a.mu.Unlock()
		return
	}
	a.lastNew = time.Now()
	a.mu.Unlock()
	if err := a.openWindow(""); err != nil {
		a.emit("problem", err.Error())
	}
}

// NewWindow is Ctrl+N from the page (on a Mac the menu has Cmd+N; Ctrl+N works there too).
func (a *App) NewWindow() { a.newWindow() }

// showAbout opens the About window. It is closed together with the app, and is not opened twice.
func (a *App) showAbout() {
	a.mu.Lock()
	defer a.mu.Unlock()
	if c := a.infoWins["about"]; c != nil && c.ProcessState == nil {
		return
	}
	exe, err := os.Executable()
	if err != nil {
		return
	}
	args := []string{"--about"}
	if a.theme != "" {
		args = append([]string{"--theme", a.theme}, args...)
	}
	cmd := exec.Command(exe, args...)
	if cmd.Start() != nil {
		return
	}
	if a.infoWins == nil {
		a.infoWins = map[string]*exec.Cmd{}
	}
	a.infoWins["about"] = cmd
	go func() {
		cmd.Wait()
		a.mu.Lock()
		delete(a.infoWins, "about")
		a.mu.Unlock()
	}()
}

// shutdown closes the About window together with the app.
func (a *App) shutdown(ctx context.Context) {
	a.mu.Lock()
	defer a.mu.Unlock()
	for _, c := range a.infoWins {
		if c.Process != nil {
			c.Process.Kill()
		}
	}
}
