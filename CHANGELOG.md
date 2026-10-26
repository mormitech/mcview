# Changelog

All notable changes to mcView are listed here, newest first. The format follows [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and the versions are the [releases](https://github.com/mormitech/mcview/releases).

## 0.30.0 – 2026-10-26

### Added
- **Circle and ellipse** tool: drag the box; Shift makes a circle. It has the width, style and outline/fill options of the rectangle in the ribbon, and the left button and the right button work as there.
- **Fill** tool (the paint bucket): a click fills the connected area of the same color with the first color, a right click with the second. The ribbon has a tolerance (0 to 100%). It is one step of the undo history, and no step when the area has that color already.

## 0.29.8 – 2026-10-26

### Changed
- The "Quit mcView?" question no longer says that Esc was pressed twice.

## 0.29.7 – 2026-10-26

### Changed
- Esc twice on the home screen asks "Quit mcView?" (Quit, Cancel) before it quits. Ctrl/Cmd+W on the home screen still quits at once.

## 0.29.6 – 2026-10-26

### Added
- Two more zoom modes, as in IrfanView: **5** zooms until the left and right edges of the picture touch the window (the top and bottom go over it, or stay short), **6** until the top and bottom edges touch it. Like the others they stay while you go through the pictures, and a double click toggles between 100% and the mode.

## 0.29.5 – 2026-10-26

### Fixed
- The page now claims the Esc key (`preventDefault`), so that the system does not get it to leave full screen, in addition to the Mac option of 0.29.4.

## 0.29.4 – 2026-10-26

### Fixed
- Esc really does not leave full screen on a Mac any more: the system itself took it (the window's own full screen), whatever the page did. **F** and **Enter** are the way out.

## 0.29.3 – 2026-10-26

### Fixed
- **F** toggles full screen on the home screen too (it did only with a picture open).

## 0.29.2 – 2026-10-26

### Changed
- Esc no longer leaves full screen, because it gets in the way of Esc twice (close the picture). **F** or **Enter** toggles full screen.

## 0.29.1 – 2026-10-26

### Fixed
- The home screen in full screen (black background): the title, the text and the logo are light now, so they can be read in the light theme too.

## 0.29.0 – 2026-10-26

### Changed
- The home screen has three big buttons with an icon, a title and a small line under it: Open file… (**O**), Open from clipboard (Ctrl/Cmd+V) and Create image (**N**). The key is written on each button.
- **O** opens the Open dialog and **N** creates a blank image, from any picture or from the home screen. If the picture has unsaved changes, the question (Save, Discard, Cancel) comes first, before the Open dialog (it used to come after a file was chosen). This goes for File → Open… too.

## 0.28.2 – 2026-10-26

### Added
- Help → mcView Help opens the manual (mormitech.com/tools/mcview/docs) in the browser. On macOS it is a Help menu after Window; on Windows and Linux it is the first item of Help, with F1 (About mcView follows it).

## 0.28.1 – 2026-10-26

### Changed
- A pencil stroke, line or rectangle that is being drawn is dropped at once by Esc or by a click of the other mouse button, while the first is still held down. Nothing is drawn into the picture.

## 0.28.0 – 2026-10-26

### Added
- The Line, Rectangle, Text and Pick tools of the edit bar work, in the way of Paint. Each has its options in the ribbon above the bottom bar.
  - Line and Rectangle: drag it out; Shift makes the line go at a multiple of 45° and the rectangle a square; Esc drops it. The left button uses the first color, the right button the second. Options: width, and a solid, dashed or dotted style. The rectangle also has outline, outline and fill (the second color fills) and fill only (the first color).
  - Text: click to put a box on the picture and type; click elsewhere, pick another tool or press Ctrl/Cmd+Enter to draw it in, Esc to drop it. Options: font, size, bold and italic.
  - Pick: a click takes the color of the pixel as the first color, a right click as the second, and the tool you had comes back.
- Every one of them is a step of the undo history. The drawing goes to the picture as a small PNG (`DrawLayer`), so big pictures stay quick.

## 0.27.0 – 2026-10-26

### Added
- Esc pressed twice within a quarter of a second closes the picture (asking about unsaved changes first), and closes the window when there is no picture. The first Esc does what it did before.

## 0.26.4 – 2026-10-26

### Changed
- A new blank image (Create Image) is not modified: no dot in the title, and closing it (Cmd+W, Cmd+Q, the window's close button, the ✕, going to another picture) asks nothing. The first change to it, a stroke or a pasted or moved part, makes it a picture with unsaved changes; undoing back to the blank one makes it blank again.

## 0.26.3 – 2026-10-26

### Changed
- Ctrl+W (Cmd+W) closes the picture when there is one, and the window when the window is empty. File → Close Window is now File → Close.

## 0.26.2 – 2026-10-26

### Changed
- The pencil cursor shows only over the picture; beside it the pointer is the ordinary one.
- The bottom bar no longer repeats the size, place in the folder and zoom (the title has them); it shows them in full screen only, where there is no title. The size of a selection is still there.

## 0.26.1 – 2026-10-26

### Changed
- Close Image on a picture with changes asks the same question as closing the window: Save, Discard or Cancel.

## 0.26.0 – 2026-10-26

### Changed
- S on a picture from a file that has changes (or is turned) asks first: Overwrite, Save As… or Cancel. Overwrite writes the changes into the file; Save As… opens the dialog with the name followed by the date and time (`photo-20261026-153012.jpg`).

## 0.25.3 – 2026-10-26

### Fixed
- Letting go of a selection (Esc, a click) ends the clipboard highlight at once, and so does starting a new one. Before, the clipboard icon of the old selection could appear again over the next one while its button was still held.

## 0.25.2 – 2026-10-26

### Fixed
- A selection starts only when the mouse covers more than one pixel. Holding the button down no longer shows a 1 × 1 frame, and a drag that stays on one pixel selects and copies nothing.

## 0.25.1 – 2026-10-26

### Fixed
- The pencil stops the moment the mouse button is let go. Before, it kept following the pointer until the stroke was drawn into the picture, which took longer on big pictures.

## 0.25.0 – 2026-10-26

### Changed
- While a selection is drawn or resized, its size is written in the middle of it (white with a black outline, so that it can be read on any picture).
- A copy to the clipboard lights things up for a moment: the border of the selection fades to orange and white and back to its usual black and dirty white, a clipboard fades in and out in the middle, and the message in the status bar flashes orange and settles. The border is black and dirty white at rest again.

## 0.24.0 – 2026-10-26

### Changed
- Zoom mode 2 (the key 2) zooms every image, up or down, so that it fills 85% of the window. It is the mode every window starts with (it was 3), so a new image, a small one and a big one all show whole with a margin around them.

## 0.23.2 – 2026-10-26

### Fixed
- Edit mode cannot be turned on in the window without an image, and it ends whenever the window gets empty (File → Close Image, the close button, deleting the last image, undo back to the empty window).

## 0.23.1 – 2026-10-26

### Added
- In edit mode the status bar has a close (x) button at its right end. It closes the image (any image) and goes back to the empty window, and edit mode ends. Undo brings the image back, as with File → Close Image.

## 0.23.0 – 2026-10-26

### Added
- File → Create Image makes a new white 1280 × 720 image and starts edit mode. The empty window's "Create image" button does the same (it replaces "New picture…" and its size fields; the size will be configurable later).

### Changed
- Names: "Open file…" (the button of the empty window, File → Open File…), "Delete Image…" and "Close Image".

## 0.22.1 – 2026-10-26

### Added
- File → Delete Picture… does the same as the key D.

## 0.22.0 – 2026-10-26

### Added
- D deletes the picture after a question (on macOS the file goes to the Trash). The folder is counted again and the picture that comes next in the folder is shown at the same place (3 / 8 becomes 3 / 7); after the last picture the one before it, and the empty window when no picture is left. Undo does not bring a deleted picture back.

### Changed
- F toggles full screen, the same as Enter (it used to choose zoom mode 2; the key 2 still does).
- US English spelling: "color", not "colour".
- Keys typed into a field of the empty window (the size of a new picture) are not taken as commands.

## 0.21.0 – 2026-10-26

### Changed
- The pencil draws a round stroke 5 pixels wide by default. While the pencil is the tool, a ribbon at the bottom has a slider for its width (1 to 64 pixels) and a dot that shows it; the color palette takes its place while it is open.

## 0.20.0 – 2026-10-26

### Added
- The window without a picture offers "Open from clipboard" when the clipboard has a picture (checked when the window opens and when it gets focus again), and "New picture…", which asks for a size and whether the picture is white or transparent.

## 0.19.1 – 2026-10-26

### Changed
- The border of the selection is black and dirty white while it is being drawn, and orange and dirty white once the button is released.

## 0.19.0 – 2026-10-26

### Changed
- A finished selection (the button is released after drawing it, or after moving or resizing its frame) is put on the clipboard at once, with a short note in the status bar.
- The moving border of the selection is black and orange instead of black and white.
- S (Save As) with a selection saves only the selected part, as a picture of its own; the window keeps showing its picture. Without a selection the whole picture is offered, as before.

## 0.18.4 – 2026-10-26

### Fixed
- The pencil no longer freezes the window: the line between two points of a stroke could go on for ever.

## 0.18.3 – 2026-10-26

### Fixed
- The status bar has one height. It used to get lower when the pixel readout was not shown (the pointer outside the picture).

## 0.18.2 – 2026-10-26

### Fixed
- The status bar no longer takes the pointer. Moving the pointer to the bottom of the window in edit mode used to leave the picture, which cleared the pixel readout and made the bar flicker.

## 0.18.1 – 2026-10-26

### Fixed
- In edit mode the pointer at the bottom of the window no longer does anything to the status bar, which edit mode shows already.

## 0.18.0 – 2026-10-26

### Added
- Edit bar: a Select tool at the top, which is the one chosen whenever edit mode starts (the selection works as before).
- Edit bar: the pencil draws. A stroke is one pixel wide, in the first color with the left button and in the second color with the right button, and is one step of the undo history. Choosing the pencil lets go of the selection.

## 0.17.0 – 2026-10-26

### Changed
- View mode has no status bar. What it said about the picture (its size, its place in the folder, the zoom) is in the title of the window. The bar shows in edit mode, while there is a selection, for a message, and in full screen while Shift is held or the pointer is at the bottom.

## 0.16.1 – 2026-10-26

### Fixed
- With the edit bar open, the status bar starts at its right edge instead of running under it, so the two colors at the bottom of the edit bar are not covered.

## 0.16.0 – 2026-10-26

### Added
- Edit bar: the two colors (the drawing color and a second one) are at the bottom. Clicking one slides a palette up above the status bar, with the 20 basic colors and "Edit colors…" for the system picker. It closes with a click anywhere else, and when the pointer leaves it after a color was changed.
- Edit bar: buttons for a pencil and a color picker (they do nothing yet, like the line, rectangle and text buttons).

## 0.15.0 – 2026-10-26

### Added
- The title of the window says the size of the picture, after its name.
- In full screen the bottom bar shows while the pointer is at the bottom of the screen, and goes when it leaves.
- In full screen with edit mode off, the bottom bar stays for as long as there is a selection.

## 0.14.4 – 2026-10-26

### Fixed
- The items of the status bar are centred vertically, so the pixel readout lines up with the file name and the picture info.

## 0.14.3 – 2026-10-26

### Changed
- The pixel readout in the status bar is easier to read: the position is written x and y, each color channel has a small meter, and the hex code sits in a pill with the color swatch.

## 0.14.2 – 2026-10-26

### Changed
- Tab toggles edit mode, the same as E.

## 0.14.1 – 2026-10-26

### Changed
- Every window starts in zoom mode 3 (contain). The mode chosen with the keys 1 to 4 is kept while the window is open, but no longer for the next start or for a new window (Ctrl+N).

## 0.14.0 – 2026-10-26

### Changed
- Edit mode (E) shows in the title ("[edit]") and in the status bar ("EDIT"). Turning it on switches to the pixelated view and turning it off goes back to the smoothed one; Tab does nothing for now.
- In full screen, edit mode shows the status bar along with the edit bar; with it off both are hidden.

## 0.13.1 – 2026-10-26

### Changed
- With edit mode off, paste always makes the clipboard picture the picture of the window, dropping the open one (undo brings it back). With edit mode on, paste works as before.

## 0.13.0 – 2026-10-26

### Changed
- Edit mode (E) now decides what the selection does. With edit mode off, dragging the inside of the selection moves only the frame, and dragging an edge resizes only the frame; the picture is not touched and nothing goes into the undo history. With edit mode on, dragging the inside lifts and moves the part, and dragging an edge stretches it, like Paint; these steps can be undone.

## 0.12.29 – 2026-10-25

### Fixed
- On Windows the menu shortcuts Ctrl+O, Ctrl+W, Ctrl+Shift+W and Ctrl+Q also work while the menu bar is hidden in full screen (the others were already handled by the page).

## 0.12.28 – 2026-10-25

### Changed
- Full screen has a black background. On Windows the menu bar is hidden in full screen and shows while the pointer is at the top edge of the screen.

## 0.12.27 – 2026-10-25

### Added
- **Installers for macOS and Linux.** A `.dmg` for macOS (open it and drag mcView to Applications), and a `.deb` (Debian, Ubuntu) and an `.rpm` (Fedora, openSUSE) for Linux, which also put mcView in the application menu and in Open With.
  The release workflow builds them and tests them (the disk image is opened and its app is checked; the .deb is installed and removed; the .rpm is checked for its files and requirements), so they come with every release from now on.
  No AppImage yet: it would have to carry WebKit with it.

## 0.12.26 – 2026-10-25

The steps 0.12.0 to 0.12.25 were never released one by one; this is everything since 0.11.3.

### Added
- **Undo and redo** (Ctrl+Z, Ctrl+Shift+Z / Ctrl+Y; Cmd+Z, Cmd+Shift+Z on a Mac, and Edit → Undo, Redo). Every change of what the window shows is a step: another picture (Open, the arrow keys, paste, save to another
  folder), the mode (1 to 4, F), the pixelated view (Tab), the turn (R), a moved or stretched part, a pasted layer, Close Image. A step brings the view back the way it was when it was made (the zoom, its mode, the part that is visible);
  zooming and panning themselves are not steps. Going back to the first state gives the empty window.
- **Rotate the view** with R (Shift+R the other way), a quick animation; it keeps the pixel that was in the middle in the middle, and so does full screen (Enter) and back. Only the view turns, until you save.
- **Select by dragging** with the mouse (no Shift needed); a click takes the selection away. The border is like Paint's: still dashes in black and dirty white while the button is down, "marching ants" once it is set; the selected part is a
  little lighter. With Shift, the edges of the frame move the frame only. The picture is panned with the middle mouse button (the pointer shows the grabbing hand), the trackpad and the wheel as before.
- **Move and stretch the selected part, like Paint.** Drag an edge or corner to stretch it, drag the inside (or the top left corner) to move it. The place it came from turns white (transparent in a picture that has transparency) and the part
  stays a layer until the selection is let go (Esc, a click beside it, a new selection, R, copy, paste, save, closing); then it is applied. Every drop and the applying are steps of undo. The picture is not recoded for this: it is kept as decoded pixels and only
  encoded again by Save As, and the color profile of the file (ICC of a JPEG or WebP, the color chunks of a PNG) is carried along, so the colors do not change when you edit.
- **Paste is a layer.** Ctrl+V (Cmd+V, Edit → Paste) lays the clipboard picture over the picture at its own size, on pixels near the top left corner of what is visible, selected and floating like a lifted part: drag it, stretch it, take it away with
  Delete, Backspace or Esc, or lay it down with a click. A clipboard picture that covers about half of the picture or more (or does not fit in it) is probably a new image, so mcView asks "Large image content is about to be pasted":
  Paste as new image, Paste into current image, Cancel. A window with no picture takes the clipboard picture as its picture. It works on JPEG, PNG, GIF, BMP, WebP and TIFF pictures.
- **A ● after the name of a modified picture** in the title and, in full screen, in the bottom bar.
- **A native alert before unsaved changes are lost.** Going to another picture or closing the window with a modified picture asks first, in the system's own alert: on macOS Save, Discard, Cancel; on Windows Yes, No, Cancel; on Linux GTK's two-button question
  twice. Save starts Save As and goes on once the file is written.
- **Transparent pictures get squares** (small bluish and white ones, as Gimp does) instead of a flat color. Save As offers such a picture as PNG, writes GIF with a transparent color, and says when JPEG made the transparent parts white.
- **File → Close Image** (Ctrl+Shift+W, Cmd+Shift+W): back to the way the window starts, with no picture; undo brings the picture back. **File → Close Window** (Cmd+W) on a Mac.
- **The window comes up where and how big it was left** (`window.json` in the settings folder of the user); full screen, maximised and minimised windows are not remembered, and a place that is no longer on a screen is ignored.
- **The edit bar** (key E) slides in a bar of tools from the left, like the bottom bar in full screen; Line, Rectangle and Text are there but do nothing yet.
- **In full screen the bottom bar is hidden** until Shift is held, slides and fades in and out, and shows the file name.
- With MCVIEW_TRACE set, a failure of the page shows in the trace.

### Changed
- **Save As saves what you see:** the picture as it is turned and with what was pasted or moved, starting in the folder of the picture (the Pictures folder for one from the clipboard). The saved file is the picture of the window afterwards. A picture of a type mcView does not write
  (WebP, SVG…) is offered as PNG.
- The next or the previous picture is decoded off screen and takes the place of the old one in one step, so nothing flickers; the same for an edit that is applied.

### Fixed
- Windows: the installer registered the "Open with" entries with values that were never created (fixed in 0.11.3 as well).
- A selection that mcView made itself (a pasted picture, a moved part, the selection after an undo) was one pixel too big in the pixelated view.
- The dashed border of a selection left traces behind while it was dragged.
- The circle of the title no longer goes out for a moment on undo and redo of a moved part, and undo no longer flashes the untouched picture.

## 0.11.3 – 2026-10-25

### Fixed
- **Windows: "Open with…" did not list mcView.** The installer wrote the picture types as registry values of type "none", which Inno Setup does not
  create at all; they are empty strings now. (Found by the installer test of the first GitHub build.)

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
  whole picture. The part is cut out of the decoded file, so it has the colors of the file (transparency included). The web view's own
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
- **Pixel readout.** In the pixelated view the bottom bar shows the color of the pixel under the pointer: its position, R G B and
  alpha (0 to 255), the hex code and a swatch (a checkerboard shows through a transparent pixel).
  - The numbers are those of the file: mcView decodes the picture a second time in Go, without any color conversion (the web view
    converts colors for the screen, and its JPEG decoder can be one off in the last digit). JPEG, PNG, GIF, BMP, WebP and TIFF.
    A JPEG turned by its EXIF Orientation tag is read the way it is shown.
  - For types Go cannot decode (AVIF, SVG, ICO) the color is read from what the web view drew, and the bar says `from the screen`.
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
