// workstation_portfolio.js
document.addEventListener('DOMContentLoaded', function () {
  const panel = document.querySelector('.ws-panel[data-panel="defi-portfolio"]');
  if (!panel) return;

  const refreshBtn = document.getElementById('btn-portfolio-refresh');
  if (refreshBtn) {
    refreshBtn.addEventListener('click', loadPortfolio);
  }

  // Load once when panel is first shown
  const observer = new MutationObserver(() => {
    if (panel.style.display !== 'none') {
      loadPortfolio();
      observer.disconnect();
    }
  });
  observer.observe(panel, { attributes: true, attributeFilter: ['style'] });
});

async function loadPortfolio() {
  try {
    const [vaults] = await Promise.all([
      safeFetchJSON('/api/vault.php?action=list')
      // Later we can add: safeFetchJSON('/api/treasury.php')
    ]);

    const summary = buildPortfolioSummaryFromVaults(vaults || []);
    renderPortfolioSummary(summary);
    renderPortfolioPositions([]); // wired once trading engine positions are in place
    renderPortfolioBalances(summary.balances || []);
    renderPortfolioActivity([]);
  } catch (e) {
    console.error('Failed to load portfolio', e);
    if (window.UINotifier) {
      UINotifier.error('Failed to load portfolio: ' + (e.message || e), e);
    } else {
      alert('Failed to load portfolio: ' + (e.message || e));
    }
  }
}

function buildPortfolioSummaryFromVaults(vaults) {
  let total = 0;
  const balances = [];

  (vaults || []).forEach(v => {
    const label = v.label || v.vault_id || 'Vault';
    const bal = v.balance && typeof v.balance.GRC === 'number' ? v.balance.GRC : 0;
    total += bal;
    balances.push({
      source: 'vault:' + label,
      asset: 'GRC',
      amount: bal,
    });
  });

  return {
    net_value_grc: Math.round(total * 1e8),
    onchain_balance_grc: Math.round(total * 1e8),
    open_pnl_grc: 0,
    position_count: 0,
    balances,
  };
}

function renderPortfolioSummary(s) {
  const netEl = document.getElementById('pf-net-value');
  const onchainEl = document.getElementById('pf-onchain-balance');
  const pnlEl = document.getElementById('pf-open-pnl');
  const pnlSubEl = document.getElementById('pf-open-pnl-sub');

  const net = (s.net_value_grc ?? 0) / 1e8;
  const onchain = (s.onchain_balance_grc ?? 0) / 1e8;
  const pnl = (s.open_pnl_grc ?? 0) / 1e8;

  if (netEl) netEl.textContent = net.toFixed(4) + ' GRC';
  if (onchainEl) onchainEl.textContent = onchain.toFixed(4) + ' GRC';
  if (pnlEl) {
    pnlEl.textContent = (pnl >= 0 ? '+' : '') + pnl.toFixed(4) + ' GRC';
    pnlEl.style.color = pnl >= 0 ? '#22c55e' : '#ef4444';
  }
  if (pnlSubEl) {
    pnlSubEl.textContent = 'Unrealized PnL across ' + (s.position_count ?? 0) + ' open positions';
  }
}

function renderPortfolioPositions(list) {
  const tbody = document.getElementById('pf-positions-body');
  const countEl = document.getElementById('pf-positions-count');
  if (!tbody) return;

  tbody.innerHTML = '';

  if (!list || !list.length) {
    const tr = document.createElement('tr');
    const td = document.createElement('td');
    td.colSpan = 7;
    td.textContent = 'No open positions.';
    tr.appendChild(td);
    tbody.appendChild(tr);
    if (countEl) countEl.textContent = '0 open';
    return;
  }

  list.forEach(p => {
    const tr = document.createElement('tr');
    const pnl = (p.pnl_grc ?? 0) / 1e8;
    const side = (p.side || '').toUpperCase();
    const liq = p.liq_price ?? null;
    tr.innerHTML = `
      <td>${p.symbol || ''}</td>
      <td class="mono ${side === 'LONG' ? 'text-long' : 'text-short'}">${side}</td>
      <td>${(p.size ?? 0).toFixed ? p.size.toFixed(4) : p.size}</td>
      <td>${(p.entry_price ?? 0).toFixed ? p.entry_price.toFixed(4) : p.entry_price}</td>
      <td>${(p.mark_price ?? 0).toFixed ? p.mark_price.toFixed(4) : p.mark_price}</td>
      <td>${liq ? (liq.toFixed ? liq.toFixed(4) : liq) : '—'}</td>
      <td class="${pnl >= 0 ? 'text-long' : 'text-short'}">${(pnl >= 0 ? '+' : '') + pnl.toFixed(4)} GRC</td>
    `;
    tbody.appendChild(tr);
  });

  if (countEl) countEl.textContent = list.length + ' open';
}

function renderPortfolioBalances(balances) {
  const tbody = document.getElementById('pf-balances-body');
  if (!tbody) return;
  tbody.innerHTML = '';

  if (!balances || !balances.length) {
    const tr = document.createElement('tr');
    const td = document.createElement('td');
    td.colSpan = 3;
    td.textContent = 'No balances.';
    tr.appendChild(td);
    tbody.appendChild(tr);
    return;
  }

  balances.forEach(b => {
    const tr = document.createElement('tr');
    tr.innerHTML = `
      <td>${b.source || ''}</td>
      <td>${b.asset || 'GRC'}</td>
      <td>${(b.amount ?? 0).toFixed ? b.amount.toFixed(4) : b.amount}</td>
    `;
    tbody.appendChild(tr);
  });
}

function renderPortfolioActivity(items) {
  const tbody = document.getElementById('pf-activity-body');
  if (!tbody) return;
  tbody.innerHTML = '';

  if (!items || !items.length) {
    const tr = document.createElement('tr');
    const td = document.createElement('td');
    td.colSpan = 4;
    td.textContent = 'No recent activity.';
    tr.appendChild(td);
    tbody.appendChild(tr);
    return;
  }

  items.forEach(ev => {
    const tr = document.createElement('tr');
    const delta = (ev.delta_grc ?? 0) / 1e8;
    tr.innerHTML = `
      <td>${ev.time || ''}</td>
      <td>${ev.type || ''}</td>
      <td>${ev.details || ''}</td>
      <td class="${delta >= 0 ? 'text-long' : 'text-short'}">
        ${(delta >= 0 ? '+' : '') + delta.toFixed(4)} GRC
      </td>
    `;
    tbody.appendChild(tr);
  });
}
