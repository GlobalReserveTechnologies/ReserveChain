// Wallet popup + high-level API wrapper for the workstation
// This sits on top of the simple browser keystore in wallet_keystore.js
// and exposes a richer interface for the workstation + trading terminal.

(function () {
  const toggleBtn = document.getElementById('wallet-popup-toggle');
  const modal = document.getElementById('wallet-popup-modal');
  if (!toggleBtn || !modal) return;

  const addressChip = document.getElementById('wallet-popup-address');
  const totalBalanceEl = document.getElementById('wallet-total-balance');
  const availBalanceEl = document.getElementById('wallet-available-balance');
  const assetsListEl = document.getElementById('wallet-assets-list');
  const settingsLabelEl = document.getElementById('wallet-settings-label');
  const settingsMaxTxEl = document.getElementById('wallet-settings-max-tx');
  const settingsMaxDayEl = document.getElementById('wallet-settings-max-day');
  const settingsRequireConfirmEl = document.getElementById('wallet-settings-require-confirm');
  const settingsResetBtn = document.getElementById('wallet-settings-reset');

  let currentAccount = null;
  let sessionInfo = {
    unlocked: true,
    remaining_minutes: null
  };

  function shortAddr(addr) {
    if (!addr) return 'Not connected';
    if (addr.length <= 12) return addr;
    return addr.slice(0, 6) + '…' + addr.slice(-4);
  }

  function showModal() {
    modal.style.display = 'flex';
    refreshWalletView();
  }

  function hideModal() {
    modal.style.display = 'none';
  }

  function activateTab(name) {
    const tabs = modal.querySelectorAll('.wallet-tab');
    const panels = modal.querySelectorAll('.wallet-tab-panel');
    tabs.forEach((t) => {
      if (t.getAttribute('data-wallet-tab') === name) {
        t.classList.add('is-active');
      } else {
        t.classList.remove('is-active');
      }
    });
    panels.forEach((p) => {
      if (p.getAttribute('data-wallet-tab-panel') === name) {
        p.style.display = '';
      } else {
        p.style.display = 'none';
      }
    });
  }

  function attachTabHandlers() {
    const tabs = modal.querySelectorAll('.wallet-tab');
    tabs.forEach((t) => {
      t.addEventListener('click', () => {
        const name = t.getAttribute('data-wallet-tab');
        if (name) activateTab(name);
      });
    });
  }

  function ensureAccount() {
    if (!window.ReserveWallet) return null;
    const accounts = window.ReserveWallet.listAccounts();
    if (accounts && accounts.length) {
      currentAccount = accounts[0];
      return currentAccount;
    }
    currentAccount = window.ReserveWallet.createAccount('Workstation Wallet');
    return currentAccount;
  }

  async function refreshWalletView() {
    const acct = ensureAccount();
    if (!acct) return;
    if (addressChip) addressChip.textContent = shortAddr(acct.address);

    // For DevNet, balances are not yet wired; show placeholder using a fake API if available
    let balance = null;
    try {
      if (window.WalletChainClient && WalletChainClient.getBalance) {
        balance = await WalletChainClient.getBalance(acct.address);
      }
    } catch (e) {
      console.warn('[WalletPopup] error fetching balance', e);
    }

    const totalVal = balance && typeof balance.total === 'number' ? balance.total : 0;
    if (totalBalanceEl) {
      totalBalanceEl.textContent = totalVal.toLocaleString(undefined, { maximumFractionDigits: 4 }) + ' GRC';
    }
    if (availBalanceEl) {
      availBalanceEl.textContent = totalVal.toLocaleString(undefined, { maximumFractionDigits: 4 }) + ' GRC';
    }

    if (assetsListEl) {
      assetsListEl.innerHTML = '';
      const row = document.createElement('div');
      row.textContent = 'GRC — ' + totalVal.toLocaleString(undefined, { maximumFractionDigits: 4 });
      assetsListEl.appendChild(row);
    }
  }

  function loadSettings() {
    try {
      const raw = window.localStorage.getItem('reservechain_wallet_settings_v1');
      if (!raw) return null;
      return JSON.parse(raw);
    } catch (e) {
      console.warn('[WalletPopup] failed to load settings', e);
      return null;
    }
  }

  function saveSettings(settings) {
    try {
      window.localStorage.setItem('reservechain_wallet_settings_v1', JSON.stringify(settings));
    } catch (e) {
      console.warn('[WalletPopup] failed to save settings', e);
    }
  }

  function initSettingsUI() {
    const settings = loadSettings() || {};
    if (settingsLabelEl && settings.label) settingsLabelEl.value = settings.label;
    if (settingsMaxTxEl && settings.max_tx) settingsMaxTxEl.value = settings.max_tx;
    if (settingsMaxDayEl && settings.max_day) settingsMaxDayEl.value = settings.max_day;
    if (settingsRequireConfirmEl) {
      settingsRequireConfirmEl.checked = settings.require_confirm !== false;
    }

    function syncSettings() {
      const s = {
        label: settingsLabelEl ? settingsLabelEl.value : '',
        max_tx: settingsMaxTxEl && settingsMaxTxEl.value ? Number(settingsMaxTxEl.value) : null,
        max_day: settingsMaxDayEl && settingsMaxDayEl.value ? Number(settingsMaxDayEl.value) : null,
        require_confirm: settingsRequireConfirmEl ? !!settingsRequireConfirmEl.checked : true
      };
      saveSettings(s);
    }

    if (settingsLabelEl) settingsLabelEl.addEventListener('change', syncSettings);
    if (settingsMaxTxEl) settingsMaxTxEl.addEventListener('change', syncSettings);
    if (settingsMaxDayEl) settingsMaxDayEl.addEventListener('change', syncSettings);
    if (settingsRequireConfirmEl) settingsRequireConfirmEl.addEventListener('change', syncSettings);

    if (settingsResetBtn) {
      settingsResetBtn.addEventListener('click', () => {
        if (!window.confirm('Reset dev wallet keystore and settings?')) return;
        window.localStorage.removeItem('reservechain_keystore_v1');
        window.localStorage.removeItem('reservechain_wallet_settings_v1');
        currentAccount = null;
        refreshWalletView();
        initSettingsUI();
      });
    }
  }

  // Toggle button
  toggleBtn.addEventListener('click', showModal);
  const closeBtn = modal.querySelector('[data-wallet-close]');
  if (closeBtn) closeBtn.addEventListener('click', hideModal);

  // Close when clicking backdrop
  modal.addEventListener('click', (e) => {
    if (e.target === modal) hideModal();
  });

  attachTabHandlers();
  initSettingsUI();

  // Expose a lightweight programmatic API for other modules (trading terminal, etc.).
  window.reserveWalletApi = {
    async connect() {
      const acct = ensureAccount();
      if (!acct) throw new Error('No wallet available');
      return {
        wallet_id: 'dev-browser-wallet',
        address: acct.address
      };
    },
    async getCurrentWallet() {
      const acct = ensureAccount();
      if (!acct) return null;
      return {
        wallet_id: 'dev-browser-wallet',
        address: acct.address
      };
    },
    async getBalances() {
      const acct = ensureAccount();
      if (!acct) return { total: 0, assets: [] };
      let balance = null;
      try {
        if (window.WalletChainClient && WalletChainClient.getBalance) {
          balance = await WalletChainClient.getBalance(acct.address);
        }
      } catch (e) {
        console.warn('[WalletPopup] getBalances error', e);
      }
      const total = balance && typeof balance.total === 'number' ? balance.total : 0;
      return {
        total: total,
        assets: [
          { symbol: 'GRC', balance: total }
        ]
      };
    },
    async requestSignature(req) {
      const acct = ensureAccount();
      if (!acct) throw new Error('No account');
      const payload = req && req.payload ? req.payload : {};
      const sig = await window.ReserveWallet.signMessage(acct.id, payload);
      return {
        tx: payload,
        signature: sig.signature,
        address: sig.address,
        account_id: sig.account_id
      };
    }
  };

})();