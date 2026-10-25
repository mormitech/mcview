import './style.css';
import {Version, Mode, Image, ChooseImage, Step, Trace, Reveal, ToggleFullscreen, ExitFullscreen, IsFullscreen, CloseWindow, PreparePixels, PixelAt, SelectionPNG, RotatePNG, RestoreState, KeepStates, CopyImageNative, PasteImage, SaveAs, NewWindow} from '../wailsjs/go/main/App';
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

document.addEventListener('click', (ev) => {
  const a = ev.target.closest('a[href^="http"]');
  if (a) { ev.preventDefault(); BrowserOpenURL(a.href); }
});

// ---- Viewer ----------------------------------------------------------------------------------------------
// How the picture sits in the window is one of four modes, the way IrfanView has them on the keys 1 to 4, and the mode stays when
// the next or the previous picture is shown (and the next time mcView starts):
//   1  original  - the pixels of the picture, 100%, however small or big it is (a big one is centred and can be moved)
//   2  large     - a picture bigger than the window is shrunk to fit it, a smaller one stays 100% (the default)
//   3  contain   - every picture is zoomed, up or down, to touch the borders of the window, keeping its proportions
//   4  stretch   - every picture is stretched border to border, proportions or not
// The scroll wheel (or a pinch on the trackpad) zooms around the pointer (on a Mac two fingers on the trackpad move the picture instead),
// dragging moves the picture, + and - zoom: that leaves
// the mode until the next picture or the next of the keys 1 to 4. A double click toggles between 100% and the mode.
// The arrow keys go to the previous and the next picture of the folder (so do Backspace and Space, as in IrfanView), Enter toggles full screen (Esc leaves it), Tab the pixelated view, Shift and drag selects, Ctrl+C copies, Ctrl+V pastes, S saves, R turns the picture (Shift+R the other way), Ctrl+Z undoes and Ctrl+Shift+Z redoes, Ctrl+N opens a new window.

const MODES = {'1': 'original', '2': 'large', '3': 'contain', '4': 'stretch'};
const pic = $('pic');
const view = {name: '', dir: '', index: 0, count: 0, mode: savedMode(), pixelated: false, state: 0, token: 0, sel: null, selDrag: null, rot: 0, angle: 0, free: false, ox: 0, oy: 0, sx: 1, sy: 1, x: 0, y: 0, drag: null};
let problemTimer = 0;

function savedMode() {
  try { const m = localStorage.getItem('mcview.mode'); if (Object.values(MODES).includes(m)) return m; } catch {}
  return 'large';
}

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
  view.rot = (view.rot + dir + 4) % 4; view.angle += dir * 90;
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
    case 'large': return [Math.min(1, fit), Math.min(1, fit)];
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
  setStatus(view.dir, `${place}${dispW()} × ${dispH()}  ·  ${zoom}${view.pixelated ? '  ·  pixels' : ''}${sel ? `  ·  selection ${sel.w} × ${sel.h}` : ''}`, view.name);
  renderSelection();
  if (view.pixelated && lastPointer) readPixel(lastPointer.x, lastPointer.y); // the picture moved under the pointer
}

function setStatus(folder, info = '', name = '') {
  $('status').hidden = !folder && !info;
  $('name').textContent = name; $('name').title = name;
  $('folder').textContent = folder; $('folder').title = folder;
  $('info').textContent = info;
  $('status').classList.remove('bad');
}

