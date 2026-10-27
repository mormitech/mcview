package main

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"

	wruntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

// The place and size of the window are kept in a small file of the user's settings folder, so that the next start comes up the way the
// window was left. A window that is full screen, maximised or minimised is not remembered: that is not where the user put it.

type winState struct {
	X int `json:"x"`
	Y int `json:"y"`
	W int `json:"w"`
	H int `json:"h"`
}

const minWinW, minWinH = 300, 200

func winStatePath() string {
	if dir := configDir(); dir != "" {
		return filepath.Join(dir, "window.json")
	}
	return ""
}

// loadWinState is the window as it was left, and whether there is one.
func loadWinState() (winState, bool) {
	var st winState
	path := winStatePath()
	if path == "" {
		return st, false
	}
	b, err := os.ReadFile(path)
	if err != nil || json.Unmarshal(b, &st) != nil || st.W < minWinW || st.H < minWinH {
		return winState{}, false
	}
	return st, true
}

func saveWinState(st winState) {
	path := winStatePath()
	if path == "" || st.W < minWinW || st.H < minWinH {
		return
	}
	b, err := json.Marshal(st)
	if err != nil || os.MkdirAll(filepath.Dir(path), 0o755) != nil {
		return
	}
	tmp := path + ".tmp"
	if os.WriteFile(tmp, b, 0o644) == nil {
		os.Rename(tmp, path)
	}
}

// onScreen is whether a window at x, y with this size can be reached on screens that are totalW wide (a screen that is gone since the
// last start must not leave the window out of reach).
func onScreen(st winState, totalW int) bool {
	return st.Y >= 0 && st.X+st.W > 100 && st.X < totalW-100
}

// SaveWindow remembers the place and size of the window (the UI calls it after the window was resized; it is also done when the window closes).
func (a *App) SaveWindow() {
	if a.ctx == nil || a.mode != "" {
		return
	}
	if wruntime.WindowIsFullscreen(a.ctx) || wruntime.WindowIsMaximised(a.ctx) || wruntime.WindowIsMinimised(a.ctx) {
		return
	}
	x, y := wruntime.WindowGetPosition(a.ctx)
	w, h := wruntime.WindowGetSize(a.ctx)
	saveWinState(winState{X: x, Y: y, W: w, H: h})
}

// restorePosition puts the window where it was left, when that is still on a screen.
func restorePosition(ctx context.Context, st winState) {
	total := 0
	if screens, err := wruntime.ScreenGetAll(ctx); err == nil {
		for _, s := range screens {
			total += s.Size.Width
		}
	}
	if total > 0 && !onScreen(st, total) {
		return
	}
	wruntime.WindowSetPosition(ctx, st.X, st.Y)
}
