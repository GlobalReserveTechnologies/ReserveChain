// Stealth Addresses panel
(function() {
  const vaultSelect   = document.getElementById('stealth-vault-select');
  const labelInput    = document.getElementById('stealth-label-input');
  const genBtn        = document.getElementById('stealth-generate-btn');
  const tableBody     = document.getElementById('stealth-table-body');
  const lastGenerated = document.getElementById('stealth-last-generated');

  if (!vaultSelect || !window.VaultAPI) return;

  let vaults = [];

  async function initVaultOptions() {
    vaults = await VaultAPI.listVaults();
    vaultSelect.innerHTML = '';

    if (!vaults.length) {
      const opt = document.createElement('option');
      opt.value = '';
      opt.textContent = 'No vaults available';
      vaultSelect.appendChild(opt);
      vaultSelect.disabled = true;
      genBtn.disabled = true;
      return;
    }

    vaultSelect.disabled = false;
    genBtn.disabled = false;

    vaults.forEach(v => {
      const opt = document.createElement('option');
      opt.value = v.vault_id;
      opt.textContent = `${v.label} (${v.type === 'multisig' ? 'Multi-sig' : 'Single'})`;
      vaultSelect.appendChild(opt);
    });

    if (vaults[0]) {
      vaultSelect.value = vaults[0].vault_id;
      await loadStealthTable(vaults[0].vault_id);
    }
  }

  async function loadStealthTable(vaultId) {
    tableBody.innerHTML = '';
    if (!vaultId) return;

    const rows = await VaultAPI.stealthList(vaultId);

    if (!rows.length) {
      const tr = document.createElement('tr');
      const td = document.createElement('td');
      td.colSpan = 7;
      td.textContent = 'No stealth addresses yet for this vault.';
      td.style.color = '#9aa7c0';
      tr.appendChild(td);
      tableBody.appendChild(tr);
      return;
    }

    rows.forEach(r => {
      const tr = document.createElement('tr');
      tr.innerHTML = `
        <td>${r.label || '—'}</td>
        <td class="mono">${r.stealth_address}</td>
        <td class="mono small">${r.ephemeral_pubkey}</td>
        <td>${r.is_active ? 'Yes' : 'No'}</td>
        <td>${r.created_at || '—'}</td>
        <td>${r.last_used_at || '—'}</td>
        <td>
          <button class="link-button js-stealth-copy" data-value="${r.stealth_address}">Copy</button>
        </td>
      `;
      tableBody.appendChild(tr);
    });
  }

  async function handleGenerate() {
    const vaultId = vaultSelect.value;
    if (!vaultId) return;

    const label = labelInput.value.trim() || null;

    genBtn.disabled = true;
    genBtn.textContent = 'Generating...';

    try {
      const created = await VaultAPI.stealthCreate({ vaultId, label });

      await loadStealthTable(vaultId);

      lastGenerated.innerHTML = `
        <div style="font-size:13px; color:#e9f0ff;">
          <div style="margin-bottom:6px;"><strong>Vault:</strong> ${vaultId}</div>
          <div style="margin-bottom:4px;"><strong>Label:</strong> ${created.label || '—'}</div>
          <div style="margin-bottom:4px;"><strong>Stealth Address:</strong> <span class="mono">${created.stealth_address}</span></div>
          <div style="margin-bottom:4px;"><strong>Ephemeral Pubkey:</strong> <span class="mono small">${created.ephemeral_pubkey}</span></div>
          <button class="btn btn-secondary js-stealth-copy-last" data-value="${created.stealth_address}">Copy Address</button>
        </div>
      `;
    } catch (e) {
      console.error(e);
      alert('Failed to generate stealth address: ' + e.message);
    } finally {
      genBtn.disabled = false;
      genBtn.textContent = 'Generate Stealth Address';
    }
  }

  vaultSelect.addEventListener('change', () => {
    const vaultId = vaultSelect.value;
    if (vaultId) loadStealthTable(vaultId);
  });

  genBtn.addEventListener('click', () => {
    handleGenerate();
  });

  document.addEventListener('click', (e) => {
    const btn = e.target.closest('.js-stealth-copy, .js-stealth-copy-last');
    if (!btn) return;
    const val = btn.dataset.value;
    if (!val) return;
    navigator.clipboard?.writeText(val).then(() => {
      const orig = btn.textContent;
      btn.textContent = 'Copied';
      setTimeout(() => { btn.textContent = orig; }, 900);
    });
  });

  document.addEventListener('DOMContentLoaded', () => {
    initVaultOptions().catch(console.error);
  });
})();
