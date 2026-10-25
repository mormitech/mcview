package main

import "testing"

func TestVersion(t *testing.T) {
	if v := version(); v == "" || v == "dev" {
		t.Fatalf("version() = %q, want the productVersion from wails.json", v)
	}
}

func TestWindowFlags(t *testing.T) {
	for _, c := range []struct {
		args  []string
		about bool
		theme string
		file  string
		casc  int
	}{
		{nil, false, "", "", 0},
		{[]string{"--about"}, true, "", "", 0},
		{[]string{"--theme", "dark", "--about"}, true, "dark", "", 0},
		{[]string{"-psn_0_12345"}, false, "", "", 0}, // what macOS adds when it launches an app
		{[]string{"/home/me/pic.jpg"}, false, "", "/home/me/pic.jpg", 0},
		{[]string{"--cascade", "3"}, false, "", "", 3},
		{[]string{"--cascade", "2", "--theme", "dark", "/p/a b.png"}, false, "dark", "/p/a b.png", 2},
		{[]string{"--theme", "light", `C:\Users\me\My Pictures\a b.png`}, false, "light", `C:\Users\me\My Pictures\a b.png`, 0},
	} {
		about, theme, file, casc := windowFlags(c.args)
		if about != c.about || theme != c.theme || file != c.file || casc != c.casc {
			t.Errorf("windowFlags(%v) = %v, %q, %q, %d; want %v, %q, %q, %d", c.args, about, theme, file, casc, c.about, c.theme, c.file, c.casc)
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
