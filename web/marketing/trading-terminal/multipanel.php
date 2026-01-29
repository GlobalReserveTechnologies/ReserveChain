<?php ?>
<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <title>ReserveChain — Trading Multi-Panel</title>
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <link rel="stylesheet" href="/trading-terminal/assets/css/app.css">
  <link rel="stylesheet" href="/trading-terminal/assets/css/terminal.css">
  <style>
    body.multi-root {
      margin: 0;
      height: 100vh;
      display: flex;
      flex-direction: column;
      background: radial-gradient(circle at top left, #020617, #000);
      color: #e2e8f0;
      font-family: system-ui, -apple-system, BlinkMacSystemFont, "SF Pro Text", sans-serif;
    }
    .multi-bar {
      height: 40px;
      padding: 0 12px;
      display: flex;
      align-items: center;
      justify-content: space-between;
      border-bottom: 1px solid rgba(148, 163, 184, 0.24);
      background: linear-gradient(90deg, rgba(15,23,42,0.98), rgba(15,23,42,0.7));
      font-size: 13px;
    }
    .multi-bar .title {
      font-weight: 500;
      letter-spacing: 0.04em;
      text-transform: uppercase;
      color: #94a3b8;
    }
    .multi-bar .subtitle {
      margin-left: 8px;
      color: #38bdf8;
      font-weight: 500;
    }
    .multi-bar-buttons button {
      margin-left: 8px;
      padding: 4px 10px;
      font-size: 12px;
      border-radius: 999px;
      border: 1px solid rgba(148, 163, 184, 0.5);
      background: rgba(15,23,42,0.9);
      color: #e2e8f0;
      cursor: pointer;
    }
    .multi-bar-buttons button:hover {
      border-color: #38bdf8;
      box-shadow: 0 0 10px rgba(56,189,248,0.4);
    }
    .multi-grid {
      flex: 1;
      display: grid;
      grid-template-columns: 2fr 1fr;
      grid-template-rows: 1fr 1fr;
      gap: 8px;
      padding: 8px;
      box-sizing: border-box;
      background: radial-gradient(circle at top left, #020617, #000);
    }
    .multi-panel {
      border-radius: 8px;
      border: 1px solid rgba(30, 64, 175, 0.7);
      overflow: hidden;
      background: #020617;
    }
    .multi-panel iframe {
      border: none;
      width: 100%;
      height: 100%;
    }
.multi-panel-label {
  position: absolute;
  top: 8px;
  left: 12px;
  right: 12px;
  z-index: 2;
  padding: 2px 8px;
  border-radius: 999px;
  background: rgba(15,23,42,0.86);
  color: #e2e8f0;
  font-size: 11px;
  letter-spacing: 0.06em;
  text-transform: uppercase;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
}
.multi-panel-label-text {
  white-space: nowrap;
}
.multi-panel-select {
  min-width: 160px;
  font-size: 11px;
  border-radius: 999px;
  border: 1px solid rgba(148,163,184,0.6);
  background: rgba(15,23,42,0.9);
  color: #e2e8f0;
  padding: 2px 6px;
}
.multi-panel-select:focus {
  outline: none;
  border-color: #38bdf8;
  box-shadow: 0 0 8px rgba(56,189,248,0.4);
}
    .multi-panel-wrapper {
      position: relative;
      width: 100%;
      height: 100%;
    }
  </style>
</head>
<body class="multi-root">
  <div class="multi-bar">
    <div>
      <span class="title">ReserveChain Multi-Panel</span>
      <span class="subtitle">Terminal + Treasury + Vault</span>
    </div>
    <div class="multi-bar-buttons">
      <button onclick="location.reload()">Reload Layout</button>
      <button onclick="window.close()">Close Window</button>
    </div>
  </div>
  <div class="multi-grid">
    <div class="multi-panel" style="grid-row: 1 / span 2;">
      <div class="multi-panel-wrapper">
        <div class="multi-panel-label">Trading Terminal</div>
        <iframe src="/trading-terminal/trading_terminal.php"></iframe>
      </div>
    </div>
    <div class="multi-panel">
      <div class="multi-panel-wrapper">
        <div class="multi-panel-label">Treasury &amp; Reserves</div>
        <iframe src="/workstation/index.php?panel=treasury-overview"></iframe>
      </div>
    </div>
    <div class="multi-panel">
      <div class="multi-panel-wrapper">
        <div class="multi-panel-label">Vault Dashboard</div>
        <iframe src="/workstation/index.php?panel=vault-dashboard"></iframe>
      </div>
    </div>
  </div>
  <script>
    (function () {
      const MODULES = {
        treasury: {
          label: 'Treasury & Reserves',
          url: '/workstation/index.php?panel=treasury-overview'
        },
        vaults: {
          label: 'Vault Dashboard',
          url: '/workstation/index.php?panel=vault-dashboard'
        },
        explorer: {
          label: 'Block Explorer',
          url: '/workstation/index.php?panel=network-explorer'
        },
        positions: {
          label: 'Portfolio / Positions',
          url: '/workstation/index.php?panel=defi-portfolio'
        },
        orders: {
          label: 'Orders & Fills',
          url: '/workstation/index.php?panel=defi-orders-fills'
        },
        risk: {
          label: 'Risk Dashboard',
          url: '/workstation/index.php?panel=network-risk-dashboard'
        }
      };

      const STORAGE_KEY = 'reservechain_multipanel_layout_v1';

      function loadLayout() {
        try {
          const raw = window.localStorage.getItem(STORAGE_KEY);
          if (!raw) return null;
          return JSON.parse(raw);
        } catch (e) {
          console.warn('Failed to load multipanel layout', e);
          return null;
        }
      }

      function saveLayout(layout) {
        try {
          window.localStorage.setItem(STORAGE_KEY, JSON.stringify(layout));
        } catch (e) {
          console.warn('Failed to save multipanel layout', e);
        }
      }

      function applyModuleToSlot(slotId, moduleId) {
        const cfg = MODULES[moduleId];
        const frame = document.querySelector('iframe[data-slot-frame="' + slotId + '"]');
        const label = document.querySelector('[data-slot-id="' + slotId + '"] .multi-panel-label-text');
        const select = document.querySelector('select[data-slot-select="' + slotId + '"]');
        if (!cfg || !frame || !label) return;
        frame.src = cfg.url;
        label.textContent = cfg.label;
        if (select) {
          select.value = moduleId;
        }
      }

      document.addEventListener('DOMContentLoaded', function () {
        const layout = loadLayout() || {
          'right-top': 'treasury',
          'right-bottom': 'vaults'
        };

        Object.keys(layout).forEach(slotId => {
          if (MODULES[layout[slotId]]) {
            applyModuleToSlot(slotId, layout[slotId]);
          }
        });

        document.querySelectorAll('.multi-panel-select').forEach(select => {
          select.addEventListener('change', function () {
            const slotId = this.getAttribute('data-slot-select');
            const moduleId = this.value;
            applyModuleToSlot(slotId, moduleId);
            const current = loadLayout() || {};
            current[slotId] = moduleId;
            saveLayout(current);
          });
        });
      });
    })();
  </script>
</body>
</html>
