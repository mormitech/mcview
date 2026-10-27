package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestConfigDirMovesTheOldFolder(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("XDG_CONFIG_HOME", home)
	t.Setenv("AppData", home)
	base, err := os.UserConfigDir()
	if err != nil {
		t.Skip(err)
	}
	old := filepath.Join(base, "mcview")
	if err := os.MkdirAll(old, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(old, "recent.json"), []byte("[]"), 0o644); err != nil {
		t.Fatal(err)
	}
	dir := configDir()
	if filepath.Base(dir) != "mormiview" {
		t.Fatalf("config folder is %q", dir)
	}
	if _, err := os.Stat(filepath.Join(dir, "recent.json")); err != nil {
		t.Errorf("the old settings were not moved: %v", err)
	}
	if _, err := os.Stat(old); !os.IsNotExist(err) {
		t.Errorf("the old folder is still there")
	}
}

func TestDemoUsesAFreshFolder(t *testing.T) {
	defer func() { demoDir = "" }()
	startDemo([]string{"--demo"})
	if demoDir == "" || configDir() != demoDir {
		t.Fatalf("demo folder %q, config folder %q", demoDir, configDir())
	}
	if entries, _ := os.ReadDir(demoDir); len(entries) != 0 {
		t.Errorf("the demo folder is not empty")
	}
	os.RemoveAll(demoDir)
}