// The keys 1 to 4: choose a mode, and keep it.
function setMode(mode) {
  stopBounce();
  view.mode = mode; view.free = false;
  try { localStorage.setItem('mcview.mode', mode); } catch {}
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
function showImage(info, opts = {}) {
  view.state = info?.state || 0;
  clearTimeout(turnTimer); pic.classList.remove('turning'); view.rot = opts.rot || 0; view.angle = view.rot * 90;
  if (!info?.url) { // no picture: the window as it is when it starts
    stopBounce(); pic.onload = null; pic.removeAttribute('src');
    view.name = ''; view.dir = ''; view.index = 0; view.count = 0; view.token++; view.sel = null; view.selDrag = null; view.hl = null; $('stage').style.cursor = ''; clearPixel();
    pic.hidden = true; $('empty').hidden = false; setStatus('');
    return;
  }
  view.name = info.name; view.dir = info.dir; view.index = info.index; view.count = info.count; view.free = false;
  stopBounce();
  view.token++; pixelState = ''; pixelCanvas = null; pixelPromise = null; view.sel = null; view.selDrag = null; view.hl = null; $('stage').style.cursor = ''; clearPixel();
  pic.onload = () => {
    Trace(`img onload, visibility ${document.visibilityState}, focus ${document.hasFocus()}`);
    pic.hidden = false; $('empty').hidden = true; draw();
    if (view.pixelated) preparePixels();
    pic.decode().then(() => Trace('decoded'), () => {});
    requestAnimationFrame(() => Trace('first paint after load'));
  };
  pic.onerror = () => problem(`Cannot show ${info.name}`);
  Trace('img src set');
  pic.src = info.url;
}

function problem(text, bad = true) {
  $('status').hidden = false;
  $('name').textContent = ''; $('folder').textContent = text; $('folder').title = ''; $('info').textContent = '';
  $('status').classList.toggle('bad', bad);
  clearTimeout(problemTimer);
  document.body.classList.add('peek');
  problemTimer = setTimeout(() => { document.body.classList.remove('peek'); draw(); }, 4000);
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
// the bottom bar shows the colour of the pixel under the pointer. The numbers come from Go, which decodes the file a second time
// (the web view converts colours for the screen, so what it shows is not the number in the file); for types Go cannot decode they are
// read from what the web view drew, and the bar says so. The pixelated view stays when the next picture is shown, but not the next
// time mcView starts.

let pixelState = '';   // '' not asked, 'preparing', 'go' or 'browser' (where the numbers come from), 'none' (cannot be read)
let pixelCanvas = null;
let lastPointer = null;
let pixelPending = null, pixelBusy = false;

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
  // [x, y]  (r, g, b) - #rrggbb; a pixel that is not opaque has its alpha as a fourth number and at the end of the hex code
  const rgb = p.a < 255 ? `(${p.r}, ${p.g}, ${p.b}, ${p.a})` : `(${p.r}, ${p.g}, ${p.b})`;
  const code = `#${hex2(p.r)}${hex2(p.g)}${hex2(p.b)}${p.a < 255 ? hex2(p.a) : ''}`;
  const from = pixelState === 'browser' ? '  ·  from the screen' : '';
  const sw = document.createElement('span'); sw.className = 'sw';
  const fill = document.createElement('b'); fill.style.background = `rgba(${p.r},${p.g},${p.b},${p.a / 255})`; sw.append(fill);
  const text = document.createElement('span');
  text.textContent = `[${p.x}, ${p.y}]  ${rgb} - ${code}${from}`;
  $('pixel').replaceChildren(sw, text);
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
// picture when nothing is selected. The part is cut out of the decoded file (see SelectionPNG), so it has the colours of the file.
// In the pixelated view the selection snaps to whole pixels; in the smoothed view it follows the pointer, and what is copied is the
// pixels it touches. The selection is kept in picture pixels, so it stays on its place when the picture is zoomed or moved.

// The selection as whole pixels of the picture: {x, y, w, h}, or null.
function selectionRect() {
  const s = view.sel, nw = dispW(), nh = dispH();
  if (!s || !nw) return null;
  const lo = (a, b) => Math.min(a, b), hi = (a, b) => Math.max(a, b);
  let x0, x1, y0, y1;
  if (view.pixelated) { // the pixel under each end, both included
    x0 = Math.floor(lo(s.x0, s.x1)); x1 = Math.floor(hi(s.x0, s.x1)) + 1;
    y0 = Math.floor(lo(s.y0, s.y1)); y1 = Math.floor(hi(s.y0, s.y1)) + 1;
  } else { // every pixel the selection touches
    x0 = Math.floor(lo(s.x0, s.x1)); x1 = Math.ceil(hi(s.x0, s.x1));
    y0 = Math.floor(lo(s.y0, s.y1)); y1 = Math.ceil(hi(s.y0, s.y1));
  }
  x0 = clamp(x0, 0, nw); x1 = clamp(x1, 0, nw); y0 = clamp(y0, 0, nh); y1 = clamp(y1, 0, nh);
  return x1 > x0 && y1 > y0 ? {x: x0, y: y0, w: x1 - x0, h: y1 - y0} : null;
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
  el.classList.toggle('copied', copiedKey === boxKey(box));
  const {x0, x1, y0, y1} = box;
  for (const e of ['t', 'b', 'l', 'r']) el.querySelector(`.edge.${e}`).classList.toggle('on', !!view.hl?.[e]);
  el.style.left = `${view.x + view.ox + x0 * view.sx}px`; el.style.top = `${view.y + view.oy + y0 * view.sy}px`;
  el.style.width = `${(x1 - x0) * view.sx}px`; el.style.height = `${(y1 - y0) * view.sy}px`;
  el.hidden = false;
}

// Shift held over an edge (or a corner) of the frame: the edge is lit, the pointer says "resize", and Shift and drag moves the edge.
const EDGE_GRAB = 6; // screen pixels

// Which edges of the frame are within reach of a position of the window: {t, b, l, r} (corners have two), or null.
function edgeAt(clientX, clientY) {
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
  return hit.l || hit.r || hit.t || hit.b ? hit : null;
}

const RESIZE_CURSORS = {t: 'ns-resize', b: 'ns-resize', l: 'ew-resize', r: 'ew-resize', tl: 'nwse-resize', br: 'nwse-resize', tr: 'nesw-resize', bl: 'nesw-resize'};

// Lights the edge under the pointer while Shift is down (and while an edge is being dragged).
function updateHover(shift) {
  if (view.selDrag?.edge) return;
  const hit = shift && lastPointer && !view.drag ? edgeAt(lastPointer.x, lastPointer.y) : null;
  const was = view.hl;
  view.hl = hit;
  const key = hit ? (hit.t ? 't' : hit.b ? 'b' : '') + (hit.l ? 'l' : hit.r ? 'r' : '') : '';
  $('stage').style.cursor = key ? RESIZE_CURSORS[key] : '';
  if (JSON.stringify(was) !== JSON.stringify(hit)) renderSelection();
}

function clearSelection() {
  if (!view.sel) return;
  view.sel = null; view.selDrag = null; view.hl = null; $('stage').style.cursor = '';
  draw();
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

let lastCopy = 0;

async function copySelection() {
  if (!pic.naturalWidth) return;
  const now = performance.now();
  if (now - lastCopy < 400) return; // the key and the system's copy command can both come for one press
  lastCopy = now;
  const nw = pic.naturalWidth, nh = pic.naturalHeight;
  const r = selectionRect() || {x: 0, y: 0, w: dispW(), h: dispH()};
  const n = toNaturalRect(r); // the same part in the file's picture, which is cut out and then turned the way it is shown
  try {
    await preparePixels();
    let b64 = '';
    if (pixelState === 'go') b64 = await SelectionPNG(n.x, n.y, n.w, n.h, nw, nh);
    if (!b64) { // a type Go cannot decode: cut it out of what the web view drew
      if (!pixelCanvas) throw new Error('this picture cannot be copied');
      const c = document.createElement('canvas'); c.width = n.w; c.height = n.h;
      c.getContext('2d').putImageData(pixelCanvas.getImageData(n.x, n.y, n.w, n.h), 0, 0);
      b64 = await blobToBase64(await new Promise((res) => c.toBlob(res, 'image/png')));
    }
    if (view.rot) b64 = await RotatePNG(b64, view.rot);
    try { // the web view's own clipboard, and the system's way when it does not take the picture
      await navigator.clipboard.write([new ClipboardItem({'image/png': base64ToBlob(b64, 'image/png')})]);
    } catch { await CopyImageNative(b64); }
    copiedKey = view.sel && selectionBox() ? boxKey(selectionBox()) : null;
    renderSelection();
    problem(`Copied ${r.w} × ${r.h} pixels${view.sel ? ' (the selection)' : ''} to the clipboard`, false);
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

// Ctrl+V (Cmd+V, Shift+Insert, Edit → Paste): the picture on the clipboard becomes the picture of the window. It has no file yet:
// the title says [clipboard image], and S saves it.
async function pasteImage() {
  try { await PasteImage(); } catch (e) { problem(errText(e)); }
}

// S (Ctrl+S, Cmd+S, File → Save As…): asks where to save the picture. Saving it to another folder makes that folder the one the
// window works in, and the arrow keys go through its pictures.
async function saveAs() {
  try { const note = await SaveAs(); if (note) problem(note, false); } catch (e) { problem(errText(e)); }
}

// When the window changes its size (full screen with Enter, and back, or the window dragged bigger), a picture that was zoomed or
// moved keeps the pixel that was in the middle of the window in the middle. The pixel is remembered, because a picture near its edge
// cannot be held in the middle while the window is small (it is pushed back to the edge) but should be again when the window is big
// again; the memory is only good while the picture has not been moved or zoomed since. A picture a mode places is fitted again.
let lastStage = null, anchor = null;

// Full screen has no bar at the bottom; Shift shows it for as long as it is held (and a message for a few seconds).
async function refreshFullscreen() {
  try { document.body.classList.toggle('full', await IsFullscreen()); } catch {}
}

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

const snapshot = () => ({state: view.state, mode: view.mode, pixelated: view.pixelated, rot: view.rot});
const sameSnapshot = (a, b) => a.state === b.state && a.mode === b.mode && a.pixelated === b.pixelated && a.rot === b.rot;

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
  try {
    if (e.state !== view.state) { // another picture (or none): Go brings it back, the view is set as it was
      const info = await RestoreState(e.state);
      view.mode = e.mode; view.free = false; view.pixelated = e.pixelated; pic.classList.toggle('pixelated', e.pixelated);
      try { localStorage.setItem('mcview.mode', e.mode); } catch {}
      showImage(info, {rot: e.rot});
    } else {
      if (e.mode !== view.mode) setMode(e.mode);
      if (e.pixelated !== view.pixelated) setPixelated(e.pixelated);
      turnTo(e.rot);
    }
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

function startViewer() {
  const cmd = navigator.platform.startsWith('Mac') ? '⌘' : 'Ctrl+';
  $('hint').textContent = `Open a picture with File → Open… (${cmd}O), paste one from the clipboard (${cmd}V), or open one with mcView from your file manager.`;
  $('open').onclick = choose;
  EventsOn('image', (info) => { showImage(info); commit(); });
  EventsOn('menu-undo', undo);
  EventsOn('menu-redo', redo);
  EventsOn('problem', (t) => problem(errText(t)));
  EventsOn('note', (t) => problem(t, false));
  EventsOn('menu-copy', copySelection); // Edit → Copy
  lastStage = stageSize();
  window.addEventListener('resize', onResize);
  $('stage').addEventListener('wheel', onWheel, {passive: false});
  $('stage').addEventListener('pointermove', (ev) => { lastPointer = {x: ev.clientX, y: ev.clientY}; readPixel(ev.clientX, ev.clientY); document.body.classList.toggle('shift', ev.shiftKey); if (!view.selDrag) updateHover(ev.shiftKey); });
  $('stage').addEventListener('pointerleave', () => { lastPointer = null; clearPixel(); updateHover(false); });
  window.addEventListener('keyup', (ev) => { if (ev.key === 'Shift') { document.body.classList.remove('shift'); updateHover(false); } });
  window.addEventListener('blur', () => { document.body.classList.remove('shift'); updateHover(false); });
  $('stage').addEventListener('pointerdown', (ev) => {
    if (!pic.naturalWidth || ev.button !== 0) return;
    $('stage').setPointerCapture(ev.pointerId);
    if (ev.shiftKey) { // Shift and drag: select
      const edge = edgeAt(ev.clientX, ev.clientY);
      if (edge) { // an edge of the frame: it is moved, not a new frame drawn
        const box = selectionBox();
        view.sel = {x0: box.x0, x1: box.x1, y0: box.y0, y1: box.y1}; // the first corner is the left top one, so l and t are x0 and y0
        view.selDrag = {edge, from: [ev.clientX, ev.clientY]};
        view.hl = edge;
        draw();
        return;
      }
      const [x, y] = pointInPicture(ev.clientX, ev.clientY);
      view.selDrag = {x0: x, y0: y, from: [ev.clientX, ev.clientY]};
      view.sel = {x0: x, y0: y, x1: x, y1: y};
      draw();
      return;
    }
    stopBounce(); draw(); // a picture that was still coming back stops where it is, inside its bounds
    view.drag = {x: ev.clientX - view.x, y: ev.clientY - view.y, from: [ev.clientX, ev.clientY]};
  });
  $('stage').addEventListener('pointermove', (ev) => {
    if (view.selDrag?.edge) {
      const [px, py] = pointInPicture(ev.clientX, ev.clientY), e = view.selDrag.edge;
      if (e.l) view.sel.x0 = px; if (e.r) view.sel.x1 = px;
      if (e.t) view.sel.y0 = py; if (e.b) view.sel.y1 = py;
      draw();
      return;
    }
    if (view.selDrag) {
      [view.sel.x1, view.sel.y1] = pointInPicture(ev.clientX, ev.clientY);
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
    if (view.selDrag) { // a selection dragged less than 3 screen pixels is a click, not a selection
      const [fx, fy] = view.selDrag.from, moved = !!view.selDrag.edge;
      view.selDrag = null;
      if (moved) updateHover(ev.shiftKey);
      else if (!view.pixelated && Math.abs(ev.clientX - fx) < 3 && Math.abs(ev.clientY - fy) < 3) { view.sel = null; draw(); }
    } else if (view.drag) { // a click on the picture (no Shift, no drag) takes the selection away
      const [fx, fy] = view.drag.from;
      if (Math.abs(ev.clientX - fx) < 3 && Math.abs(ev.clientY - fy) < 3) clearSelection();
    }
    if (view.drag) startSpring(); // let go: the spring takes the pull away
    view.drag = null;
  };
  $('stage').addEventListener('pointerup', endDrag);
  $('stage').addEventListener('pointercancel', endDrag);
  $('stage').addEventListener('dblclick', () => {
    if (!pic.naturalWidth) return;
    if (!view.free && (Math.abs(view.sx - 1) > 0.005 || Math.abs(view.sy - 1) > 0.005)) zoomTo100(); else { view.free = false; draw(); }
  });
  window.addEventListener('keydown', (ev) => {
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
    if (ev.metaKey || ev.ctrlKey || ev.altKey) return;
    if (ev.key === 'Enter') { ev.preventDefault(); ToggleFullscreen(); settleFullscreen(); return; }
    if (ev.key === 'Escape') { ExitFullscreen(); settleFullscreen(); clearSelection(); return; }
    if (!pic.naturalWidth) return;
    if (ev.key === 'Tab') { ev.preventDefault(); setPixelated(!view.pixelated); commit(); return; }
    const b = stageSize();
    if (ev.key === 'ArrowLeft') { ev.preventDefault(); step(-1); }
    else if (ev.key === 'ArrowRight' || ev.key === ' ') { ev.preventDefault(); step(1); }
    else if (ev.key === 'Backspace') { ev.preventDefault(); step(-1); }
    else if (MODES[ev.key]) { setMode(MODES[ev.key]); commit(); }
    else if (ev.key === 'r' || ev.key === 'R') { rotate(ev.shiftKey ? -1 : 1); commit(); }
    else if (ev.key === 's' || ev.key === 'S') saveAs();
    else if (ev.key === 'f' || ev.key === 'F') { setMode('large'); commit(); }
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
