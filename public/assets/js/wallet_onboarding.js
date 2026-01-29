(function() {
  function qs(sel){ return document.querySelector(sel); }
  function qsa(sel){ return Array.from(document.querySelectorAll(sel)); }

  const modal = qs('#rc-onboarding-modal');
  const btnOpen = qs('#btn-create-account');
  const btnCancel = qs('#rc-btn-cancel');
  const btnGen = qs('#rc-btn-generate-wallet');
  const btnDl = qs('#rc-btn-download-keystore');
  const btnLaunch = qs('#rc-btn-launch-ws');

  const labelEl = qs('#rc-wallet-label');
  const passEl = qs('#rc-wallet-pass');
  const pass2El = qs('#rc-wallet-pass2');
  const addrEl = qs('#rc-wallet-address');

  let lastKeystore = null;

  function openModal() {
    if (!modal) return;
    modal.setAttribute('aria-hidden', 'false');
    showStep(1);
    lastKeystore = null;
    if (addrEl) addrEl.textContent = '—';
    if (labelEl) labelEl.value = '';
    if (passEl) passEl.value = '';
    if (pass2El) pass2El.value = '';
  }

  function closeModal() {
    if (!modal) return;
    modal.setAttribute('aria-hidden', 'true');
  }

  function showStep(step) {
    qsa('.rc-step').forEach(el => el.classList.remove('rc-step--active'));
    const target = qs('.rc-step[data-step="' + step + '"]');
    if (target) target.classList.add('rc-step--active');
  }

  function downloadJson(filename, obj) {
    const blob = new Blob([JSON.stringify(obj, null, 2)], { type: 'application/json' });
    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = filename;
    document.body.appendChild(a);
    a.click();
    a.remove();
    URL.revokeObjectURL(url);
  }

  async function generateWallet() {
    const label = (labelEl && labelEl.value || '').trim();
    const p1 = (passEl && passEl.value || '');
    const p2 = (pass2El && pass2El.value || '');

    if (!p1 || p1.length < 8) {
      alert('Choose a password of at least 8 characters.');
      return;
    }
    if (p1 !== p2) {
      alert('Passwords do not match.');
      return;
    }

    btnGen && (btnGen.disabled = true);

    try {
      const res = await window.ReserveWallet.generateWallet(label, p1);
      lastKeystore = res.keystore;
      if (addrEl) addrEl.textContent = res.wallet.address;

      showStep(2);
    } catch (e) {
      console.error(e);
      alert('Failed to generate wallet: ' + (e && e.message ? e.message : String(e)));
    } finally {
      btnGen && (btnGen.disabled = false);
    }
  }

  function downloadKeystore() {
    if (!lastKeystore) {
      alert('Generate a wallet first.');
      return;
    }
    const addr = (lastKeystore.wallet && lastKeystore.wallet.address) ? lastKeystore.wallet.address : 'wallet';
    downloadJson('reservechain-wallet-' + addr + '.json', lastKeystore);
  }

  function launchWorkstation() {
    // Keep the workstation as a separate surface
    window.location.href = '/workstation/';
  }

  // Wire open/close
  if (btnOpen) btnOpen.addEventListener('click', openModal);

  qsa('[data-close="1"]').forEach(el => {
    el.addEventListener('click', closeModal);
  });

  if (btnCancel) btnCancel.addEventListener('click', closeModal);
  if (btnGen) btnGen.addEventListener('click', (e) => { e.preventDefault(); generateWallet(); });
  if (btnDl) btnDl.addEventListener('click', (e) => { e.preventDefault(); downloadKeystore(); });
  if (btnLaunch) btnLaunch.addEventListener('click', (e) => { e.preventDefault(); launchWorkstation(); });

  // Also open modal when CTA inside hero is clicked (if present)
  const heroCtas = qsa('[data-action="create-account"]');
  heroCtas.forEach(btn => btn.addEventListener('click', openModal));
})();
