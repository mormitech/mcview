import './style.css';
import {Version, Mode, Image, ChooseImage, Step, Trace, Reveal, ToggleFullscreen, ExitFullscreen, IsFullscreen, CloseWindow, PreparePixels, PixelAt, SelectionPNG, RotatePNG, SetTurn, SetFloat, ClearFloat, BakeFloat, HasTransparency, RestoreState, RestoreFloat, KeepStates, CopyImageNative, PasteLayer, PasteNew, DrawStroke, DeleteImage, DiscardOK, ClipboardHasImage, NewBlank, SaveSelectionAs, SetEditMode, SetTitleInfo, SetPasted, MovePasted, RestoreLayer, SaveAs, SaveWindow, SetMenuBar, NewWindow} from '../wailsjs/go/main/App';
import {BrowserOpenURL, EventsOn} from '../wailsjs/runtime/runtime';
import logo from './assets/mormitech-mark.svg';

const $ = (id) => document.getElementById(id);
const esc = (s) => String(s ?? '').replace(/[&<>"]/g, (c) => ({'&': '&amp;', '<': '&lt;', '>': '&gt;', '"': '&quot;'}[c]));
const errText = (e) => {
  const t = String(e?.message ?? e).replace(/^Error:\s*/, '');
  return t.charAt(0).toUpperCase() + t.slice(1);
};

// ---- About window ----------------------------------------------------------------------------------------
// The program started with --about shows this. Links open in the system's browser.

const COFFEE = '<svg viewBox="0 0 24 24" aria-hidden="true"><path d="M4.5 9.5h12v5a5 5 0 0 1-5 5h-2a5 5 0 0 1-5-5zM16.5 11h1.5a2.5 2.5 0 0 1 0 5h-1.8M8 3.5c-.9 1 .9 1.8 0 2.8M12 3.5c-.9 1 .9 1.8 0 2.8"/></svg>';
const REPO = 'https://github.com/mormitech/mcview';

async function showAbout() {
  document.body.classList.add('about-mode');
  window.addEventListener('keydown', (ev) => { // the About window has no menu: Cmd+W (Ctrl+W) and Esc close it
    if (ev.key === 'Escape' || ((ev.metaKey || ev.ctrlKey) && !ev.altKey && (ev.key === 'w' || ev.key === 'W'))) { ev.preventDefault(); CloseWindow(); }
  });
  $('about').hidden = false;
  $('about').innerHTML = `<img class="logo" src="${logo}" alt="" width="132">
    <h1>mcView</h1>
    <p class="muted">Version ${esc(await Version())}</p>
    <p>A picture viewer for macOS, Windows and Linux. Open a picture from your file manager (Open with…) or with File → Open.</p>
    <p class="muted">Free and open source, MIT license. No accounts, no ads, no telemetry.</p>
    <p class="links"><a href="${REPO}">Website and source code</a><a href="${REPO}/issues">Report a problem</a><a href="${REPO}/blob/main/LICENSE">License</a><a class="coffee" href="https://mormitech.com/support-us/">${COFFEE}Buy me a coffee</a></p>`;
}

window.addEventListener('error', (e) => Trace(`error: ${e.message} at ${e.lineno}:${e.colno}`)); // (with MCVIEW_TRACE set, a failure of the page shows in the trace)
window.addEventListener('unhandledrejection', (e) => Trace(`rejection: ${e.reason}`));
document.addEventListener('click', (ev) => {
  const a = ev.target.closest('a[href^="http"]');
  if (a) { ev.preventDefault(); BrowserOpenURL(a.href); }
});

// ---- Viewer ----------------------------------------------------------------------------------------------
// How the picture sits in the window is one of four modes, the way IrfanView has them on the keys 1 to 4, and the mode stays when
// the next or the previous picture is shown (until the window is closed; every new window starts with 2):
//   1  original  - the pixels of the picture, 100%, however small or big it is (a big one is centred and can be moved)
//   2  large     - every picture is zoomed, up or down, to fill 85% of the window, keeping its proportions (the default)
//   3  contain   - every picture is zoomed, up or down, to touch the borders of the window, keeping its proportions
//   4  stretch   - every picture is stretched border to border, proportions or not
// The scroll wheel (or a pinch on the trackpad) zooms around the pointer (on a Mac two fingers on the trackpad move the picture instead),
// dragging moves the picture, + and - zoom: that leaves
// the mode until the next picture or the next of the keys 1 to 4. A double click toggles between 100% and the mode.
// The arrow keys go to the previous and the next picture of the folder (so do Backspace and Space, as in IrfanView), Enter toggles full screen (Esc leaves it), E or Tab the edit mode (with the pixelated view), Shift and drag selects, Ctrl+C copies, Ctrl+V pastes, S saves, R turns the picture (Shift+R the other way), Ctrl+Z undoes and Ctrl+Shift+Z redoes, Ctrl+N opens a new window.

const MODES = {'1': 'original', '2': 'large', '3': 'contain', '4': 'stretch'};
const pic = $('pic');
const view = {tool: 'select', name: '', dir: '', index: 0, count: 0, mode: 'large', pixelated: false, state: 0, modified: false, canAlpha: false, fl: null, tdrag: null, hlMode: '', token: 0, sel: null, selDrag: null, rot: 0, angle: 0, free: false, ox: 0, oy: 0, sx: 1, sy: 1, x: 0, y: 0, drag: null};
const CHECKER = 'repeating-conic-gradient(#cdd7ec 0 25%, #fff 0 50%)'; // the squares of a picture with transparent parts, a bluish one and a white one
let problemTimer = 0;
let pointerDown = false;

// R turns the picture clockwise by a quarter, Shift+R counterclockwise: only the view (the file is not touched). view.rot is the number
// of quarter turns, 0 to 3; view.angle counts them without wrapping, so that the animation always turns the short way. Everything below
// that is about the picture on the screen (the mode, the pan limits, the selection, the readout) works in the turned picture, whose
// size is dispW × dispH; the file's own pixels (natural size) are only needed to read a pixel or cut a part out.
const dispW = () => view.rot % 2 ? pic.naturalHeight : pic.naturalWidth;
const dispH = () => view.rot % 2 ? pic.naturalWidth : pic.naturalHeight;

// A position of the turned picture (fractions included) as a position of the file's picture.
function toNatural(X, Y) {
  const nw = pic.naturalWidth, nh = pic.naturalHeight;
  switch (view.rot) {
    case 1: return [Y, nh - X];
    case 2: return [nw - X, nh - Y];
    case 3: return [nw - Y, X];
    default: return [X, Y];
  }
}

// The same for a rectangle {x, y, w, h}.
function toNaturalRect(r) {
  const [x0, y0] = toNatural(r.x, r.y), [x1, y1] = toNatural(r.x + r.w, r.y + r.h);
  return {x: Math.min(x0, x1), y: Math.min(y0, y1), w: Math.abs(x1 - x0), h: Math.abs(y1 - y0)};
}

let turnTimer = 0;

function rotate(dir) {
  if (!pic.naturalWidth) return;
  stopBounce();
  // A picture that was zoomed or moved keeps its scale, and the pixel in the middle of the window stays in the middle. (One that
  // the mode places is placed by the mode again, fitted to the turned picture.)
  const b = stageSize(), W = dispW(), H = dispH();
  const X = (b.w / 2 - view.x) / view.sx, Y = (b.h / 2 - view.y) / view.sy; // the middle of the window, in the picture as shown
  const mid = dir > 0 ? [H - Y, X] : [Y, W - X]; // and where that point is in the picture after the turn
  view.rot = (view.rot + dir + 4) % 4; view.angle += dir * 90; SetTurn(view.rot);
  view.sel = null; view.selDrag = null; view.hl = null; clearPixel();
  if (view.free) {
    [view.sx, view.sy] = [view.sy, view.sx]; // what was the vertical scale is the horizontal one now
    view.x = b.w / 2 - mid[0] * view.sx; view.y = b.h / 2 - mid[1] * view.sy;
  }
  pic.classList.add('turning'); void pic.offsetWidth; // the style changes of draw() are animated, for a moment
  draw();
  clearTimeout(turnTimer);
  turnTimer = setTimeout(() => pic.classList.remove('turning'), 240);
}

const stageSize = () => ({w: $('stage').clientWidth, h: $('stage').clientHeight});
const clamp = (v, lo, hi) => Math.min(hi, Math.max(lo, v));

// The scale of the picture, for each axis, that the mode asks for.
function modeScale() {
  const b = stageSize(), w = dispW(), h = dispH(), fit = Math.min(b.w / w, b.h / h);
  switch (view.mode) {
    case 'original': return [1, 1];
    case 'large': return [fit * 0.85, fit * 0.85]; // 85% of the window, up or down
    case 'contain': return [fit, fit];
    default: return [b.w / w, b.h / h];
  }
}

// Keeps a picture that was moved or zoomed on the screen: smaller than the window it is centred, bigger it cannot leave a gap.
function constrain() {
  const b = stageSize(), w = dispW() * view.sx, h = dispH() * view.sy;
  view.x = w <= b.w ? (b.w - w) / 2 : clamp(view.x, b.w - w, 0);
  view.y = h <= b.h ? (b.h - h) / 2 : clamp(view.y, b.h - h, 0);
}

// ---- Bounce at the edge ----------------------------------------------------------------------------------
// The picture can be pulled a little past its edge (at most BOUNCE_MAX pixels, harder the further it goes, the way a list on iOS gives
// at its end) with the mouse button down or two fingers on the trackpad. It is shown at its position plus a "pull" (view.ox, view.oy),
// and a spring takes the pull away again, every frame, as soon as nothing pushes. A picture that fits the window on an axis cannot be
// moved there, but it gives all the same, so that it is plain that there is nowhere to go.
//
// The trackpad keeps sending scroll events for about a second after the fingers have left, each smaller than the one before (the
// system lets the movement coast). Those must not keep the picture out: a pull that is already there is not made bigger by events that
// only get smaller, so it comes back at once.

const BOUNCE_MAX = 20;  // the most the picture can be pulled past its edge, in pixels
const BOUNCE_TAU = 70;  // ms: the spring takes away 63% of the pull in this time
let springRaf = 0, springLast = 0;
const bounce = {peak: 0, lastPush: 0, over: false, lastWheel: 0, swipeOver: false, prev: {x: 0, y: 0}, dec: {x: 0, y: 0}};

// Where the top left corner of the picture may be, per axis: [lowest, highest]. A picture that fits has one place only.
function bounds() {
  const b = stageSize(), w = dispW() * view.sx, h = dispH() * view.sy;
  return {x: w <= b.w ? [(b.w - w) / 2, (b.w - w) / 2] : [b.w - w, 0], y: h <= b.h ? [(b.h - h) / 2, (b.h - h) / 2] : [b.h - h, 0]};
}

// What a pull of d pixels past the edge shows: less than d, and never more than BOUNCE_MAX.
const pullOffset = (d) => BOUNCE_MAX * (1 - Math.exp(-d / (BOUNCE_MAX * 2.5)));

function stopBounce() {
  cancelAnimationFrame(springRaf); springRaf = 0;
  view.ox = view.oy = 0; bounce.over = false; bounce.peak = 0;
}

function startSpring() {
  if (!springRaf) { springLast = performance.now(); springRaf = requestAnimationFrame(springStep); }
}

function springStep(now) {
  springRaf = 0;
  if (!view.drag) { // while the mouse button is down the pointer decides; the spring only works after it is let go
    const k = Math.exp(-clamp(now - springLast, 0, 50) / BOUNCE_TAU); // (the frame's time can be older than an event's, so never negative)
    view.ox *= k; view.oy *= k;
  }
  springLast = now;
  const pull = Math.max(Math.abs(view.ox), Math.abs(view.oy));
  bounce.peak = Math.max(bounce.peak, pull);
  if (!view.drag && pull < 0.4) { // back
    if (bounce.over) Trace(`bounce: peak ${bounce.peak.toFixed(1)} px, back ${Math.round(now - bounce.lastPush)} ms after the last push`);
    view.ox = view.oy = 0; bounce.over = false; bounce.peak = 0;
    draw();
    return;
  }
  draw();
  springRaf = requestAnimationFrame(springStep);
}

function draw() {
  if (!pic.naturalWidth) return;
  if (view.free) {
    constrain();
  } else { // the mode decides, and the picture is centred (a big one at 100% shows its middle)
    [view.sx, view.sy] = modeScale();
    const b = stageSize();
    view.x = (b.w - dispW() * view.sx) / 2; view.y = (b.h - dispH() * view.sy) / 2;
  }
  // The picture gets its size on screen, not a scale: scaling a 36 megapixel picture with a CSS transform made WebKit draw
  // all of it, 5 seconds before anything showed. At its displayed size it is decoded and drawn at that size.
  // A turned picture is the box bw × bh on the screen; the image itself is as wide as that box is high, and is turned around its middle.
  const bw = Math.max(1, Math.round(dispW() * view.sx)), bh = Math.max(1, Math.round(dispH() * view.sy));
  const ew = view.rot % 2 ? bh : bw, eh = view.rot % 2 ? bw : bh;
  pic.style.width = `${ew}px`;
  pic.style.height = `${eh}px`;
  pic.style.transform = `translate(${Math.round(view.x + view.ox) + (bw - ew) / 2}px, ${Math.round(view.y + view.oy) + (bh - eh) / 2}px) rotate(${view.angle}deg)`; // the pull included
  // the name of the picture is in the title of the window; down here is where it lives and which one of the folder it is
  const place = view.count ? `${view.index} / ${view.count}  ·  ` : '';
  const pct = (v) => `${Math.round(v * 100)}%`;
  const zoom = Math.abs(view.sx - view.sy) > 0.005 ? `${pct(view.sx)} × ${pct(view.sy)}` : pct(view.sx);
  const sel = selectionRect();
  // the window title has the size, place and zoom; the bar repeats them only in full screen, where there is no title
  const full = document.body.classList.contains('full');
  const info = `${place}${dispW()} × ${dispH()}  ·  ${zoom}${view.pixelated ? '  ·  pixels' : ''}`;
  setStatus(view.dir, full ? `${info}${sel ? `  ·  selection ${sel.w} × ${sel.h}` : ''}` : sel ? `selection ${sel.w} × ${sel.h}` : '', view.name + (view.modified || floatChanges(view.fl) ? ' ●' : ''));
  document.body.classList.toggle('sel', !!(sel || view.fl)); // in full screen the bar stays for as long as there is a selection
  tellInfo(pic.naturalWidth ? `${dispW()} × ${dispH()}${view.count ? `  ·  ${view.index} / ${view.count}` : ''}  ·  ${zoom}${view.pixelated ? '  ·  pixels' : ''}` : '');
  renderFloat();
  renderInk();
  renderSelection();
  if (view.pixelated && lastPointer) readPixel(lastPointer.x, lastPointer.y); // the picture moved under the pointer
}

// What the status bar used to say about the picture (size, place in the folder, zoom) goes to Go, for the title of the window.
let toldInfo = '';
function tellInfo(text) {
  if (text === toldInfo) return;
  toldInfo = text;
  SetTitleInfo(text).catch(() => {});
}

function setStatus(folder, info = '', name = '') {
  $('status').hidden = !folder && !info;
  $('name').textContent = name; $('name').title = name;
  $('folder').textContent = folder; $('folder').title = folder;
  $('info').textContent = info;
  $('status').classList.remove('bad', 'ok-flash');
}

// The keys 1 to 4: choose a mode, and keep it.
function setMode(mode) {
  stopBounce();
  view.mode = mode; view.free = false;
  draw();
}

function zoomAt(factor, px, py) {
  stopBounce();
  const sx = clamp(view.sx * factor, 0.02, 32), f = sx / view.sx; // the same factor on both axes: a stretched picture stays stretched
  view.x = px - (px - view.x) * f; view.y = py - (py - view.y) * f;
  view.sx = sx; view.sy = clamp(view.sy * f, 0.02, 32);
  view.free = true;
  draw();
}

// 100% around the middle of the window, without changing the mode.
function zoomTo100() {
  const b = stageSize();
  view.free = true; view.sx = view.sy = 1;
  view.x = (b.w - dispW()) / 2; view.y = (b.h - dispH()) / 2;
  draw();
}

// opts.rot: the turn the picture is shown with (undo brings a picture back as it was turned); a new picture is not turned.
// The empty window offers the picture that is on the clipboard (if there is one) next to Open file…
function offerClipboard() {
  ClipboardHasImage().then((has) => { if (!$('empty').hidden) $('open-clip').hidden = !has; }).catch(() => {});
}

function showImage(info, opts = {}) {
  let done; const loaded = new Promise((r) => { done = r; }); // resolved when the picture is on the screen
  view.state = info?.state || 0;
  view.fl = opts.fl || null; view.tdrag = null; if (!opts.keepView && !opts.fl) renderFloat(); // (the same picture with a change applied: what is lifted stays on the screen until the new one is there, so nothing flashes)
  clearTimeout(turnTimer); pic.classList.remove('turning'); view.rot = opts.rot || 0; view.angle = view.rot * 90; SetTurn(view.rot);
  if (!info?.url) { // no picture: the window as it is when it starts
    stopBounce(); pic.onload = null; pic.removeAttribute('src');
    view.name = ''; view.modified = false; view.dir = ''; view.index = 0; view.count = 0; view.token++; view.sel = null; view.selDrag = null; view.hl = null; $('stage').style.cursor = ''; clearPixel();
    pic.hidden = true; $('empty').hidden = false; setStatus(''); renderSelection(); offerClipboard();
    if (editOn()) setEdit(false); // no image: no edit mode (closing it, deleting the last one, undo back to the empty window)
    done();
    return loaded;
  }
  view.name = info.name; view.modified = !!info.modified; view.canAlpha = !!info.canAlpha; view.dir = info.dir; view.index = info.index; view.count = info.count; if (!opts.keepView) view.free = false; // (a picture that is the same, with something applied, keeps its zoom)
  stopBounce();
  view.token++; pixelState = ''; pixelCanvas = null; pixelPromise = null; view.sel = opts.sel ? rectSel(opts.sel) : null; view.selDrag = null; view.hl = null; $('stage').style.cursor = ''; clearPixel();
  pic.onload = () => {
    Trace(`img onload, visibility ${document.visibilityState}, focus ${document.hasFocus()}`);
    pic.hidden = false; $('empty').hidden = true; draw();
    checkTransparency(view.token);
    if (view.pixelated) preparePixels();
    pic.decode().then(() => Trace('decoded'), () => {});
    requestAnimationFrame(() => Trace('first paint after load'));
    done();
  };
  pic.onerror = () => { problem(`Cannot show ${info.name}`); done(); };
  Trace('img src set');
  // The new picture is decoded off screen first, and then it takes the place of the old one in a single step: the old one stays on the screen until
  // then, so nothing is cleared in between (that flashed when the next or the previous picture of a folder was shown).
  const pre = document.createElement('img'), token = view.token; // (not new Image(): that name is the binding of Go's Image)
  pre.src = info.url;
  pre.decode().catch(() => {}).then(() => {
    if (token !== view.token) { done(); return; } // another picture was asked for meanwhile
    if (!opts.keepView) $('stage').classList.remove('checker');
    pic.src = info.url;
  });
  return loaded;
}

function problem(text, bad = true, ms = 4000) {
  $('status').hidden = false;
  $('name').textContent = ''; $('folder').textContent = text; $('folder').title = ''; $('info').textContent = '';
  $('status').classList.toggle('bad', bad);
  clearTimeout(problemTimer);
  document.body.classList.add('peek');
  problemTimer = setTimeout(() => { document.body.classList.remove('peek'); draw(); }, ms);
}

// Left and right arrow: the previous and the next picture of the folder, in alphabetical order.
async function step(delta) {
  try {
    const note = await Step(delta);
    if (note) problem(note, false);
  } catch (e) { problem(errText(e)); }
}

async function choose() {
  try { await ChooseImage(); } catch (e) { problem(errText(e)); }
}

// ---- Pixel readout ---------------------------------------------------------------------------------------
// Tab switches the picture between smoothed and pixelated (every pixel a sharp square, nothing blended), and in the pixelated view
// the bottom bar shows the color of the pixel under the pointer. The numbers come from Go, which decodes the file a second time
// (the web view converts colors for the screen, so what it shows is not the number in the file); for types Go cannot decode they are
// read from what the web view drew, and the bar says so. The pixelated view stays when the next picture is shown, but not the next
// time mcView starts.

let pixelState = '';   // '' not asked, 'preparing', 'go' or 'browser' (where the numbers come from), 'none' (cannot be read)
let pixelCanvas = null;
let lastPointer = null;
let pixelPending = null, pixelBusy = false;

// Edit mode (E): the edit bar, the status bar in full screen, [edit] in the title and the status bar, and the pixelated view (leaving it smooths the picture again).
// The pixelated view follows it, so it is not a step of the history of its own: the steps always show the view the mode has.
const editOn = () => document.body.classList.contains('edit');
function setEdit(on) {
  if (on && !pic.naturalWidth) return; // nothing to edit
  document.body.classList.toggle('edit', on);
  if (!on) closeDock();
  if (on) setTool('select');
  syncPenDock();
  $('editmark').hidden = !on;
  SetEditMode(on).catch(() => {});
  if (pic.naturalWidth && view.pixelated !== on) setPixelated(on); else view.pixelated = on;
  updateHover(false);
}

// ---- Tools of the edit bar: select (the default when edit mode starts) and the pencil ----------------------------------
// With edit mode off the selection is the only thing there is. The pencil draws one pixel wide strokes with the left button in the first color,
// with the right button in the second; a stroke is shown on #ink while the button is down and then drawn into the picture by Go (a step of the history).
const tool = () => (editOn() ? view.tool : 'select');

function setTool(t) {
  if (t !== 'select' && view.tool !== t) clearSelection(); // drawing has no selection
  view.tool = t;
  syncPenDock();
  for (const k of ['select', 'pen']) $(`tool-${k}`).classList.toggle('on', k === t);
  updateHover(false);
}

let ink = null; // the stroke that is being drawn: {pts: [x, y, x, y, …], color}
let fin = null; // the stroke that is let go and is being drawn into the picture by Go: only its overlay is still shown

function inkPoint(ev) {
  const [px, py] = pointInPicture(ev.clientX, ev.clientY);
  return [clamp(Math.floor(px), 0, dispW() - 1), clamp(Math.floor(py), 0, dispH() - 1)];
}

function renderInk() {
  const cv = $('ink');
  if (!ink && !fin) { cv.hidden = true; return; }
  cv.style.left = `${view.x + view.ox}px`; cv.style.top = `${view.y + view.oy}px`;
  cv.style.width = `${dispW() * view.sx}px`; cv.style.height = `${dispH() * view.sy}px`;
  cv.hidden = false;
}

// The round tip of the pencil: the pixels around the point that are inside a circle of the width (Go draws the same, see DrawStroke).
function stampOffsets(width) {
  const lim = (width / 2 - 0.25) ** 2, mid = (width - 1) / 2, c = Math.floor((width - 1) / 2), out = [];
  for (let j = 0; j < width; j++) for (let i = 0; i < width; i++) if ((i - mid) ** 2 + (j - mid) ** 2 <= lim) out.push([i - c, j - c]);
  return out;
}

function startInk(ev) {
  const cv = $('ink');
  cv.width = dispW(); cv.height = dispH();
  ink = {pts: [], width: paint.width, color: ev.button === 2 ? paint.bg : paint.fg, ctx: cv.getContext('2d')};
  ink.tip = stampOffsets(ink.width);
  ink.ctx.fillStyle = ink.color;
  addInk(inkPoint(ev));
  renderInk();
}

function addInk([x, y]) {
  const n = ink.pts.length;
  if (n && ink.pts[n - 2] === x && ink.pts[n - 1] === y) return;
  let [x1, y1] = n ? [ink.pts[n - 2], ink.pts[n - 1]] : [x, y];
  const dx = Math.abs(x - x1), dy = -Math.abs(y - y1), sx = x1 < x ? 1 : -1, sy = y1 < y ? 1 : -1;
  let e = dx + dy;
  for (;;) { // the same line Go draws (Bresenham)
    for (const [i, j] of ink.tip) ink.ctx.fillRect(x1 + i, y1 + j, 1, 1);
    if (x1 === x && y1 === y) break;
    const e2 = 2 * e;
    if (e2 >= dy) { e += dy; x1 += sx; }
    if (e2 <= dx) { e += dx; y1 += sy; }
  }
  ink.pts.push(x, y);
}

async function finishInk(k) {
  fin = k; // the overlay stays until the picture has the stroke, but the stroke itself is over
  try {
    const info = await DrawStroke(view.rot, k.pts, k.color, k.width);
    await showImage(info, {keepView: true, rot: 0});
    commit();
  } catch (e) { problem(errText(e)); }
  if (fin === k) fin = null; // (a stroke that started meanwhile has the overlay now)
  renderInk();
}

// ---- Colors: the two wells at the bottom of the edit bar, and the palette that slides up from the status bar ----------
// A click on a well opens the palette for that color; a click anywhere else closes it, and so does the pointer leaving it once a color was changed in it.
const PALETTE = ['#000000', '#7F7F7F', '#880015', '#ED1C24', '#FF7F27', '#FFF200', '#22B14C', '#00A2E8', '#3F48CC', '#A349A4',
  '#FFFFFF', '#C3C3C3', '#B97A57', '#FFAEC9', '#FFC90E', '#EFE4B0', '#B5E61D', '#99D9EA', '#7092BE', '#C8BFE7'];
const paint = {fg: '#ED1C24', bg: '#FFFFFF', target: 'fg', changed: false, native: false, width: 5};

// The ribbon with the width of the pencil: there while the pencil is the tool (and the palette is not).
function syncPenDock() {
  const open = editOn() && view.tool === 'pen' && !$('dock').classList.contains('open');
  if (open) $('pendock').style.bottom = `${$('status').hidden ? 0 : $('status').offsetHeight}px`;
  $('pendock').classList.toggle('open', open);
}

function showPenWidth() {
  $('pen-wv').textContent = `${paint.width} px`;
  const d = Math.min(paint.width, 28);
  $('pen-dot').style.width = $('pen-dot').style.height = `${d}px`;
  $('pen-dot').style.background = paint.fg;
}

function showColors() {
  showPenWidth();
  $('well-fg').firstElementChild.style.background = paint.fg;
  $('well-bg').firstElementChild.style.background = paint.bg;
  for (const c of $('palette').children) c.classList.toggle('sel', c.dataset.color === paint[paint.target]);
}

function openDock(target) {
  paint.target = target; paint.changed = false;
  $('dock').style.bottom = `${$('status').hidden ? 0 : $('status').offsetHeight}px`;
  $('dock-native').value = paint[target].toLowerCase();
  $('dock').classList.add('open');
  syncPenDock();
  showColors();
}

function closeDock() {
  if (paint.native) return; // the system picker is open: it closes the palette when it is done
  $('dock').classList.remove('open');
  syncPenDock();
}

function setColor(c) {
  paint[paint.target] = c.toUpperCase(); paint.changed = true;
  showColors();
}

function initDock() {
  for (const c of PALETTE) {
    const b = document.createElement('button'); b.className = 'cell'; b.dataset.color = c; b.title = c; b.style.background = c;
    b.onclick = () => setColor(c);
    $('palette').append(b);
  }
  $('pen-w').addEventListener('input', () => { paint.width = +$('pen-w').value; showPenWidth(); });
  showPenWidth();
  for (const t of ['fg', 'bg']) {
    $(`well-${t}`).onclick = () => { if ($('dock').classList.contains('open') && paint.target === t) closeDock(); else openDock(t); };
  }
  $('dock-more').onclick = () => { paint.native = true; $('dock-native').click(); };
  $('dock-native').addEventListener('input', (ev) => setColor(ev.target.value));
  $('dock-native').addEventListener('change', () => { paint.native = false; });
  window.addEventListener('focus', () => { paint.native = false; }); // back from the system picker, however it was closed
  $('dock').addEventListener('mouseleave', () => { if (paint.changed) closeDock(); });
  window.addEventListener('pointerdown', (ev) => { if (!ev.target.closest('#dock, #wells, #pendock')) closeDock(); }, true);
  showColors();
}

function setPixelated(on) {
  view.pixelated = on;
  pic.classList.toggle('pixelated', on);
  clearPixel();
  if (on && pic.naturalWidth) preparePixels();
  draw();
}

let pixelPromise = null;

// Decodes the picture for the readout and for copying (once per picture); everyone who needs it waits for the same answer.
function preparePixels() {
  if (!pixelPromise) pixelPromise = doPreparePixels();
  return pixelPromise;
}

async function doPreparePixels() {
  const token = view.token;
  pixelState = 'preparing';
  let source = '';
  try { source = await PreparePixels(); } catch (e) { problem(errText(e), false); }
  if (token !== view.token) return; // another picture meanwhile
  if (source === 'browser') {
    try {
      const w = pic.naturalWidth, h = pic.naturalHeight;
      if (w * h > 100e6) throw new Error('too big');
      const c = document.createElement('canvas'); c.width = w; c.height = h;
      pixelCanvas = c.getContext('2d', {willReadFrequently: true, colorSpace: 'srgb'});
      pixelCanvas.drawImage(pic, 0, 0);
    } catch { pixelCanvas = null; source = ''; }
  }
  pixelState = source || 'none';
  if (lastPointer) readPixel(lastPointer.x, lastPointer.y);
}

function clearPixel() { pixelPending = null; $('pixel').hidden = true; }

const hex2 = (v) => v.toString(16).toUpperCase().padStart(2, '0');

function showPixel(p) {
  if (!view.pixelated || !p?.ready) { if (pixelState !== 'preparing') clearPixel(); return; }
  // x 12  y 34 | R ▬ 255  G ▬ 128  B ▬ 0 | (#rrggbb, swatch); a pixel that is not opaque has its alpha as a fourth meter and at the end of the hex code
  const code = `#${hex2(p.r)}${hex2(p.g)}${hex2(p.b)}${p.a < 255 ? hex2(p.a) : ''}`;
  const el = (tag, cls, text) => { const e = document.createElement(tag); if (cls) e.className = cls; if (text !== undefined) e.textContent = text; return e; };
  const pos = el('span', 'pos');
  for (const [k, v] of [['x', p.x], ['y', p.y]]) { const s = el('span'); s.append(el('i', '', k), String(v)); pos.append(s); }
  const chans = el('span', 'chans');
  const channels = [['R', p.r, 'r'], ['G', p.g, 'g'], ['B', p.b, 'b']];
  if (p.a < 255) channels.push(['A', p.a, 'a']);
  for (const [name, v, cls] of channels) {
    const meter = el('span', 'm'), fill = el('s'); fill.style.width = `${v / 255 * 100}%`; meter.append(fill);
    const ch = el('span', `ch ${cls}`); ch.append(el('em', '', name), meter, el('span', 'n', String(v))); chans.append(ch);
  }
  const sw = el('span', 'sw');
  const fill = el('b'); fill.style.background = `rgba(${p.r},${p.g},${p.b},${p.a / 255})`; sw.append(fill);
  const chip = el('span', 'chip'); chip.append(el('span', '', code), sw);
  const items = [pos, el('span', 'sep'), chans, el('span', 'sep'), chip];
  if (pixelState === 'browser') items.push(el('span', 'from', 'from the screen'));
  $('pixel').replaceChildren(...items);
  $('pixel').hidden = false;
}

// The pixel of the picture under a position of the window, or nothing when the pointer is beside the picture.
function pixelUnder(clientX, clientY) {
  const r = $('stage').getBoundingClientRect();
  const x = Math.floor((clientX - r.left - view.x - view.ox) / view.sx), y = Math.floor((clientY - r.top - view.y - view.oy) / view.sy);
  return x < 0 || y < 0 || x >= dispW() || y >= dispH() ? null : [x, y];
}

// One question at a time: while Go answers, only the newest position is kept.
async function readPixel(clientX, clientY) {
  if (!view.pixelated || !pic.naturalWidth) return;
  const at = pixelUnder(clientX, clientY);
  if (!at) { clearPixel(); return; }
  if (pixelState === 'preparing') {
    $('pixel').replaceChildren(document.createTextNode('reading the pixels…')); $('pixel').hidden = false; return;
  }
  pixelPending = at;
  if (pixelBusy) return;
  pixelBusy = true;
  const token = view.token;
  while (pixelPending) {
    const [X, Y] = pixelPending; pixelPending = null;
    const [fx, fy] = toNatural(X + 0.5, Y + 0.5), x = Math.floor(fx), y = Math.floor(fy); // the pixel in the file; the readout says where it is on the screen
    let p = null;
    try {
      if (pixelState === 'go') p = await PixelAt(x, y, pic.naturalWidth, pic.naturalHeight);
      else if (pixelState === 'browser' && pixelCanvas) {
        const d = pixelCanvas.getImageData(x, y, 1, 1).data;
        p = {ready: true, x, y, r: d[0], g: d[1], b: d[2], a: d[3]};
      }
      if (p) { p.x = X; p.y = Y; }
    } catch {}
    if (token === view.token) showPixel(p);
  }
  pixelBusy = false;
}

// ---- Selection and copy ----------------------------------------------------------------------------------
// Shift and drag selects a part of the picture; Ctrl+C (Cmd+C on a Mac, Ctrl+Insert) puts just that part on the clipboard, or the whole
// picture when nothing is selected. The part is cut out of the decoded file (see SelectionPNG), so it has the colors of the file.
// In the pixelated view the selection snaps to whole pixels; in the smoothed view it follows the pointer, and what is copied is the
// pixels it touches. The selection is kept in picture pixels, so it stays on its place when the picture is zoomed or moved.

// view.sel is the frame, {x0, y0, x1, y1} in picture pixels, and its end is the edge of the frame, not a pixel: 3 to 6 is the pixels 3, 4 and 5.
// With the mouse in the pixelated view it is made of the pixel under each end, both included (see dragFrame), so it is whole pixels already.

// The selection as whole pixels of the picture: {x, y, w, h}, or null. In the smoothed view a frame that is not on whole pixels
// covers every pixel it touches.
function selectionRect() {
  const s = view.sel, nw = dispW(), nh = dispH();
  if (!s || !nw) return null;
  let x0 = Math.floor(Math.min(s.x0, s.x1)), x1 = Math.ceil(Math.max(s.x0, s.x1));
  let y0 = Math.floor(Math.min(s.y0, s.y1)), y1 = Math.ceil(Math.max(s.y0, s.y1));
  x0 = clamp(x0, 0, nw); x1 = clamp(x1, 0, nw); y0 = clamp(y0, 0, nh); y1 = clamp(y1, 0, nh);
  return x1 > x0 && y1 > y0 ? {x: x0, y: y0, w: x1 - x0, h: y1 - y0} : null;
}

// The frame that a drag from (ax, ay) to (bx, by), in picture pixels, makes: exactly where the pointer went when smoothed, and the pixel under each
// end (both included) when pixelated.
function dragFrame(ax, ay, bx, by) {
  const x0 = Math.min(ax, bx), x1 = Math.max(ax, bx), y0 = Math.min(ay, by), y1 = Math.max(ay, by);
  if (!view.pixelated) return {x0, y0, x1, y1};
  return {x0: Math.floor(x0), y0: Math.floor(y0), x1: Math.floor(x1) + 1, y1: Math.floor(y1) + 1};
}

// The frame on the screen: on whole pixels when pixelated, exactly where the pointer went when smoothed.
// The frame in picture pixels: {x0, x1, y0, y1}, or null.
function selectionBox() {
  const s = view.sel;
  if (!s || !pic.naturalWidth) return null;
  if (view.pixelated) {
    const r = selectionRect();
    return r && {x0: r.x, x1: r.x + r.w, y0: r.y, y1: r.y + r.h};
  }
  return {x0: Math.min(s.x0, s.x1), x1: Math.max(s.x0, s.x1), y0: Math.min(s.y0, s.y1), y1: Math.max(s.y0, s.y1)};
}

// What was put on the clipboard last, as the frame it was (null: the whole picture, or something that is not here any more). The frame
// of that part is drawn with marching ants for as long as it is the frame that is there; clearing the selection does not touch the clipboard.
let copiedKey = null;
const boxKey = (b) => `${view.state}:${view.rot}:${b.x0},${b.x1},${b.y0},${b.y1}`;

function renderSelection() {
  const el = $('sel'), box = selectionBox();
  if (!box) { el.hidden = true; return; }
  el.classList.toggle('drawing', !!(view.selDrag || view.tdrag)); // while the button is down the dashes stand still; set, they move
  el.classList.toggle('copied', copiedKey === boxKey(box));
  el.classList.toggle('content', view.hlMode === 'content');
  const dr = selectionRect(); el.querySelector('.dims').textContent = dr ? `${dr.w} × ${dr.h}` : '';
  const {x0, x1, y0, y1} = box;
  for (const e of ['t', 'b', 'l', 'r']) el.querySelector(`.edge.${e}`).classList.toggle('on', !!view.hl?.[e]);
  el.style.left = `${view.x + view.ox + x0 * view.sx}px`; el.style.top = `${view.y + view.oy + y0 * view.sy}px`;
  const w = Math.max(3, Math.round((x1 - x0) * view.sx)), h = Math.max(3, Math.round((y1 - y0) * view.sy));
  el.style.width = `${w}px`; el.style.height = `${h}px`;
  const svg = el.querySelector('.ants'); // the dashes: a rectangle of numbers, so that nothing of an earlier size is left on the screen
  svg.setAttribute('width', w); svg.setAttribute('height', h);
  for (const r of svg.querySelectorAll('rect')) { r.setAttribute('x', 1); r.setAttribute('y', 1); r.setAttribute('width', w - 2); r.setAttribute('height', h - 2); }
  el.hidden = false;
}

// Shift held over an edge (or a corner) of the frame: the edge is lit, the pointer says "resize", and Shift and drag moves the edge.
const EDGE_GRAB = 6; // screen pixels

// Which edges of the frame are within reach of a position of the window: {t, b, l, r} (corners have two), or null.
function edgeAt(clientX, clientY, inside = false) {
  const box = selectionBox();
  if (!box) return null;
  const r = $('stage').getBoundingClientRect(), ox = r.left + view.x + view.ox, oy = r.top + view.y + view.oy;
  const l = ox + box.x0 * view.sx, rt = ox + box.x1 * view.sx, t = oy + box.y0 * view.sy, bt = oy + box.y1 * view.sy;
  const near = (v, a, b) => Math.abs(v - a) <= EDGE_GRAB && Math.abs(v - a) <= Math.abs(v - b); // the closer of two edges that are near
  const inX = clientX >= l - EDGE_GRAB && clientX <= rt + EDGE_GRAB, inY = clientY >= t - EDGE_GRAB && clientY <= bt + EDGE_GRAB;
  const hit = {
    l: inY && near(clientX, l, rt), r: inY && near(clientX, rt, l) && !(Math.abs(clientX - l) === Math.abs(clientX - rt)),
    t: inX && near(clientY, t, bt), b: inX && near(clientY, bt, t) && !(Math.abs(clientY - t) === Math.abs(clientY - bt)),
  };
  if (hit.l || hit.r || hit.t || hit.b) return hit;
  // inside the frame, away from the edges: where Paint moves the part (without Shift)
  return inside && clientX > l && clientX < rt && clientY > t && clientY < bt ? {move: true} : null;
}

const RESIZE_CURSORS = {t: 'ns-resize', b: 'ns-resize', l: 'ew-resize', r: 'ew-resize', tl: 'nwse-resize', br: 'nwse-resize', tr: 'nesw-resize', bl: 'nesw-resize'};

// Lights the edge under the pointer, in two ways: with Shift ("frame": the frame is moved, the picture is not touched) and without
// ("content", like Paint: the selected part is stretched by its edges and moved by its top left corner; see below).
function updateHover(shift) {
  if (view.selDrag?.edge || view.tdrag || view.drag) return;
  if (tool() === 'pen') { // the pencil cursor is for the picture; beside it the pointer is an ordinary one (while a stroke is drawn it stays)
    view.hl = null; view.hlMode = '';
    $('stage').style.cursor = ink || (lastPointer && pixelUnder(lastPointer.x, lastPointer.y)) ? 'crosshair' : '';
    return;
  }
  const editOn = document.body.classList.contains('edit');
  const hit = lastPointer && !view.drag ? edgeAt(lastPointer.x, lastPointer.y, true) : null;
  const was = JSON.stringify([view.hl, view.hlMode]);
  view.hl = hit;
  view.hlMode = hit ? (editOn ? 'content' : 'frame') : '';
  const key = hit ? (hit.t ? 't' : hit.b ? 'b' : '') + (hit.l ? 'l' : hit.r ? 'r' : '') : '';
  $('stage').style.cursor = hit?.move || (view.hlMode === 'content' && key === 'tl') ? 'move' : key ? RESIZE_CURSORS[key] : '';
  if (was !== JSON.stringify([view.hl, view.hlMode])) renderSelection();
}

function clearSelection() {
  stopSelectionFlash();
  if (view.fl) { flushFloat({clear: true}); return; } // a lifted part is applied: the selection is let go
  if (!view.sel) return;
  view.sel = null; view.selDrag = null; view.hl = null; view.hlMode = ''; $('stage').style.cursor = '';
  draw();
}

// ---- Move and stretch the selected part, like Paint ----------------------------------------------------------
// Without Shift, the edges of the selection stretch the part, and its inside (and its top left corner) moves it. The first time, the part is lifted: the
// place it came from is white, and the part is shown where it is dropped (#float, over the picture, which is not changed yet; #hole is the
// white). It is worked out from the part as it was lifted, never from the result of the last drag, so stretching and going back loses
// nothing. It stays like that, and can be dragged again, until the selection is let go (Esc, a click, a new selection, a picture that
// changes): then it is applied and the picture is modified. Every drop is a step of the undo history, and so is the applying; undoing the
// applying brings the lifted part back, selected, and undoing the first drag selects the frame it started from.
// view.fl = {src: {x, y, w, h} where it came from, dest: {x, y, w, h} where it is, url: its pixels as a PNG}, in the picture as shown.

const rectSel = (r) => ({x0: r.x, y0: r.y, x1: r.x + r.w, y1: r.y + r.h});
const sameRect = (a, b) => a.x === b.x && a.y === b.y && a.w === b.w && a.h === b.h;
const floatChanges = (f) => !!f && (f.pasted || !sameRect(f.src, f.dest)); // a pasted layer is always a change; a lifted part is one once it is moved

function renderFloat() {
  const f = view.fl, el = $('float'), hole = $('hole');
  if (!f || !pic.naturalWidth) { el.hidden = true; hole.hidden = true; return; }
  const ox = view.x + view.ox, oy = view.y + view.oy;
  const place = (e, r) => { e.style.left = `${ox + r.x * view.sx}px`; e.style.top = `${oy + r.y * view.sy}px`; e.style.width = `${r.w * view.sx}px`; e.style.height = `${r.h * view.sy}px`; };
  place(hole, f.src); hole.hidden = !!f.pasted; // (a pasted picture leaves no hole)
  // what it leaves is white, or transparent where the type can be: the squares of the window, in step with them
  hole.style.background = $('stage').classList.contains('checker') ? `${CHECKER} ${-parseFloat(hole.style.left)}px ${-parseFloat(hole.style.top)}px / 16px 16px` : '#fff';
  el.classList.toggle('pixelated', view.pixelated);
  place(el, f.dest); el.hidden = !f.url;
}

// The lifted pixels, for #float.
async function loadFloatPixels(f) {
  const b64 = await cutPng(f.src);
  if (view.fl !== f) return; // let go or changed meanwhile
  f.url = `data:image/png;base64,${b64}`;
  for (const en of history.list) if (en.fl && !en.fl.url && en.state === f.state && en.rot === f.rot && sameRect(en.fl.src, f.src)) en.fl.url = f.url; // the history keeps the pixels, so that undo can show them at once
  $('float').onload = () => renderFloat();
  $('float').src = f.url;
}

// A picture with transparent parts gets the squares as the background of the window (worked out after it is shown: Go decodes the file).
async function checkTransparency(token) {
  let has = false;
  try { has = await HasTransparency(); } catch {}
  if (token === view.token) $('stage').classList.toggle('checker', has);
}

function dropFloat() {
  view.fl = null; view.tdrag = null;
  renderFloat();
  ClearFloat().catch(() => {});
}

// Tells Go where the part is, so that it can apply it whenever the picture is saved, pasted into or closed.
function sendFloat() {
  const f = view.fl;
  if (f.pasted) return MovePasted(f.dest.x, f.dest.y, f.dest.w, f.dest.h, !view.pixelated);
  return SetFloat(view.rot, f.src.x, f.src.y, f.src.w, f.src.h, f.dest.x, f.dest.y, f.dest.w, f.dest.h, !view.pixelated);
}

let floatBusy = false;

// Applies the lifted part: the picture becomes the one with it dropped (a new step); opts.clear lets the selection go too.
async function flushFloat(opts = {}) {
  const f = view.fl;
  if (!f || floatBusy) return;
  floatBusy = true;
  try {
    const dest = f.dest;
    if (floatChanges(f)) {
      await sendFloat(); // (the pixelated view may have been switched since: stretched in whole pixels or smoothed)
      const info = await BakeFloat();
      await showImage(info, {keepView: true, rot: 0});
      if (opts.clear) { // the drops and the applying are one step: undoing it selects the frame again, redoing it moves the part and keeps it selected
        history.list.length = history.at + 1;
        while (history.at > 0 && history.list[history.at].fl) { history.list.pop(); history.at--; }
      }
      commit();
      if (opts.clear) { const e = history.list[history.at]; if (e) e.sel = {...dest}; }
      if (!opts.clear) view.sel = rectSel(dest);
    } else {
      dropFloat();
      if (opts.clear) view.sel = null;
    }
    draw();
  } catch (e) { problem(errText(e)); }
  floatBusy = false;
}

function startTransform(ev, edge) {
  const r = selectionRect();
  if (!r) return;
  const first = !view.fl;
  if (first) { view.fl = {src: {...r}, dest: {...r}, url: '', state: view.state, rot: view.rot}; loadFloatPixels(view.fl).catch((e) => problem(errText(e))); }
  const [px, py] = pointInPicture(ev.clientX, ev.clientY);
  view.tdrag = {edge, px, py, start: {...view.fl.dest}, first};
  view.sel = rectSel(view.fl.dest); view.hl = edge; view.hlMode = 'content';
  draw();
}

function dragTransform(ev) {
  const t = view.tdrag, s = t.start, W = dispW(), H = dispH();
  const [px, py] = pointInPicture(ev.clientX, ev.clientY);
  const dx = Math.round(px - t.px), dy = Math.round(py - t.py);
  let x0 = s.x, y0 = s.y, x1 = s.x + s.w, y1 = s.y + s.h;
  if (t.edge.move || (t.edge.t && t.edge.l)) { // the inside and the top left corner move it
    x0 = clamp(s.x + dx, 0, Math.max(0, W - s.w)); y0 = clamp(s.y + dy, 0, Math.max(0, H - s.h)); x1 = x0 + s.w; y1 = y0 + s.h;
  } else { // the others stretch it (the other edges stay where they are)
    if (t.edge.l) x0 = clamp(s.x + dx, 0, x1 - 1);
    if (t.edge.r) x1 = clamp(s.x + s.w + dx, x0 + 1, W);
    if (t.edge.t) y0 = clamp(s.y + dy, 0, y1 - 1);
    if (t.edge.b) y1 = clamp(s.y + s.h + dy, y0 + 1, H);
  }
  view.fl.dest = {x: x0, y: y0, w: x1 - x0, h: y1 - y0};
  view.sel = rectSel(view.fl.dest);
  draw();
}

async function finishTransform() {
  const t = view.tdrag, f = view.fl;
  view.tdrag = null;
  if (!f) return;
  if (sameRect(f.dest, t.start)) { // not moved
    if (t.first) dropFloat();
    draw();
    return;
  }
  try { await sendFloat(); } catch (e) { problem(errText(e)); }
  if (t.first) { const prev = history.list[history.at]; if (prev) prev.undoSel = {...f.src}; } // undo (not redo) selects the frame it started from again
  commit();
}

// A position of the window as a position in the picture (in pixels, fractions included), kept inside the picture.
function pointInPicture(clientX, clientY) {
  const r = $('stage').getBoundingClientRect();
  return [clamp((clientX - r.left - view.x - view.ox) / view.sx, 0, dispW()), clamp((clientY - r.top - view.y - view.oy) / view.sy, 0, dispH())];
}

function base64ToBlob(b64, type) {
  const bin = atob(b64), bytes = new Uint8Array(bin.length);
  for (let i = 0; i < bin.length; i++) bytes[i] = bin.charCodeAt(i);
  return new Blob([bytes], {type});
}

const blobToBase64 = (blob) => new Promise((resolve, reject) => {
  const fr = new FileReader();
  fr.onload = () => resolve(String(fr.result).split(',')[1]);
  fr.onerror = () => reject(fr.error);
  fr.readAsDataURL(blob);
});

// The part r {x, y, w, h} of the picture as it is shown, as a PNG in base64 (cut out of the decoded file; see SelectionPNG).
async function cutPng(r) {
  const nw = pic.naturalWidth, nh = pic.naturalHeight;
  const n = toNaturalRect(r); // the same part in the file's picture, which is cut out and then turned the way it is shown
  await preparePixels();
  let b64 = '';
  if (pixelState === 'go') b64 = await SelectionPNG(n.x, n.y, n.w, n.h, nw, nh);
  if (!b64) { // a type Go cannot decode: cut it out of what the web view drew
    if (!pixelCanvas) throw new Error('this picture cannot be copied');
    const c = document.createElement('canvas'); c.width = n.w; c.height = n.h;
    c.getContext('2d').putImageData(pixelCanvas.getImageData(n.x, n.y, n.w, n.h), 0, 0);
    b64 = await blobToBase64(await new Promise((res) => c.toBlob(res, 'image/png')));
  }
  return view.rot ? RotatePNG(b64, view.rot) : b64;
}

let lastCopy = 0;

// A copy lights things up for a moment: the border of the selection (orange and white, with a clipboard in the middle), and the message in the status bar.
let flashTimer = 0, statusFlashTimer = 0;
function flashSelection() {
  const el = $('sel');
  el.classList.remove('flash'); void el.offsetWidth; el.classList.add('flash'); // (taken off and put on again, so that a second copy starts it again)
  clearTimeout(flashTimer); flashTimer = setTimeout(() => el.classList.remove('flash'), 1700);
}
// A selection that is let go, or a new one begun, ends the flash of the old one: #sel is the same element, and would show it again.
function stopSelectionFlash() {
  clearTimeout(flashTimer); $('sel').classList.remove('flash');
}
function flashStatus() {
  const el = $('status');
  el.classList.remove('ok-flash'); void el.offsetWidth; el.classList.add('ok-flash');
  clearTimeout(statusFlashTimer); statusFlashTimer = setTimeout(() => el.classList.remove('ok-flash'), 1700);
}

async function copySelection(auto = false) {
  if (!pic.naturalWidth) return;
  const now = performance.now();
  if (now - lastCopy < 400) return; // the key and the system's copy command can both come for one press
  lastCopy = now;
  if (view.fl) await flushFloat(); // what was lifted is applied first, so that the copy is what is seen
  const r = selectionRect() || {x: 0, y: 0, w: dispW(), h: dispH()};
  try {
    const b64 = await cutPng(r);
    try { // the web view's own clipboard, and the system's way when it does not take the picture
      await navigator.clipboard.write([new ClipboardItem({'image/png': base64ToBlob(b64, 'image/png')})]);
    } catch { await CopyImageNative(b64); }
    copiedKey = view.sel && selectionBox() ? boxKey(selectionBox()) : null;
    renderSelection();
    if (auto) { problem(`Selection copied to the clipboard  ·  ${r.w} × ${r.h}`, false, 1700); flashStatus(); } // (a finished selection is copied by itself: a short note)
    if (view.sel) flashSelection();
    else problem(`Copied ${r.w} × ${r.h} pixels${view.sel ? ' (the selection)' : ''} to the clipboard`, false);
  } catch (e) { problem(errText(e)); }
}

// On a Mac the wheel events of the trackpad (two fingers up and down) move the picture, and those of a mouse wheel zoom it; a pinch
// (which the web view reports as a wheel event with Ctrl) zooms. The two are told apart by their size: a mouse wheel moves in whole
// lines, which come as multiples of 120 in wheelDelta, and the trackpad in small steps of any size. A trackpad step that happens to be
// a multiple of 120 stays a move while the trackpad is still going. Other systems: the wheel zooms, as before.
const MAC = navigator.platform.startsWith('Mac');
let lastTrackpad = 0;

function isWheelNotch(ev) {
  return ev.deltaMode !== 0 || ((ev.wheelDeltaY % 120 === 0) && (ev.wheelDeltaX % 120 === 0) && (ev.wheelDeltaY !== 0 || ev.wheelDeltaX !== 0));
}

function panBy(dx, dy) {
  const bd = bounds(), now = performance.now();
  if (now - bounce.lastWheel > 120) { // a new swipe
    bounce.prev.x = bounce.prev.y = 0; bounce.dec.x = bounce.dec.y = 0; bounce.swipeOver = false;
  }
  bounce.lastWheel = now;
  let moved = false;
  for (const ax of ['x', 'y']) {
    const o = ax === 'x' ? 'ox' : 'oy', [lo, hi] = bd[ax];
    let d = -(ax === 'x' ? dx : dy); // how far this event wants the picture to go
    const mag = Math.abs(d);
    bounce.dec[ax] = mag <= bounce.prev[ax] ? bounce.dec[ax] + 1 : 0; // how many events in a row were no bigger than the one before
    bounce.prev[ax] = mag;
    if (view[o] !== 0 && d !== 0 && Math.sign(d) === -Math.sign(view[o])) { // going back first takes the pull away
      const use = Math.min(mag, Math.abs(view[o]));
      view[o] += Math.sign(d) * use; d -= Math.sign(d) * use;
    }
    const target = view[ax] + d, at = clamp(target, lo, hi), excess = target - at; // what goes past the edge is a push
    if (at !== view[ax]) { view[ax] = at; moved = true; }
    // the trackpad running out (events that only get smaller): once this swipe has pulled the picture out they do not push it again,
    // not even when the spring has taken the pull away already; coasting that first reaches the edge on its own does bounce, once
    const coasting = bounce.dec[ax] >= 2 && (view[o] !== 0 || bounce.swipeOver);
    if (Math.abs(excess) >= 0.5 && !coasting) {
      view[o] = clamp(view[o] + Math.sign(excess) * Math.abs(excess) * (1 - Math.abs(view[o]) / BOUNCE_MAX) ** 2, -BOUNCE_MAX, BOUNCE_MAX);
      bounce.over = true; bounce.swipeOver = true; bounce.lastPush = now;
    }
  }
  if (moved) view.free = true;
  draw();
  startSpring();
}

function onWheel(ev) {
  ev.preventDefault();
  if (!pic.naturalWidth) return;
  const r = $('stage').getBoundingClientRect(), now = performance.now();
  const zoom = () => zoomAt(Math.exp(-ev.deltaY * (ev.ctrlKey ? 0.01 : 0.0025)), ev.clientX - r.left, ev.clientY - r.top);
  if (ev.ctrlKey || !MAC) { zoom(); return; }
  if (!isWheelNotch(ev)) lastTrackpad = now;
  if (isWheelNotch(ev) && now - lastTrackpad > 120) zoom(); else panBy(ev.deltaX, ev.deltaY);
}

// With edit mode off, paste always makes the clipboard picture the picture of the window (undo brings the old one back); the rest is for edit mode on.
// Ctrl+V (Cmd+V, Shift+Insert, Edit → Paste): the picture on the clipboard becomes the picture of the window. It has no file yet:
// the title says [clipboard image], and S saves it.
// Paste (Ctrl+V, Cmd+V, Edit → Paste): the picture on the clipboard is laid over the picture, near the top left corner of what is visible, as a part that
// floats and is selected, like a lifted part: it can be dragged and stretched, Delete, Backspace or Esc take it away again, and a click beside it lays
// it down. A part that was lifted and moved is applied first. A clipboard picture that is big (about as big as the picture) is probably a whole new
// image: Go asks whether it becomes the picture of the window or is laid over this one, and then the window shows it by itself. A window without a picture takes it as its picture.
async function pasteImage() {
  try {
    if (view.fl) await flushFloat({clear: false});
    if (!document.body.classList.contains('edit')) { await PasteNew(); return; } // edit mode off: a whole new image (the 'image' event shows it and makes it a step)
    let px = 0, py = 0;
    if (pic.naturalWidth) { const r = $('stage').getBoundingClientRect(); [px, py] = pointInPicture(r.left + 12, r.top + 12); }
    const L = await PasteLayer(view.rot, Math.ceil(px), Math.ceil(py)); // (rounded up: the first pixel that is seen whole, 12 screen pixels in from the corner)
    if (!L.w) return;
    const f = {src: {x: 0, y: 0, w: L.w, h: L.h}, dest: {x: L.x, y: L.y, w: L.w, h: L.h}, url: `data:image/png;base64,${L.png}`, pasted: true, state: view.state, rot: view.rot};
    view.fl = f;
    $('float').onload = () => renderFloat();
    $('float').src = f.url;
    view.sel = rectSel(f.dest);
    draw();
    commit();
  } catch (e) { problem(errText(e)); }
}

// Delete, Backspace and Esc take a pasted picture away again (a step of the history, like laying it).
function removeLayer() {
  dropFloat();
  view.sel = null; view.selDrag = null; view.hl = null; view.hlMode = ''; $('stage').style.cursor = '';
  draw();
  commit();
}

// File → Close Image: the window goes back to the way it starts, with no picture. It is a step of the history like any other, so undo brings
// the picture back with its changes, but a picture that has some is asked about first (Save, Discard, Cancel), as when the window is closed.
async function closeImage() {
  if (!pic.naturalWidth) return;
  try { if (!await DiscardOK('close the image')) return; } catch (e) { problem(errText(e)); return; }
  if (view.fl) await flushFloat({clear: true}); // what is lifted is part of the picture that comes back
  if (!view.state || history.busy) return;
  try {
    const info = await RestoreState(0);
    await showImage(info);
    commit();
  } catch (e) { problem(errText(e)); }
}

// Ctrl+W, Cmd+W (File → Close): the picture is closed; the window, when there is none.
function closeOrQuit() { if (pic.naturalWidth) closeImage(); else CloseWindow(); }

// S (Ctrl+S, Cmd+S, File → Save As…): asks where to save the picture. Saving it to another folder makes that folder the one the
// window works in, and the arrow keys go through its pictures.
// Create Image (File menu, and the empty window): a new white 1280 × 720 image, and edit mode starts with it.
let editAfterImage = false;
async function createImage() {
  try { editAfterImage = true; await NewBlank(1280, 720, false); } catch (e) { editAfterImage = false; problem(errText(e)); }
}

// D: the picture is deleted (the system asks first); the next picture of the folder is shown in its place, with the folder's new count.
async function deleteImage() {
  try { const note = await DeleteImage(); if (note) problem(note, false, 2000); } catch (e) { problem(errText(e)); }
}

async function saveAs() {
  try {
    if (view.fl) await flushFloat(); // what was lifted is part of what is saved
    const r = selectionRect(); // with a selection only that part is saved, as a picture of its own; the window keeps its picture
    const note = r ? await SaveSelectionAs(view.rot, r.x, r.y, r.w, r.h) : await SaveAs(view.rot); if (note) problem(note, false); } catch (e) { problem(errText(e)); }
}

// When the window changes its size (full screen with Enter, and back, or the window dragged bigger), a picture that was zoomed or
// moved keeps the pixel that was in the middle of the window in the middle. The pixel is remembered, because a picture near its edge
// cannot be held in the middle while the window is small (it is pushed back to the edge) but should be again when the window is big
// again; the memory is only good while the picture has not been moved or zoomed since. A picture a mode places is fitted again.
let lastStage = null, anchor = null;

// Full screen has no bar at the bottom; Shift shows it for as long as it is held (and a message for a few seconds).
// On Windows the menu bar is hidden in full screen and comes back while the pointer is at the top edge.
let menuShown = true;
function showMenuBar(on) { if (on !== menuShown) { menuShown = on; SetMenuBar(on).catch(() => {}); } }
async function refreshFullscreen() {
  try {
    const full = await IsFullscreen();
    document.body.classList.toggle('full', full);
    showMenuBar(!full);
    if (pic.naturalWidth) draw(); // the bar says the size and zoom in full screen only
  } catch {}
}
window.addEventListener('mousemove', (ev) => {
  if (!document.body.classList.contains('full')) return;
  document.body.classList.toggle('nearbottom', !editOn() && ev.clientY >= window.innerHeight - 44); // the bar comes while the pointer is at the bottom (edit mode has the bar already)
  if (ev.clientY <= 1) showMenuBar(true); else if (ev.clientY > 40) showMenuBar(false);
});

document.documentElement.addEventListener('mouseleave', () => document.body.classList.remove('nearbottom'));

// The system takes a moment to get in and out of full screen; ask again until it is done.
function settleFullscreen() { for (const ms of [100, 400, 1000, 1800]) setTimeout(refreshFullscreen, ms); }

function onResize() {
  refreshFullscreen();
  const b = stageSize();
  if (view.free && pic.naturalWidth && lastStage && b.w > 0 && b.h > 0) {
    const fresh = !anchor || anchor.x !== view.x || anchor.y !== view.y || anchor.sx !== view.sx || anchor.sy !== view.sy || anchor.rot !== view.rot;
    const a = fresh ? {X: (lastStage.w / 2 - view.x) / view.sx, Y: (lastStage.h / 2 - view.y) / view.sy} : anchor;
    view.x = b.w / 2 - a.X * view.sx; view.y = b.h / 2 - a.Y * view.sy;
    draw(); // keeps it inside the window (changes view.x and view.y)
    anchor = {X: a.X, Y: a.Y, x: view.x, y: view.y, sx: view.sx, sy: view.sy, rot: view.rot};
  } else {
    draw(); anchor = null;
  }
  lastStage = b;
}

// ---- Undo and redo ---------------------------------------------------------------------------------------
// Every change of what the window shows is a step that can be undone with Ctrl+Z (Cmd+Z) and done again with Ctrl+Shift+Z (Cmd+Shift+Z,
// Ctrl+Y): another picture (Open, the arrow keys, paste, save to another folder, a file opened with mcView), the mode (1 to 4, F), the
// pixelated view (Tab) and the turn (R). Not steps: how far it is zoomed, where it is moved, the selection.
// The history is a list of whole states, {picture, mode, pixelated, turn}, and an undo shows the one before. The picture is a number
// that Go knows (ImageInfo.state; 0 is no picture at all), because a picture pasted from the clipboard exists only there.
const HISTORY_MAX = 100;
const history = {list: [], at: -1, busy: false};
let lastUndo = 0;

const snapshot = () => ({state: view.state, mode: view.mode, pixelated: view.pixelated, rot: view.rot, pic: `${view.dir}|${view.name}`, view: {free: view.free, sx: view.sx, sy: view.sy, x: view.x, y: view.y}, fl: view.fl ? {src: {...view.fl.src}, dest: {...view.fl.dest}, url: view.fl.url, pasted: !!view.fl.pasted} : null});
const flKey = (f) => (f ? JSON.stringify([f.src, f.dest]) : '');
const sameSnapshot = (a, b) => a.state === b.state && a.mode === b.mode && a.pixelated === b.pixelated && a.rot === b.rot && flKey(a.fl) === flKey(b.fl);

// Makes the state the window is in now a step of the history (and ends what could be redone).
function commit() {
  if (history.busy) return;
  const s = snapshot();
  if (history.at >= 0 && sameSnapshot(history.list[history.at], s)) return;
  history.list.length = history.at + 1;
  history.list.push(s);
  if (history.list.length > HISTORY_MAX) history.list.shift();
  history.at = history.list.length - 1;
  KeepStates([...new Set(history.list.map((e) => e.state))]).catch(() => {}); // Go keeps no other picture
}

function turnTo(rot) {
  const d = (rot - view.rot + 4) % 4;
  if (d === 1 || d === 3) rotate(d === 1 ? 1 : -1); else if (d === 2) { rotate(1); rotate(1); }
}

// Shows the state at place `to` of the history.
async function travel(to) {
  if (history.busy || to < 0 || to >= history.list.length) return;
  history.busy = true;
  const e = history.list[to];
  const hi = history.list[Math.max(to, history.at)], lo = history.list[Math.min(to, history.at)]; // the step that is undone or done again, and the one next to it
  const hadFloat = !!view.fl;
  if (hadFloat) { view.fl = null; view.tdrag = null; } // (what is shown of it stays on the screen until the picture that follows is there)
  const sel = e.fl ? e.fl.dest : e.sel || (e.undoSel && to < history.at ? e.undoSel : null);
  const fl = e.fl ? {src: {...e.fl.src}, dest: {...e.fl.dest}, url: e.fl.url || '', state: e.state, rot: e.rot, pasted: !!e.fl.pasted} : null;
  try {
    if (e.state !== view.state) { // another picture (or none): Go brings it back, the view is set as it was
      // Go brings the picture back with its lifted part in one call, so that the circle of the title does not go out for a moment
      const info = fl?.pasted ? await RestoreLayer(e.state, e.rot, fl.url.split(',')[1], fl.dest.x, fl.dest.y, fl.dest.w, fl.dest.h, !editOn()) : fl ? await RestoreFloat(e.state, e.rot, fl.src.x, fl.src.y, fl.src.w, fl.src.h, fl.dest.x, fl.dest.y, fl.dest.w, fl.dest.h, !editOn()) : await RestoreState(e.state);
      const same = view.state !== 0 && info?.url && info.name === view.name && info.dir === view.dir && e.mode === view.mode; // the same picture, with a change undone or redone: the zoom stays
      view.mode = e.mode; if (!same) view.free = false; view.pixelated = editOn(); pic.classList.toggle('pixelated', editOn());
      if (fl?.url) { const fe = $('float'); fe.src = fl.url; await fe.decode().catch(() => {}); } // the lifted part is ready before the picture changes
      await showImage(info, {rot: e.rot, keepView: !!same, fl: fl?.url ? fl : null, sel});
    } else {
      if (hadFloat && !e.fl) ClearFloat().catch(() => {});
      if (e.mode !== view.mode) setMode(e.mode);
      if (editOn() !== view.pixelated) setPixelated(editOn());
      turnTo(e.rot);
    }
    if (hi.view && hi.pic === lo.pic && hi.rot === lo.rot && view.state) { // the view (zoom and which part is visible) the step was made in
      Object.assign(view, hi.view); draw();
    }
    if (e.fl) { // the lifted part is back where it was dropped, and selected
      if (!view.fl) { view.fl = fl; if (fl.pasted) { $('float').onload = () => renderFloat(); $('float').src = fl.url; } else loadFloatPixels(view.fl).catch(() => {}); } // (without its pixels in the history they are cut out again)
      view.sel = rectSel(e.fl.dest);
      if (fl.pasted) await SetPasted(e.rot, fl.url.split(',')[1], fl.dest.x, fl.dest.y, fl.dest.w, fl.dest.h, !editOn()); else await sendFloat();
    } else view.sel = sel ? rectSel(sel) : null; // the frame that was selected before it was lifted, or the part that was dropped and let go
    draw();
    history.at = to;
  } catch (err) { problem(errText(err)); }
  history.busy = false;
}

// The key and the menu item can both come for one press: that is one step.
function undoRedo(delta) {
  const now = performance.now();
  if (now - lastUndo < 120) return;
  lastUndo = now;
  travel(history.at + delta);
}
const undo = () => undoRedo(-1), redo = () => undoRedo(1);

// Drag to select (with Shift on an edge of the frame, the frame is moved without touching the picture).
function beginSelection(clientX, clientY) {
  stopSelectionFlash();
  const edge = document.body.classList.contains('edit') ? null : edgeAt(clientX, clientY, true); // edit mode off: the frame alone is moved and resized
  const [qx, qy] = pointInPicture(clientX, clientY);
  if (edge) { // an edge of the frame: it is moved, not a new frame drawn
    const box = selectionBox();
    view.sel = {x0: box.x0, x1: box.x1, y0: box.y0, y1: box.y1}; // the first corner is the left top one, so l and t are x0 and y0
    view.selDrag = {edge, from: [clientX, clientY], p0: [qx, qy], start: {...view.sel}};
    view.hl = edge;
    draw();
    return;
  }
  const [x, y] = pointInPicture(clientX, clientY);
  view.selDrag = {x0: x, y0: y, from: [clientX, clientY]};
  view.sel = null; // the frame appears once the drag covers more than one pixel
  draw();
}

function startViewer() {
  initDock();
  $('tool-select').onclick = () => setTool('select');
  $('tool-pen').onclick = () => setTool('pen');
  $('stage').addEventListener('contextmenu', (ev) => { if (tool() === 'pen') ev.preventDefault(); });
  const cmd = navigator.platform.startsWith('Mac') ? '⌘' : 'Ctrl+';
  $('hint').textContent = `Open a picture with File → Open… (${cmd}O), paste one from the clipboard (${cmd}V), or open one with mcView from your file manager.`;
  $('open').onclick = choose;
  $('open-clip').onclick = () => PasteNew().catch((e) => problem(errText(e)));
  $('new-pic').onclick = createImage;
  $('close-img').onclick = closeImage; // back to the home screen (which ends edit mode)
  EventsOn('menu-create', createImage); // File → Create Image
  window.addEventListener('focus', () => { if (!$('empty').hidden) offerClipboard(); });
  offerClipboard();
  EventsOn('history-reset', () => { history.list.length = 0; history.at = -1; }); // a deleted picture cannot be brought back by undo
  EventsOn('image', (info) => { const edit = editAfterImage; editAfterImage = false; showImage(info).then(() => { if (edit) setEdit(true); }); commit(); });
  EventsOn('menu-undo', undo);
  EventsOn('menu-redo', redo);
  EventsOn('problem', (t) => problem(errText(t)));
  EventsOn('note', (t) => problem(t, false));
  EventsOn('menu-copy', copySelection); // Edit → Copy
  EventsOn('menu-paste', pasteImage); // Edit → Paste
  EventsOn('menu-save', saveAs); // File → Save As…
  EventsOn('menu-close', closeOrQuit); // File → Close
  EventsOn('menu-close-image', closeImage); // File → Close Image
  EventsOn('menu-delete', deleteImage); // File → Delete Image…
  lastStage = stageSize();
  window.addEventListener('resize', onResize);
  let saveTimer = 0; // the place and size of the window are remembered for the next start (Go also does it when the window closes)
  const rememberWindow = () => { clearTimeout(saveTimer); saveTimer = setTimeout(() => SaveWindow().catch(() => {}), 600); };
  window.addEventListener('resize', rememberWindow);
  window.addEventListener('blur', rememberWindow);
  $('stage').addEventListener('wheel', onWheel, {passive: false});
  $('stage').addEventListener('pointermove', (ev) => { lastPointer = {x: ev.clientX, y: ev.clientY}; readPixel(ev.clientX, ev.clientY); document.body.classList.toggle('shift', ev.shiftKey); if (!view.selDrag && !view.tdrag) updateHover(ev.shiftKey); });
  $('stage').addEventListener('pointerleave', () => { lastPointer = null; clearPixel(); updateHover(false); });
  window.addEventListener('keyup', (ev) => { if (ev.key === 'Shift') { document.body.classList.remove('shift'); updateHover(false); } });
  window.addEventListener('blur', () => { document.body.classList.remove('shift'); updateHover(false); });
  $('stage').addEventListener('pointerdown', (ev) => {
    if (!pic.naturalWidth || (ev.button !== 0 && ev.button !== 1 && !(ev.button === 2 && tool() === 'pen'))) return;
    try { $('stage').setPointerCapture(ev.pointerId); } catch {}
    if (ev.button === 1) { // the middle button pans the picture
      ev.preventDefault();
      stopBounce(); draw(); // a picture that was still coming back stops where it is, inside its bounds
      view.drag = {x: ev.clientX - view.x, y: ev.clientY - view.y, from: [ev.clientX, ev.clientY]};
      $('stage').style.cursor = 'grabbing';
      return;
    }
    pointerDown = true;
    if (tool() === 'pen') { // what is lifted is applied first, then the stroke starts (if the button is still down)
      (view.fl ? flushFloat({clear: true}) : Promise.resolve()).then(() => { if (pointerDown && !view.fl) startInk(ev); });
      return;
    }
    if (document.body.classList.contains('edit') && view.sel) { // an edge of the selection: stretch the selected part, or move it (see above)
      const edge = edgeAt(ev.clientX, ev.clientY, true);
      if (edge) { startTransform(ev, edge); return; }
    }
    if (view.fl) { // what is lifted is applied first; then it is a selection like any other, or, when the button is up by then, a click that lets go of it
      const [cx, cy] = [ev.clientX, ev.clientY];
      flushFloat({clear: false}).then(() => {
        if (view.fl) return;
        if (pointerDown) beginSelection(cx, cy); else clearSelection();
      });
      return;
    }
    beginSelection(ev.clientX, ev.clientY);
  });
  $('stage').addEventListener('pointermove', (ev) => {
    if (ink) { addInk(inkPoint(ev)); return; }
    if (view.tdrag) { dragTransform(ev); return; }
    if (view.selDrag?.edge) {
      const [px, py] = pointInPicture(ev.clientX, ev.clientY), e = view.selDrag.edge;
      if (e.move) { // the inside of the frame: the frame moves, the picture is not touched
        const d = view.selDrag, st = d.start, W = dispW(), H = dispH();
        const dx = clamp(Math.round(px - d.p0[0]), -st.x0, W - st.x1), dy = clamp(Math.round(py - d.p0[1]), -st.y0, H - st.y1);
        view.sel = {x0: st.x0 + dx, x1: st.x1 + dx, y0: st.y0 + dy, y1: st.y1 + dy};
        draw();
        return;
      }
      const snap = view.pixelated ? Math.round : (v) => v; // (an edge of the frame is between two pixels)
      if (e.l) view.sel.x0 = snap(px); if (e.r) view.sel.x1 = snap(px);
      if (e.t) view.sel.y0 = snap(py); if (e.b) view.sel.y1 = snap(py);
      draw();
      return;
    }
    if (view.selDrag) {
      const [px, py] = pointInPicture(ev.clientX, ev.clientY);
      view.sel = dragFrame(view.selDrag.x0, view.selDrag.y0, px, py);
      const r = selectionRect();
      if (!r || r.w * r.h < 2) view.sel = null; // one pixel is not a selection (also not when the drag comes back to it)
      draw();
      return;
    }
    if (!view.drag) return;
    const bd = bounds(), rx = ev.clientX - view.drag.x, ry = ev.clientY - view.drag.y; // where the pointer would put the picture
    const cx = clamp(rx, bd.x[0], bd.x[1]), cy = clamp(ry, bd.y[0], bd.y[1]);
    if (cx !== view.x || cy !== view.y) view.free = true;
    view.x = cx; view.y = cy;
    view.ox = Math.sign(rx - cx) * pullOffset(Math.abs(rx - cx)); view.oy = Math.sign(ry - cy) * pullOffset(Math.abs(ry - cy));
    if (view.ox || view.oy) { bounce.over = true; bounce.lastPush = performance.now(); }
    draw();
  });
  const endDrag = (ev) => {
    pointerDown = false;
    setTimeout(renderSelection, 0); // the button is up: the dashes move again
    if (ink) { const k = ink; ink = null; finishInk(k).then(() => updateHover(ev.shiftKey)); return; } // the button is up: drawing stops now, not when Go is done
    if (view.tdrag) { finishTransform().then(() => updateHover(ev.shiftKey)); return; }
    if (view.selDrag) { // a selection dragged less than 3 screen pixels is a click, not a selection
      const [fx, fy] = view.selDrag.from, moved = !!view.selDrag.edge;
      view.selDrag = null;
      if (moved) { updateHover(ev.shiftKey); if (selectionRect()) copySelection(true); }
      else if (Math.abs(ev.clientX - fx) < 3 && Math.abs(ev.clientY - fy) < 3) { view.sel = null; draw(); }
      else if (selectionRect()) copySelection(true); // a finished selection goes to the clipboard at once
    }
    if (view.drag) { startSpring(); view.drag = null; updateHover(ev.shiftKey); } // let go: the spring takes the pull away
  };
  $('stage').addEventListener('mousedown', (ev) => { if (ev.button === 1) ev.preventDefault(); }); // no auto-scroll cursor
  $('stage').addEventListener('pointerup', endDrag);
  $('stage').addEventListener('pointercancel', endDrag);
  $('stage').addEventListener('dblclick', () => {
    if (!pic.naturalWidth) return;
    if (!view.free && (Math.abs(view.sx - 1) > 0.005 || Math.abs(view.sy - 1) > 0.005)) zoomTo100(); else { view.free = false; draw(); }
  });
  window.addEventListener('keydown', (ev) => {
    if (ev.target instanceof HTMLInputElement && ev.target.type !== 'range') return; // typing in a field is not a command
    if (ev.key === 'Shift') { document.body.classList.add('shift'); updateHover(true); return; }
    if ((ev.ctrlKey || ev.metaKey) && !ev.altKey && (ev.key === 'c' || ev.key === 'C' || ev.key === 'Insert')) {
      ev.preventDefault(); copySelection(); return; // Ctrl+C, Cmd+C and Ctrl+Insert
    }
    if ((ev.ctrlKey || ev.metaKey) && !ev.altKey && (ev.key === 'v' || ev.key === 'V')) { ev.preventDefault(); pasteImage(); return; }
    if ((ev.ctrlKey || ev.metaKey) && !ev.altKey && (ev.key === 'z' || ev.key === 'Z')) { ev.preventDefault(); if (ev.shiftKey) redo(); else undo(); return; }
    if (ev.ctrlKey && !ev.metaKey && !ev.altKey && (ev.key === 'y' || ev.key === 'Y')) { ev.preventDefault(); redo(); return; }
    if (ev.shiftKey && ev.key === 'Insert') { ev.preventDefault(); pasteImage(); return; } // Shift+Insert
    if ((ev.ctrlKey || ev.metaKey) && !ev.altKey && (ev.key === 's' || ev.key === 'S')) { ev.preventDefault(); saveAs(); return; }
    if ((ev.ctrlKey || ev.metaKey) && !ev.altKey && (ev.key === 'n' || ev.key === 'N')) { ev.preventDefault(); NewWindow(); return; } // Ctrl+N, Cmd+N
    if (!menuShown && ev.ctrlKey && !ev.metaKey && !ev.altKey) { // the menu bar is hidden (Windows full screen), and with it the shortcuts of its items that this page does not handle
      const k = ev.key.toLowerCase();
      if (k === 'o') { ev.preventDefault(); choose(); return; }
      if (k === 'w' && ev.shiftKey) { ev.preventDefault(); closeImage(); return; }
      if (k === 'w') { ev.preventDefault(); closeOrQuit(); return; }
      if (k === 'q') { ev.preventDefault(); CloseWindow(); return; }
    }
    if (ev.metaKey || ev.ctrlKey || ev.altKey) return;
    if (ev.key === 'Enter') { ev.preventDefault(); ToggleFullscreen(); settleFullscreen(); return; }
    if (view.fl?.pasted && (ev.key === 'Escape' || ev.key === 'Delete' || ev.key === 'Backspace')) { ev.preventDefault(); removeLayer(); return; }
    if (ev.key === 'Escape') { ExitFullscreen(); settleFullscreen(); clearSelection(); return; }
    if (ev.key === 'e' || ev.key === 'E' || ev.key === 'Tab') { ev.preventDefault(); setEdit(!editOn()); return; } // the edit bar (Tab does the same)
    if (!pic.naturalWidth) return;
    const b = stageSize();
    if (ev.key === 'ArrowLeft') { ev.preventDefault(); step(-1); }
    else if (ev.key === 'ArrowRight' || ev.key === ' ') { ev.preventDefault(); step(1); }
    else if (ev.key === 'Backspace') { ev.preventDefault(); step(-1); }
    else if (MODES[ev.key]) { setMode(MODES[ev.key]); commit(); }
    else if (ev.key === 'r' || ev.key === 'R') {
      const dir = ev.shiftKey ? -1 : 1;
      if (view.fl) flushFloat({clear: true}).then(() => { rotate(dir); commit(); }); // what is lifted is applied before the picture is turned
      else { rotate(dir); commit(); }
    }
    else if (ev.key === 's' || ev.key === 'S') saveAs();
    else if (ev.key === 'f' || ev.key === 'F') { ToggleFullscreen(); settleFullscreen(); }
    else if (ev.key === 'd' || ev.key === 'D') deleteImage();
    else if (ev.key === '+' || ev.key === '=') zoomAt(1.25, b.w / 2, b.h / 2);
    else if (ev.key === '-') zoomAt(0.8, b.w / 2, b.h / 2);
  });
}

document.addEventListener('copy', (ev) => { ev.preventDefault(); copySelection(); });
document.addEventListener('visibilitychange', () => Trace(`visibility ${document.visibilityState}`));
// macOS may still be hiding the app when the page starts (see App.startup): ask to be shown until the page is visible.
async function revealWhenHidden() {
  for (let i = 0; i < 40 && document.visibilityState === 'hidden'; i++) { await Reveal(); await new Promise((r) => setTimeout(r, 150)); }
}

(async () => {
  Trace(`js start, visibility ${document.visibilityState}`);
  if (await Mode() === 'about') { showAbout(); return; }
  revealWhenHidden();
  startViewer();
  showImage(await Image());
  commit(); // the way the window starts is where undo ends
})();
