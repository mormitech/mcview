#!/usr/bin/env python3
"""Talk to an mcView that was started in control mode (see docs/control.md).

As a library:
    from mcview_ctl import McView
    mc = McView.start("build/bin/mcview.app", "/tmp/Pictures/coast.jpg")
    mc.cmd("key", key="e")
    mc.shot("edit.png")
    mc.quit()

From the shell (the app has to be running with --control):
    scripts/mcview_ctl.py '{"cmd": "state"}'
    scripts/mcview_ctl.py key e          # the same as {"cmd": "key", "key": "e"}
    scripts/mcview_ctl.py --batch commands.json
"""
import json
import os
import subprocess
import sys
import tempfile
import time
import urllib.request


class McViewError(Exception):
    pass


def control_file():
    return os.environ.get("MCVIEW_CONTROL_FILE") or os.path.join(tempfile.gettempdir(), "mcview-control.json")


class McView:
    def __init__(self, file=None):
        with open(file or control_file()) as f:
            info = json.load(f)
        self.url, self.token = info["url"], info["token"]

    @classmethod
    def start(cls, app, picture="", file=None, timeout=20):
        """Starts the app (a .app on macOS, otherwise the program) in control mode and waits until it answers."""
        file = file or control_file()
        app = os.path.abspath(app)  # (`open -a` needs the whole path)
        if os.path.exists(file):
            os.remove(file)
        args = ["--control", "0"] + ([picture] if picture else [])
        env = dict(os.environ, MCVIEW_CONTROL_FILE=file)
        if app.endswith(".app"):
            subprocess.run(["open", "-n", "--env", "MCVIEW_CONTROL_FILE=" + file, "-a", app, "--args"] + args, check=True)
        else:
            subprocess.Popen([app] + args, env=env)
        end = time.time() + timeout
        while time.time() < end:
            try:
                mc = cls(file)
                mc.info()
                return mc
            except (OSError, KeyError, ValueError, McViewError):
                time.sleep(0.3)
        raise McViewError("mcView did not answer in %d seconds" % timeout)

    def _post(self, path, body):
        req = urllib.request.Request(self.url + path, json.dumps(body).encode(), {"Authorization": "Bearer " + self.token})
        with urllib.request.urlopen(req, timeout=60) as r:
            return json.load(r)

    def info(self):
        req = urllib.request.Request(self.url + "/info", headers={"Authorization": "Bearer " + self.token})
        with urllib.request.urlopen(req, timeout=10) as r:
            return json.load(r)

    def cmd(self, _name, /, **args):
        """One command; returns the answer, or raises McViewError with what the app said."""
        res = self._post("/cmd", dict(args, cmd=_name))
        if res.get("error"):
            raise McViewError("%s: %s" % (_name, res["error"]))
        return res

    def batch(self, commands):
        res = self._post("/batch", commands)
        for r in res:
            if isinstance(r, dict) and r.get("error"):
                raise McViewError(r["error"])
        return res

    def answer(self, kind, *values):
        """Queues answers for the dialogs of a kind: discard (save, discard, cancel), overwrite (overwrite, saveas, cancel),
        large (new, over, cancel), delete (yes, no), quit (yes, no), open (a path), save (a path)."""
        return self.cmd("answer", kind=kind, values=list(values))

    def state(self):
        return self.cmd("state")

    def shot(self, path, delay=0.4):
        """A picture of the window (macOS: screencapture -l; the window can be behind others, and it has its shadow)."""
        time.sleep(delay)
        wid = self.info().get("windowId", 0)
        if not wid:
            raise McViewError("the window number is only known on macOS")
        os.makedirs(os.path.dirname(os.path.abspath(path)), exist_ok=True)
        subprocess.run(["screencapture", "-x", "-l", str(wid), path], check=True)

    def quit(self, discard=True):
        if discard:
            self.answer("discard", "discard")
        try:
            self.cmd("quit")
        except OSError:
            pass


def main(argv):
    if len(argv) < 2:
        print(__doc__)
        return 2
    mc = McView()
    if argv[1] == "--batch":
        with open(argv[2]) as f:
            print(json.dumps(mc.batch(json.load(f)), indent=1))
        return 0
    if argv[1].lstrip().startswith("{"):
        body = json.loads(argv[1])
        print(json.dumps(mc.cmd(body.pop("cmd"), **body), indent=1))
        return 0
    name, rest = argv[1], argv[2:]
    args = {}
    positional = {"key": ["key"], "tool": ["name"], "mode": ["mode"], "menu": ["name"], "open": ["path"], "edit": [], "wait": ["ms"]}.get(name, [])
    for k, v in zip(positional, rest):
        args[k] = int(v) if v.isdigit() and k in ("ms",) else v
    print(json.dumps(mc.cmd(name, **args), indent=1))
    return 0


if __name__ == "__main__":
    try:
        sys.exit(main(sys.argv))
    except McViewError as e:
        print("mcview:", e, file=sys.stderr)
        sys.exit(1)
