
// Workstation Vault Settings panel
(function () {
  const root = document.getElementById('vault-settings-root');
  if (!root || !window.VaultAPI) return;

  const visLabels = {
    A: 'Visible',
    B: 'Hidden Toggle',
    C: 'Passphrase',
    D: 'Audit Only',
  };

  const pnlModes = {
    source: 'Settle to Source (current)',
    target: 'Settle to Target',
  };

  const durationLabels = {
    short: 'Short — liquidity first',
    medium: 'Medium — balanced',
    long: 'Long — yield focus',
  };

  let currentVaults = [];
  let selectedVaultId = null;

  async function loadVaults() {
    root.textContent = 'Loading vault settings...';
    try {
      const vaults = await VaultAPI.listVaults();
      currentVaults = vaults;
      if (!vaults.length) {
        root.innerHTML = '<div class="vault-list-placeholder">No vaults yet. Use the Vault Dashboard to create one.</div>';
        return;
      }
      if (!selectedVaultId || !vaults.find(v => v.vault_id === selectedVaultId)) {
        selectedVaultId = vaults[0].vault_id;
      }
      render();
    } catch (e) {
      console.error('Failed to load vaults for settings', e);
      root.innerHTML = '<div class="vault-list-placeholder">Failed to load vaults. Please try again.</div>';
    }
  }

  function render() {
    const vault = currentVaults.find(v => v.vault_id === selectedVaultId);
    if (!vault) {
      root.innerHTML = '<div class="vault-list-placeholder">Select a vault to configure.</div>';
      return;
    }

    const vis = vault.visibility_mode || 'A';
    const pnlMode = vault.pnl_settlement_mode || 'source';
    const durationTier =
      (vault.yield_policy && vault.yield_policy.duration_tier) || 'medium';

    const optionsVault = currentVaults
      .map(v => {
        const sel = v.vault_id === selectedVaultId ? ' selected' : '';
        const label = v.label || v.vault_id;
        return '<option value="' + v.vault_id + '"' + sel + '>' + label + '</option>';
      })
      .join('');

    const visOptions = Object.entries(visLabels)
      .map(([key, label]) => {
        const sel = key === vis ? ' selected' : '';
        return '<option value="' + key + '"' + sel + '>' + label + '</option>';
      })
      .join('');

    const pnlOptions = Object.entries(pnlModes)
      .map(([key, label]) => {
        const sel = key === pnlMode ? ' selected' : '';
        return '<option value="' + key + '"' + sel + '>' + label + '</option>';
      })
      .join('');

    const durationOptions = Object.entries(durationLabels)
      .map(([key, label]) => {
        const sel = key === durationTier ? ' selected' : '';
        return '<option value="' + key + '"' + sel + '>' + label + '</option>';
      })
      .join('');

    root.innerHTML = `
      <div class="vault-settings-form">
        <div class="field">
          <span>Vault</span>
          <select id="vs-vault-select">${optionsVault}</select>
        </div>

        <div class="field">
          <span>Label</span>
          <input type="text" id="vs-label" value="${escapeHtml(vault.label || '')}" placeholder="Primary Vault" />
        </div>

        <div class="field">
          <span>Visibility</span>
          <select id="vs-visibility">${visOptions}</select>
        </div>

        <div class="field">
          <span>PnL Settlement</span>
          <select id="vs-pnl-mode">${pnlOptions}</select>
        </div>

        <div class="field">
          <span>Yield Duration Profile</span>
          <select id="vs-duration">${durationOptions}</select>
          <div class="modal-note">
            Duration controls how the vault prefers liquidity vs. yield when allocating to reserve tiers.
          </div>
        </div>

        <div class="vault-settings-actions">
          <button class="btn btn-secondary" id="vs-refresh">Reload</button>
          <button class="btn btn-primary" id="vs-save">Save Settings</button>
        </div>
      </div>
    `;

    wireEvents();
  }

  function wireEvents() {
    const selectVault = document.getElementById('vs-vault-select');
    const btnSave = document.getElementById('vs-save');
    const btnRefresh = document.getElementById('vs-refresh');

    if (selectVault) {
      selectVault.addEventListener('change', (e) => {
        selectedVaultId = e.target.value;
        render();
      });
    }

    if (btnRefresh) {
      btnRefresh.addEventListener('click', (e) => {
        e.preventDefault();
        loadVaults();
      });
    }

    if (btnSave) {
      btnSave.addEventListener('click', async (e) => {
        e.preventDefault();
        const vaultId = selectedVaultId;
        const labelEl = document.getElementById('vs-label');
        const visEl = document.getElementById('vs-visibility');
        const pnlEl = document.getElementById('vs-pnl-mode');
        const durEl = document.getElementById('vs-duration');

        const payload = {
          vaultId,
          visibilityMode: visEl ? visEl.value : 'A',
          pnlSettlementMode: pnlEl ? pnlEl.value : 'source',
          label: labelEl ? labelEl.value : '',
          durationTier: durEl ? durEl.value : 'medium',
        };

        try {
          await VaultAPI.updateSettings(payload);
          if (window.UINotifier) {
            UINotifier.info('Vault settings saved.');
          }
          // Reload to reflect changes across views
          loadVaults();
        } catch (err) {
          console.error('Failed to save vault settings', err);
          if (window.UINotifier) {
            UINotifier.error('Failed to save vault settings: ' + err.message, err);
          } else {
            alert('Failed to save vault settings: ' + err.message);
          }
        }
      });
    }
  }

  function escapeHtml(str) {
    return String(str)
      .replace(/&/g, '&amp;')
      .replace(/</g, '&lt;')
      .replace(/>/g, '&gt;')
      .replace(/"/g, '&quot;')
      .replace(/'/g, '&#039;');
  }

  // Load immediately when panel exists
  loadVaults();
})();
