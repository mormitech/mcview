package main

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	wruntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

// Control mode: mcView started with `--control PORT` (or MCVIEW_CONTROL=PORT; 0 takes a free port) listens on 127.0.0.1 for commands from
// another program, so that every feature can be triggered from outside: demo screenshots, scripted tests, a manual in pictures. Nothing
// of it exists without the flag. The commands are JSON (see docs/control.md): the ones about the window and the dialogs are done here,
// the rest go to the page, which does them with the very code the mouse and the keys use (see "Control mode" in main.js).
//
// Every request needs the token that is written, with the address, to the control file (MCVIEW_CONTROL_FILE, default
// mcview-control.json in the temporary folder, readable by the user only). Requests with an Origin header or another Host are refused,
// so that a web page cannot reach it.
//
// The dialogs of the system cannot be clicked from outside, so in this mode they are answered from a queue the script fills (the
// command "answer"); a question with no answer queued is answered with Cancel (or No), which is the safe one.

type control struct {
	app     *App
	token   string
	addr    string
	send    func(id int, cmd map[string]any) // gives a command to the page; the page answers with ControlReply
	mu      sync.Mutex
	seq     int
	waiting map[int]chan string
	answers map[string][]string // by the kind of dialog: discard, overwrite, large, delete, quit, open, save
	asked   []string            // the questions asked since the last "dialogs" command
}

// controlPort is the port from --control or MCVIEW_CONTROL; the second result is false when control mode is not asked for.
func controlPort(args []string, env string) (string, bool) {
	for i, a := range args {
		if a == "--control" || a == "-control" {
			if i+1 < len(args) && !strings.HasPrefix(args[i+1], "-") {
				return args[i+1], true
			}
			return "0", true
		}
		if v, ok := strings.CutPrefix(a, "--control="); ok {
			return v, true
		}
	}
	if env != "" {
		return env, true
	}
	return "", false
}

func newToken() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}

// startControl starts listening. It returns the address (127.0.0.1:port).
func (a *App) startControl(port string) (string, error) {
	ln, err := net.Listen("tcp", "127.0.0.1:"+port)
	if err != nil {
		return "", err
	}
	c := &control{app: a, token: newToken(), addr: ln.Addr().String(), waiting: map[int]chan string{}, answers: map[string][]string{}}
	c.send = func(id int, cmd map[string]any) { a.emit("control", id, cmd) }
	a.mu.Lock()
	a.ctl = c
	a.mu.Unlock()
	mux := http.NewServeMux()
	mux.HandleFunc("/cmd", c.guard(c.handleCmd))
	mux.HandleFunc("/batch", c.guard(c.handleBatch))
	mux.HandleFunc("/info", c.guard(c.handleInfo))
	go http.Serve(ln, mux)
	file := os.Getenv("MCVIEW_CONTROL_FILE")
	if file == "" {
		file = filepath.Join(os.TempDir(), "mcview-control.json")
	}
	info, _ := json.Marshal(map[string]any{"url": "http://" + c.addr, "token": c.token, "pid": os.Getpid()})
	if err := os.WriteFile(file, info, 0o600); err != nil {
		return c.addr, err
	}
	return c.addr, nil
}

// guard lets a request through when it has the token, comes to the address of this program and was not made by a web page.
func (c *control) guard(next func(http.ResponseWriter, *http.Request)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		host := r.Host
		if h, _, err := net.SplitHostPort(host); err == nil {
			host = h
		}
		if r.Header.Get("Origin") != "" || (host != "127.0.0.1" && host != "localhost") {
			http.Error(w, "refused", http.StatusForbidden)
			return
		}
		if r.Header.Get("Authorization") != "Bearer "+c.token {
			http.Error(w, "token missing or wrong", http.StatusUnauthorized)
			return
		}
		next(w, r)
	}
}

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(v)
}

func (c *control) handleCmd(w http.ResponseWriter, r *http.Request) {
	var m map[string]any
	if err := json.NewDecoder(r.Body).Decode(&m); err != nil {
		writeJSON(w, 400, map[string]any{"error": "not JSON: " + err.Error()})
		return
	}
	res, err := c.do(m)
	if err != nil {
		writeJSON(w, 200, map[string]any{"error": err.Error()})
		return
	}
	writeJSON(w, 200, res)
}

// handleBatch runs a list of commands one after the other and stops at the first that fails (unless "continue" is in the query).
func (c *control) handleBatch(w http.ResponseWriter, r *http.Request) {
	var list []map[string]any
	if err := json.NewDecoder(r.Body).Decode(&list); err != nil {
		writeJSON(w, 400, map[string]any{"error": "not a JSON list: " + err.Error()})
		return
	}
	out := make([]any, 0, len(list))
	for i, m := range list {
		res, err := c.do(m)
		if err != nil {
			out = append(out, map[string]any{"error": err.Error(), "at": i})
			if r.URL.Query().Get("continue") == "" {
				break
			}
			continue
		}
		out = append(out, res)
	}
	writeJSON(w, 200, out)
}

func (c *control) handleInfo(w http.ResponseWriter, r *http.Request) {
	res, _ := c.do(map[string]any{"cmd": "info"})
	writeJSON(w, 200, res)
}

func str(m map[string]any, k string) string { s, _ := m[k].(string); return s }
func num(m map[string]any, k string, def int) int {
	if f, ok := m[k].(float64); ok {
		return int(f)
	}
	return def
}

