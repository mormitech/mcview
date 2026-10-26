//go:build darwin

package main

/*
#cgo LDFLAGS: -framework CoreGraphics -framework CoreFoundation
#include <CoreGraphics/CoreGraphics.h>
#include <CoreFoundation/CoreFoundation.h>

// The number of the first normal window of the process, which `screencapture -l` takes.
static int windowOfPid(int pid) {
	CFArrayRef list = CGWindowListCopyWindowInfo(kCGWindowListOptionOnScreenOnly | kCGWindowListExcludeDesktopElements, kCGNullWindowID);
	if (!list) return 0;
	int found = 0;
	for (CFIndex i = 0; i < CFArrayGetCount(list) && !found; i++) {
		CFDictionaryRef w = CFArrayGetValueAtIndex(list, i);
		int owner = 0, layer = -1, number = 0;
		CFNumberRef o = CFDictionaryGetValue(w, kCGWindowOwnerPID), l = CFDictionaryGetValue(w, kCGWindowLayer), n = CFDictionaryGetValue(w, kCGWindowNumber);
		if (o) CFNumberGetValue(o, kCFNumberIntType, &owner);
		if (l) CFNumberGetValue(l, kCFNumberIntType, &layer);
		if (n) CFNumberGetValue(n, kCFNumberIntType, &number);
		if (owner == pid && layer == 0) found = number;
	}
	CFRelease(list);
	return found;
}
*/
import "C"

import "os"

// windowID is the number of the window of this program on the screen, for `screencapture -l ID file.png` (0 when there is none).
func windowID() int { return int(C.windowOfPid(C.int(os.Getpid()))) }
