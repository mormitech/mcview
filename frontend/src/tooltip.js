// Tooltips: every element with a title shows it in a small dark bubble after a second (the system one is not used).
const DELAY = 1000;

export function initTips() {
  const tip = document.createElement('div');
  tip.id = 'tip';
  tip.setAttribute('role', 'tooltip');
  document.body.appendChild(tip);
  let timer = 0, from = null;

  const hide = () => {
    clearTimeout(timer);
    tip.classList.remove('on');
    if (from && from.dataset.tip) { from.title = from.dataset.tip; delete from.dataset.tip; } // (the title goes back, so that it can be changed or read later)
    from = null;
  };
  const show = (el) => {
    tip.textContent = el.dataset.tip;
    const r = el.getBoundingClientRect(), t = tip.getBoundingClientRect();
    const x = Math.max(8, Math.min(innerWidth - t.width - 8, r.left + r.width / 2 - t.width / 2));
    const below = r.bottom + 8 + t.height < innerHeight;
    tip.style.left = `${x}px`;
    tip.style.top = `${below ? r.bottom + 8 : Math.max(8, r.top - 8 - t.height)}px`;
    tip.classList.add('on');
  };

  document.addEventListener('pointerover', (ev) => {
    const el = ev.target.closest?.('[title],[data-tip]');
    if (el === from) return;
    hide();
    if (!el || ev.pointerType === 'touch') return;
    const text = el.title || el.dataset.tip;
    if (!text) return;
    if (!el.getAttribute('aria-label')) el.setAttribute('aria-label', text);
    el.dataset.tip = text;
    el.removeAttribute('title'); // (the system tooltip would come on top)
    from = el;
    timer = setTimeout(() => show(el), DELAY);
  });
  document.addEventListener('pointerout', (ev) => { if (from && !from.contains(ev.relatedTarget)) hide(); });
  for (const t of ['pointerdown', 'keydown', 'wheel', 'blur']) window.addEventListener(t, hide, true);
}
