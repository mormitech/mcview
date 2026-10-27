package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestVersion(t *testing.T) {
	if v := version(); v == "" || v == "dev" {
		t.Fatalf("version() = %q, want the productVersion from wails.json", v)
	}
}

func TestWindowFlags(t *testing.T) {
	for _, c := range []struct {
		args  []string
		mode  string
		theme string
		file  string
		casc  int
	}{
		{nil, "", "", "", 0},
		{[]string{"--about"}, "about", "", "", 0},
		{[]string{"--theme", "light", "--about"}, "about", "light", "", 0},
		{[]string{"--theme", "dark", "--about"}, "about", "dark", "", 0},
		{[]string{"-psn_0_12345"}, "", "", "", 0}, // what macOS adds when it launches an app
		{[]string{"/home/me/pic.jpg"}, "", "", "/home/me/pic.jpg", 0},
		{[]string{"--cascade", "3"}, "", "", "", 3},
		{[]string{"--cascade", "2", "--theme", "dark", "/p/a b.png"}, "", "dark", "/p/a b.png", 2},
		{[]string{"--theme", "light", `C:\Users\me\My Pictures\a b.png`}, "", "light", `C:\Users\me\My Pictures\a b.png`, 0},
	} {
		mode, theme, file, casc := windowFlags(c.args)
		if mode != c.mode || theme != c.theme || file != c.file || casc != c.casc {
			t.Errorf("windowFlags(%v) = %q, %q, %q, %d; want %q, %q, %q, %d", c.args, mode, theme, file, casc, c.mode, c.theme, c.file, c.casc)
		}
	}
}

func TestThemes(t *testing.T) {
	if macAppearance("dark") == macAppearance("light") {
		t.Error("macAppearance: dark and light must differ")
	}
	if windowsTheme("dark") == windowsTheme("light") || windowsTheme("") == windowsTheme("dark") {
		t.Error("windowsTheme: dark, light and the system's setting must differ")
	}
}

func TestResolvedPathFollowsLinks(t *testing.T) {
	dir := t.TempDir()
	real := filepath.Join(dir, "app", "prog")
	if err := os.MkdirAll(filepath.Dir(real), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(real, []byte("x"), 0o755); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(dir, "prog")
	if err := os.Symlink(real, link); err != nil {
		t.Skip("no symbolic links here")
	}
	want, _ := filepath.EvalSymlinks(real)
	if got, changed := resolvedPath(link); !changed || got != want {
		t.Errorf("resolvedPath(link) = %q, %v; want %q, true", got, changed, want)
	}
	if got, changed := resolvedPath(want); changed || got != want {
		t.Errorf("resolvedPath(real) = %q, %v; want it unchanged", got, changed)
	}
}
