<?php ?>
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <title>ReserveChain Trading Terminal</title>
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <link rel="stylesheet" href="/trading-terminal/assets/css/app.css">
</head>
<body>
<div class="app-shell">
    <aside class="sidebar" id="sidebar">
      <div class="sidebar-glow"></div>
      <div class="sidebar-toggle" id="sidebar-toggle">
        <span class="sidebar-toggle-icon">⟨⟩</span>
      </div>

      <div class="brand">
        <div class="brand-logo">
          <div class="brand-logo-inner">WT</div>
        </div>
        <div class="brand-text">
          <div class="brand-name">ReserveChain Trading Terminal</div>
          <div class="brand-subtitle">AI-ASSISTED CRYPTO DESK</div>
        </div>
      </div>

      <div class="sidebar-block pair-selector">
        <label>Active Market</label>
        <div class="pair-row">
          <div class="pair-main">
            <div class="pair-symbol" id="pair-symbol">WEB / USDT</div>
            <div class="pair-price" id="pair-price">$2.50</div>
          </div>
          <div class="pair-main" style="align-items: flex-end;">
            <span class="pair-change" id="pair-change">+0.00%</span>
            <select class="pair-dropdown" id="pair-select">
              <option value="WEBUSDT" selected>WEB / USDT</option>
              <option value="BTCUSDT">BTC / USDT</option>
              <option value="ETHUSDT">ETH / USDT</option>
              <option value="SOLUSDT">SOL / USDT</option>
            </select>
          </div>
        </div>
      </div>

      <div>
        <div class="sidebar-section-title">Workspace</div>
        <nav class="sidebar-nav">
          <div class="nav-item">
            <span class="icon">🏠</span><span class="label">Overview</span>
          </div>
          <div class="nav-item active">
            <span class="icon">📈</span><span class="label">Trading Terminal</span>
          </div>
          <div class="nav-item">
            <span class="icon">📊</span><span class="label">Portfolio</span>
          </div>
          <div class="nav-item">
            <span class="icon">⚙️</span><span class="label">Risk & Settings</span>
          </div>
        </nav>
      </div>

      <div class="sidebar-footer">
        <div class="latency-pill">
          <span class="latency-dot"></span>
          <span>Latency: <span id="latency-value">11</span> ms</span>
        </div>
        <div>Session: <span id="session-label">ReserveChain DevNet</span></div>
      </div>
    </aside>

    <main class="main">
      <div class="views-container">
        <section class="view view-active" id="view-standard">
          <!-- reuse existing standard/pro content via iframe pointing to old defi index? -->
          <iframe src="/defi/standard-pro.html" frameborder="0" class="view-frame"></iframe>
        </section>

        <section class="view" id="view-pro">
          <iframe src="/defi/standard-pro.html#pro" frameborder="0" class="view-frame"></iframe>
        </section>

        <section class="view" id="view-terminal1">
          <iframe src="/defi/terminal-v1.html" frameborder="0" class="view-frame"></iframe>
        </section>

        <section class="view" id="view-terminal2">
          <iframe src="/defi/terminal-v2.html" frameborder="0" class="view-frame"></iframe>
        </section>

        <section class="view" id="view-lending">
          <iframe src="/defi/lending.html" frameborder="0" class="view-frame"></iframe>
        </section>

        <section class="view" id="view-accounts">
          <iframe src="/defi/accounts.html" frameborder="0" class="view-frame"></iframe>
        </section>

        <section class="view" id="view-activity">
          <iframe src="/defi/activity.html" frameborder="0" class="view-frame"></iframe>
        </section>

        <section class="view" id="view-explorer">
          <iframe src="/explorer.html" frameborder="0" class="view-frame"></iframe>
        </section>

        <section class="view" id="view-health">
          <iframe src="/treasury.html" frameborder="0" class="view-frame"></iframe>
        </section>
        <section class="view" id="view-portfolio">
          <iframe src="/defi/portfolio.html" frameborder="0" class="view-frame"></iframe>
        </section>
        <section class="view" id="view-wallet">
          <iframe src="/defi/wallet.html" frameborder="0" class="view-frame"></iframe>
        </section>
        <section class="view" id="view-orders">
          <iframe src="/defi/orders.html" frameborder="0" class="view-frame"></iframe>
        </section>
        <section class="view" id="view-transfers">
          <iframe src="/defi/transfers.html" frameborder="0" class="view-frame"></iframe>
        </section>
        <section class="view" id="view-funding">
          <iframe src="/defi/funding.html" frameborder="0" class="view-frame"></iframe>
        </section>
        <section class="view" id="view-staking">
          <iframe src="/defi/staking.html" frameborder="0" class="view-frame"></iframe>
        </section>
        <section class="view" id="view-liquidity">
          <iframe src="/defi/liquidity.html" frameborder="0" class="view-frame"></iframe>
        </section>
        <section class="view" id="view-shielded-vault">
          <iframe src="/defi/shielded-vault.html" frameborder="0" class="view-frame"></iframe>
        </section>
        <section class="view" id="view-shielded-amm">
          <iframe src="/defi/shielded-amm.html" frameborder="0" class="view-frame"></iframe>
        </section>
        <section class="view" id="view-shielded-lending">
          <iframe src="/defi/shielded-lending.html" frameborder="0" class="view-frame"></iframe>
        </section>
        <section class="view" id="view-audit-proofs">
          <iframe src="/defi/audit-proofs.html" frameborder="0" class="view-frame"></iframe>
        </section>
        <section class="view" id="view-treasury-defi">
          <iframe src="/defi/treasury-defi.html" frameborder="0" class="view-frame"></iframe>
        </section>
        <section class="view" id="view-epochs">
          <iframe src="/defi/epochs.html" frameborder="0" class="view-frame"></iframe>
        </section>
        <section class="view" id="view-governance-defi">
          <iframe src="/defi/governance-defi.html" frameborder="0" class="view-frame"></iframe>
        </section>
        <section class="view" id="view-analytics">
          <iframe src="/defi/analytics.html" frameborder="0" class="view-frame"></iframe>
        </section>
        <section class="view" id="view-risk">
          <iframe src="/defi/risk.html" frameborder="0" class="view-frame"></iframe>
        </section>
        <section class="view" id="view-benchmarks">
          <iframe src="/defi/benchmarks.html" frameborder="0" class="view-frame"></iframe>
        </section>
        <section class="view" id="view-account">
          <iframe src="/defi/account.html" frameborder="0" class="view-frame"></iframe>
        </section>
        <section class="view" id="view-api-keys">
          <iframe src="/defi/api-keys.html" frameborder="0" class="view-frame"></iframe>
        </section>
        <section class="view" id="view-security">
          <iframe src="/defi/security.html" frameborder="0" class="view-frame"></iframe>
        </section>        
      </div>
    </main>
  </div>

  <script>
    // Sidebar collapse
    const sidebar = document.querySelector('.sidebar');
    const sidebarToggle = document.getElementById('sidebar-toggle');
    if (sidebar && sidebarToggle) {
      sidebarToggle.addEventListener('click', () => {
        sidebar.classList.toggle('sidebar-expanded');
      });
    }

    // View routing
    const navItems = document.querySelectorAll('.nav-item[data-view]');
    const views = {
      standard: document.getElementById('view-standard'),
      terminal2: document.getElementById('view-terminal2'),
      lending: document.getElementById('view-lending'),
      portfolio: document.getElementById('view-portfolio'),
      wallet: document.getElementById('view-wallet'),
      orders: document.getElementById('view-orders'),
      transfers: document.getElementById('view-transfers'),
      funding: document.getElementById('view-funding'),
      staking: document.getElementById('view-staking'),
      liquidity: document.getElementById('view-liquidity'),
      activity: document.getElementById('view-activity'),
      explorer: document.getElementById('view-explorer'),
      'treasury-defi': document.getElementById('view-treasury-defi'),
      epochs: document.getElementById('view-epochs'),
      'governance-defi': document.getElementById('view-governance-defi'),
      analytics: document.getElementById('view-analytics'),
      risk: document.getElementById('view-risk'),
      benchmarks: document.getElementById('view-benchmarks'),
      'shielded-vault': document.getElementById('view-shielded-vault'),
      'shielded-amm': document.getElementById('view-shielded-amm'),
      'shielded-lending': document.getElementById('view-shielded-lending'),
      'audit-proofs': document.getElementById('view-audit-proofs'),
      health: document.getElementById('view-health'),
      account: document.getElementById('view-account'),
      'api-keys': document.getElementById('view-api-keys'),
      security: document.getElementById('view-security'),
    };

    function setActiveView(key) {
      Object.values(views).forEach(v => {
        if (!v) return;
        if (v.id === 'view-' + key) {
          v.classList.add('view-active');
        } else {
          v.classList.remove('view-active');
        }
      });
      navItems.forEach(item => {
        if (item.getAttribute('data-view') === key) {
          item.classList.add('active');
        } else {
          item.classList.remove('active');
        }
      });
    }

    navItems.forEach(item => {
      item.addEventListener('click', () => {
        const key = item.getAttribute('data-view');
        setActiveView(key);
      });
    });

    // Node status polling from existing PHP APIs
    async function refreshStatus() {
      try {
        const res = await fetch('/api/status.php');
        if (!res.ok) throw new Error('status fail');
        const status = await res.json();
        const pill = document.getElementById('statusPill');
        const text = document.getElementById('statusPillText');
        if (status && typeof status.height !== 'undefined') {
          pill.classList.remove('offline');
          text.textContent = 'Online · h=' + status.height;
        } else {
          pill.classList.add('offline');
          text.textContent = 'Offline';
        }
      } catch (e) {
        const pill = document.getElementById('statusPill');
        const text = document.getElementById('statusPillText');
        pill.classList.add('offline');
        text.textContent = 'Offline';
      }
    }
    refreshStatus();
    setInterval(refreshStatus, 4000);
  </script>
<script src="/assets/js/ui_notifier.js"></script>
<script src="/assets/js/api_client.js"></script>
<script src="/assets/js/vault_client.js"></script>
<script src="/assets/js/wallet_keystore.js"></script>
<script src="/assets/js/session_manager.js"></script>
<script src="/trading-terminal/assets/js/terminal_margin.js"></script>
</body>
</html>
