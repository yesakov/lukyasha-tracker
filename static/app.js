(function () {
  const root = document.documentElement;
  const key = 'theme';
  const mq = window.matchMedia('(prefers-color-scheme: light)');

  const applyPrefFlag = () => {
    if (!root.getAttribute('data-theme')) {
      if (mq.matches) root.setAttribute('data-prefers-light', '');
      else root.removeAttribute('data-prefers-light');
    }
  };

  const currentTheme = () => root.getAttribute('data-theme') || (mq.matches ? 'light' : 'dark');

  const setTheme = (t) => {
    if (t === 'light' || t === 'dark') root.setAttribute('data-theme', t);
    else root.removeAttribute('data-theme');
    try { localStorage.setItem(key, t); } catch (_) {}
    updateIcons();
    applyPrefFlag();
  };

  const updateIcons = () => {
    const cur = currentTheme();
    document.querySelectorAll('.theme-toggle i').forEach((i) => {
      i.className = cur === 'dark' ? 'bi bi-sun' : 'bi bi-moon-stars';
    });
  };

  const bindToggle = () => {
    document.querySelectorAll('.theme-toggle').forEach((btn) => {
      btn.addEventListener('click', () => {
        const cur = currentTheme();
        setTheme(cur === 'dark' ? 'light' : 'dark');
      });
    });
  };

  document.addEventListener('DOMContentLoaded', () => {
    try {
      const stored = localStorage.getItem(key);
      if (stored === 'light' || stored === 'dark') root.setAttribute('data-theme', stored);
    } catch (_) {}
    applyPrefFlag();
    updateIcons();
    bindToggle();
    try { mq.addEventListener('change', () => { applyPrefFlag(); updateIcons(); }); } catch (_) {}

    // Toast helpers
    const toastEl = document.getElementById('app-toast');
    const showToast = (msg) => {
      if (!toastEl) return;
      toastEl.textContent = msg;
      toastEl.classList.add('show');
      setTimeout(() => toastEl.classList.remove('show'), 2000);
    };

    // htmx custom events
    document.body.addEventListener('game-removed', () => showToast('Game deleted'));
    document.body.addEventListener('toast', (e) => {
      const msg = (e && e.detail) ? e.detail : 'Done';
      showToast(msg);
    });
  });
})();
