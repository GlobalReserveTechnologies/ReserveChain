
// Reserve Monitor card — calls /api/valuation/latest and displays NAV state.
(function () {
  const totalEl   = document.getElementById('rm-total-reserve');
  const supplyEl  = document.getElementById('rm-supply-grc');
  const navEl     = document.getElementById('rm-nav');
  const pegEl     = document.getElementById('rm-peg-status');
  const detailEl  = document.getElementById('rm-reserve-detail');
  const corrEl    = document.getElementById('rm-corridor');
  if (!totalEl || !window.fetch) return;

  async function refresh() {
    try {
      const res = await fetch('/api/valuation/latest', { credentials: 'include' });
      if (!res.ok) {
        throw new Error('HTTP ' + res.status);
      }
      const data = await res.json();

      const reserve = data.reserve_usd != null ? data.reserve_usd : 0;
      const supply  = data.supply_grc  != null ? data.supply_grc  : 0;
      const nav     = data.nav         != null ? data.nav         : 1.0;
      const lower   = data.corridor_lower != null ? data.corridor_lower : 0;
      const upper   = data.corridor_upper != null ? data.corridor_upper : 0;
      const peg     = data.peg_status  || 'INSIDE';

      totalEl.textContent  = reserve.toLocaleString(undefined, { maximumFractionDigits: 2 }) + ' USD-eq';
      supplyEl.textContent = supply.toLocaleString(undefined, { maximumFractionDigits: 4 }) + ' GRC';
      navEl.textContent    = nav.toFixed(data.decimals != null ? data.decimals : 4);

      if (corrEl) {
        corrEl.textContent = 'Corridor ' + lower.toFixed(4) + ' – ' + upper.toFixed(4);
      }

      if (detailEl) {
        detailEl.textContent = 'Backed by multi-asset reserves (USDC/USDT/DAI/ETH/WBTC).';
      }

      if (pegEl) {
        pegEl.textContent = peg;
      }
    } catch (e) {
      console.warn('[ReserveMonitor] failed to load valuation', e);
      totalEl.textContent  = 'Error';
      supplyEl.textContent = 'Error';
      navEl.textContent    = 'Error';
      if (detailEl) detailEl.textContent = 'Could not reach node valuation API.';
      if (pegEl) pegEl.textContent = 'UNKNOWN';
      if (corrEl) corrEl.textContent = '–';
    }
  }

  // Simple one-shot refresh on page load; Workstation itself is still launched in a popup.
  refresh();
})();
