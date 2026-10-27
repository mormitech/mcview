import {ChatInfo, ChatLogin, ChatReset, ChatSend, ChatStop} from '../wailsjs/go/main/App';
import {BrowserOpenURL, EventsOn} from '../wailsjs/runtime/runtime';
import {renderMarkdown} from './md.js';

// The chat panel: Claude Code, started in the background by Go (see chat.go), with the pictures of this window as its tool. Three states: Claude Code is
// not installed, the user is not signed in (the sign-in opens the browser and Claude Code keeps the token), and the chat itself.

const $ = (id) => document.getElementById(id);
const errText = (e) => (e && e.message) || String(e);
let isOpen = false, info = null, busy = false, bubble = null, streamed = false, signingIn = false, loginUrl = '', act = null;

function add(cls, text) {
  const d = document.createElement('div');
  d.className = cls; d.textContent = text;
  $('chat-log').append(d);
  scroll();
  return d;
}
// The answer is Markdown: it is drawn again, whole, each time a piece comes (the pieces can cut a word or a ** in half).
function say(text, append) {
  if (!bubble) { bubble = add('msg bot', ''); bubble.dataset.raw = ''; }
  bubble.dataset.raw = append ? bubble.dataset.raw + text : text;
  bubble.replaceChildren(renderMarkdown(bubble.dataset.raw));
  scroll();
}

// What Claude does with the tools is one line, not a list: a spinner while it works, the number of actions and how many failed; the list opens on a click.
function startActivity() {
  const d = document.createElement('details');
  d.className = 'act working';
  d.innerHTML = '<summary><span class="spin"></span><span class="lbl">Processing…</span><span class="cnt"></span></summary><div class="list"></div>';
  $('chat-log').append(d);
  scroll();
  act = {el: d, list: d.querySelector('.list'), lbl: d.querySelector('.lbl'), cnt: d.querySelector('.cnt'), n: 0, bad: 0};
}
function showCount() {
  const a = act, bits = [];
  if (a.n) bits.push(`${a.n} action${a.n === 1 ? '' : 's'}`);
  if (a.bad) bits.push(`${a.bad} failed`);
  a.cnt.textContent = bits.length ? '· ' + bits.join(' · ') : '';
  a.el.classList.toggle('hasbad', a.bad > 0);
}
function endActivity(stopped) {
  if (!act) return;
  if (!act.n && !act.bad) act.el.remove();
  else { act.el.classList.remove('working'); act.lbl.textContent = stopped ? 'Stopped' : act.bad ? 'Done with problems' : 'Done'; showCount(); }
  act = null;
}

const scroll = () => { const l = $('chat-log'); l.scrollTop = l.scrollHeight; };

function setBusy(b) {
  busy = b;
  $('chat-send').hidden = b; $('chat-stop').hidden = !b;
}

function showSetup(text, buttons = [], extra = '') {
  $('chat-setup').hidden = false; $('chat-main').hidden = true;
  $('chat-msg').textContent = text;
  const box = $('chat-actions'); box.replaceChildren();
  for (const [label, fn, primary] of buttons) {
    const b = document.createElement('button'); b.textContent = label; b.onclick = fn; if (primary) b.className = 'primary'; box.append(b);
  }
  const u = $('chat-url'); u.replaceChildren();
  if (extra) {
    const a = document.createElement('a'); a.textContent = extra; a.href = '#'; a.onclick = (ev) => { ev.preventDefault(); BrowserOpenURL(extra); }; u.append('If no browser tab opened: ', a);
  }
}

async function login() {
  try { signingIn = true; await ChatLogin(); render(); } catch (e) { signingIn = false; render(errText(e)); }
}

function render(problem = '') {
  if (!info) { showSetup('Checking Claude Code…'); return; }
  if (!info.claude) {
    showSetup(`${problem ? problem + '\n\n' : ''}Claude Code is not installed. The chat uses your own Claude Code and your Claude subscription. Install it from claude.com/product/claude-code, then check again.`, [['Check again', refresh, true]]);
  } else if (signingIn) {
    showSetup('Waiting for the sign-in in your browser…', [['Cancel', () => { ChatStop(); signingIn = false; loginUrl = ''; render(); }]], loginUrl);
  } else if (!info.loggedIn) {
    showSetup(`${problem ? problem + '\n\n' : ''}Sign in to Claude to chat here. This opens your browser; mcView never sees your password. Claude Code keeps the sign-in.`, [['Sign in with Claude', login, true], ['Check again', refresh]]);
  } else {
    $('chat-setup').hidden = true; $('chat-main').hidden = false;
    const who = [info.email, info.plan].filter(Boolean).join(' · ');
    $('chat-who').textContent = who;
    $('chat-who').title = info.method && info.method !== 'claude.ai' ? 'Not a Claude subscription: this is billed by usage' : '';
    $('chat-who').classList.toggle('warn', !!info.method && info.method !== 'claude.ai');
    if (!$('chat-log').children.length) add('hint', 'Ask Claude to change this picture, for example “draw a red circle in the middle”.');
    setBusy(!!info.busy);
    $('chat-text').focus();
  }
}

async function refresh() {
  info = null; render();
  try { info = await ChatInfo(); signingIn = signingIn && !!info.signingIn; } catch (e) { info = {claude: false}; render(errText(e)); return; }
  render();
}