// do runs one command: the ones about the window and the dialogs here, anything else in the page.
func (c *control) do(m map[string]any) (map[string]any, error) {
	a := c.app
	switch str(m, "cmd") {
	case "info":
		x, y, w, h := 0, 0, 0, 0
		if a.ctx != nil {
			x, y = wruntime.WindowGetPosition(a.ctx)
			w, h = wruntime.WindowGetSize(a.ctx)
		}
		return map[string]any{"ok": true, "version": version(), "pid": os.Getpid(), "windowId": windowID(), "x": x, "y": y, "w": w, "h": h}, nil
	case "open": // a picture of the folder, without the Open dialog
		path := str(m, "path")
		if err := a.setImage(path); err != nil {
			return nil, err
		}
		time.Sleep(600 * time.Millisecond) // the page loads it
		return map[string]any{"ok": true}, nil
	case "window": // size and place (the window is centred when no x and y), for screenshots that look alike
		if a.ctx == nil {
			return nil, fmt.Errorf("no window yet")
		}
		if w, h := num(m, "w", 0), num(m, "h", 0); w > 0 && h > 0 {
			wruntime.WindowSetSize(a.ctx, w, h)
		}
		if _, ok := m["x"]; ok {
			wruntime.WindowSetPosition(a.ctx, num(m, "x", 0), num(m, "y", 0))
		} else {
			wruntime.WindowCenter(a.ctx)
		}
		time.Sleep(400 * time.Millisecond)
		return c.do(map[string]any{"cmd": "info"})
	case "answer": // the answer for the next question of a kind: "value" or "values" (a list, one for each question)
		kind := str(m, "kind")
		c.mu.Lock()
		defer c.mu.Unlock()
		if v, ok := m["values"].([]any); ok {
			for _, x := range v {
				c.answers[kind] = append(c.answers[kind], fmt.Sprint(x))
			}
		} else {
			c.answers[kind] = append(c.answers[kind], str(m, "value"))
		}
		return map[string]any{"ok": true, "queued": len(c.answers[kind])}, nil
	case "dialogs": // the questions asked since the last time
		c.mu.Lock()
		defer c.mu.Unlock()
		asked := c.asked
		c.asked = nil
		return map[string]any{"ok": true, "asked": asked}, nil
	case "sleep":
		time.Sleep(time.Duration(num(m, "ms", 100)) * time.Millisecond)
		return map[string]any{"ok": true}, nil
	case "quit": // without asking anything
		if a.ctx != nil {
			go func() { time.Sleep(100 * time.Millisecond); wruntime.Quit(a.ctx) }()
		}
		return map[string]any{"ok": true}, nil
	case "":
		return nil, fmt.Errorf("no cmd")
	}
	return c.toPage(m)
}

// toPage gives a command to the page and waits for its answer.
func (c *control) toPage(m map[string]any) (map[string]any, error) {
	c.mu.Lock()
	c.seq++
	id := c.seq
	ch := make(chan string, 1)
	c.waiting[id] = ch
	c.mu.Unlock()
	c.send(id, m)
	select {
	case s := <-ch:
		var res map[string]any
		if err := json.Unmarshal([]byte(s), &res); err != nil {
			return nil, fmt.Errorf("the page answered something else: %s", s)
		}
		if e, ok := res["error"].(string); ok && e != "" {
			return nil, fmt.Errorf("%s", e)
		}
		return res, nil
	case <-time.After(30 * time.Second):
		c.mu.Lock()
		delete(c.waiting, id)
		c.mu.Unlock()
		return nil, fmt.Errorf("the page did not answer %q in 30 seconds", str(m, "cmd"))
	}
}

// ControlReply is the answer of the page to a command (JSON).
func (a *App) ControlReply(id int, result string) {
	a.mu.Lock()
	c := a.ctl
	a.mu.Unlock()
	if c == nil {
		return
	}
	c.mu.Lock()
	ch := c.waiting[id]
	delete(c.waiting, id)
	c.mu.Unlock()
	if ch != nil {
		ch <- result
	}
}

// mockAnswer is what the dialogs ask first: in control mode the answer comes from the queue ("" when there is none, which every
// dialog takes as Cancel or No), and the question is noted. ok is false outside control mode, and the system's dialog is shown.
func (a *App) mockAnswer(kind, detail string) (answer string, ok bool) {
	a.mu.Lock()
	c := a.ctl
	a.mu.Unlock()
	if c == nil {
		return "", false
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	c.asked = append(c.asked, kind+": "+detail)
	if q := c.answers[kind]; len(q) > 0 {
		answer, c.answers[kind] = q[0], q[1:]
	}
	return answer, true
}

// pickSave is the Save dialog of the system, or the answer queued for it in control mode (the path to save to).
func (a *App) pickSave(opts wruntime.SaveDialogOptions) (string, error) {
	if v, ok := a.mockAnswer("save", opts.DefaultFilename); ok {
		return v, nil
	}
	return wruntime.SaveFileDialog(a.ctx, opts)
}

// pickOpen is the Open dialog of the system, or the answer queued for it in control mode (the path of the picture).
func (a *App) pickOpen(opts wruntime.OpenDialogOptions) (string, error) {
	if v, ok := a.mockAnswer("open", opts.DefaultDirectory); ok {
		return v, nil
	}
	return wruntime.OpenFileDialog(a.ctx, opts)
}
