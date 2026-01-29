
// vault_client.js
// Front-end client for vault-related API calls with error handling.

(function () {
  if (!window.safeFetchJSON) {
    console.error('safeFetchJSON missing, include api_client.js first');
  }

  async function callVaultAPI(action, opts = {}) {
    const method = opts.method || 'GET';
    const url =
      '/api/vault.php?action=' +
      encodeURIComponent(action) +
      (method === 'GET' && opts.query ? '&' + new URLSearchParams(opts.query).toString() : '');
    const fetchOpts = {
      method,
    };
    if (opts.body) {
      fetchOpts.headers = { 'Content-Type': 'application/json' };
      fetchOpts.body = JSON.stringify(opts.body);
    }
    try {
      return await safeFetchJSON(url, fetchOpts);
    } catch (e) {
      const msg = `Vault API error (${action})`;
      if (window.UINotifier) UINotifier.error(msg + ': ' + e.message, e);
      throw e;
    }
  }

  const VaultAPI = {
    async listVaults() {
      const data = await callVaultAPI('list');
      if (typeof data.nav === 'number') {
        VaultAPI.lastNav = data.nav;
      }
      return data.vaults || [];
    },

    async createVault(opts) {
      const body = {
        label: opts.label || 'New Vault',
        type: opts.type || 'single',
        visibility_mode: opts.visibility_mode || 'A',
        pnl_settlement_mode: opts.pnl_settlement_mode || 'source',
        template: opts.template || 'custom',
      };
      return callVaultAPI('create', { method: 'POST', body });
    },

    async transferBetweenVaults({ fromVaultId, toVaultId, amount, assetSymbol = 'GRC' }) {
      const body = {
        from_vault_id: fromVaultId,
        to_vault_id: toVaultId,
        amount: Number(amount),
        asset_symbol: assetSymbol,
      };
      return callVaultAPI('transfer_internal', { method: 'POST', body });
    },

    async updateSettings({ vaultId, visibilityMode, pnlSettlementMode, label, durationTier }) {
      const body = {
        vault_id: vaultId,
        visibility_mode: visibilityMode,
        pnl_settlement_mode: pnlSettlementMode,
        label,
      };
      if (durationTier) {
        body.duration_tier = durationTier;
      }
      return callVaultAPI('update_settings', { method: 'POST', body });
    },

    async stealthList(vaultId) {
      if (!vaultId) return [];
      const data = await callVaultAPI('stealth_list', {
        method: 'GET',
        query: { vault_id: vaultId },
      });
      return data.addresses || [];
    },

    async stealthCreate({ vaultId, label }) {
      function randomHex(len) {
        const bytes = new Uint8Array(len);
        crypto.getRandomValues(bytes);
        return Array.from(bytes)
          .map((b) => b.toString(16).padStart(2, '0'))
          .join('');
      }

      const stealthAddress = 'grc1stealth' + randomHex(24);
      const ephemeralPubkey = randomHex(64);

      const body = {
        vault_id: vaultId,
        label,
        stealth_address: stealthAddress,
        ephemeral_pubkey: ephemeralPubkey,
      };
      return callVaultAPI('stealth_create', { method: 'POST', body });
    },
  };

  window.VaultAPI = VaultAPI;
})();
