// wallet_chain.js
// Helper for submitting L1 TX_TRANSFER from the browser keystore.

(function () {
  if (!window.ReserveWallet) {
    console.warn('ReserveWallet keystore missing; wallet_chain.js will be inert.');
    return;
  }

  async function fetchJSON(url, opts) {
    const res = await fetch(url, Object.assign({ credentials: 'same-origin' }, opts || {}));
    if (!res.ok) {
      const txt = await res.text();
      throw new Error('HTTP ' + res.status + ': ' + txt);
    }
    return res.json();
  }

  async function getNonce(address) {
    const data = await fetchJSON('/api/account/nonce?address=' + encodeURIComponent(address));
    return Number(data.nonce || 0);
  }

  async function submitTransfer({ fromAccountId, toAddress, amount, asset, memo }) {
    const accounts = ReserveWallet.listAccounts();
    const acct = accounts.find(a => a.id === fromAccountId);
    if (!acct) {
      throw new Error('Account not found: ' + fromAccountId);
    }
    if (!toAddress) {
      throw new Error('Missing destination address');
    }
    const amt = Number(amount);
    if (!isFinite(amt) || amt <= 0) {
      throw new Error('Invalid amount');
    }

    const currentNonce = await getNonce(acct.address);
    const nextNonce = currentNonce + 1;

    const tx = {
      from: acct.address,
      to: toAddress,
      asset: asset || 'GRC',
      amount: amt,
      nonce: nextNonce,
      memo: memo || ''
    };

    // Placeholder "signature" using keystore; backend does not verify yet,
    // but we keep it so we can harden this path later.
    const sig = ReserveWallet.signMessage(fromAccountId, tx);

    const body = {
      type: 'TX_TRANSFER',
      tx: Object.assign({}, tx, { signature: sig.signature })
    };

    const data = await fetchJSON('/api/tx/transfer', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(body)
    });

    return data;
  }

  window.ChainWallet = {
    submitTransfer
  };
})();
