package main

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"

	"golang.org/x/image/tiff"
)

func TestIsImage(t *testing.T) {
	for path, want := range map[string]bool{
		"a.jpg": true, "/x/y/A.JPEG": true, "c.png": true, "d.tiff": true, "e.webp": true, "f.svg": true,
		"g.txt": false, "h": false, "i.jpg.exe": false, ".jpg.bak": false,
	} {
		if got := isImage(path); got != want {
			t.Errorf("isImage(%q) = %v, want %v", path, got, want)
		}
	}
}

func TestFileArg(t *testing.T) {
	if fileArg(nil) != "" || fileArg([]string{"a.png", "b.png"}) != "a.png" {
		t.Error("fileArg: want the first argument, and nothing for no arguments")
	}
	_, theme, file, _ := windowFlags([]string{"--theme", "dark", "pic.jpg"})
	if theme != "dark" || file != "pic.jpg" {
		t.Errorf("windowFlags: theme %q, file %q", theme, file)
	}
}

func writePNG(t *testing.T, path string) {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, 4, 3))
	img.Set(1, 1, color.RGBA{R: 255, A: 255})
	var b bytes.Buffer
	if err := png.Encode(&b, img); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, b.Bytes(), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestSetImage(t *testing.T) {
	dir := t.TempDir()
	pic := filepath.Join(dir, "pic.png")
	writePNG(t, pic)
	os.WriteFile(filepath.Join(dir, "notes.txt"), []byte("hi"), 0o644)
	os.Mkdir(filepath.Join(dir, "dir.png"), 0o755)

	a := NewApp()
	if a.Image().URL != "" {
		t.Error("a new window has no picture")
	}
	for _, bad := range []string{filepath.Join(dir, "notes.txt"), filepath.Join(dir, "missing.png"), filepath.Join(dir, "dir.png")} {
		if a.setImage(bad) == nil {
			t.Errorf("setImage(%q) = nil, want an error", bad)
		}
	}
	if err := a.setImage(pic); err != nil {
		t.Fatal(err)
	}
	first := a.Image()
	if first.Name != "pic.png" || first.Path != pic || first.URL == "" {
		t.Errorf("Image() = %+v", first)
	}
	a.setImage(pic)
	if a.Image().URL == first.URL {
		t.Error("every picture needs a URL of its own, or the web view shows the old one")
	}
}

func get(h http.Handler, path string) *httptest.ResponseRecorder {
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest("GET", path, nil))
	return rec
}

func TestImageHandler(t *testing.T) {
	dir := t.TempDir()
	pic := filepath.Join(dir, "pic.png")
	writePNG(t, pic)
	secret := filepath.Join(dir, "secret.txt")
	os.WriteFile(secret, []byte("secret"), 0o644)

	a := NewApp()
	h := a.imageHandler()
	if get(h, "/image/0").Code != http.StatusNotFound {
		t.Error("no picture yet: want 404")
	}
	a.setImage(pic)
	url := a.Image().URL
	rec := get(h, url)
	if rec.Code != 200 || rec.Header().Get("Content-Type") != "image/png" {
		t.Fatalf("GET %s = %d %q", url, rec.Code, rec.Header().Get("Content-Type"))
	}
	if _, err := png.Decode(rec.Body); err != nil {
		t.Errorf("the body is not the picture: %v", err)
	}
	// only the picture of the window is served: not an older one, not any other path or file
	for _, p := range []string{"/image/99", "/image/", "/", "/image/1/../../secret.txt", secret} {
		if get(h, p).Code != http.StatusNotFound {
			t.Errorf("GET %s: want 404", p)
		}
	}
}

func TestTiffIsServedAsPNG(t *testing.T) {
	path := filepath.Join(t.TempDir(), "scan.tif")
	var b bytes.Buffer
	if err := tiff.Encode(&b, image.NewRGBA(image.Rect(0, 0, 5, 2)), nil); err != nil {
		t.Fatal(err)
	}
	os.WriteFile(path, b.Bytes(), 0o644)
	a := NewApp()
	a.setImage(path)
	rec := get(a.imageHandler(), a.Image().URL)
	if rec.Header().Get("Content-Type") != "image/png" {
		t.Fatalf("Content-Type = %q, want image/png", rec.Header().Get("Content-Type"))
	}
	img, err := png.Decode(rec.Body)
	if err != nil || img.Bounds().Dx() != 5 || img.Bounds().Dy() != 2 {
		t.Errorf("decoded %v, %v; want a 5 x 2 picture", img, err)
	}
}

