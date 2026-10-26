<h1 align="center">mcView</h1>

<p align="center">
  <b>A picture viewer for macOS, Windows and Linux.</b><br>
  Early preview: it shows pictures, and it is offered in Open with… for them.
</p>

<p align="center">
  <a href="https://github.com/mormitech/mcview/actions/workflows/build.yml"><img alt="build" src="https://github.com/mormitech/mcview/actions/workflows/build.yml/badge.svg"></a>
  <a href="LICENSE"><img alt="MIT license" src="https://img.shields.io/badge/license-MIT-blue.svg"></a>
  <img alt="platforms" src="https://img.shields.io/badge/platforms-macOS%20%7C%20Windows%20%7C%20Linux-lightgrey.svg">
  <img alt="status" src="https://img.shields.io/badge/status-early%20preview-orange.svg">
</p>

---

## What it does

mcView shows a picture, quickly, and nothing else (it is going to be an [IrfanView](https://www.irfanview.com)-like viewer). It is built
the same way as [mcPrint](https://github.com/mormitech/mcprint): Go and [Wails](https://wails.io) v2 for the window, a vanilla JS
frontend built with Vite.

- **Shows pictures.** JPEG, PNG, GIF, BMP, WebP, ICO, SVG, AVIF and TIFF (AVIF and WebP need a web view that has them; TIFF is
  decoded by mcView itself, so it also works on Windows and Linux). Four zoom modes on the keys 1 to 4, as in IrfanView, which stay when you
  go to another picture (and the next time you start mcView): **1** original size (100%, however small or big), **2** shrink big
  pictures to fit the window, small ones stay 100% (the default), **3** fit every picture to the window borders, keeping its proportions,
  **4** stretch every picture to the window borders. **+** and **-** zoom, dragging moves the picture, double-click toggles 100%.
  A mouse wheel zooms; on a Mac two fingers on the trackpad move the picture, and a pinch zooms.
- **Arrow keys** (and Space for next, Backspace for previous). Left and right show the previous and the next picture of the folder, in alphabetical order, and go round (after the last comes the
  first). The bar at the bottom shows the folder, and which picture of it you are looking at.
- **Pixel view and color readout.** Tab switches between smoothed and pixelated (every pixel a sharp square); in the pixelated view the
  bar at the bottom shows position, R G B A and hex of the pixel under the pointer, as `[x, y]  (r, g, b) - #rrggbb`. The numbers are the ones in the file (mcView decodes
  it itself, with no color conversion), except for AVIF, SVG and ICO, where they are read from what the web view drew and the bar says so.
- **Rotate the view.** R turns the picture a quarter clockwise, Shift+R counterclockwise (a quick animation). Only the view turns; the file is not changed.
- **Paste.** Ctrl+V lays the clipboard picture over the picture, near the top left corner of what you see, as a floating part you can drag and stretch (Delete, Backspace or Esc take it away; a click beside it lays it down); the name gets a ● until you save it. A clipboard picture that is about as big as the picture asks whether it is a new image; a window with no picture takes it as its picture.
- **Move and stretch a selection (like Paint).** Without Shift, drag an edge of the selection to stretch the selected part, or its top left corner to move it (the place it
  came from turns white). Shift and drag an edge still moves just the frame. Undo steps back through every drop.
- **Transparency.** A picture with transparent parts has chequered squares behind it; a part lifted out of one leaves a transparent hole, and Save As offers PNG.
- **Undo and redo.** Ctrl+Z / Ctrl+Shift+Z (Cmd+Z / Cmd+Shift+Z on a Mac). Every change of the picture shown, of the mode (1–4), the pixelated view and the turn can be undone, including
  the arrow keys and paste; zooming, moving and the selection are not steps.
- **Select and copy.** Drag with the mouse to select a part of the picture (it snaps to whole pixels in the pixelated view); Ctrl+C
  (Cmd+C on a Mac, Ctrl+Insert) puts just that part on the clipboard, or the whole picture when nothing is selected. Shift and drag an edge or corner of the frame (it lights up) to resize it. The copied frame has moving dashes. Esc or a click takes the
  selection away (the clipboard keeps the picture). On Linux copying a picture needs `xclip` (X11) or `wl-clipboard` (Wayland).
- **New window, paste, save.** File → New Window (Ctrl+N; Cmd+N on a Mac) opens an empty window; Cmd/Ctrl+V pastes the picture on the clipboard into it (the title says
  `[clipboard image]`, the folder is your Pictures folder); **S** saves the picture as PNG, JPEG, GIF, BMP or TIFF. Saving to another folder makes
  that folder the one the window works in, so the arrow keys go through its pictures. On Linux pasting and copying pictures needs `xclip` or `wl-clipboard`.
- **Bounce.** Pull a big picture past its edge with the mouse or the trackpad and it gives a little (up to 20 pixels), then eases back.
- **Full screen.** Enter toggles it, Esc leaves it.
- **Open with… offers mcView.** [See below](#open-with). No code signing needed.
- **File → Open…** (⌘O or Ctrl+O). A second picture opened from the file manager gets a window of its own.
- **One window, three systems.** The same code builds for macOS, Windows and Linux.
- **A window that fits the screen.** At start the window is cut down to 90% of the screen and centred, so its title bar is never out of reach, also
  with a high scaling setting in Windows.
- **An About window.** *mcView → About mcView* on macOS, *Help → About mcView* on Windows and Linux. It has the Mormitech logo, the version and links.
- **Follows your system.** Light or dark mode.
- **Tests and CI.** Builds, vets and tests on macOS, Windows and Linux, and packages the app on a release tag.

## Install

Download the package for your system from the [releases](https://github.com/mormitech/mcview/releases):

| | Package |
|---|---|
| **macOS** (Apple silicon) | `mcview-<version>-macos-arm64.dmg` (open it and drag mcView to Applications), or the `.zip` with the app |
| **Windows** (x64) | `mcview-setup-<version>-windows-x64.exe` (per user, no admin rights), or the `.zip` |
| **Linux** (x64) | `mcview-<version>-linux-amd64.deb` (Debian, Ubuntu: `sudo apt install ./mcview-….deb`), `mcview-<version>-linux-x86_64.rpm` (Fedora, openSUSE: `sudo dnf install ./mcview-….rpm`), or the `.tar.gz` with `install.sh` |

The app is not code-signed or notarized yet, so macOS asks you to confirm the first start (right-click → Open). Or build it yourself (below).

<a id="open-with"></a>
## Open with…

mcView is offered for the picture types above. It does not take them over from your current viewer: you choose it.

| | How |
|---|---|
| **macOS** | Right-click a picture → **Open With** → mcView. To make it the default for a type: **Get Info** → **Open with** → mcView → **Change All…**. The app must be in Applications (or have been opened once) for macOS to know it. |
| **Windows** | Right-click a picture → **Open with** → mcView (the first time **Choose another app**). To make it the default: **Settings → Apps → Default apps → mcView**. The installer adds this per user, without admin rights. |
| **Linux** | The .deb and .rpm do this by themselves. From the tar.gz, run `./install.sh` (`./install.sh --remove` undoes it). Then right-click a picture → **Open With** → mcView. |

The list of types is written in four places, because each system wants it in its own form: `image.go`, `wails.json` (macOS Info.plist),
`packaging/windows/installer.iss` and `packaging/linux/mcview.desktop`. A test keeps them equal.

## Develop

```
cd frontend && npm install && cd ..
wails dev          # live-reloading window
go vet ./... && go test ./...
wails build        # the app ends up in build/bin
```

You need [Go](https://go.dev), Node and the [Wails](https://wails.io) CLI (`go install github.com/wailsapp/wails/v2/cmd/wails@latest`).
On Linux also `libgtk-3-dev libwebkit2gtk-4.1-dev`, and build with `-tags webkit2_41`. The Go code embeds the built frontend, so run
`npm run build` in `frontend` before `go vet` and `go test`.

| Path | What |
|---|---|
| `main.go`, `about.go` | the window, the menus per system, the `--about` window |
| `window.go` | the screen-size gate: `fitSize` and `fitWindow` |
| `app.go` | the methods the UI calls (the version, the window mode, Open…) and the extra windows |
| `image.go` | the picture types, opening a file (command line and the macOS open-file event), serving the picture to the web view |
| `frontend/` | the UI (vanilla JS, built with Vite) |
| `build/` | icons, macOS Info.plist, Windows manifest and version info |
| `packaging/windows` | the Inno Setup installer script, with the Open with… registry entries |
| `packaging/linux` | the desktop entry and `install.sh` (the .deb and .rpm are made by the workflow with fpm) |
| `.github/workflows/build.yml` | builds, tests and packages on macOS, Windows and Linux; publishes the release on a tag |

The window has one process per window (Wails v2): the About window is the same program started with `--about` (or with a picture: `mcview picture.jpg`). `--theme dark` (or `light`)
forces the look instead of the system's.

**Slow start?** Start the app with `MCVIEW_TRACE=/tmp/mcview-trace.txt` (macOS: `open --env MCVIEW_TRACE=/tmp/t.txt -a mcview.app picture.jpg`): each step of the
start, in Go and in the UI, is written to that file with a time in milliseconds.

**Versioning.** The version lives in one place, `wails.json` (`info.productVersion`). It feeds the macOS Info.plist, the Windows file
properties, the window title and the About window. A release is cut by pushing a `v*` tag that matches it; CI checks that, builds the
packages and attaches them to a GitHub release.

## License

[MIT](LICENSE): use it, change it, ship it, forever.
