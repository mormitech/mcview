import {BrowserOpenURL} from '../wailsjs/runtime/runtime';

// A small, safe Markdown renderer for the chat: it builds DOM nodes (no innerHTML), so nothing in an answer can run. It knows paragraphs, headings,
// bold, italic, strike-through, `code`, fenced code blocks, bullet and numbered lists, quotes, rules, tables and [links](https://…).

const el = (tag, cls, text) => { const e = document.createElement(tag); if (cls) e.className = cls; if (text !== undefined) e.textContent = text; return e; };

const INLINE = /(`+)([^`]|[^`][\s\S]*?[^`])\1(?!`)|\*\*([^*]|[^*][\s\S]*?[^*])\*\*|__([^_]|[^_][\s\S]*?[^_])__|~~([^~]|[^~][\s\S]*?[^~])~~|\*([^*\s]|[^*\s][^*]*?[^*\s])\*|\b_([^_\s]|[^_\s][^_]*?[^_\s])_\b|\[([^\]]+)\]\((https?:\/\/[^)\s]+)\)/;

function inline(text, into) {
  while (text) {
    const m = INLINE.exec(text);
    if (!m) { into.append(text); return; }
    if (m.index) into.append(text.slice(0, m.index));
    if (m[2] !== undefined) into.append(el('code', '', m[2].trim()));
    else if (m[3] !== undefined || m[4] !== undefined) { const b = el('strong'); inline(m[3] ?? m[4], b); into.append(b); }
    else if (m[5] !== undefined) { const s = el('del'); inline(m[5], s); into.append(s); }
    else if (m[6] !== undefined || m[7] !== undefined) { const i = el('em'); inline(m[6] ?? m[7], i); into.append(i); }
    else {
      const a = el('a'); a.href = m[9]; a.title = m[9]; inline(m[8], a);
      a.onclick = (ev) => { ev.preventDefault(); BrowserOpenURL(m[9]); };
      into.append(a);
    }
    text = text.slice(m.index + m[0].length);
  }
}

const cells = (line) => line.trim().replace(/^\|/, '').replace(/\|$/, '').split('|').map((c) => c.trim());
const isRule = (l) => /^ {0,3}([-*_])( *\1){2,} *$/.test(l);
const isTableSep = (l) => /^\s*\|?\s*:?-{2,}:?\s*(\|\s*:?-{2,}:?\s*)*\|?\s*$/.test(l) && l.includes('-') && (l.includes('|') || false);

export function renderMarkdown(src) {
  const out = document.createDocumentFragment();
  const lines = src.replace(/\r\n?/g, '\n').split('\n');
  let i = 0;
  while (i < lines.length) {
    const line = lines[i];
    if (!line.trim()) { i++; continue; }
    const fence = /^ {0,3}(`{3,}|~{3,})\s*([\w+-]*)/.exec(line);
    if (fence) { // a code block, up to the closing fence (or the end: the answer may still be written)
      const code = []; i++;
      while (i < lines.length && !lines[i].trim().startsWith(fence[1])) code.push(lines[i++]);
      i++;
      const pre = el('pre'); pre.append(el('code', '', code.join('\n'))); out.append(pre);
      continue;
    }
    const h = /^ {0,3}(#{1,6})\s+(.*?)\s*#*\s*$/.exec(line);
    if (h) { const e = el('h' + Math.min(h[1].length + 1, 5)); inline(h[2], e); out.append(e); i++; continue; }
    if (isRule(line)) { out.append(el('hr')); i++; continue; }
    if (/^ {0,3}>/.test(line)) {
      const q = []; while (i < lines.length && /^ {0,3}>/.test(lines[i])) q.push(lines[i++].replace(/^ {0,3}> ?/, ''));
      const b = el('blockquote'); b.append(renderMarkdown(q.join('\n'))); out.append(b);
      continue;
    }
    if (line.includes('|') && i + 1 < lines.length && isTableSep(lines[i + 1])) {
      const t = el('table'), head = cells(line); i += 2;
      const tr = el('tr'); for (const c of head) { const th = el('th'); inline(c, th); tr.append(th); } t.append(tr);
      while (i < lines.length && lines[i].includes('|') && lines[i].trim()) {
        const r = el('tr'); for (const c of cells(lines[i++])) { const td = el('td'); inline(c, td); r.append(td); } t.append(r);
      }
      const w = el('div', 'tablewrap'); w.append(t); out.append(w);
      continue;
    }
    const li = /^( *)([-*+]|\d+[.)])\s+(.*)$/.exec(line);
    if (li) {
      const ordered = /\d/.test(li[2]), list = el(ordered ? 'ol' : 'ul');
      if (ordered && li[2] !== '1.' && li[2] !== '1)') list.start = parseInt(li[2], 10);
      while (i < lines.length) {
        const m = /^( *)([-*+]|\d+[.)])\s+(.*)$/.exec(lines[i]);
        if (!m || /\d/.test(m[2]) !== ordered) break;
        const item = [m[3]]; i++;
        while (i < lines.length && lines[i].trim() && /^ {2,}\S/.test(lines[i]) ) { // lines that belong to the item (indented), also a nested list
          item.push(lines[i++].replace(/^ {2,4}/, ''));
        }
        const e = el('li');
        const nested = item.findIndex((l, k) => k > 0 && /^( *)([-*+]|\d+[.)])\s+/.test(l));
        inline(item.slice(0, nested < 0 ? item.length : nested).join(' '), e);
        if (nested >= 0) e.append(renderMarkdown(item.slice(nested).join('\n')));
        list.append(e);
        while (i < lines.length && !lines[i].trim() && i + 1 < lines.length && /^( *)([-*+]|\d+[.)])\s+/.test(lines[i + 1])) i++; // a blank line between items
      }
      out.append(list);
      continue;
    }
    const para = [];
    while (i < lines.length && lines[i].trim() && !/^ {0,3}(#{1,6}\s|>|`{3,}|~{3,})/.test(lines[i]) && !isRule(lines[i]) && !(para.length && /^( *)([-*+]|\d+[.)])\s+/.test(lines[i]))) para.push(lines[i++]);
    const p = el('p');
    para.forEach((l, k) => { if (k) p.append(el('br')); inline(l.trim(), p); });
    out.append(p);
  }
  return out;
}
