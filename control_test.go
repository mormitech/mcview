package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestControlModeNeedsTheTokenAndAnswersDialogs(t *testing.T) {
	file := filepath.Join(t.TempDir(), "control.json")
	t.Setenv("MCVIEW_CONTROL_FILE", file)
	a := NewApp()
	addr, err := a.startControl("0")
	if err != nil {
		t.Fatal(err)
	}
	c := a.ctl
	c.send = func(id int, cmd map[string]any) { // the page, which answers at once
		go a.ControlReply(id, `{"ok":true,"echo":"`+str(cmd, "cmd")+`"}`)
	}
	call := func(token, body string) (int, map[string]any) {
		req, _ := http.NewRequest("POST", "http://"+addr+"/cmd", bytes.NewBufferString(body))
		if token != "" {
			req.Header.Set("Authorization", "Bearer "+token)
		}
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()
		var m map[string]any
		json.NewDecoder(resp.Body).Decode(&m)
		return resp.StatusCode, m
	}
	if code, _ := call("", `{"cmd":"info"}`); code != 401 {
		t.Errorf("no token: %d", code)
	}
	if code, _ := call("wrong", `{"cmd":"info"}`); code != 401 {
		t.Errorf("wrong token: %d", code)
	}
	raw, _ := os.ReadFile(file)
	var info struct{ Token string }
	json.Unmarshal(raw, &info)
	if info.Token != c.token {
		t.Fatalf("the control file has token %q, want %q", info.Token, c.token)
	}
	if fi, _ := os.Stat(file); runtime.GOOS != "windows" && fi.Mode().Perm() != 0o600 { // (Windows has no such permission bits)
		t.Errorf("the control file is %v", fi.Mode().Perm())
	}
	if code, m := call(c.token, `{"cmd":"wait","ms":1}`); code != 200 || m["echo"] != "wait" {
		t.Errorf("a command for the page: %d %v", code, m)
	}
	// a browser page cannot reach it (it sends an Origin)
	req, _ := http.NewRequest("POST", "http://"+addr+"/cmd", bytes.NewBufferString(`{"cmd":"info"}`))
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Origin", "https://example.com")
	if resp, _ := http.DefaultClient.Do(req); resp == nil || resp.StatusCode != 403 {
		t.Error("a request with an Origin must be refused")
	}
	// dialogs: the queue answers, and an empty queue is a Cancel
	if v, ok := a.mockAnswer("discard", "x"); !ok || v != "" || a.askDiscard("x.jpg", "close") != choiceCancel {
		t.Error("with nothing queued a dialog is cancelled")
	}
	call(c.token, `{"cmd":"answer","kind":"discard","values":["discard","save"]}`)
	if a.askDiscard("x.jpg", "close") != choiceDiscard || a.askDiscard("x.jpg", "close") != choiceSave {
		t.Error("the queued answers are used in order")
	}
	_, m := call(c.token, `{"cmd":"dialogs"}`)
	if asked, _ := m["asked"].([]any); len(asked) < 3 {
		t.Errorf("the questions are noted: %v", m)
	}
}

func TestControlPort(t *testing.T) {
	for _, tc := range []struct {
		args []string
		env  string
		port string
		ok   bool
	}{
		{nil, "", "", false},
		{[]string{"--control", "8123"}, "", "8123", true},
		{[]string{"--control=9"}, "", "9", true},
		{[]string{"--control", "--theme", "dark"}, "", "0", true},
		{nil, "7000", "7000", true},
	} {
		if port, ok := controlPort(tc.args, tc.env); port != tc.port || ok != tc.ok {
			t.Errorf("controlPort(%v, %q) = %q, %v", tc.args, tc.env, port, ok)
		}
	}
}