// The panel floats over the picture's bottom right corner, like a chat window: it grows out of the chat button and shrinks back into it.
export function toggleChat(force) {
  isOpen = force ?? !isOpen;
  $('chat').classList.toggle('open', isOpen);
  $('chat').setAttribute('aria-hidden', String(!isOpen));
  document.body.classList.toggle('chat-open', isOpen);
  if (isOpen) { refresh(); setTimeout(grow, 0); }
}

function send() {
  const t = $('chat-text').value.trim();
  if (!t || busy) return;
  $('chat-text').value = ''; grow();
  const hint = $('chat-log').querySelector('.hint'); if (hint) hint.remove();
  add('msg user', t);
  bubble = null; streamed = false; setBusy(true); startActivity();
  ChatSend(t).catch((e) => { endActivity(); add('msg err', errText(e)); setBusy(false); });
}

// The box grows with what is typed, from one line up to about eight, and only then gets a scrollbar. (A box with a fixed row count and a scrollbar that shows for one
// line is what this replaces: the height is measured with the box at its smallest, and the border is part of it.)
const CHAT_MAX_H = 168;
function grow() {
  const t = $('chat-text');
  t.style.height = 'auto';
  const border = t.offsetHeight - t.clientHeight, want = t.scrollHeight + border;
  t.style.height = `${Math.min(want, CHAT_MAX_H)}px`;
  t.style.overflowY = want > CHAT_MAX_H ? 'auto' : 'hidden';
}

function onEvent(ev) {
  switch (ev.kind) {
    case 'start': bubble = null; streamed = false; break;
    case 'delta': say(ev.text, true); streamed = true; break;
    case 'text': if (!streamed && ev.text.trim()) { bubble = null; say(ev.text, false); } break;
    case 'tool': bubble = null; if (!act) startActivity(); act.n++; { const t = document.createElement('div'); t.className = 'tool'; t.textContent = `${ev.name} ${ev.input}`; act.list.append(t); } act.lbl.textContent = 'Processing…'; showCount(); break;
    case 'toolerror': if (!act) startActivity(); act.bad++; { const t = document.createElement('div'); t.className = 'tool bad'; t.textContent = ev.text || 'A tool failed'; act.list.append(t); } showCount(); break;
    case 'done': if (ev.error) add('msg err', ev.text || 'Claude could not answer'); break;
    case 'error': endActivity(); add('msg err', ev.text); break;
    case 'end': endActivity(ev.stopped); setBusy(false); bubble = null; $('chat-text').focus(); break;
    case 'loginurl': loginUrl = ev.url; if (signingIn) render(); break;
    case 'login': signingIn = false; loginUrl = ''; if (ev.ok) refresh(); else { refresh().then(() => render(ev.message ? `The sign-in did not finish: ${ev.message}` : 'The sign-in did not finish.')); } break;
  }
}

// The width of the panel: 460 pixels to start with, then what the user dragged it to (kept for next time), never under 280 or over what leaves 320 pixels for the picture.
const CHAT_MIN = 280;
const chatMax = () => Math.max(CHAT_MIN, window.innerWidth - 320);
function setChatWidth(w, save) {
  w = Math.round(Math.min(chatMax(), Math.max(CHAT_MIN, w)));
  document.documentElement.style.setProperty('--chat-w', `${w}px`);
  if (save) { try { localStorage.setItem('chatWidth', String(w)); } catch {} }
  return w;
}

function initResize() {
  try { const w = parseInt(localStorage.getItem('chatWidth') || '', 10); if (w) setChatWidth(w, false); } catch {}
  const handle = $('chat-resize');
  handle.addEventListener('pointerdown', (ev) => {
    if (ev.button !== 0) return;
    ev.preventDefault();
    handle.setPointerCapture(ev.pointerId);
    document.body.classList.add('chat-resizing');
  });
  handle.addEventListener('pointermove', (ev) => {
    if (!document.body.classList.contains('chat-resizing')) return;
    setChatWidth(window.innerWidth - 16 - ev.clientX, false);
  });
  const end = (ev) => {
    if (!document.body.classList.contains('chat-resizing')) return;
    document.body.classList.remove('chat-resizing');
    setChatWidth(window.innerWidth - 16 - ev.clientX, true);
  };
  handle.addEventListener('pointerup', end);
  handle.addEventListener('pointercancel', (ev) => { document.body.classList.remove('chat-resizing'); });
  window.addEventListener('resize', () => { // a window that gets smaller takes the panel with it
    const cur = parseInt(getComputedStyle(document.documentElement).getPropertyValue('--chat-w'), 10);
    if (cur && cur > chatMax()) setChatWidth(chatMax(), false);
  });
}

export function initChat() {
  initResize();
  EventsOn('chat', onEvent);
  $('chat-toggle').onclick = () => toggleChat();
  document.addEventListener('pointerdown', (ev) => { if (isOpen && !ev.target.closest('#chat, #chat-toggle')) toggleChat(false); }, true); // a click next to the panel minimizes it (and still does what it was for)
  $('chat-close').onclick = () => toggleChat(false);
  $('chat-new').onclick = () => { ChatReset(); $('chat-log').replaceChildren(); bubble = null; setBusy(false); add('hint', 'New conversation. Ask Claude to change this picture.'); };
  $('chat-send').onclick = send;
  $('chat-stop').onclick = () => ChatStop();
  const t = $('chat-text');
  t.addEventListener('input', grow);
  t.addEventListener('keydown', (ev) => {
    if (ev.key === 'Enter' && !ev.shiftKey && !ev.isComposing) { ev.preventDefault(); send(); }
    if (ev.key === 'Escape') { ev.preventDefault(); t.blur(); }
  });
}
