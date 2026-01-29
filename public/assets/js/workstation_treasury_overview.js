
// Treasury & Reserves overview (DevNet-backed)
(function () {
  const totalEl = document.getElementById('treasury-total-reserve');
  const covEl = document.getElementById('treasury-coverage-ratio');
  const durEl = document.getElementById('treasury-duration-split');
  const yieldEl = document.getElementById('treasury-yield-sources');
  if (!totalEl) return;

  async function refresh() {
    try {
      // 1) Load vaults for duration split (local PHP state)
      let vaults = [];
      if (window.VaultAPI) {
        try {
          vaults = await VaultAPI.listVaults();
        } catch (e) {
          console.warn('[Treasury] failed to load vaults, continuing with DevNet only', e);
        }
      }

      let shortCount = 0;
      let mediumCount = 0;
      let longCount = 0;
      if (vaults && vaults.length) {
        vaults.forEach(v => {
          const tier = (v.yield_policy && v.yield_policy.duration_tier) || 'medium';
          if (tier === 'short') shortCount++;
          else if (tier === 'long') longCount++;
          else mediumCount++;
        });
      }

      const durParts = [];
      if (shortCount) durParts.push(shortCount + ' short');
      if (mediumCount) durParts.push(mediumCount + ' medium');
      if (longCount) durParts.push(longCount + ' long');
      durEl.textContent = durParts.length ? durParts.join(' · ') : 'No vaults';

      // 2) Load DevNet treasury snapshot from Go backend via PHP proxy
      let tsnap = null;
      try {
        const res = await fetch('/api/treasury.php');
        if (res.ok) {
          tsnap = await res.json();
        } else {
          console.warn('[Treasury] non-200 from treasury.php', res.status);
        }
      } catch (e) {
        console.warn('[Treasury] error fetching treasury.php', e);
      }

      if (tsnap && typeof tsnap.TotalUSD === 'number') {
        totalEl.textContent =
          tsnap.TotalUSD.toLocaleString(undefined, { maximumFractionDigits: 2 }) + ' USD';

        // For now, coverage ratio is implied 100%+ for DevNet synthetic + optional external.
        covEl.textContent = 'DevNet synthetic + external mix';
      } else {
        // Fall back to local vault balances if DevNet is not running.
        let totalGRC = 0;
        if (vaults && vaults.length) {
          vaults.forEach(v => {
            const bal = parseFloat((v.balance && v.balance.GRC) || 0);
            totalGRC += bal;
          });
        }
        totalEl.textContent =
          totalGRC.toLocaleString(undefined, { maximumFractionDigits: 4 }) + ' GRC';
        covEl.textContent = 'Local vault-only view';
      }

      yieldEl.textContent = 'External + internal (mixed yield model)';
    } catch (e) {
      console.error('[Treasury] failed to load reserve overview', e);
      totalEl.textContent = 'Error';
      covEl.textContent = 'Error';
      durEl.textContent = 'Error';
      yieldEl.textContent = 'Error';
    }
  }

  // Refresh on load; in future we could add a refresh button or WS events.
  refresh();
})();
