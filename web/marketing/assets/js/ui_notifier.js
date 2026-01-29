
// ui_notifier.js
// Simple global notification + logger used by workstation and terminal.

(function () {
  const QUEUE_LIMIT = 5;
  let queue = [];

  function ensureContainer() {
    let c = document.getElementById('rc-toast-container');
    if (c) return c;
    c = document.createElement('div');
    c.id = 'rc-toast-container';
    c.style.position = 'fixed';
    c.style.right = '16px';
    c.style.bottom = '16px';
    c.style.zIndex = '99999';
    c.style.display = 'flex';
    c.style.flexDirection = 'column';
    c.style.gap = '8px';
    document.body.appendChild(c);
    return c;
  }

  function render() {
    const c = ensureContainer();
    c.innerHTML = '';
    queue.forEach((msg) => {
      const el = document.createElement('div');
      el.style.minWidth = '260px';
      el.style.maxWidth = '360px';
      el.style.padding = '8px 10px';
      el.style.borderRadius = '8px';
      el.style.fontSize = '13px';
      el.style.display = 'flex';
      el.style.justifyContent = 'space-between';
      el.style.alignItems = 'flex-start';
      el.style.gap = '8px';
      el.style.boxShadow = '0 12px 24px rgba(0,0,0,0.45)';
      el.style.border = '1px solid rgba(148,163,184,0.8)';
      el.style.background =
        msg.type === 'error'
          ? 'linear-gradient(135deg, #7f1d1d, #111827)'
          : msg.type === 'warn'
          ? 'linear-gradient(135deg, #92400e, #111827)'
          : 'linear-gradient(135deg, #0f766e, #020617)';
      el.style.color = '#e5e7eb';

      const text = document.createElement('div');
      text.textContent = msg.text;

      const closeBtn = document.createElement('button');
      closeBtn.textContent = '×';
      closeBtn.style.border = 'none';
      closeBtn.style.background = 'transparent';
      closeBtn.style.color = '#e5e7eb';
      closeBtn.style.cursor = 'pointer';
      closeBtn.style.fontSize = '15px';
      closeBtn.addEventListener('click', () => dismiss(msg.id));

      el.appendChild(text);
      el.appendChild(closeBtn);
      c.appendChild(el);
    });
  }

  function dismiss(id) {
    queue = queue.filter((m) => m.id !== id);
    render();
  }

  function push(type, text) {
    const id = Date.now() + ':' + Math.random().toString(16).slice(2);
    queue.push({ id, type, text });
    if (queue.length > QUEUE_LIMIT) queue.shift();
    render();
    setTimeout(() => {
      dismiss(id);
    }, 8000);
  }

  const UINotifier = {
    info(msg) {
      console.log('[INFO]', msg);
      push('info', msg);
    },
    warn(msg) {
      console.warn('[WARN]', msg);
      push('warn', msg);
    },
    error(msg, err) {
      console.error('[ERROR]', msg, err);
      push('error', msg);
    },
  };

  window.UINotifier = UINotifier;
})();
