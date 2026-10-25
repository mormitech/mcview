package main

import (
	"path/filepath"
	"testing"
)

func TestDiscardOnlyAsksForModifiedPictures(t *testing.T) {
	dir := t.TempDir()
	one, two := filepath.Join(dir, "a.png"), filepath.Join(dir, "b.png")
	writePNG(t, one)
	writePNG(t, two)
	a := NewApp()
	asked := 0
	answer := choiceCancel
	a.confirmDiscard = func(name, action string) discardChoice { asked++; return answer }
	a.setImage(one)

	if !a.discardOK("close the window") || asked != 0 {
		t.Fatalf("an unmodified picture asks nothing: asked %d", asked)
	}
	if _, err := a.Step(1); err != nil || asked != 0 || a.current().Name != "b.png" {
		t.Fatalf("stepping from an unmodified picture: asked %d, now %q, %v", asked, a.current().Name, err)
	}

	modify(a) // b.png, modified
	if !a.current().Modified {
		t.Fatal("the paste must modify the picture")
	}
	// cancel: nothing moves
	if _, err := a.Step(1); err != nil || asked != 1 || a.current().Name != "b.png" || !a.current().Modified {
		t.Errorf("Cancel must stay: asked %d, now %q", asked, a.current().Name)
	}
	if a.discardOK("close the window") || asked != 2 {
		t.Error("Cancel is not OK to close")
	}
	// discard: on to the other picture, the changes are gone
	answer = choiceDiscard
	if _, err := a.Step(1); err != nil || a.current().Name != "a.png" || a.current().Modified {
		t.Errorf("Discard must go on: now %q, modified %v, %v", a.current().Name, a.current().Modified, err)
	}
	// the alert says what the picture is called, and what is about to happen
	var gotName, gotAction string
	modify(a)
	a.confirmDiscard = func(name, action string) discardChoice { gotName, gotAction = name, action; return choiceDiscard }
	a.Step(-1)
	if gotName != "a.png" || gotAction != "go to the previous picture" {
		t.Errorf("the alert was asked %q, %q", gotName, gotAction)
	}
}

func TestPastedPictureIsNotAskedAboutWhenStepping(t *testing.T) {
	a := NewApp()
	asked := 0
	a.confirmDiscard = func(string, string) discardChoice { asked++; return choiceDiscard }
	a.setClip(solid(2, 2, blue))
	if note, _ := a.Step(1); note == "" || asked != 0 {
		t.Errorf("a picture without a file keeps its note (%q) and asks nothing (%d)", note, asked)
	}
	// closing the window is where it is asked
	if !a.discardOK("close the window") || asked != 1 {
		t.Errorf("closing asks once: %d", asked)
	}
}
