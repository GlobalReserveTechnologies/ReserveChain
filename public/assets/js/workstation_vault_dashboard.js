// Vault Dashboard logic
(function() {
  const totalEl       = document.getElementById('vault-total-assets');
  const countEl       = document.getElementById('vault-count');
  const approvalsEl   = document.getElementById('vault-pending-approvals');
  const netFlowEl     = document.getElementById('vault-net-flow');
  const tableBodyEl   = document.getElementById('vault-table-body');
  const approvalsList = document.getElementById('vault-approvals-list');
  const activityList  = document.getElementById('vault-activity-list');

  const btnNew       = document.getElementById('btn-vault-new');
  const btnTransfer  = document.getElementById('btn-vault-transfer');
  const btnDeposit   = document.getElementById('btn-vault-deposit');
  const btnWithdraw  = document.getElementById('btn-vault-withdraw');

  if (!tableBodyEl || !window.VaultAPI) return;

  async function loadVaultDashboard() {
    try {
      const vaults = await VaultAPI.listVaults();

      // Determine current NAV so we can present vaults in client currency (USD-eq).
      let nav = 1.0;
      if (typeof VaultAPI.lastNav === 'number' && !Number.isNaN(VaultAPI.lastNav)) {
        nav = VaultAPI.lastNav;
      } else if (window.fetch) {
        try {
          const res = await fetch('/api/valuation/latest', { credentials: 'include' });
          if (res.ok) {
            const data = await res.json();
            if (typeof data.nav === 'number') nav = data.nav;
          }
        } catch (e) {
          console.warn('[VaultDashboard] failed to fetch NAV, defaulting to 1.0', e);
        }
      }

      let totalGrc = 0;
      let totalUsd = 0;
      vaults.forEach(v => {
        const bal = parseFloat((v.balance && v.balance.GRC) || 0);
        if (!Number.isNaN(bal)) {
          totalGrc += bal;
          const usd = (typeof v.balance_usd === 'number') ? v.balance_usd : bal * nav;
          totalUsd += usd;
        }
      });

      if (totalEl) {
        const usdText = totalUsd.toLocaleString(undefined, { maximumFractionDigits: 2 });
        const grcText = totalGrc.toLocaleString(undefined, { maximumFractionDigits: 4 });
        totalEl.textContent = usdText + ' USD-eq (' + grcText + ' GRC)';
      }
      if (countEl) countEl.textContent = vaults.length.toString();
      if (approvalsEl) approvalsEl.textContent = '0';
      if (netFlowEl) netFlowEl.textContent = '–';

      tableBodyEl.innerHTML = '';

      if (!vaults.length) {
        const tr = document.createElement('tr');
        const td = document.createElement('td');
        td.colSpan = 6;
        td.textContent = 'No vaults yet. Create your first private vault to start segmenting client or strategy balances.';
        td.style.color = '#9aa7c0';
        tr.appendChild(td);
        tableBodyEl.appendChild(tr);
        return;
      }

      const visLabels = { A: 'Visible', B: 'Hidden Toggle', C: 'Passphrase', D: 'Audit Only' };
      const templateLabels = {
        client: 'Client vault',
        strategy: 'Strategy vault',
        treasury: 'Treasury bucket',
        custom: 'Custom vault',
      };

      vaults.forEach(v => {
        const tr = document.createElement('tr');
        const bal = parseFloat((v.balance && v.balance.GRC) || 0);
        const usd = (typeof v.balance_usd === 'number') ? v.balance_usd : bal * nav;
        const durationTier = (v.yield_policy && v.yield_policy.duration_tier) || 'medium';
        const durationLabelMap = {
          short: 'Short',
          medium: 'Medium',
          long: 'Long',
        };
        const durationLabel = durationLabelMap[durationTier] || durationTier;

        const usdText = usd.toLocaleString(undefined, { maximumFractionDigits: 2 });
        const grcText = bal.toLocaleString(undefined, { maximumFractionDigits: 4 });
        const templateKey = v.template || 'custom';
        const templateLabel = templateLabels[templateKey] || 'Custom vault';

        tr.innerHTML = `
          <td>${v.label}<br><span class="vault-template-sub">${templateLabel}</span></td>
          <td>${v.type === 'multisig' ? 'Multi-sig' : 'Single'}</td>
          <td>${usdText} USD-eq<br><span class="vault-grc-sub">${grcText} GRC</span></td>
          <td>${v.pnl_settlement_mode || 'source'}</td>
          <td>${durationLabel}</td>
          <td>${visLabels[v.visibility_mode] || v.visibility_mode || 'A'}</td>
          <td>
            <button class="link-button js-vault-quick-transfer" data-vault="${v.vault_id}">Transfer</button>
            <button class="link-button js-vault-quick-deposit" data-vault="${v.vault_id}">Deposit</button>
            <button class="link-button js-vault-quick-withdraw" data-vault="${v.vault_id}">Withdraw</button>
          </td>
        `;
        tableBodyEl.appendChild(tr);
      });

      approvalsList.textContent = 'Policy engine not wired yet.';
      activityList.textContent  = 'Activity log not wired yet.';
    } catch (e) {
      console.error('Failed to load vault dashboard', e);
      if (totalEl) totalEl.textContent = 'Error';
    }
  }

  // expose globally so transfer modal can refresh
  window.loadVaultDashboard = loadVaultDashboard;

  // load on DOM ready
  document.addEventListener('DOMContentLoaded', () => {
    loadVaultDashboard();
  });

  // wire header buttons (very simple for now)
  btnNew && btnNew.addEventListener('click', () => {
    const label = prompt('Vault label?');
    if (!label) return;

    const choice = (prompt('Vault template?\n1 = Client vault\n2 = Strategy vault\n3 = Treasury bucket', '1') || '').trim();
    let template = 'client';
    let type = 'single';
    let visibility = 'A';
    let pnlMode = 'source';

    if (choice === '2') {
      template = 'strategy';
      type = 'single';
      visibility = 'B'; // slightly more private by default
      pnlMode = 'target';
    } else if (choice === '3') {
      template = 'treasury';
      type = 'multisig';
      visibility = 'D'; // audit-only
      pnlMode = 'target';
    } else {
      template = 'client';
      type = 'single';
      visibility = 'A';
      pnlMode = 'source';
    }

    VaultAPI.createVault({
      label,
      type,
      visibility_mode: visibility,
      pnl_settlement_mode: pnlMode,
      template,
    }).then(() => loadVaultDashboard())
      .catch(err => alert('Failed to create vault: ' + err.message));
  });

  // Transfer button handled by modal script based on id=btn-vault-transfer
})();
