// Risk Dashboard wiring (frontend scaffolding)
(function () {
  const networkLevelEl = document.getElementById('risk-network-level');
  const networkMetaEl = document.getElementById('risk-network-meta');
  const walletLevelEl = document.getElementById('risk-wallet-level');
  const walletMetaEl = document.getElementById('risk-wallet-meta');
  const vaultExposureEl = document.getElementById('risk-vault-exposure');
  const vaultMetaEl = document.getElementById('risk-vault-meta');
  const insLevelEl = document.getElementById('risk-insurance-level');
  const insMetaEl = document.getElementById('risk-insurance-meta');
  const vaultTableBody = document.getElementById('risk-vault-table-body');
  const instrTableBody = document.getElementById('risk-instrument-table-body');
  if (!networkLevelEl) return;

  function loadTablesPlaceholder() {
    if (vaultTableBody) {
      vaultTableBody.innerHTML = '<tr><td colspan=\"6\" class=\"risk-table-empty\">Risk engine will populate vaults when available.</td></tr>';
    }
    if (instrTableBody) {
      instrTableBody.innerHTML = '<tr><td colspan=\"6\" class=\"risk-table-empty\">Risk engine will populate instruments when available.</td></tr>';
    }
  }

  async function loadGlobal() {
    try {
      const res = await fetch('/api/risk_summary.php');
      if (!res.ok) throw new Error('HTTP ' + res.status);
      const data = await res.json();
      const lvl = data.network_risk_level || 'GREEN';
      networkLevelEl.textContent = lvl;
      const lev = data.global_leverage_factor != null ? data.global_leverage_factor : '–';
      const util = data.corridor_utilization_pct != null ? Math.round(data.corridor_utilization_pct * 100) : 0;
      const height = data.chain_height != null ? data.chain_height : '–';
      const mpLen = data.mempool_len != null ? data.mempool_len : 0;
      networkMetaEl.textContent = 'Leverage ' + lev + '×, corridor ' + util + '% · height ' + height + ' · mempool ' + mpLen;
      const vt = data.vault_exposure_total != null ? data.vault_exposure_total : 0;
      vaultExposureEl.textContent = vt.toLocaleString(undefined, { maximumFractionDigits: 2 }) + ' GRC';
      vaultMetaEl.textContent = 'Backs margin & PoP scoring across the network.';
      insLevelEl.textContent = data.stabilizer_mode || 'NORMAL';
      const ins = data.insurance_fund != null ? data.insurance_fund : 0;
      insMetaEl.textContent = 'Insurance fund ' + ins.toLocaleString(undefined, { maximumFractionDigits: 2 }) + ' GRC';
    } catch (e) {
      console.warn('[RiskDashboard] failed to load global summary', e);
      networkLevelEl.textContent = '–';
      networkMetaEl.textContent = 'Risk engine not available (DevNet offline?).';
    }
  }

  async function loadWallet() {
    try {
      if (!window.reserveWalletApi || !reserveWalletApi.getCurrentWallet) {
        walletLevelEl.textContent = '–';
        walletMetaEl.textContent = 'Connect a wallet to see risk.';
        return;
      }
      const w = await reserveWalletApi.getCurrentWallet();
      if (!w) {
        walletLevelEl.textContent = '–';
        walletMetaEl.textContent = 'Connect a wallet to see risk.';
        return;
      }
      const balances = await reserveWalletApi.getBalances();
      const total = balances.total != null ? balances.total : 0;
      let lvl = 'GREEN';
      if (total > 100000) lvl = 'AMBER';
      if (total > 250000) lvl = 'RED';
      walletLevelEl.textContent = lvl;
      walletMetaEl.textContent = 'Approximate risk bucket based on devnet balance only.';
    } catch (e) {
      console.warn('[RiskDashboard] failed to compute wallet risk', e);
      walletLevelEl.textContent = '–';
      walletMetaEl.textContent = 'Error calculating wallet risk.';
    }
  }

  loadTablesPlaceholder();
  loadGlobal();
  loadWallet();
})();