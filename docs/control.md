# Control mode

mcView can be driven from another program, so that every feature can be triggered without a mouse or a keyboard: for demo
screenshots, for the manual in pictures, for scripted tests. It is off unless you ask for it.

```
mcview --control 0 picture.jpg          # 0 takes a free port; or MCVIEW_CONTROL=0 in the environment
```

On start the app writes `mcview-control.json` (in the temporary folder, or where `MCVIEW_CONTROL_FILE` says; readable by you only):

```json
{"url": "http://127.0.0.1:53211", "token": "…", "pid": 4711}
```

Every request needs `Authorization: Bearer <token>`. Requests with an `Origin` header (a web page) or for another host are refused, and
the server only listens on 127.0.0.1.

## Requests

| | |
|---|---|
| `POST /cmd` | one command, a JSON object with `"cmd"`; the answer is JSON, `{"error": "…"}` when it failed |
| `POST /batch` | a JSON list of commands, done one after the other; stops at the first error (`?continue=1` goes on) |
| `GET /info` | version, pid, and the window: `windowId` (macOS: for `screencapture -l`), `x`, `y`, `w`, `h` |

`scripts/mcview_ctl.py` is a small client (and a library) for it; `scripts/demo_screenshots.py` makes the screenshots of the manual with it.

## Positions

Commands take **pixels of the picture, as it is shown**: `{"x": 120, "y": 80}` is the pixel in the 121st column and 81st row. Nothing here
moves a pointer or presses a key: a drawing command puts its result into the picture at once, like the other edits.

## Commands

| cmd | arguments | does |
|---|---|---|
| `state` | | what the window shows: name, index and count, size, mode, zoom, edit, tool, selection, colors and options, modified, the history |
| `open` | `path` | opens a picture of a folder (no dialog) |
| `window` | `w`, `h`, `x`, `y` | size and place of the window (centred without `x`) |
| `info` | | see above |
| `menu` | `name` | `open`, `save`, `close`, `close-image`, `create`, `delete`, `undo`, `redo`, `copy`, `paste` |
| `edit` | `on` | edit mode on (default) or off |
| `tool` | `name` | `select`, `pen`, `line`, `rect`, `ellipse`, `fill`, `text`, `pick` (turns edit mode on) |
| `mode` | `mode` | the zoom modes `1` to `6` |
| `zoom` | `scale`, `x`, `y` | an exact zoom (`1` is 100%) with the pixel `x`, `y` (default: the middle) in the middle of the window |
| `zoomby` | `factor`, `x`, `y` | zoom by a factor around a pixel (the wheel) |
| `pan` | `x`, `y` | the pixel in the middle of the window |
| `rotate` | `dir` | a quarter turn, `1` clockwise, `-1` counterclockwise |
| `fullscreen` | `on` | full screen on or off |
| `color` | `fg`, `bg` | the two colors, `#rrggbb` |
| `option` | `width`, `style` (`solid`, `dashed`, `dotted`), `fill` (`outline`, `both`, `fill`), `tolerance` (0 to 100), `font`, `size`, `bold`, `italic` | the options of the tools |
| `palette` | `open: "fg"` or `"bg"` | the palette slides up (`open` left out: it closes) |
| `shape` | `kind` (`line`, `rect`, `ellipse`, `path`), `x0`, `y0`, `x1`, `y1` or `pts: [[x, y], …]`, `color`, `bg`, `width`, `style`, `fill` | puts a shape into the picture at once (one step of the history); no pointer is used |
| `fill` | `x`, `y`, `color`, `tolerance` (0 to 100) | the paint bucket at the pixel |
| `select` | `x0`, `y0`, `x1`, `y1` | a selection with these two pixels as its corners |
| `text` | `x`, `y`, `value`, `color` | puts a text into the picture with its top left corner at the pixel (font and size: `option`) |
| `pixel` | `x`, `y` | the color of the pixel, from the file |
| `snapshot` | `x`, `y`, `w`, `h`, `max` | the picture as a PNG (base64 in `png`, with `scale`, `region`, `picture`); used by the [chat panel](chat.md) |
| `saveto` | `path`, `overwrite` | saves the picture to a full path without the Save dialog |
| `wait` / `sleep` | `ms` | a pause |
| `answer` | `kind`, `value` or `values` | queues answers for the dialogs (below) |
| `dialogs` | | the questions the dialogs have asked since the last time |
| `quit` | | quits, without asking |

## Dialogs

The system's dialogs cannot be clicked from outside, so in this mode they are answered from queues that you fill with `answer`. The
question that has no answer queued is answered with the safe one (Cancel, No, no file). Every question is noted; `dialogs` tells.

| kind | the answers |
|---|---|
| `discard` (unsaved changes) | `save`, `discard`, `cancel` |
| `overwrite` (S on a changed file) | `overwrite`, `saveas`, `cancel` |
| `large` (a big clipboard picture) | `new`, `over`, `cancel` |
| `delete` | `yes`, `no` |
| `quit` (Esc twice on the home screen) | `yes`, `no` |
| `open` | the path of the picture the Open dialog returns |
| `save` | the path the Save dialog returns |

## Example

```python
from mcview_ctl import McView
mc = McView.start("build/bin/mcview.app", "/tmp/Pictures/coast.jpg")
mc.cmd("window", w=1500, h=960)
mc.cmd("shape", kind="path", pts=[[3200, 900], [3300, 850], [3400, 920]], color="#ED1C24", width=14)
mc.cmd("zoom", scale=4, x=3300, y=900)
mc.shot("zoomed.png")
mc.quit()
```
