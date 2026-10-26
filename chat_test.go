package main

import (
	"net/http"
	"strings"
	"testing"
)

func TestChatEvents(t *testing.T) {
	kinds := func(line string) []string {
		var k []string
		for _, e := range chatEvents([]byte(line)) {
			k = append(k, e["kind"].(string))
		}
		return k
	}
	eq := func(got []string, want ...string) {
		t.Helper()
		if strings.Join(got, ",") != strings.Join(want, ",") {
			t.Errorf("got %v, want %v", got, want)
		}
	}
	eq(kinds(`{"type":"system","subtype":"init","session_id":"abc"}`))
	eq(kinds(`{"type":"stream_event","event":{"type":"message_start","message":{}}}`), "start")
	delta := chatEvents([]byte(`{"type":"stream_event","event":{"type":"content_block_delta","index":0,"delta":{"type":"text_delta","text":"ok"}}}`))
	if len(delta) != 1 || delta[0]["kind"] != "delta" || delta[0]["text"] != "ok" {
		t.Errorf("delta = %v", delta)
	}
	eq(kinds(`{"type":"stream_event","event":{"type":"content_block_delta","index":0,"delta":{"type":"thinking_delta","thinking":"hm"}}}`))
	asst := chatEvents([]byte(`{"type":"assistant","message":{"content":[{"type":"text","text":"Drawing."},{"type":"tool_use","id":"t1","name":"mcp__mcview__draw_rect","input":{"x":1,"y":2,"w":30,"h":40}}]}}`))
	if len(asst) != 2 || asst[0]["kind"] != "text" || asst[1]["kind"] != "tool" || asst[1]["name"] != "draw_rect" || asst[1]["input"] != `{"x":1,"y":2,"w":30,"h":40}` {
		t.Errorf("assistant = %v", asst)
	}
	long := chatEvents([]byte(`{"type":"assistant","message":{"content":[{"type":"tool_use","name":"mcp__mcview__set_pixels","input":{"pixels":"` + strings.Repeat("x", 500) + `"}}]}}`))
	if s := long[0]["input"].(string); len([]rune(s)) > 170 || !strings.HasSuffix(s, "…") {
		t.Errorf("a long input is cut: %d runes", len([]rune(s)))
	}
	bad := chatEvents([]byte(`{"type":"user","message":{"content":[{"type":"tool_result","is_error":true,"content":[{"type":"text","text":"outside the picture"}]},{"type":"tool_result","is_error":false,"content":"fine"}]}}`))
	if len(bad) != 1 || bad[0]["kind"] != "toolerror" || bad[0]["text"] != "outside the picture" {
		t.Errorf("tool errors = %v", bad)
	}
	done := chatEvents([]byte(`{"type":"result","is_error":true,"result":"Not logged in","total_cost_usd":0.5}`))
	if len(done) != 1 || done[0]["kind"] != "done" || done[0]["error"] != true || done[0]["text"] != "Not logged in" {
		t.Errorf("result = %v", done)
	}
	eq(kinds(`not json`))
}

func TestChatArgs(t *testing.T) {
	a := strings.Join(chatArgs("", "/tmp/c.json"), "\x00")
	for _, want := range []string{"-p", "stream-json", "--strict-mcp-config", "/tmp/c.json", "mcp__mcview", "call undo once", "note exactly which area"} {
		if !strings.Contains(a, want) {
			t.Errorf("the arguments lack %q: %q", want, a)
		}
	}
	if strings.Contains(a, "--resume") {
		t.Errorf("a first message resumes nothing")
	}
	args := chatArgs("sess-1", "/tmp/c.json")
	if args[len(args)-2] != "--resume" || args[len(args)-1] != "sess-1" {
		t.Errorf("the next message continues the session: %v", args[len(args)-2:])
	}
	for i, x := range args { // no file, shell or web tools
		if x == "--tools" && args[i+1] != "" {
			t.Errorf("tools = %q", args[i+1])
		}
	}
}

func TestChatToolsServerIsPrivate(t *testing.T) {
	a := NewApp()
	defer a.stopChat()
	url, token, err := a.startChatTools()
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(url, "http://127.0.0.1:") || len(token) < 32 {
		t.Fatalf("url %q, token %q", url, token)
	}
	if u2, t2, _ := a.startChatTools(); u2 != url || t2 != token {
		t.Errorf("the server is started once")
	}
	post := func(auth, origin string) int {
		req, _ := http.NewRequest("POST", url, strings.NewReader(`{"jsonrpc":"2.0","id":1,"method":"ping"}`))
		if auth != "" {
			req.Header.Set("Authorization", "Bearer "+auth)
		}
		if origin != "" {
			req.Header.Set("Origin", origin)
		}
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		resp.Body.Close()
		return resp.StatusCode
	}
	if c := post("", ""); c != 401 {
		t.Errorf("no token: %d", c)
	}
	if c := post("wrong", ""); c != 401 {
		t.Errorf("wrong token: %d", c)
	}
	if c := post(token, "https://evil.example"); c != 403 {
		t.Errorf("a web page: %d", c)
	}
	if c := post(token, ""); c != 200 {
		t.Errorf("the token: %d", c)
	}
	a.stopChat()
	if _, err := http.Post(url, "application/json", strings.NewReader("{}")); err == nil {
		t.Errorf("the server must be gone after stopChat")
	}
}
