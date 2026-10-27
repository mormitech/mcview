package main

import "testing"

func TestAISettingsDefaultOn(t *testing.T) {
	t.Setenv("HOME", t.TempDir()) // (macOS and Linux settings folder)
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	t.Setenv("AppData", t.TempDir())
	if loadAI().NoWeb {
		t.Fatal("the web is on until it is turned off")
	}
	saveAI(aiSettings{NoWeb: true})
	if !loadAI().NoWeb {
		t.Fatal("turning the web off is remembered")
	}
	saveAI(aiSettings{})
	if loadAI().NoWeb {
		t.Fatal("turning it on again is remembered")
	}
}
