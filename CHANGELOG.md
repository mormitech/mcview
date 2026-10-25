# Changelog

All notable changes to mcView are listed here, newest first. The format follows [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and the versions are the [releases](https://github.com/mormitech/mcview/releases).

## 0.11.2 – 2026-10-25

### Changed
- **The file name is in the bar in full screen**, with its extension, at the left (the title bar that normally says it is not there).

## 0.11.1 – 2026-10-25

### Fixed
- **Cmd+W did nothing on a Mac.** File → Close Window (Cmd+W) closes the window, like the red button does (which ends mcView, as it is one window per
  program). In the About window Cmd+W (Ctrl+W elsewhere) and Esc close only that window.

## 0.11.0 – 2026-10-25

### Added
- **Marching ants on the copied selection.** After Ctrl+C (Cmd+C) the frame of the selection that was copied is drawn as a dashed line that
  moves, to show that this part is on the clipboard. It stays so until the frame is changed or moved; clearing the selection (a click, Esc) does not
  touch the clipboard, the frame is just not marked any more.

## 0.10.2 – 2026-10-25

### Changed
- **The bar in full screen slides and fades** in and out (0.12 s) instead of appearing at once.

## 0.10.1 – 2026-10-25

### Changed
- **No bar at the bottom in full screen.** The bar with the folder, the picture's place and size and the pixel readout is hidden in full screen,
  and shown for as long as Shift is held (a message, like "Copied …", also shows it for a few seconds).

## 0.10.0 – 2026-10-25

### Added
- **Resize the selection.** With a selection, hold Shift and move over one of its edges (or a corner): the edge lights up blue and the pointer
  becomes a resize arrow. Shift and drag then moves that edge (a corner moves two), in the pixelated view snapped to whole pixels.

## 0.9.1 – 2026-10-25

### Changed
- **Full screen (Enter) and back keep the middle.** When the picture is zoomed or moved, the pixel that was in the middle of the window is in the
  middle of the full screen, and again in the middle of the window when you leave it. This holds for any resize of the window. A picture that
  cannot be held in the middle (it would show a gap at its edge) stays at its edge, and comes back to the middle when there is room again.

## 0.9.0 – 2026-10-25

### Added
- **Undo and redo.** Ctrl+Z (Cmd+Z on a Mac) undoes, Ctrl+Shift+Z (Cmd+Shift+Z; also Ctrl+Y) redoes, and Edit → Undo / Redo do the same. Undoable are every
  change of what the window shows: opening a picture (also from the file manager, and from an empty window back to empty), the arrow keys
  and Space/Backspace through the folder, pasting, saving to another folder, the modes 1 to 4 and F, the pixelated view (Tab) and the turn (R).
  Not steps: how far it is zoomed, where it is moved, the selection. The history keeps 100 steps; a pasted picture stays in it, so
  undoing a paste brings the picture before it back, and redoing brings the pasted one back.

## 0.8.1 – 2026-10-25

### Changed
- **Turning a zoomed picture keeps the zoom and the middle.** When the picture was zoomed or moved, R and Shift+R no longer go back to the
  mode's fit: the scale stays, and the pixel that was in the middle of the window is in the middle after the turn. A picture that
  the mode (1 to 4) places is still fitted to the turned picture.

## 0.8.0 – 2026-10-25

### Added
- **Rotate the view.** **R** turns the picture a quarter clockwise, **Shift+R** a quarter counterclockwise, with a quick 0.2 s animation. It is only the
  view: no file is changed, and the next picture is shown upright again. The mode (1 to 4) fits the turned picture to the window; the
  size in the bar, the selection (cleared by a turn), the pixel readout (positions as on the screen) and Ctrl+C (the copy is turned the way it is shown)
  all follow the turned picture.

## 0.7.3 – 2026-10-25

### Changed
- **The bounce at the edge is a spring now, and comes back at once.** After the fingers have left, the trackpad keeps sending scroll events for
  about a second, each smaller than the one before (it lets the movement coast), and the old bounce waited for them to stop, so the
  picture stayed out long after the swipe was over. Now a spring takes the pull away every frame, and events that only get smaller do
  not push the picture out again: in a test with a swipe followed by 0.6 s of coasting, the picture was back **275 ms** after the last
  push (it took 600 ms and more, with a second small bounce, before).
- **A picture that fits the window bounces too** when you swipe or drag it, so that it is plain there is nowhere to pan.

## 0.7.2 – 2026-10-25

### Added
- **Bounce at the edge.** A picture bigger than the window can be pulled a little past its edge (up to 20 pixels, and the harder the further
  it goes) with the mouse button held or two fingers on the trackpad, and eases back when it is let go. A picture that fits the window on an
  axis does not move on that axis.

### Fixed
- **Ghost lines under or beside the picture after zooming out fast.** One-pixel lines of the old edges stayed on the screen, because the
  picture had fractional sizes and positions and the web view did not repaint the sliver at the old edge. The picture now has whole
  pixel sizes and positions. In a test that zooms out from 300% to the minimum with a burst of wheel events, 8 of 8 windows had
  stale lines before and none has now.

## 0.7.1 – 2026-10-25

### Added
- **Ctrl+N opens a new window** on every system, as well as in the File menu. On Windows and Linux that is the usual key; on a Mac the menu
  has Cmd+N and Ctrl+N works too. Both together never open two windows.

## 0.7.0 – 2026-10-25

### Added
- **File → New Window** (Cmd+N, Ctrl+N) opens an empty window, a little down and to the right of the one it came from.
- **Paste a picture from the clipboard** (Cmd+V, Ctrl+V, Shift+Insert, Edit → Paste): the window shows it. The title says
  `[clipboard image]` instead of a file name, and the bar shows the Pictures folder of the user. The pixel readout, the selection and
  Ctrl+C work on it. The arrow keys do nothing but say so while the picture is not saved, so that it is not lost by accident, and a file
  opened from the system gets a window of its own instead of replacing it.
- **S saves the picture** (also Cmd+S, Ctrl+S and File → Save As…) in a dialog that starts in the folder of the picture (the Pictures
  folder for a pasted one). The type follows the name: PNG, JPEG, GIF, BMP or TIFF; a name without one gets `.png`. A file saved as the
  type it already is is copied byte for byte; any other is decoded and written again, turned the way it is shown. A pasted picture
  becomes the file it was saved to, and **saving to another folder makes that folder the one the window works in**: the bar shows it,
  and the arrow keys go through its pictures.
- An **Edit menu** with Copy and Paste on every system (the system's own work on text in the page, not on a picture).

### Changed
- The first window is the same as before; the hint in an empty window now mentions pasting.

## 0.6.0 – 2026-10-25

### Added
- **Select a part of the picture: hold Shift and drag.** The selection is a white frame with a soft dark shadow, so it shows on a white
  picture as well. In the pixelated view it snaps to whole pixels; in the smoothed view it follows the pointer, and the pixels it touches
  are selected. It stays on its place when the picture is zoomed or moved. A click on the picture, Esc, or another picture takes it away.
  The bar shows its size (`selection 120 × 80`).
- **Ctrl+C (Cmd+C on a Mac, and Ctrl+Insert) copies the selected part to the clipboard** as a picture; with nothing selected it copies the
  whole picture. The part is cut out of the decoded file, so it has the colours of the file (transparency included). The web view's own
  clipboard is tried first, then the system's way (osascript on macOS, PowerShell on Windows, `xclip` or `wl-copy` on Linux, which have to
  be installed). AVIF, SVG and ICO are cut from what the web view drew.

### Changed
- The pixel readout reads `[x, y]  (r, g, b) - #rrggbb`, with the alpha as a fourth number and at the end of the hex code when the pixel
  is not opaque.

## 0.5.0 – 2026-10-25

### Added
- **Tab switches between the smoothed view and the pixelated view**, where every pixel of the picture is a sharp square and nothing is
  blended. Tab again goes back. The pixelated view stays when the next or the previous picture is shown, but not the next time mcView
  starts (it always starts smoothed). The bar says `pixels` while it is on.
- **Pixel readout.** In the pixelated view the bottom bar shows the colour of the pixel under the pointer: its position, R G B and
  alpha (0 to 255), the hex code and a swatch (a checkerboard shows through a transparent pixel).
  - The numbers are those of the file: mcView decodes the picture a second time in Go, without any colour conversion (the web view
    converts colours for the screen, and its JPEG decoder can be one off in the last digit). JPEG, PNG, GIF, BMP, WebP and TIFF.
    A JPEG turned by its EXIF Orientation tag is read the way it is shown.
  - For types Go cannot decode (AVIF, SVG, ICO) the colour is read from what the web view drew, and the bar says `from the screen`.
  - A picture over 150 megapixels is not decoded for the readout.

## 0.4.1 – 2026-10-25

### Added
- **Space** shows the next picture of the folder, like the right arrow, and **Backspace** the previous one, like the left arrow.

## 0.4.0 – 2026-10-25

### Added
- **Enter** toggles full screen, and **Esc** leaves it. On macOS this is the system's full screen, in a Space of its own.

## 0.3.0 – 2026-10-25

### Added
- **Zoom modes on the keys 1 to 4**, the way IrfanView has them: **1** original size (100%, even when the picture is small or bigger than
  the window), **2** big pictures are shrunk to fit and small ones stay 100% (the default, and what **F** does), **3** every picture
  is zoomed up or down to touch the window borders, **4** every picture is stretched to the window borders. The mode stays when the
  next or the previous picture is shown, and the next time mcView starts. Zooming or moving with the mouse leaves the mode until the
  next picture. In the stretch mode the bar shows both zooms (`34% × 91%`).
- **Mac trackpad:** two fingers up and down (or sideways) move the picture instead of zooming it. A mouse wheel still zooms, and so does a
  pinch. The two are told apart by the size of the scroll steps.

### Changed
- **Left and right arrow go round**: after the last picture of the folder comes the first (8/8 → 1/8), and before the first comes the last.
  The "last picture" message is gone.

## 0.2.0 – 2026-10-25

### Added
- **Left and right arrow** show the previous and the next picture of the folder of the current one, in alphabetical order (upper and
  lower case alike). At the first or the last picture the bar says so and nothing moves.
- The bottom bar shows the folder of the picture (the name is in the title), and on the right which picture of the folder it is
  (`2 / 3`), its size in pixels and the zoom.
- `MCVIEW_TRACE=<file>` writes timestamps of the start and of showing a picture to that file, for tracking down slow starts.

### Fixed
- **A picture opened with Open With, a double-click or `open -a` took about 5 seconds to show.** macOS starts an app hidden when it is
  launched with a file, and Wails did not unhide it, so the window (and the picture) came up only after a timeout. The window is now
  shown at once: a 36 megapixel JPEG is on the screen in under a second, a small one in about 0.4 s.
- The hand ("grab") cursor over the picture is gone. It was also showing over the title bar buttons.

### Changed
- The picture is shown at its size on the screen instead of scaled with a CSS transform.

## 0.1.0 – 2026-10-25

### Added
- **mcView shows pictures.** JPEG (jpg, jpeg, jpe, jfif), PNG, GIF, BMP, WebP, ICO, SVG, AVIF and TIFF. The picture is fitted to the window;
  scroll to zoom, drag to move, double-click for 100% and back, **F** fits, **1** is 100%, **+** and **-** zoom. The window title and a
  status line show the name, the size in pixels and the zoom.
- **Open with…** offers mcView for these file types, with no signing needed.
  - macOS: the app declares the types in its Info.plist. It is offered in Open With and never takes a type over by itself;
    choose it in Get Info → Open with → Change All to make it the default.
  - Windows: the installer adds the types per user (no admin rights); mcView is in Open with… and in Settings → Apps → Default apps.
    Windows does not let an installer make a program the default.
  - Linux: `install.sh` in the tar.gz puts a desktop entry in `~/.local/share/applications`.
- File → Open… (⌘O, Ctrl+O). A second picture opened from the file manager gets a window of its own.

### Changed
- The Hello World window is the empty state of the viewer.

## 0.0.1 – 2026-10-25

### Added
- The app stub: a window that says Hello, world. No functionality yet.
- An About window (About mcView in the application menu on macOS, Help → About mcView on Windows and Linux) with the Mormitech logo,
  the version and links.
- The window is cut down to 90% of the screen at start and centred, so its title bar stays reachable with a high scaling setting.
- `--about` and `--theme light|dark` command line options.
- Builds and tests on macOS, Windows and Linux, and a Windows installer (Inno Setup).
