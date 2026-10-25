package main

import "testing"

func TestFitSize(t *testing.T) {
	for _, c := range []struct{ w, h, sw, sh, fw, fh int }{
		{1440, 800, 1920, 1080, 1440, 800}, // fits: left alone
		{1440, 800, 1093, 614, 983, 552},   // 1366x768 at 125% scaling: cut to 90%
		{1440, 800, 1440, 1000, 1296, 800}, // only the width is too big
		{440, 520, 0, 0, 440, 520},         // unknown screen: left alone
	} {
		fw, fh := fitSize(c.w, c.h, c.sw, c.sh)
		if fw != c.fw || fh != c.fh {
			t.Errorf("fitSize(%d, %d, %d, %d) = %d x %d, want %d x %d", c.w, c.h, c.sw, c.sh, fw, fh, c.fw, c.fh)
		}
	}
}
