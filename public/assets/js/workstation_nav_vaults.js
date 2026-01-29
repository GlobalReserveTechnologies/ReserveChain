// Workstation navigation for Vaults sidebar
(function() {
  const sidebar = document.querySelector('.ws-sidebar');
  const panels  = document.querySelectorAll('.ws-panel');
  if (!sidebar || !panels.length) return;

  const items   = sidebar.querySelectorAll('.ws-sidebar-item[data-panel]');
  const groups  = sidebar.querySelectorAll('.ws-sidebar-group[data-group]');

  function showPanel(panelId) {
    panels.forEach(p => {
      p.style.display = (p.dataset.panel === panelId) ? 'block' : 'none';
    });
    items.forEach(i => {
      i.classList.toggle('is-active', i.dataset.panel === panelId);
    });
  }

  sidebar.addEventListener('click', (e) => {
    const item = e.target.closest('.ws-sidebar-item[data-panel]');
    if (item) {
      showPanel(item.dataset.panel);
      return;
    }

    const groupBtn = e.target.closest('.ws-sidebar-group[data-group]');
    if (groupBtn) {
      const group = groupBtn.dataset.group;
      const submenu = sidebar.querySelector('.ws-sidebar-submenu[data-group="' + group + '"]');
      if (submenu) submenu.classList.toggle('is-collapsed');
    }
  });

  // allow small "View X" links to jump panels
  document.addEventListener('click', (e) => {
    const jump = e.target.closest('[data-panel-jump]');
    if (!jump) return;
    const target = jump.dataset.panelJump;
    const item = sidebar.querySelector('.ws-sidebar-item[data-panel="' + target + '"]');
    if (item) item.click();
  });

  
  // expose for other scripts to call
  window.WorkstationNav = { showPanel };

  // default / deep link support
  const params = new URLSearchParams(window.location.search);
  const panelParam = params.get('panel');
  const hashPanel = window.location.hash ? window.location.hash.replace('#', '') : null;
  const initialPanel = panelParam || hashPanel || 'vault-dashboard';
  showPanel(initialPanel);
})();
