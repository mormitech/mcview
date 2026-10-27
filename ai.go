package main

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/wailsapp/wails/v2/pkg/menu"
)

// The AI menu holds what the Claude of the chat panel may use besides the picture tools. Every permission is on until the user turns it off, and
// the choice is kept in a small file of the user's settings folder (the fields say "no", so that a missing file or field means on).

type aiSettings struct {
	NoWeb bool `json:"noWeb"` // no web search and fetch
}

func aiSettingsPath() string {
	dir, err := os.UserConfigDir()
	if err != nil {
		return ""
	}
	return filepath.Join(dir, "mcview", "ai.json")
}

// loadAI is read for every message, so that a change made in another window counts at once.
func loadAI() aiSettings {
	var s aiSettings
	if path := aiSettingsPath(); path != "" {
		if b, err := os.ReadFile(path); err == nil {
			json.Unmarshal(b, &s)
		}
	}
	return s
}

func saveAI(s aiSettings) {
	path := aiSettingsPath()
	b, err := json.Marshal(s)
	if path == "" || err != nil || os.MkdirAll(filepath.Dir(path), 0o755) != nil {
		return
	}
	tmp := path + ".tmp"
	if os.WriteFile(tmp, b, 0o644) == nil {
		os.Rename(tmp, path)
	}
}

func aiMenu(m *menu.Menu) {
	ai := m.AddSubmenu("AI")
	ai.AddCheckbox("Web search and fetch", !loadAI().NoWeb, nil, func(cd *menu.CallbackData) {
		s := loadAI()
		s.NoWeb = !cd.MenuItem.Checked
		saveAI(s)
	})
}
