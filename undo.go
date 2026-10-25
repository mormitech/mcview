package main

import (
	"fmt"
	"image"
)

// Undo. The history itself (what the picture was, the mode, the pixelated view, the turn) is kept by the UI, which knows the view.
// What only this side has is the picture: a file, or a picture pasted from the clipboard that is nowhere else. Every time the picture
// of the window changes, it gets a number here (ImageInfo.State) and what is needed to show it again is kept; the UI keeps the numbers
// in its history and asks for one with RestoreState. State 0 is the window with no picture.

// picState is a picture of the window as it was: everything that setImage and setClip change.
type picState struct {
	image   string
	clip    image.Image
	clipPNG []byte
	workDir string
}

// newStateLocked numbers the picture the window has now, App.mu held.
func (a *App) newStateLocked() {
	a.statesLocked()
	a.nextState++
	a.stateID = a.nextState
	a.states[a.stateID] = picState{image: a.image, clip: a.clip, clipPNG: a.clipPNG, workDir: a.workDir}
}

func (a *App) statesLocked() map[int]picState {
	if a.states == nil {
		a.states = map[int]picState{0: {workDir: a.workDir}}
	}
	return a.states
}

// RestoreState makes the picture state n the one of this window again, and returns it. The UI shows it (nothing is emitted).
func (a *App) RestoreState(n int) (ImageInfo, error) {
	a.mu.Lock()
	st, ok := a.statesLocked()[n]
	if !ok {
		a.mu.Unlock()
		return ImageInfo{}, fmt.Errorf("that picture is not remembered any more")
	}
	a.image, a.clip, a.clipPNG, a.workDir = st.image, st.clip, st.clipPNG, st.workDir
	a.stateID = n
	a.imgSeq++ // its own URL, so that the web view loads it again
	if st.clip != nil {
		a.pix = pixelCache{seq: a.imgSeq, img: st.clip, orient: 1}
	} else {
		a.pix = pixelCache{}
	}
	a.mu.Unlock()
	a.titleWindow()
	return a.Image(), nil
}

// KeepStates forgets the picture states that are not in n (and not the current one), so that a long history does not keep
// every picture that was pasted in memory.
func (a *App) KeepStates(n []int) {
	a.mu.Lock()
	defer a.mu.Unlock()
	keep := map[int]bool{0: true, a.stateID: true}
	for _, i := range n {
		keep[i] = true
	}
	for i := range a.statesLocked() {
		if !keep[i] {
			delete(a.states, i)
		}
	}
}