func TestOnFileOpen(t *testing.T) {
	pic := filepath.Join(t.TempDir(), "pic.png")
	writePNG(t, pic)
	pic2 := filepath.Join(t.TempDir(), "pic2.png")
	writePNG(t, pic2)
	a := NewApp()
	var opened []string
	a.openWindow = func(p string) error { opened = append(opened, p); return nil } // never start the test binary again

	a.onFileOpen(filepath.Join(t.TempDir(), "notes.txt")) // not a picture: ignored
	if a.Image().URL != "" || len(opened) != 0 {
		t.Error("a file that is not a picture must be ignored")
	}
	a.onFileOpen(pic)
	if a.Image().Path != pic || len(opened) != 0 {
		t.Errorf("the first file goes into the window, got %q, windows %v", a.Image().Path, opened)
	}
	// macOS reports the file a window was started with once more: no new window (this once started windows without end)
	a.onFileOpen(pic)
	if len(opened) != 0 {
		t.Errorf("the picture of the window again must not open a window, got %v", opened)
	}
	a.onFileOpen(pic2)
	if len(opened) != 1 || opened[0] != pic2 {
		t.Errorf("another picture goes into a window of its own, got %v", opened)
	}
}

func TestStartedWindowNeverStartsWindows(t *testing.T) {
	pic, pic2 := filepath.Join(t.TempDir(), "a.png"), filepath.Join(t.TempDir(), "b.png")
	writePNG(t, pic)
	writePNG(t, pic2)
	a := NewApp()
	a.openWindow = func(p string) error {
		t.Errorf("a window that was started for a picture started a window for %s", p)
		return nil
	}
	a.child = true
	a.setImage(pic)
	a.onFileOpen(pic)
	a.onFileOpen(pic2)
	if a.Image().Path != pic2 {
		t.Errorf("the picture goes into this window, got %q", a.Image().Path)
	}
}

// The list of file types is written in four places, because each system wants it in its own form. They must not drift apart.
func TestFileTypeListsAgree(t *testing.T) {
	want := make([]string, 0, len(imageTypes))
	for e := range imageTypes {
		want = append(want, e)
	}
	sort.Strings(want)

	// macOS: wails.json fileAssociations
	var exts []string
	for _, m := range regexp.MustCompile(`"ext":\s*"([^"]+)"`).FindAllStringSubmatch(string(projectJSON), -1) {
		exts = append(exts, "."+m[1])
	}
	check(t, "wails.json", exts, want)

	// Windows: the installer registers each type three times (OpenWithProgids, SupportedTypes, Capabilities)
	iss := readFile(t, "packaging/windows/installer.iss")
	check(t, "installer.iss OpenWithProgids", pick(iss, `Classes\\(\.[a-z0-9]+)\\OpenWithProgids`), want)
	check(t, "installer.iss SupportedTypes", pick(iss, `SupportedTypes"; ValueType: string; ValueName: "(\.[a-z0-9]+)"`), want)
	check(t, "installer.iss Capabilities", pick(iss, `FileAssociations"; ValueType: string; ValueName: "(\.[a-z0-9]+)"`), want)

	// Linux: every type we serve has its MIME type in the desktop file
	desktop := readFile(t, "packaging/linux/mormiview.desktop")
	for _, typ := range imageTypes {
		if !strings.Contains(desktop, typ+";") {
			t.Errorf("mormiview.desktop has no MimeType %s", typ)
		}
	}
}

func readFile(t *testing.T, path string) string {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

func pick(text, pattern string) []string {
	var out []string
	for _, m := range regexp.MustCompile(pattern).FindAllStringSubmatch(text, -1) {
		out = append(out, m[1])
	}
	return out
}

func check(t *testing.T, where string, got, want []string) {
	t.Helper()
	sort.Strings(got)
	if strings.Join(got, " ") != strings.Join(want, " ") {
		t.Errorf("%s lists %v, want %v (the types in image.go)", where, got, want)
	}
}
