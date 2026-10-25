package main

import "testing"

func TestWinStateRoundTrip(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	t.Setenv("AppData", t.TempDir())
	if _, ok := loadWinState(); ok {
		t.Fatal("nothing is remembered yet")
	}
	saveWinState(winState{X: 120, Y: 80, W: 900, H: 600})
	if st, ok := loadWinState(); !ok || st != (winState{X: 120, Y: 80, W: 900, H: 600}) {
		t.Errorf("loaded %+v, %v", st, ok)
	}
	saveWinState(winState{X: 1, Y: 1, W: 10, H: 10}) // a window that small is not a real one: the last good one stays
	if st, _ := loadWinState(); st.W != 900 {
		t.Errorf("a tiny window overwrote the state: %+v", st)
	}
}

func TestOnScreen(t *testing.T) {
	for _, c := range []struct {
		st   winState
		want bool
	}{
		{winState{X: 100, Y: 50, W: 800, H: 600}, true},
		{winState{X: 3000, Y: 50, W: 800, H: 600}, false}, // the second screen is gone
		{winState{X: -2000, Y: 50, W: 800, H: 600}, false},
		{winState{X: 100, Y: -300, W: 800, H: 600}, false},
	} {
		if got := onScreen(c.st, 2560); got != c.want {
			t.Errorf("onScreen(%+v) = %v, want %v", c.st, got, c.want)
		}
	}
}
