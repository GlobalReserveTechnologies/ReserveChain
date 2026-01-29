// Vault Transfer modal wiring
(function() {
  const modal      = document.getElementById('vault-transfer-modal');
  const fromSelect = document.getElementById('vault-transfer-from');
  const toSelect   = document.getElementById('vault-transfer-to');
  const amountInput= document.getElementById('vault-transfer-amount');
  const submitBtn  = document.getElementById('vault-transfer-submit');

  if (!modal || !window.VaultAPI) return;

  let vaultsCache = [];

  async function loadVaultOptions(defaultFromVaultId) {
    vaultsCache = await VaultAPI.listVaults();

    fromSelect.innerHTML = '';
    toSelect.innerHTML   = '';

    if (!vaultsCache.length) {
      const opt = document.createElement('option');
      opt.value = '';
      opt.textContent = 'No vaults';
      fromSelect.appendChild(opt);
      toSelect.appendChild(opt.cloneNode(true));
      submitBtn.disabled = true;
      return;
    }

    submitBtn.disabled = false;

    vaultsCache.forEach(v => {
      const optFrom = document.createElement('option');
      optFrom.value = v.vault_id;
      optFrom.textContent = v.label;
      fromSelect.appendChild(optFrom);

      const optTo = optFrom.cloneNode(true);
      toSelect.appendChild(optTo);
    });

    if (defaultFromVaultId) {
      fromSelect.value = defaultFromVaultId;
    }

    if (vaultsCache.length > 1) {
      const firstDifferent = vaultsCache.find(v => v.vault_id !== fromSelect.value) || vaultsCache[0];
      toSelect.value = firstDifferent.vault_id;
    }
  }

  function openModal(defaultFromVaultId) {
    modal.style.display = 'flex';
    document.body.classList.add('modal-open');
    loadVaultOptions(defaultFromVaultId);
  }

  function closeModal() {
    modal.style.display = 'none';
    document.body.classList.remove('modal-open');
    amountInput.value = '';
  }

  async function handleSubmit() {
    const fromVaultId = fromSelect.value;
    const toVaultId   = toSelect.value;
    const amount      = parseFloat(amountInput.value || '0');

    if (!fromVaultId || !toVaultId || fromVaultId === toVaultId) {
      alert('Select two different vaults.');
      return;
    }
    if (!amount || amount <= 0) {
      alert('Enter a valid amount.');
      return;
    }

    submitBtn.disabled = true;
    submitBtn.textContent = 'Transferring...';

    try {
      await VaultAPI.transferBetweenVaults({
        fromVaultId,
        toVaultId,
        amount
      });
      alert('Transfer completed.');
      closeModal();
      if (window.loadVaultDashboard) {
        window.loadVaultDashboard();
      }
    } catch (e) {
      console.error(e);
      alert('Vault transfer failed: ' + e.message);
    } finally {
      submitBtn.disabled = false;
      submitBtn.textContent = 'Transfer';
    }
  }

  // header button
  document.addEventListener('click', (e) => {
    if (e.target.matches('#btn-vault-transfer')) {
      openModal(null);
    }
    const quick = e.target.closest('.js-vault-quick-transfer');
    if (quick) {
      const vid = quick.dataset.vault;
      openModal(vid || null);
    }
    const close = e.target.closest('[data-close-modal]');
    if (close && modal.contains(close)) {
      closeModal();
    }
  });

  submitBtn.addEventListener('click', handleSubmit);

  document.addEventListener('keydown', (e) => {
    if (e.key === 'Escape' && modal.style.display === 'flex') {
      closeModal();
    }
  });
})();
