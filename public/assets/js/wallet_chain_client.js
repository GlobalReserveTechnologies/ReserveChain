// wallet_chain_client.js
// Bridge between the browser and the ReserveChain HTTP API for balances.
//
// This is intentionally very small: for DevNet it just calls /api/balances
// and filters for a single address. In production you should front this
// through Nginx / HAProxy so /api/* is routed to the Go node.

(function () {
  if (!window.safeFetchJSON) {
    console.warn('[WalletChainClient] safeFetchJSON missing; include api_client.js first.');
    return;
  }

  async function getBalance(address) {
    if (!address) {
      throw new Error('Address is required for getBalance');
    }
    const data = await window.safeFetchJSON('/api/balances');
    const accounts = (data && data.accounts) || [];
    const match = accounts.find(a => a.address === address || a.Address === address);
    const balances = (match && (match.balances || match.Balances)) || {};
    // Normalise keys to upper case asset symbols
    const out = {};
    for (const k in balances) {
      if (!Object.prototype.hasOwnProperty.call(balances, k)) continue;
      const v = Number(balances[k]) || 0;
      out[k.toUpperCase()] = v;
    }
    const total = typeof out.GRC === 'number' ? out.GRC : 0;
    return {
      address,
      total,
      balances: out
    };
  }

  window.WalletChainClient = {
    getBalance
  };
})();
